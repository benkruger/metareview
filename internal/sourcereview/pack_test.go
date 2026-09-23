package sourcereview

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestPackWholeFilesSmallest(t *testing.T) {
	files := []File{
		{Path: "a.go", Body: bytes.Repeat([]byte("x"), 100000)},
		{Path: "b.go", Body: bytes.Repeat([]byte("y"), 50000)},
		{Path: "c.go", Body: bytes.Repeat([]byte("z"), 60000)},
	}
	prompts, err := Pack(files, MaxPromptBytes)
	if err != nil {
		t.Fatal(err)
	}
	if len(prompts) != 2 {
		t.Fatalf("prompts = %d, want 2", len(prompts))
	}
	seen := map[string]int{}
	for _, p := range prompts {
		if len(p) > MaxPromptBytes {
			t.Fatalf("prompt is %d bytes", len(p))
		}
		if !bytes.Contains(p, []byte(instruction)) {
			t.Fatal("missing instruction")
		}
		if bytes.Contains(p, []byte("byte slice of")) {
			t.Fatal("whole file prompt has a slice note")
		}
		for _, f := range files {
			needle := append([]byte(f.Path+"\n"), f.Body...)
			if bytes.Contains(p, needle) {
				seen[f.Path]++
			}
		}
	}
	for _, f := range files {
		if seen[f.Path] != 1 {
			t.Fatalf("%s seen %d times", f.Path, seen[f.Path])
		}
	}
}

func TestPackSlicesPartitionUTF8(t *testing.T) {
	body := bytes.Repeat([]byte("世"), 70000)
	f := File{Path: "pkg/wide.go", Body: body}
	prompts, err := Pack([]File{f}, MaxPromptBytes)
	if err != nil {
		t.Fatal(err)
	}
	if len(prompts) < 2 {
		t.Fatalf("prompts = %d", len(prompts))
	}
	off := 0
	files := map[string][]byte{f.Path: body}
	for _, p := range prompts {
		if len(p) > MaxPromptBytes {
			t.Fatalf("prompt is %d bytes", len(p))
		}
		header := sliceHeader(f.Path, lineAt(body, off))
		prefix := header + f.Path + "\n"
		if !strings.HasPrefix(string(p), prefix) {
			t.Fatalf("prompt missing slice header at offset %d", off)
		}
		part := p[len(prefix):]
		if len(part) == 0 {
			t.Fatal("empty slice")
		}
		if !utf8.Valid(part) {
			t.Fatal("slice is not valid UTF-8")
		}
		if !bytes.HasPrefix(body[off:], part) {
			t.Fatal("slice is not the next bytes of the file")
		}
		line := lineAt(body, off)
		text := `{"findings":[{"tag":"bug","file":"pkg/wide.go","start_line":` + strconv.Itoa(line) + `,"end_line":` + strconv.Itoa(line) + `,"issue":"i","consequence":"c","confidence":80,"severity":"P1"}]}`
		got, _, err := Parse(text, files)
		if err != nil {
			t.Fatalf("finding at original line %d: %v", line, err)
		}
		if got[0].File != f.Path || got[0].StartLine != line {
			t.Fatalf("citation %+v", got[0])
		}
		off += len(part)
	}
	if off != len(body) {
		t.Fatalf("partition covered %d of %d", off, len(body))
	}
}

func TestPackMixedAndSeparators(t *testing.T) {
	small := File{Path: "small.go", Body: []byte("package small\n")}
	empty := File{Path: "empty.go", Body: nil}
	nonl := File{Path: "nonl.go", Body: []byte("package nonl")}
	big := File{Path: "big.go", Body: bytes.Repeat([]byte("世"), 70000)}
	prompts, err := Pack([]File{small, empty, nonl, big}, MaxPromptBytes)
	if err != nil {
		t.Fatal(err)
	}
	var whole, slices int
	for _, p := range prompts {
		if bytes.Contains(p, []byte("byte slice of")) {
			slices++
			if !bytes.Contains(p, []byte("big.go\n")) {
				t.Fatal("slice prompt missing path")
			}
			continue
		}
		whole++
		for _, f := range []File{small, empty, nonl} {
			if !bytes.Contains(p, []byte(f.Path+"\n")) {
				t.Fatalf("whole prompt missing %s", f.Path)
			}
		}
		if bytes.Contains(p, []byte("big.go\n")) {
			t.Fatal("big file packed whole")
		}
	}
	if whole != 1 || slices < 2 {
		t.Fatalf("whole=%d slices=%d", whole, slices)
	}
}

func TestCutEndAndSliceErrors(t *testing.T) {
	if _, err := cutEnd([]byte("世"), 0, 0); err == nil {
		t.Fatal("room 0")
	}
	if _, err := cutEnd([]byte("世"), 0, 1); err == nil {
		t.Fatal("rune does not fit")
	}
	end, err := cutEnd([]byte("世世"), 0, 4)
	if err != nil || end != 3 {
		t.Fatalf("end=%d err=%v", end, err)
	}
	end, err = cutEnd([]byte("ab"), 0, 10)
	if err != nil || end != 2 {
		t.Fatalf("short end=%d err=%v", end, err)
	}
	header := sliceHeader("p.go", 1)
	if _, err := sliceFile(File{Path: "p.go", Body: []byte("世")}, len(header)+len("p.go")+1+1); err == nil {
		t.Fatal("sliceFile should reject a rune that does not fit")
	}
	if _, err := sliceFile(File{Path: "p.go", Body: []byte("hello")}, 10); err == nil {
		t.Fatal("sliceFile should reject a tiny budget")
	}
	if _, err := Pack([]File{{Path: "a.go", Body: []byte("package a\n")}}, 0); err == nil {
		t.Fatal("budget 0")
	}
	if _, err := Pack([]File{{Path: "a.go", Body: bytes.Repeat([]byte("x"), 50)}}, 10); err == nil {
		t.Fatal("pack should surface the slice error")
	}
	if prompts, err := Pack(nil, MaxPromptBytes); err != nil || len(prompts) != 0 {
		t.Fatalf("empty pack %v %v", prompts, err)
	}
	two, err := Pack([]File{
		{Path: "z.go", Body: bytes.Repeat([]byte("世"), 70000)},
		{Path: "a.go", Body: bytes.Repeat([]byte("世"), 70000)},
	}, MaxPromptBytes)
	if err != nil || len(two) < 2 {
		t.Fatalf("two big files: %v %d", err, len(two))
	}
}

func TestPromptNamesPathBeforeBody(t *testing.T) {
	body := []byte("package pkg\nfunc A() {}\n")
	p := promptWith(instruction, []File{{Path: "pkg/app.go", Body: body}})
	if !bytes.Contains(p, append([]byte("pkg/app.go\n"), body...)) {
		t.Fatalf("path is not immediately before the body:\n%s", p)
	}
	if !bytes.Contains(p, []byte(`{"findings":[...]}`)) {
		t.Fatal("instruction missing findings object")
	}
}
