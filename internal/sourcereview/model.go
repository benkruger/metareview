package sourcereview

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Runner runs one logged-in CLI. stdin is the prompt for claude and codex;
// grok receives the prompt as the argument of -p and stdin is nil.
type Runner interface {
	Run(name string, args []string, stdin []byte) ([]byte, error)
}

// OSRunner is the real process runner. It does not read an API key and does not
// add a token to the command.
type OSRunner struct{}

func (OSRunner) Run(name string, args []string, stdin []byte) ([]byte, error) {
	cmd := exec.Command(name, args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
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

// Call is the one function that runs claude, codex, or grok. The prompt text
// and the parser stay the same; only the CLI and the model id change.
func Call(runner Runner, model string, prompt []byte) (string, error) {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "opus":
		out, err := runner.Run("claude", []string{"-p", "--model", "opus", "--output-format", "json"}, prompt)
		if err != nil {
			return "", err
		}
		return claudeModelText(out)
	case "astra":
		out, err := runner.Run("codex", []string{"exec", "--json", "-m", "gpt-6-astra", "-"}, prompt)
		if err != nil {
			return "", err
		}
		return codexModelText(out)
	case "grok":
		out, err := runner.Run("grok", []string{"-p", string(prompt), "-m", "grok-4.7", "--output-format", "json"}, nil)
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
