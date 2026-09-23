package sourcereview

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// MaxPromptBytes is the prompt budget. AGENTS.md uses 120000 as the size that
// cannot be held in one review context.
const MaxPromptBytes = 120000

// instruction is the same text for every model. Only the CLI and the model id change.
const instruction = `Review the first-party source below and return only {"findings":[...]}.
Each finding's tag is bug or advisory. Each finding's severity is P0, P1, P2, or P3.
Every field is present: tag, file, start_line, end_line, issue, consequence, confidence, severity.
confidence is an integer from 0 to 100.
file is one of the paths given below. start_line and end_line are line numbers in that file.
Return {"findings":[]} when there is nothing to report.
Do not write anything outside that JSON object.`

func sliceNote(path string, line int) string {
	return fmt.Sprintf("This prompt contains a byte slice of %s starting at original line %d. Citations use the original path and the original line numbers.", path, line)
}

func sliceHeader(path string, line int) string {
	return instruction + "\n" + sliceNote(path, line) + "\n"
}

// promptWith writes the instruction header and each path on its own line
// immediately before that file's bytes. A following path stays on its own line
// when the previous body does not end in a newline.
func promptWith(header string, files []File) []byte {
	var b bytes.Buffer
	b.WriteString(header)
	if !strings.HasSuffix(header, "\n") {
		b.WriteByte('\n')
	}
	for i, f := range files {
		if i > 0 && (len(files[i-1].Body) == 0 || files[i-1].Body[len(files[i-1].Body)-1] != '\n') {
			b.WriteByte('\n')
		}
		b.WriteString(f.Path)
		b.WriteByte('\n')
		b.Write(f.Body)
	}
	return b.Bytes()
}

// Pack packs whole kept files into the smallest number of prompts that stay
// within budget (first-fit decreasing). A file that cannot fit in one prompt
// is cut into consecutive non-empty UTF-8 slices that partition it.
func Pack(files []File, budget int) ([][]byte, error) {
	if budget <= 0 {
		return nil, fmt.Errorf("prompt budget must be positive")
	}
	var whole, big []File
	for _, f := range files {
		if len(promptWith(instruction, []File{f})) <= budget {
			whole = append(whole, f)
		} else {
			big = append(big, f)
		}
	}
	prompts := packWhole(whole, budget)
	sort.Slice(big, func(i, j int) bool { return big[i].Path < big[j].Path })
	for _, f := range big {
		parts, err := sliceFile(f, budget)
		if err != nil {
			return nil, err
		}
		prompts = append(prompts, parts...)
	}
	return prompts, nil
}

func packWhole(files []File, budget int) [][]byte {
	sorted := append([]File(nil), files...)
	sort.Slice(sorted, func(i, j int) bool {
		if len(sorted[i].Body) == len(sorted[j].Body) {
			return sorted[i].Path < sorted[j].Path
		}
		return len(sorted[i].Body) > len(sorted[j].Body)
	})
	var bins [][]File
	for _, f := range sorted {
		placed := false
		for i := range bins {
			candidate := append(append([]File{}, bins[i]...), f)
			if len(promptWith(instruction, candidate)) <= budget {
				bins[i] = candidate
				placed = true
				break
			}
		}
		if !placed {
			bins = append(bins, []File{f})
		}
	}
	out := make([][]byte, 0, len(bins))
	for _, bin := range bins {
		out = append(out, promptWith(instruction, bin))
	}
	return out
}

func sliceFile(f File, budget int) ([][]byte, error) {
	var out [][]byte
	start := 0
	for start < len(f.Body) {
		line := lineAt(f.Body, start)
		header := sliceHeader(f.Path, line)
		room := budget - len(header) - len(f.Path) - 1
		end, err := cutEnd(f.Body, start, room)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f.Path, err)
		}
		out = append(out, promptWith(header, []File{{Path: f.Path, Body: f.Body[start:end]}}))
		start = end
	}
	return out, nil
}

func lineAt(content []byte, offset int) int {
	return bytes.Count(content[:offset], []byte("\n")) + 1
}

// cutEnd returns the byte end of a non-empty slice that starts at start, is at
// most room bytes, and ends on a UTF-8 character boundary.
func cutEnd(content []byte, start, room int) (int, error) {
	if room < 1 {
		return 0, fmt.Errorf("prompt overhead exceeds the budget")
	}
	end := start + room
	if end > len(content) {
		end = len(content)
	} else {
		for end > start && !utf8.RuneStart(content[end]) {
			end--
		}
	}
	if end == start {
		return 0, fmt.Errorf("a character does not fit in one prompt")
	}
	return end, nil
}
