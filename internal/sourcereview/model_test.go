package sourcereview

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

type fakeRunner struct {
	calls []fakeCall
	reply func(n int, name string, args []string, stdin []byte) ([]byte, error)
	onRun func()
}

type fakeCall struct {
	name  string
	args  []string
	stdin []byte
}

func (f *fakeRunner) Run(name string, args []string, stdin []byte) ([]byte, error) {
	if f.onRun != nil {
		f.onRun()
	}
	f.calls = append(f.calls, fakeCall{name, append([]string(nil), args...), append([]byte(nil), stdin...)})
	if f.reply == nil {
		return nil, errors.New("no reply")
	}
	return f.reply(len(f.calls)-1, name, args, stdin)
}

func TestModelID(t *testing.T) {
	cases := map[string]string{"opus": "opus", " astra ": "gpt-6-astra", "GROK": "grok-4.7"}
	for in, want := range cases {
		got, err := ModelID(in)
		if err != nil || got != want {
			t.Fatalf("%q -> %q %v", in, got, err)
		}
	}
	if _, err := ModelID("sonnet"); err == nil {
		t.Fatal("expected unknown model")
	}
}

func TestCallShapes(t *testing.T) {
	prompt := []byte("review this")
	for _, model := range []string{"opus", "astra", "grok"} {
		fr := &fakeRunner{reply: func(int, string, []string, []byte) ([]byte, error) {
			switch model {
			case "opus":
				return []byte(`{"is_error":false,"result":"{\"findings\":[]}"}`), nil
			case "astra":
				return []byte("not json\n{\"type\":\"item.completed\",\"item\":{\"type\":\"agent_message\",\"text\":\"first\"}}\n{\"type\":\"item.completed\",\"item\":{\"type\":\"agent_message\",\"text\":\"{\\\"findings\\\":[]}\"}}\n"), nil
			default:
				return []byte(`{"text":"{\"findings\":[]}"}`), nil
			}
		}}
		text, err := Call(fr, model, prompt)
		if err != nil {
			t.Fatal(model, err)
		}
		if model == "astra" {
			if text != `{"findings":[]}` {
				t.Fatalf("astra text %q", text)
			}
		} else if text != `{"findings":[]}` {
			t.Fatalf("%s text %q", model, text)
		}
		c := fr.calls[0]
		switch model {
		case "opus":
			if c.name != "claude" || !reflectArgs(c.args, []string{"-p", "--model", "opus", "--output-format", "json"}) || !bytes.Equal(c.stdin, prompt) {
				t.Fatalf("claude call %+v", c)
			}
		case "astra":
			if c.name != "codex" || !reflectArgs(c.args, []string{"exec", "--json", "-m", "gpt-6-astra", "-"}) || !bytes.Equal(c.stdin, prompt) {
				t.Fatalf("codex call %+v", c)
			}
		case "grok":
			if c.name != "grok" || !reflectArgs(c.args, []string{"-p", string(prompt), "-m", "grok-4.7", "--output-format", "json"}) || c.stdin != nil {
				t.Fatalf("grok call %+v", c)
			}
		}
	}
}

func reflectArgs(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestCallErrors(t *testing.T) {
	fr := &fakeRunner{reply: func(int, string, []string, []byte) ([]byte, error) {
		return nil, errors.New("executable file not found")
	}}
	if _, err := Call(fr, "opus", []byte("p")); err == nil {
		t.Fatal("start failure")
	}
	fr.reply = func(int, string, []string, []byte) ([]byte, error) {
		return []byte(`{"is_error":true,"result":"boom"}`), nil
	}
	if _, err := Call(fr, "opus", []byte("p")); err == nil || !strings.Contains(err.Error(), "is_error") {
		t.Fatal(err)
	}
	fr.reply = func(int, string, []string, []byte) ([]byte, error) {
		return []byte(`{"is_error":false}`), nil
	}
	if _, err := Call(fr, "opus", []byte("p")); err == nil {
		t.Fatal("missing result")
	}
	fr.reply = func(int, string, []string, []byte) ([]byte, error) {
		return []byte(`{"is_error":false,"result":""}`), nil
	}
	if _, err := Call(fr, "opus", []byte("p")); err == nil {
		t.Fatal("empty result")
	}
	fr.reply = func(int, string, []string, []byte) ([]byte, error) {
		return []byte("not-json"), nil
	}
	if _, err := Call(fr, "opus", []byte("p")); err == nil {
		t.Fatal("bad claude json")
	}
	if _, err := Call(fr, "grok", []byte("p")); err == nil {
		t.Fatal("bad grok json")
	}
	fr.reply = func(int, string, []string, []byte) ([]byte, error) {
		return []byte(`{"text":""}`), nil
	}
	if _, err := Call(fr, "grok", []byte("p")); err == nil {
		t.Fatal("empty grok text")
	}
	fr.reply = func(int, string, []string, []byte) ([]byte, error) {
		return []byte(`{}`), nil
	}
	if _, err := Call(fr, "grok", []byte("p")); err == nil {
		t.Fatal("nil grok text")
	}
	fr.reply = func(int, string, []string, []byte) ([]byte, error) {
		return []byte("{\"type\":\"item.completed\",\"item\":{\"type\":\"agent_message\",\"text\":\"\"}}\n"), nil
	}
	if _, err := Call(fr, "astra", []byte("p")); err == nil {
		t.Fatal("empty codex text")
	}
	fr.reply = func(int, string, []string, []byte) ([]byte, error) {
		return []byte("nope\n"), nil
	}
	if _, err := Call(fr, "astra", []byte("p")); err == nil {
		t.Fatal("no agent message")
	}
	if _, err := Call(fr, "nope", []byte("p")); err == nil {
		t.Fatal("unknown model")
	}
	fr.reply = func(int, string, []string, []byte) ([]byte, error) {
		return nil, errors.New("exit status 1")
	}
	if _, err := Call(fr, "astra", []byte("p")); err == nil {
		t.Fatal("nonzero")
	}
	if _, err := Call(fr, "grok", []byte("p")); err == nil {
		t.Fatal("grok nonzero")
	}
}

func TestOSRunner(t *testing.T) {
	out, err := OSRunner{}.Run("echo", []string{"hi"}, nil)
	if err != nil || !bytes.Contains(out, []byte("hi")) {
		t.Fatalf("echo: %v %q", err, out)
	}
	out, err = OSRunner{}.Run("cat", nil, []byte("yo"))
	if err != nil || !bytes.Equal(out, []byte("yo")) {
		t.Fatalf("cat: %v %q", err, out)
	}
	if _, err := (OSRunner{}).Run("false", nil, nil); err == nil {
		t.Fatal("false should fail")
	}
	if _, err := (OSRunner{}).Run("sh", []string{"-c", "echo boom >&2; exit 1"}, nil); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatal(err)
	}
	if _, err := (OSRunner{}).Run("sh", []string{"-c", "echo usage-limit; exit 1"}, nil); err == nil || !strings.Contains(err.Error(), "usage-limit") {
		t.Fatal(err)
	}
	if _, err := (OSRunner{}).Run("sh", []string{"-c", "echo out-line; echo err-line >&2; exit 1"}, nil); err == nil || !strings.Contains(err.Error(), "err-line") || !strings.Contains(err.Error(), "out-line") {
		t.Fatal(err)
	}
	if _, err := (OSRunner{}).Run("no-such-binary-xyz", nil, nil); err == nil {
		t.Fatal("missing binary")
	}
	// exec.Command is used; a lookup failure is *exec.Error, not ExitError.
	if _, err := exec.LookPath("echo"); err != nil {
		t.Fatal(err)
	}
}

func TestClaudeIsErrorWithoutResult(t *testing.T) {
	if _, err := claudeModelText([]byte(`{"is_error":true}`)); err == nil {
		t.Fatal("expected is_error")
	}
}

func TestCodexSkipsBlank(t *testing.T) {
	raw := "\n  \n{\"type\":\"turn.completed\"}\n{\"type\":\"item.completed\",\"item\":{\"type\":\"agent_message\",\"text\":\"{\\\"findings\\\":[]}\"}}\n"
	text, err := codexModelText([]byte(raw))
	if err != nil || text != `{"findings":[]}` {
		t.Fatalf("%q %v", text, err)
	}
	b, err := json.Marshal(map[string]string{"text": "ok"})
	if err != nil || !bytes.Contains(b, []byte("ok")) {
		t.Fatal(err)
	}
}
