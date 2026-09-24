package sourcereview

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// Runner runs one logged-in CLI. stdin is the prompt for claude and codex;
// grok reads the prompt from the --prompt-file file and stdin is nil. When ctx
// ends, the process and its children are killed.
type Runner interface {
	Run(ctx context.Context, name string, args []string, stdin []byte) ([]byte, error)
}

// OSRunner is the real process runner. It does not read an API key and does not
// add a token to the command.
type OSRunner struct{}

// waitDelay bounds how long Run waits for output pipes after the process group
// is killed (the internal/fsm/cmdexec pattern).
const waitDelay = 2 * time.Second

func (OSRunner) Run(ctx context.Context, name string, args []string, stdin []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = waitDelay
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		// Codex puts the turn failure on stdout and a snapshot warning on
		// stderr. Keep both so a non-zero exit records the CLI's reason.
		msg := strings.TrimSpace(stderr.String())
		if out := strings.TrimSpace(stdout.String()); out != "" {
			if msg != "" {
				msg = msg + "\n" + out
			} else {
				msg = out
			}
		}
		if msg == "" {
			msg = err.Error()
		}
		return stdout.Bytes(), fmt.Errorf("%s: %s", name, msg)
	}
	return stdout.Bytes(), nil
}

// ModelID is the id shown on the page and passed to the CLI.
func ModelID(model string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "opus":
		return "opus", nil
	case "astra":
		return "gpt-6-astra", nil
	case "grok":
		return "grok-4.7", nil
	default:
		return "", fmt.Errorf("unknown model %q (want astra, opus, or grok)", model)
	}
}

// grokSchema constrains Grok's answer to {"findings":[...]} with all eight
// fields present.
const grokSchema = `{"type":"object","properties":{"findings":{"type":"array","items":{"type":"object","properties":{"tag":{"type":"string","enum":["bug","advisory"]},"file":{"type":"string"},"start_line":{"type":"integer"},"end_line":{"type":"integer"},"issue":{"type":"string"},"consequence":{"type":"string"},"confidence":{"type":"integer"},"severity":{"type":"string","enum":["P0","P1","P2","P3"]}},"required":["tag","file","start_line","end_line","issue","consequence","confidence","severity"]}}},"required":["findings"]}`

// grokSystem replaces Grok's agent system prompt. Without it the model tries
// to call tools, and --max-turns 1 cancels the turn.
const grokSystem = "You are a code reviewer. You have no tools. Everything you need is in the user message. Answer in a single message."

// grokArgs runs Grok as one turn with no tools. -p hands a large prompt to
// Grok's agent as an excerpt plus a file to read back with tools, which made
// one hh prompt take 11 model calls; --prompt-file with --verbatim sends it whole.
func grokArgs(promptFile string) []string {
	return []string{
		"--prompt-file", promptFile, "--verbatim",
		"-m", "grok-4.7", "--output-format", "json",
		"--json-schema", grokSchema,
		"--tools", "", "--max-turns", "1", "--no-subagents", "--disable-web-search",
		"--system-prompt-override", grokSystem,
	}
}

// Temp-file seams. Production uses the os functions. Tests replace them to
// cover the error returns.
var (
	createTemp = os.CreateTemp
	removeFile = os.Remove
)

// writePromptFile writes the prompt to a new file in os.TempDir(), which is
// never inside the reviewed tree or the output directory.
func writePromptFile(prompt []byte) (string, error) {
	f, err := createTemp("", "metareview-grok-prompt-*.txt")
	if err != nil {
		return "", err
	}
	_, werr := f.Write(prompt)
	cerr := f.Close()
	if err := errors.Join(werr, cerr); err != nil {
		_ = removeFile(f.Name())
		return "", err
	}
	return f.Name(), nil
}

// Call is the one function that runs claude, codex, or grok. The prompt text
// and the parser stay the same; only the CLI and the model id change.
func Call(ctx context.Context, runner Runner, model string, prompt []byte) (string, error) {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "opus":
		out, err := runner.Run(ctx, "claude", []string{"-p", "--model", "opus", "--output-format", "json"}, prompt)
		if err != nil {
			return "", err
		}
		return claudeModelText(out)
	case "astra":
		out, err := runner.Run(ctx, "codex", []string{"exec", "--json", "-m", "gpt-6-astra", "-"}, prompt)
		if err != nil {
			return "", err
		}
		return codexModelText(out)
	case "grok":
		path, err := writePromptFile(prompt)
		if err != nil {
			return "", err
		}
		defer func() { _ = removeFile(path) }()
		out, err := runner.Run(ctx, "grok", grokArgs(path), nil)
		if err != nil {
			return "", err
		}
		return grokModelText(out)
	default:
		return "", fmt.Errorf("unknown model %q", model)
	}
}

func claudeModelText(stdout []byte) (string, error) {
	var doc struct {
		IsError bool    `json:"is_error"`
		Result  *string `json:"result"`
	}
	if err := json.Unmarshal(stdout, &doc); err != nil {
		return "", fmt.Errorf("missing model text: %w", err)
	}
	if doc.IsError {
		msg := ""
		if doc.Result != nil {
			msg = *doc.Result
		}
		return "", fmt.Errorf("claude: is_error: %s", msg)
	}
	if doc.Result == nil || *doc.Result == "" {
		return "", errors.New("missing model text")
	}
	return *doc.Result, nil
}

func codexModelText(stdout []byte) (string, error) {
	var text string
	found := false
	for _, line := range strings.Split(string(stdout), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var ev struct {
			Type string `json:"type"`
			Item struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"item"`
		}
		if json.Unmarshal([]byte(line), &ev) != nil {
			continue
		}
		if ev.Type == "item.completed" && ev.Item.Type == "agent_message" {
			text = ev.Item.Text
			found = true
		}
	}
	if !found || text == "" {
		return "", errors.New("missing model text")
	}
	return text, nil
}

func grokModelText(stdout []byte) (string, error) {
	var doc struct {
		Text *string `json:"text"`
	}
	if err := json.Unmarshal(stdout, &doc); err != nil {
		return "", fmt.Errorf("missing model text: %w", err)
	}
	if doc.Text == nil || *doc.Text == "" {
		return "", errors.New("missing model text")
	}
	return *doc.Text, nil
}
