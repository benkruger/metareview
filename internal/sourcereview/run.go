package sourcereview

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsifry/metareview/internal/lensoutput"
)

const usageText = `Usage: metareview source-review --model astra|opus|grok --output <dir> [<repo>]

Reviews the first-party source at HEAD. Prints the kept list and the dropped
list before any model call. Writes <dir>/findings.json and <dir>/review.html.
The output directory must be outside the repository.`

var errHelp = errors.New("help")

// Path seams. Production uses the stdlib functions. Tests replace them to
// cover the error returns filepath and os do not produce on this OS.
var (
	pathAbs  = filepath.Abs
	pathEval = filepath.EvalSymlinks
	pathRel  = filepath.Rel
	statPath = os.Stat
)

// Options is one source-review run. A nil Runner, Git, MkdirAll, WriteFile, or
// Remove uses the real process, git, or filesystem. Budget 0 means MaxPromptBytes.
type Options struct {
	Repo      string
	Model     string
	OutputDir string
	Stdout    io.Writer
	Stderr    io.Writer
	Budget    int
	Runner    Runner
	Git       GitFunc
	MkdirAll  func(string, os.FileMode) error
	WriteFile func(string, []byte, os.FileMode) error
	Remove    func(string) error
}

// Run reviews one repository. It writes the findings file and the page only
// after every prompt succeeds. A failed prompt leaves the output directory empty.
func Run(opts Options) error {
	if opts.Stdout == nil {
		opts.Stdout = io.Discard
	}
	if opts.Stderr == nil {
		opts.Stderr = io.Discard
	}
	if strings.TrimSpace(opts.OutputDir) == "" {
		return errors.New("missing output directory")
	}
	modelID, err := ModelID(opts.Model)
	if err != nil {
		return err
	}
	git := opts.Git
	if git == nil {
		git = realGit
	}
	top, err := resolveTop(opts.Repo, git)
	if err != nil {
		return err
	}
	if err := refuseInside(top, opts.OutputDir); err != nil {
		return err
	}
	kept, dropped, commit, err := Select(top, git)
	if err != nil {
		return err
	}
	writeLists(opts.Stdout, kept, dropped)
	files := make(map[string][]byte, len(kept))
	for _, f := range kept {
		files[f.Path] = f.Body
	}
	if len(kept) == 0 {
		return writeOutputs(opts, top, commit, modelID, nil, files)
	}
	budget := opts.Budget
	if budget == 0 {
		budget = MaxPromptBytes
	}
	prompts, err := Pack(kept, budget)
	if err != nil {
		return err
	}
	if opts.Runner == nil {
		return errors.New("no runner")
	}
	var all []lensoutput.TypedFinding
	for _, p := range prompts {
		fmt.Fprintln(opts.Stdout, "prompt-begin")
		_, _ = opts.Stdout.Write(p)
		if len(p) == 0 || p[len(p)-1] != '\n' {
			fmt.Fprintln(opts.Stdout)
		}
		fmt.Fprintln(opts.Stdout, "prompt-end")
		text, err := Call(opts.Runner, opts.Model, p)
		if err != nil {
			return err
		}
		got, skipped, err := Parse(text, files)
		if err != nil {
			return err
		}
		for _, msg := range skipped {
			fmt.Fprintln(opts.Stderr, msg)
		}
		all = append(all, got...)
	}
	return writeOutputs(opts, top, commit, modelID, all, files)
}

func writeLists(w io.Writer, kept []File, dropped []string) {
	fmt.Fprintln(w, "kept:")
	for _, f := range kept {
		fmt.Fprintln(w, f.Path)
	}
	fmt.Fprintln(w, "dropped:")
	for _, p := range dropped {
		fmt.Fprintln(w, p)
	}
}

func writeOutputs(opts Options, repo, commit, modelID string, findings []lensoutput.TypedFinding, files map[string][]byte) error {
	body := findingsJSON(findings)
	page := Render(Page{Repo: repo, Commit: commit, ModelID: modelID, Findings: findings, Files: files})
	mkdir := opts.MkdirAll
	if mkdir == nil {
		mkdir = os.MkdirAll
	}
	write := opts.WriteFile
	if write == nil {
		write = os.WriteFile
	}
	remove := opts.Remove
	if remove == nil {
		remove = os.Remove
	}
	if err := mkdir(opts.OutputDir, 0o755); err != nil {
		return err
	}
	findingsPath := filepath.Join(opts.OutputDir, "findings.json")
	pagePath := filepath.Join(opts.OutputDir, "review.html")
	if err := write(findingsPath, body, 0o644); err != nil {
		return err
	}
	if err := write(pagePath, page, 0o644); err != nil {
		_ = remove(findingsPath)
		return err
	}
	return nil
}

func findingsJSON(fs []lensoutput.TypedFinding) []byte {
	if fs == nil {
		fs = []lensoutput.TypedFinding{}
	}
	body, _ := json.Marshal(struct {
		Findings []lensoutput.TypedFinding `json:"findings"`
	}{Findings: fs})
	return append(body, '\n')
}

func refuseInside(repo, out string) error {
	repoAbs, outAbs, err := absBoth(repo, out)
	if err != nil {
		return err
	}
	repoReal, err := pathEval(repoAbs)
	if err != nil {
		return err
	}
	outReal, err := evalExisting(outAbs)
	if err != nil {
		return err
	}
	rel, err := pathRel(repoReal, outReal)
	if err != nil {
		return err
	}
	if insideRel(rel) {
		return fmt.Errorf("refusing output directory inside the tree under review: %s", out)
	}
	return nil
}

func absBoth(a, b string) (string, string, error) {
	aa, err := pathAbs(a)
	if err != nil {
		return "", "", err
	}
	bb, err := pathAbs(b)
	if err != nil {
		return "", "", err
	}
	return aa, bb, nil
}

func evalExisting(p string) (string, error) {
	cur := p
	var tail []string
	for {
		if _, err := statPath(cur); err == nil {
			real, err := pathEval(cur)
			if err != nil {
				return "", err
			}
			for i := len(tail) - 1; i >= 0; i-- {
				real = filepath.Join(real, tail[i])
			}
			return real, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", fmt.Errorf("cannot resolve %s", p)
		}
		tail = append(tail, filepath.Base(cur))
		cur = parent
	}
}

func insideRel(rel string) bool {
	if rel == "." {
		return true
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}

// CLI is the source-review command. repo defaults to cwd.
func CLI(args []string, cwd string, stdout, stderr io.Writer) int {
	model, out, repo, err := parseArgs(args)
	if err != nil {
		if errors.Is(err, errHelp) {
			fmt.Fprintln(stdout, usageText)
			return 0
		}
		fmt.Fprintln(stderr, err.Error())
		fmt.Fprintln(stderr, usageText)
		return 2
	}
	if repo == "" {
		repo = cwd
	}
	err = Run(Options{
		Repo:      repo,
		Model:     model,
		OutputDir: out,
		Stdout:    stdout,
		Stderr:    stderr,
		Runner:    OSRunner{},
	})
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	return 0
}

func parseArgs(args []string) (model, out, repo string, err error) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			return "", "", "", errHelp
		case a == "--model":
			if i+1 >= len(args) {
				return "", "", "", errors.New("missing value for --model")
			}
			i++
			model = args[i]
		case a == "--output":
			if i+1 >= len(args) {
				return "", "", "", errors.New("missing value for --output")
			}
			i++
			out = args[i]
		case strings.HasPrefix(a, "--"):
			return "", "", "", fmt.Errorf("unknown option: %s", a)
		default:
			if repo != "" {
				return "", "", "", fmt.Errorf("unexpected argument: %s", a)
			}
			repo = a
		}
	}
	if model == "" || out == "" {
		return "", "", "", errors.New("missing --model or --output")
	}
	return model, out, repo, nil
}
