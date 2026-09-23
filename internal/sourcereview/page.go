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
body { font: 16px/1.4 ui-sans-serif, system-ui, sans-serif; margin: 1.5rem; color: #1a1a1a; background: #f7f5f2; }
header { margin-bottom: 1rem; }
.meta { margin: 0.15rem 0; }
.counts { display: flex; gap: 1rem; padding: 0; list-style: none; }
.filters { display: flex; gap: 1rem; margin: 1rem 0; }
article.finding { background: #fff; border: 1px solid #ccc; margin: 0.5rem 0; padding: 0.5rem 0.75rem; }
article.finding > button { font: inherit; background: none; border: 0; padding: 0; cursor: pointer; text-align: left; }
.detail { display: none; gap: 1.5rem; margin-top: 0.75rem; }
.detail:not([hidden]) { display: flex; }
.detail pre { margin: 0; background: #f0ece6; padding: 0.5rem; overflow: auto; }
`

// Render returns one HTML document. It has no external assets. The quote is the
// excerpt of the cited lines; it is not a field of the findings file.
func Render(p Page) []byte {
	p0, p1, p2, p3 := severityCounts(p.Findings)
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html>\n<head>\n<meta charset=\"utf-8\">\n<title>Source review</title>\n<style>")
	b.WriteString(pageCSS)
	b.WriteString("</style>\n</head>\n<body>\n<header>\n")
	fmt.Fprintf(&b, "<p class=\"meta\" id=\"repo\">%s</p>\n", html.EscapeString(p.Repo))
	fmt.Fprintf(&b, "<p class=\"meta\" id=\"commit\">%s</p>\n", html.EscapeString(p.Commit))
	fmt.Fprintf(&b, "<p class=\"meta\" id=\"model\">%s</p>\n", html.EscapeString(p.ModelID))
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
		b.WriteString("<div class=\"detail\" hidden>\n")
		fmt.Fprintf(&b, "<p class=\"issue\">%s</p>\n", html.EscapeString(f.Issue))
		fmt.Fprintf(&b, "<p class=\"consequence\">%s</p>\n", html.EscapeString(f.Consequence))
		fmt.Fprintf(&b, "<pre class=\"quote\">%s</pre>\n", html.EscapeString(quote))
		b.WriteString("</div>\n</article>\n")
	}
	b.WriteString("</div>\n<script id=\"page\">")
	b.WriteString(pageScript)
	b.WriteString("</script>\n</body>\n</html>\n")
	return []byte(b.String())
}
