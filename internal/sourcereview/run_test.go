package sourcereview

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dsifry/metareview/internal/lensoutput"
)

func cliStdout(model, text string) []byte {
	switch model {
	case "opus":
		b, _ := json.Marshal(map[string]any{"is_error": false, "result": text})
		return b
	case "grok":
		b, _ := json.Marshal(map[string]any{"text": text})
		return b
	default:
		quoted, _ := json.Marshal(text)
		return []byte(`{"type":"item.completed","item":{"type":"agent_message","text":` + string(quoted) + "}}\n")
	}
}

func TestRunSpecFixtureThreeModels(t *testing.T) {
	const sentinel = "super-secret-api-key-9f3a"
	t.Setenv("ANTHROPIC_API_KEY", sentinel)
	t.Setenv("OPENAI_API_KEY", sentinel)
	t.Setenv("XAI_API_KEY", sentinel)
	root, commit := newRepoBytes(t, specFiles())
	for _, model := range []string{"opus", "astra", "grok"} {
		var buf bytes.Buffer
		var prompts [][]byte
		fr := &fakeRunner{
			onRun: func() {
				out := buf.String()
				if !strings.Contains(out, "kept:\npkg/app.go\npkg/note.go\n") || !strings.Contains(out, "dropped:\n") {
					t.Fatalf("%s lists before model call:\n%s", model, out)
				}
				idx := strings.Index(out, "prompt-begin")
				keptAt := strings.Index(out, "kept:")
				if keptAt < 0 || idx < 0 || keptAt > idx {
					t.Fatalf("%s kept list is not before the prompt", model)
				}
				for _, p := range specDropped {
					if !strings.Contains(out[:idx], p) {
						t.Fatalf("%s dropped list missing %s", model, p)
					}
				}
			},
			reply: func(_ int, name string, args []string, stdin []byte) ([]byte, error) {
				var prompt []byte
				if model == "grok" {
					if stdin != nil {
						t.Fatal("grok prompt must be an argument")
					}
					prompt = []byte(args[1])
				} else {
					prompt = stdin
				}
				prompts = append(prompts, append([]byte(nil), prompt...))
				if name == "" {
					t.Fatal("empty binary")
				}
				return cliStdout(model, `{"findings":[]}`), nil
			},
		}
		outDir := t.TempDir()
		err := Run(Options{Repo: root, Model: model, OutputDir: outDir, Stdout: &buf, Runner: fr})
		if err != nil {
			t.Fatal(model, err)
		}
		if strings.Contains(buf.String(), sentinel) {
			t.Fatal("api key leaked into stdout")
		}
		if len(prompts) != 1 {
			t.Fatalf("%s prompts %d", model, len(prompts))
		}
		p := prompts[0]
		if len(p) > MaxPromptBytes {
			t.Fatalf("%s prompt %d", model, len(p))
		}
		if !bytes.Contains(p, []byte(instruction)) || !bytes.Contains(p, []byte(`{"findings":[...]}`)) {
			t.Fatal("instruction missing")
		}
		for path, body := range specFiles() {
			needle := append([]byte(path+"\n"), body...)
			has := bytes.Contains(p, needle)
			if path == "pkg/app.go" || path == "pkg/note.go" {
				if !has {
					t.Fatalf("%s prompt missing %s immediately before its text", model, path)
				}
			} else if bytes.Contains(p, []byte(path)) || bytes.Contains(p, []byte("DROPME")) {
				t.Fatalf("%s prompt contains dropped %s", model, path)
			}
		}
		raw, err := os.ReadFile(filepath.Join(outDir, "findings.json"))
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(string(raw)) != `{"findings":[]}` {
			t.Fatalf("findings %s", raw)
		}
		page, err := os.ReadFile(filepath.Join(outDir, "review.html"))
		if err != nil {
			t.Fatal(err)
		}
		html := string(page)
		id, _ := ModelID(model)
		for _, want := range []string{id, commit, `id="severity"`, `id="directory"`, "P0 0", "P1 0", "P2 0", "P3 0"} {
			if !strings.Contains(html, want) {
				t.Fatalf("%s page missing %q", model, want)
			}
		}
		if strings.Contains(html, `class="finding"`) {
			t.Fatal("empty findings rendered a finding")
		}
		entries, _ := os.ReadDir(outDir)
		if len(entries) != 2 {
			t.Fatalf("output entries %d", len(entries))
		}
	}
	srcHasGetenv(t)
}

func srcHasGetenv(t *testing.T) {
	t.Helper()
	ents, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range ents {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		b, err := os.ReadFile(e.Name())
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(b, []byte("Getenv")) || bytes.Contains(b, []byte("API_KEY")) {
			t.Fatalf("%s reads an API key", e.Name())
		}
	}
}

func TestRunEmptyKeptWritesZeros(t *testing.T) {
	root, commit := newRepoBytes(t, droppedOnlyFiles())
	outDir := t.TempDir()
	var buf bytes.Buffer
	fr := &fakeRunner{reply: func(int, string, []string, []byte) ([]byte, error) {
		t.Fatal("model called")
		return nil, nil
	}}
	if err := Run(Options{Repo: root, Model: "opus", OutputDir: outDir, Stdout: &buf, Runner: fr}); err != nil {
		t.Fatal(err)
	}
	if len(fr.calls) != 0 {
		t.Fatal("model called")
	}
	out := buf.String()
	if !strings.HasPrefix(out, "kept:\ndropped:\n") {
		t.Fatalf("lists:\n%s", out)
	}
	for _, p := range specDropped {
		if !strings.Contains(out, p+"\n") {
			t.Fatalf("missing dropped %s", p)
		}
	}
	raw, _ := os.ReadFile(filepath.Join(outDir, "findings.json"))
	if strings.TrimSpace(string(raw)) != `{"findings":[]}` {
		t.Fatalf("%s", raw)
	}
	page, _ := os.ReadFile(filepath.Join(outDir, "review.html"))
	html := string(page)
	for _, want := range []string{commit, "opus", `id="severity"`, `id="directory"`, "P0 0", "P1 0", "P2 0", "P3 0"} {
		if !strings.Contains(html, want) {
			t.Fatalf("page missing %s\n%s", want, html)
		}
	}
}

func TestRunMergesPromptOrder(t *testing.T) {
	files := map[string][]byte{
		"a.go": bytes.Repeat([]byte("a"), 100000),
		"b.go": bytes.Repeat([]byte("b"), 50000),
		"c.go": bytes.Repeat([]byte("c"), 60000),
	}
	root, _ := newRepoBytes(t, files)
	var order []string
	fr := &fakeRunner{reply: func(_ int, _ string, _ []string, stdin []byte) ([]byte, error) {
		var path string
		for _, p := range []string{"a.go", "b.go", "c.go"} {
			if bytes.Contains(stdin, []byte(p+"\n")) {
				path = p
				break
			}
		}
		order = append(order, path)
		text := `{"findings":[{"tag":"bug","file":"` + path + `","start_line":1,"end_line":1,"issue":"i-` + path + `","consequence":"c","confidence":80,"severity":"P2"}]}`
		return cliStdout("opus", text), nil
	}}
	outDir := t.TempDir()
	if err := Run(Options{Repo: root, Model: "opus", OutputDir: outDir, Stdout: &bytes.Buffer{}, Runner: fr}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(outDir, "findings.json"))
	var doc struct {
		Findings []lensoutput.TypedFinding `json:"findings"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Findings) != len(order) {
		t.Fatalf("findings %d order %v", len(doc.Findings), order)
	}
	for i, f := range doc.Findings {
		if err := f.Validate(); err != nil {
			t.Fatal(err)
		}
		if f.File != order[i] || f.Issue != "i-"+order[i] {
			t.Fatalf("order[%d]=%s finding %+v", i, order[i], f)
		}
		blob, ok := files[f.File]
		if !ok || f.EndLine > lineCount(blob) {
			t.Fatal("line missing")
		}
		asMap := map[string]any{}
		one, _ := json.Marshal(f)
		_ = json.Unmarshal(one, &asMap)
		if _, ok := asMap["model"]; ok {
			t.Fatal("model field")
		}
	}
	page, _ := os.ReadFile(filepath.Join(outDir, "review.html"))
	if !bytes.Contains(page, []byte("opus")) || bytes.Count(page, []byte(`class="finding"`)) != len(order) {
		t.Fatal("page findings")
	}
}

func TestRunFailuresLeaveDirEmpty(t *testing.T) {
	root, _ := newRepoBytes(t, specFiles())
	cases := []struct {
		name  string
		reply func() ([]byte, error)
	}{
		{"nonzero", func() ([]byte, error) { return nil, errors.New("exit status 1") }},
		{"cannot-start", func() ([]byte, error) { return nil, errors.New("executable file not found") }},
		{"is-error", func() ([]byte, error) { return []byte(`{"is_error":true,"result":"boom"}`), nil }},
		{"null", func() ([]byte, error) { return cliStdout("opus", `{"findings":null}`), nil }},
		{"prose", func() ([]byte, error) { return cliStdout("opus", "There are no findings."), nil }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			outDir := t.TempDir()
			before, _ := os.ReadDir(outDir)
			fr := &fakeRunner{reply: func(int, string, []string, []byte) ([]byte, error) { return tc.reply() }}
			err := Run(Options{Repo: root, Model: "opus", OutputDir: outDir, Stdout: &bytes.Buffer{}, Runner: fr})
			if err == nil {
				t.Fatal("expected error")
			}
			after, _ := os.ReadDir(outDir)
			if len(before) != 0 || len(after) != 0 {
				t.Fatalf("dir before %d after %d", len(before), len(after))
			}
		})
	}
}

func TestRunSkipsABadFinding(t *testing.T) {
	root, _ := newRepoBytes(t, specFiles())
	var stderr bytes.Buffer
	fr := &fakeRunner{reply: func(int, string, []string, []byte) ([]byte, error) {
		return cliStdout("opus", `{"findings":[{"tag":"bug","file":"vite.config.js","start_line":1,"end_line":1,"issue":"i","consequence":"c","confidence":10,"severity":"P1"},{"tag":"bug","file":"pkg/app.go","start_line":1,"end_line":1,"issue":"real","consequence":"c","confidence":10,"severity":"P1"}]}`), nil
	}}
	outDir := t.TempDir()
	err := Run(Options{Repo: root, Model: "opus", OutputDir: outDir, Stdout: &bytes.Buffer{}, Stderr: &stderr, Runner: fr})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr.String(), "not kept") {
		t.Fatalf("stderr %q", stderr.String())
	}
	raw, err := os.ReadFile(filepath.Join(outDir, "findings.json"))
	if err != nil || !bytes.Contains(raw, []byte(`"file":"pkg/app.go"`)) || bytes.Contains(raw, []byte("vite.config.js")) {
		t.Fatalf("%v %s", err, raw)
	}
}

func TestRunSecondPromptFails(t *testing.T) {
	files := map[string][]byte{
		"a.go": bytes.Repeat([]byte("a"), 100000),
		"b.go": bytes.Repeat([]byte("b"), 100000),
	}
	root, _ := newRepoBytes(t, files)
	fr := &fakeRunner{reply: func(n int, _ string, _ []string, _ []byte) ([]byte, error) {
		if n == 0 {
			return cliStdout("opus", `{"findings":[]}`), nil
		}
		return nil, errors.New("exit status 1")
	}}
	outDir := t.TempDir()
	if err := Run(Options{Repo: root, Model: "opus", OutputDir: outDir, Stdout: &bytes.Buffer{}, Runner: fr}); err == nil {
		t.Fatal("expected error")
	}
	ents, _ := os.ReadDir(outDir)
	if len(ents) != 0 {
		t.Fatal("wrote files after a failed prompt")
	}
}

func TestRunRefusesInsideAndWritesOutside(t *testing.T) {
	root, _ := newRepoBytes(t, droppedOnlyFiles())
	inside := filepath.Join(root, "out")
	if err := os.MkdirAll(inside, 0o755); err != nil {
		t.Fatal(err)
	}
	err := Run(Options{Repo: root, Model: "opus", OutputDir: inside, Stdout: &bytes.Buffer{}, Runner: OSRunner{}})
	if err == nil || !strings.Contains(err.Error(), "refusing") {
		t.Fatal(err)
	}
	ents, _ := os.ReadDir(inside)
	if len(ents) != 0 {
		t.Fatal("wrote inside the tree")
	}
	nested := filepath.Join(t.TempDir(), "nested", "out")
	if err := Run(Options{Repo: root, Model: "grok", OutputDir: nested, Stdout: ioDiscard(), Runner: &fakeRunner{reply: func(int, string, []string, []byte) ([]byte, error) {
		t.Fatal("model called")
		return nil, nil
	}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(nested, "review.html")); err != nil {
		t.Fatal(err)
	}
}

func ioDiscard() *bytes.Buffer { return &bytes.Buffer{} }

func TestRunSelectError(t *testing.T) {
	root, _ := newRepoBytes(t, map[string][]byte{"a.go": []byte("package a\n")})
	git := func(dir string, args []string, stdin []byte) ([]byte, error) {
		if args[0] == "ls-tree" {
			return nil, errors.New("ls-tree")
		}
		return realGit(dir, args, stdin)
	}
	err := Run(Options{Repo: root, Model: "opus", OutputDir: t.TempDir(), Stdout: &bytes.Buffer{}, Git: git, Runner: &fakeRunner{}})
	if err == nil || !strings.Contains(err.Error(), "ls-tree") {
		t.Fatal(err)
	}
}

func TestRunBudgetAndMissingRunner(t *testing.T) {
	root, _ := newRepoBytes(t, map[string][]byte{"a.go": []byte("package a\n")})
	outDir := t.TempDir()
	if err := Run(Options{Repo: root, Model: "opus", OutputDir: outDir, Budget: -1, Stdout: &bytes.Buffer{}, Runner: &fakeRunner{}}); err == nil {
		t.Fatal("negative budget")
	}
	ents, _ := os.ReadDir(outDir)
	if len(ents) != 0 {
		t.Fatal("wrote on pack error")
	}
	if err := Run(Options{Repo: root, Model: "opus", OutputDir: outDir, Stdout: &bytes.Buffer{}}); err == nil || !strings.Contains(err.Error(), "no runner") {
		t.Fatal(err)
	}
	if err := Run(Options{Repo: root, Model: "nope", OutputDir: outDir}); err == nil {
		t.Fatal("bad model")
	}
	if err := Run(Options{Repo: root, Model: "opus", Stdout: nil}); err == nil {
		t.Fatal("missing output")
	}
	if err := Run(Options{Repo: t.TempDir(), Model: "opus", OutputDir: outDir, Stdout: nil}); err == nil {
		t.Fatal("not a repo")
	}
}

func TestRunWriteErrors(t *testing.T) {
	root, _ := newRepoBytes(t, droppedOnlyFiles())
	outDir := t.TempDir()
	err := Run(Options{
		Repo: root, Model: "opus", OutputDir: outDir, Stdout: &bytes.Buffer{},
		MkdirAll: func(string, os.FileMode) error { return errors.New("mkdir") },
	})
	if err == nil || !strings.Contains(err.Error(), "mkdir") {
		t.Fatal(err)
	}
	err = Run(Options{
		Repo: root, Model: "opus", OutputDir: outDir, Stdout: &bytes.Buffer{},
		WriteFile: func(string, []byte, os.FileMode) error { return errors.New("write") },
	})
	if err == nil || !strings.Contains(err.Error(), "write") {
		t.Fatal(err)
	}
	removed := false
	n := 0
	err = Run(Options{
		Repo: root, Model: "opus", OutputDir: outDir, Stdout: &bytes.Buffer{},
		WriteFile: func(path string, data []byte, mode os.FileMode) error {
			n++
			if n == 2 {
				return errors.New("page")
			}
			return os.WriteFile(path, data, mode)
		},
		Remove: func(string) error { removed = true; return nil },
	})
	if err == nil || !removed {
		t.Fatalf("err=%v removed=%v", err, removed)
	}
	// real Remove deletes the findings file when the page write fails
	n = 0
	err = Run(Options{
		Repo: root, Model: "opus", OutputDir: outDir, Stdout: &bytes.Buffer{},
		WriteFile: func(path string, data []byte, mode os.FileMode) error {
			n++
			if strings.HasSuffix(path, "review.html") {
				return errors.New("page")
			}
			return os.WriteFile(path, data, mode)
		},
	})
	if err == nil {
		t.Fatal("expected page error")
	}
	if _, statErr := os.Stat(filepath.Join(outDir, "findings.json")); !os.IsNotExist(statErr) {
		t.Fatal("findings file left behind")
	}
}

func TestRefuseSeams(t *testing.T) {
	root, _ := newRepoBytes(t, map[string][]byte{"a.go": []byte("package a\n")})
	out := t.TempDir()
	prevAbs, prevEval, prevRel, prevStat := pathAbs, pathEval, pathRel, statPath
	t.Cleanup(func() {
		pathAbs, pathEval, pathRel, statPath = prevAbs, prevEval, prevRel, prevStat
	})
	pathAbs = func(string) (string, error) { return "", errors.New("abs") }
	if err := refuseInside(root, out); err == nil {
		t.Fatal("abs1")
	}
	n := 0
	pathAbs = func(p string) (string, error) {
		n++
		if n == 2 {
			return "", errors.New("abs2")
		}
		return prevAbs(p)
	}
	if err := refuseInside(root, out); err == nil {
		t.Fatal("abs2")
	}
	pathAbs = prevAbs
	pathEval = func(string) (string, error) { return "", errors.New("eval") }
	if err := refuseInside(root, out); err == nil {
		t.Fatal("eval repo")
	}
	calls := 0
	pathEval = func(p string) (string, error) {
		calls++
		if calls > 1 {
			return "", errors.New("eval out")
		}
		return prevEval(p)
	}
	if err := refuseInside(root, out); err == nil {
		t.Fatal("eval out")
	}
	pathEval = prevEval
	pathRel = func(string, string) (string, error) { return "", errors.New("rel") }
	if err := refuseInside(root, out); err == nil {
		t.Fatal("rel")
	}
	pathRel = prevRel
	statPath = func(string) (os.FileInfo, error) { return nil, errors.New("stat") }
	if _, err := evalExisting(out); err == nil {
		t.Fatal("stat walk")
	}
	if !insideRel(".") || insideRel("..") || insideRel("../x") || !insideRel("foo") {
		t.Fatal("insideRel")
	}
}

func TestPromptWithoutTrailingNewline(t *testing.T) {
	root, _ := newRepoBytes(t, map[string][]byte{"nonl.go": []byte("package nonl")})
	var buf bytes.Buffer
	fr := &fakeRunner{reply: func(int, string, []string, []byte) ([]byte, error) {
		return cliStdout("opus", `{"findings":[]}`), nil
	}}
	outDir := t.TempDir()
	if err := Run(Options{Repo: root, Model: "opus", OutputDir: outDir, Stdout: &buf, Runner: fr}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "prompt-end\n") {
		t.Fatalf("prompt marker missing:\n%s", buf.String())
	}
	if !bytes.Contains(fr.calls[0].stdin, []byte("nonl.go\npackage nonl")) {
		t.Fatal("path not immediately before text")
	}
}

func TestCLI(t *testing.T) {
	code, _ := cliCapture(t, "", "--help")
	if code != 0 {
		t.Fatal(code)
	}
	code, errOut := cliCapture(t, "", "-h")
	if code != 0 || errOut != "" {
		t.Fatalf("h code=%d err=%q", code, errOut)
	}
	for _, args := range [][]string{
		{},
		{"--model"},
		{"--output"},
		{"--model", "opus"},
		{"--output", t.TempDir()},
		{"--nope"},
		{"--model", "opus", "--output", t.TempDir(), "extra", "more"},
	} {
		code, errOut = cliCapture(t, "", args...)
		if code != 2 || errOut == "" {
			t.Fatalf("args %v code=%d err=%q", args, code, errOut)
		}
	}
	root, commit := newRepoBytes(t, droppedOnlyFiles())
	outDir := t.TempDir()
	code, errOut = cliCapture(t, root, "--model", "opus", "--output", outDir)
	if code != 0 || errOut != "" {
		t.Fatalf("cli code=%d err=%q", code, errOut)
	}
	page, err := os.ReadFile(filepath.Join(outDir, "review.html"))
	if err != nil || !bytes.Contains(page, []byte(commit)) {
		t.Fatal(err)
	}
	code, errOut = cliCapture(t, root, "--model", "opus", "--output", root)
	if code != 1 || !strings.Contains(errOut, "refusing") {
		t.Fatalf("refuse code=%d err=%q", code, errOut)
	}
}

func cliCapture(t *testing.T, cwd string, args ...string) (int, string) {
	t.Helper()
	if cwd == "" {
		cwd = t.TempDir()
	}
	var stdout, stderr bytes.Buffer
	code := CLI(args, cwd, &stdout, &stderr)
	return code, stderr.String()
}

func TestFindingsJSONNil(t *testing.T) {
	if strings.TrimSpace(string(findingsJSON(nil))) != `{"findings":[]}` {
		t.Fatal(findingsJSON(nil))
	}
	body := findingsJSON([]lensoutput.TypedFinding{{
		Tag: lensoutput.TagBug, File: "a.go", StartLine: 1, EndLine: 1,
		Issue: "i", Consequence: "c", Confidence: 1, Severity: "P3",
	}})
	if !bytes.Contains(body, []byte(`"severity":"P3"`)) || bytes.Contains(body, []byte(`"model"`)) {
		t.Fatalf("%s", body)
	}
}
