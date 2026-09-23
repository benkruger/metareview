package sourcereview

import (
	"fmt"
	"html"
	"path"
	"sort"
	"strings"

	"github.com/dsifry/metareview/internal/lensoutput"
)

// Page is the input to the self-contained HTML file.
type Page struct {
	Repo     string
	Commit   string
	ModelID  string
	Findings []lensoutput.TypedFinding
	Files    map[string][]byte
}

func severityCounts(fs []lensoutput.TypedFinding) (p0, p1, p2, p3 int) {
	for _, f := range fs {
		switch f.Severity {
		case "P0":
			p0++
		case "P1":
			p1++
		case "P2":
			p2++
		case "P3":
			p3++
		}
	}
	return p0, p1, p2, p3
}

func findingDirs(fs []lensoutput.TypedFinding) []string {
	seen := map[string]bool{}
	var dirs []string
	for _, f := range fs {
		d := path.Dir(f.File)
		if !seen[d] {
			seen[d] = true
			dirs = append(dirs, d)
		}
	}
	sort.Strings(dirs)
	return dirs
}

func excerpt(body []byte, start, end int) string {
	lines := splitLines(body)
	if start < 1 || end > len(lines) || end < start {
		return ""
	}
	return strings.Join(lines[start-1:end], "\n")
}

const pageScript = `
(function () {
  var severity = document.getElementById("severity");
  var directory = document.getElementById("directory");
  var findings = document.querySelectorAll("#findings .finding");
  function apply() {
    var s = severity.value;
    var d = directory.value;
    for (var i = 0; i < findings.length; i++) {
      var el = findings[i];
      var okS = s === "" || el.getAttribute("data-severity") === s;
      var okD = d === "" || el.getAttribute("data-dir") === d;
      el.style.display = (okS && okD) ? "" : "none";
    }
  }
  severity.addEventListener("change", apply);
  directory.addEventListener("change", apply);
  var buttons = document.querySelectorAll("#findings .finding > button");
  for (var j = 0; j < buttons.length; j++) {
    buttons[j].addEventListener("click", function () {
      var detail = this.parentNode.querySelector(".detail");
      detail.hidden = !detail.hidden;
    });
  }
})();
`

const pageCSS = `
:root {
  --paper: #e4ebf0;
  --ink: #17202a;
  --muted: #4e5d6c;
  --line: #c3cdd6;
  --sheet: #f4f7f9;
  --p0: #9b2335;
  --p1: #c05621;
  --p2: #8a6d1b;
  --p3: #1f6f68;
}
* { box-sizing: border-box; }
html { background: var(--paper); }
body {
  margin: 0 auto;
  max-width: 68rem;
  padding: 2rem 1.5rem 4rem;
  color: var(--ink);
  font: 17px/1.5 Charter, "Iowan Old Style", Palatino, Georgia, serif;
}
header { padding-bottom: 1.1rem; border-bottom: 1px solid var(--line); }
#repo {
  margin: 0;
  font-family: "Avenir Next", "Segoe UI", sans-serif;
  font-size: 1.6rem;
  font-weight: 600;
  letter-spacing: -0.02em;
  line-height: 1.2;
  overflow-wrap: anywhere;
}
.identity {
  display: grid;
  gap: 0.15rem;
  margin-top: 0.45rem;
  color: var(--muted);
  font-family: "Avenir Next", "Segoe UI", sans-serif;
  font-size: 0.92rem;
}
#commit { overflow-wrap: anywhere; }
.identity p { margin: 0; }
.counts {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
  margin: 0.9rem 0 0;
  padding: 0;
  list-style: none;
  font-family: "Avenir Next", "Segoe UI", sans-serif;
  font-variant-numeric: tabular-nums;
}
.counts li {
  min-width: 4.6rem;
  padding: 0.35rem 0.7rem 0.28rem;
  background: var(--sheet);
  border-top: 3px solid var(--line);
}
#count-p0 { border-top-color: var(--p0); }
#count-p1 { border-top-color: var(--p1); }
#count-p2 { border-top-color: var(--p2); }
#count-p3 { border-top-color: var(--p3); }
.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.8rem 1.6rem;
  margin: 1.15rem 0;
  font-family: "Avenir Next", "Segoe UI", sans-serif;
  font-size: 0.95rem;
}
select {
  margin-left: 0.35rem;
  padding: 0.28rem 0.4rem;
  color: var(--ink);
  background: var(--sheet);
  border: 1px solid var(--line);
  border-radius: 2px;
  font: inherit;
}
#findings { border-bottom: 1px solid var(--line); }
article.finding {
  margin: 0;
  background: var(--sheet);
  border-top: 1px solid var(--line);
  border-left: 5px solid var(--line);
}
article.finding[data-severity="P0"] { border-left-color: var(--p0); }
article.finding[data-severity="P1"] { border-left-color: var(--p1); }
article.finding[data-severity="P2"] { border-left-color: var(--p2); }
article.finding[data-severity="P3"] { border-left-color: var(--p3); }
article.finding > button {
  display: block;
  width: 100%;
  padding: 0.72rem 1rem;
  color: var(--ink);
  background: transparent;
  border: 0;
  cursor: pointer;
  font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace;
  font-size: 0.9rem;
  text-align: left;
  white-space: normal;
  overflow-wrap: anywhere;
}
article.finding > button:hover,
article.finding > button:focus-visible { background: #e7eef3; }
article.finding > button:focus-visible { outline: 2px solid var(--ink); outline-offset: -2px; }
.detail { display: none; }
.detail:not([hidden]) {
  display: grid;
  grid-template-columns: minmax(16rem, 0.9fr) minmax(18rem, 1.15fr);
  gap: 1rem 1.5rem;
  padding: 0 1rem 1rem 1.1rem;
}
.issue { margin: 0 0 0.55rem; font-size: 1.05rem; }
.consequence { margin: 0; color: var(--muted); }
.quote {
  margin: 0;
  padding: 0.75rem 0.9rem;
  overflow: auto;
  background: #e7eef3;
  color: var(--ink);
  border-left: 1px solid var(--line);
  font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace;
  font-size: 0.8rem;
  line-height: 1.45;
  white-space: pre;
}
@media (max-width: 720px) {
  body { padding: 1rem 0.75rem 2.5rem; }
  #repo { font-size: 1.25rem; }
  .detail:not([hidden]) { grid-template-columns: 1fr; }
}
`

// Render returns one HTML document. It has no external assets. The quote is the
// excerpt of the cited lines; it is not a field of the findings file.
func Render(p Page) []byte {
	p0, p1, p2, p3 := severityCounts(p.Findings)
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html>\n<head>\n<meta charset=\"utf-8\">\n<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n<title>Source review</title>\n<style>")
	b.WriteString(pageCSS)
	b.WriteString("</style>\n</head>\n<body>\n<header>\n")
	fmt.Fprintf(&b, "<p class=\"meta\" id=\"repo\">%s</p>\n", html.EscapeString(p.Repo))
	b.WriteString("<div class=\"identity\">\n")
	fmt.Fprintf(&b, "<p class=\"meta\" id=\"commit\">%s</p>\n", html.EscapeString(p.Commit))
	fmt.Fprintf(&b, "<p class=\"meta\" id=\"model\">%s</p>\n", html.EscapeString(p.ModelID))
	b.WriteString("</div>\n")
	b.WriteString("<ul class=\"counts\">\n")
	fmt.Fprintf(&b, "<li id=\"count-p0\">P0 %d</li>\n", p0)
	fmt.Fprintf(&b, "<li id=\"count-p1\">P1 %d</li>\n", p1)
	fmt.Fprintf(&b, "<li id=\"count-p2\">P2 %d</li>\n", p2)
	fmt.Fprintf(&b, "<li id=\"count-p3\">P3 %d</li>\n", p3)
	b.WriteString("</ul>\n</header>\n<div class=\"filters\">\n")
	b.WriteString("<label>Severity <select id=\"severity\">\n")
	b.WriteString("<option value=\"\">all</option>\n<option value=\"P0\">P0</option>\n<option value=\"P1\">P1</option>\n<option value=\"P2\">P2</option>\n<option value=\"P3\">P3</option>\n")
	b.WriteString("</select></label>\n")
	b.WriteString("<label>Directory <select id=\"directory\">\n<option value=\"\">all</option>\n")
	for _, d := range findingDirs(p.Findings) {
		fmt.Fprintf(&b, "<option value=\"%s\">%s</option>\n", html.EscapeString(d), html.EscapeString(d))
	}
	b.WriteString("</select></label>\n</div>\n<div id=\"findings\">\n")
	for _, f := range p.Findings {
		dir := path.Dir(f.File)
		quote := ""
		if p.Files != nil {
			quote = excerpt(p.Files[f.File], f.StartLine, f.EndLine)
		}
		fmt.Fprintf(&b, "<article class=\"finding\" data-severity=\"%s\" data-dir=\"%s\">\n", html.EscapeString(f.Severity), html.EscapeString(dir))
		fmt.Fprintf(&b, "<button type=\"button\">%s:%d-%d</button>\n", html.EscapeString(f.File), f.StartLine, f.EndLine)
		b.WriteString("<div class=\"detail\" hidden>\n<div class=\"prose\">\n")
		fmt.Fprintf(&b, "<p class=\"issue\">%s</p>\n", html.EscapeString(f.Issue))
		fmt.Fprintf(&b, "<p class=\"consequence\">%s</p>\n", html.EscapeString(f.Consequence))
		b.WriteString("</div>\n")
		fmt.Fprintf(&b, "<pre class=\"quote\">%s</pre>\n", html.EscapeString(quote))
		b.WriteString("</div>\n</article>\n")
	}
	b.WriteString("</div>\n<script id=\"page\">")
	b.WriteString(pageScript)
	b.WriteString("</script>\n</body>\n</html>\n")
	return []byte(b.String())
}
