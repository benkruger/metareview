package sourcereview

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dsifry/metareview/internal/lensoutput"
)

func TestRenderFiltersAndQuote(t *testing.T) {
	files := map[string][]byte{
		"main.go":    []byte("package main\n"),
		"pkg/app.go": []byte("alpha\nbeta\n"),
		"a.go":       []byte("a\n"),
		"b/c.go":     []byte("c\n"),
	}
	findings := []lensoutput.TypedFinding{
		{Tag: lensoutput.TagBug, File: "main.go", StartLine: 1, EndLine: 1, Issue: "root broke", Consequence: "no entry", Confidence: 90, Severity: "P0"},
		{Tag: lensoutput.TagAdvisory, File: "pkg/app.go", StartLine: 1, EndLine: 2, Issue: "alpha broke", Consequence: "wrong answer", Confidence: 70, Severity: "P1"},
		{Tag: lensoutput.TagBug, File: "a.go", StartLine: 1, EndLine: 1, Issue: "a", Consequence: "a", Confidence: 60, Severity: "P2"},
		{Tag: lensoutput.TagBug, File: "b/c.go", StartLine: 1, EndLine: 1, Issue: "c", Consequence: "c", Confidence: 50, Severity: "P3"},
	}
	page := Render(Page{Repo: "/repo", Commit: "abc123", ModelID: "opus", Findings: findings, Files: files})
	html := string(page)
	for _, want := range []string{
		`id="repo">/repo`, `id="commit">abc123`, `id="model">opus`,
		`id="count-p0">P0 1`, `id="count-p1">P1 1`, `id="count-p2">P2 1`, `id="count-p3">P3 1`,
		`id="severity"`, `id="directory"`, `value="."`, `value="pkg"`,
		"root broke", "no entry", "package main", "alpha broke", "wrong answer", "alpha\nbeta",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("page missing %q", want)
		}
	}
	if strings.Contains(html, "http://") || strings.Contains(html, "https://") || strings.Contains(html, "fetch(") || strings.Contains(html, "<form") || strings.Contains(html, "src=") {
		t.Fatal("page is not self-contained")
	}
	if !strings.Contains(html, `<script id="page">`) {
		t.Fatal("shipped script missing")
	}
	report := runPage(t, page)
	if report.Counts["p0"] != "P0 1" || report.Counts["p3"] != "P3 1" {
		t.Fatalf("counts %+v", report.Counts)
	}
	// The two findings the filter drives are the first two, which differ in
	// severity and directory. The harness records what the shipped script did.
	if len(report.Steps) < 2 {
		t.Fatalf("steps %+v", report)
	}
	sev := report.Steps[0]
	if sev.Displays[0] != "" || sev.Displays[1] != "none" {
		t.Fatalf("severity filter displays %v", sev.Displays)
	}
	dir := report.Steps[1]
	if dir.Displays[0] != "" || dir.Displays[1] != "none" {
		t.Fatalf("directory filter displays %v", dir.Displays)
	}
	if report.Opened.Hidden || report.Opened.Issue != "root broke" || report.Opened.Consequence != "no entry" || report.Opened.Quote != "package main" {
		t.Fatalf("opened %+v", report.Opened)
	}
}

func TestRenderEmpty(t *testing.T) {
	page := Render(Page{Repo: "r", Commit: "c", ModelID: "gpt-6-astra", Findings: nil, Files: nil})
	html := string(page)
	for _, id := range []string{"count-p0", "count-p1", "count-p2", "count-p3"} {
		if !strings.Contains(html, `id="`+id+`">`) || !strings.Contains(html, "0") {
			t.Fatalf("missing zero %s", id)
		}
	}
	if strings.Contains(html, `class="finding"`) {
		t.Fatal("empty page has a finding")
	}
	if !strings.Contains(html, `id="severity"`) || !strings.Contains(html, `id="directory"`) {
		t.Fatal("filters missing")
	}
	report := runPage(t, page)
	if report.FindingCount != 0 {
		t.Fatalf("findings %d", report.FindingCount)
	}
	for _, k := range []string{"p0", "p1", "p2", "p3"} {
		if !strings.HasSuffix(report.Counts[k], " 0") {
			t.Fatalf("%s = %s", k, report.Counts[k])
		}
	}
}

type pageReport struct {
	FindingCount int               `json:"findingCount"`
	Counts       map[string]string `json:"counts"`
	Steps        []struct {
		Displays []string `json:"displays"`
	} `json:"steps"`
	Opened struct {
		Hidden      bool   `json:"hidden"`
		Issue       string `json:"issue"`
		Consequence string `json:"consequence"`
		Quote       string `json:"quote"`
	} `json:"opened"`
}

func runPage(t *testing.T, page []byte) pageReport {
	t.Helper()
	dir := t.TempDir()
	htmlPath := filepath.Join(dir, "page.html")
	scriptPath := filepath.Join(dir, "dom.js")
	if err := os.WriteFile(htmlPath, page, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptPath, []byte(pageHarness), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("node", scriptPath, htmlPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("node: %v\n%s", err, out)
	}
	var report pageReport
	if err := json.Unmarshal(out, &report); err != nil {
		t.Fatalf("report: %v\n%s", err, out)
	}
	return report
}

const pageHarness = `
const fs = require("fs");
const html = fs.readFileSync(process.argv[2], "utf8");
const m = html.match(/<script id="page">([\s\S]*)<\/script>/);
if (!m) { console.error("no shipped script"); process.exit(1); }
const script = m[1];

function decode(s) {
  return s.replace(/&lt;/g, "<").replace(/&gt;/g, ">").replace(/&quot;/g, '"').replace(/&#39;/g, "'").replace(/&amp;/g, "&");
}
function parseAttrs(s) {
  const attrs = {};
  const re = /([A-Za-z0-9:-]+)(?:\s*=\s*"([^"]*)")?/g;
  let mm;
  while ((mm = re.exec(s))) attrs[mm[1]] = decode(mm[2] ?? "");
  return attrs;
}
class El {
  constructor(name, attrs) {
    this.name = name;
    this.attrs = attrs || {};
    this.children = [];
    this.parentNode = null;
    this.style = { display: "" };
    this.hidden = Object.prototype.hasOwnProperty.call(this.attrs, "hidden");
    this.value = this.attrs.value || "";
    this._listeners = {};
  }
  getAttribute(k) { return Object.prototype.hasOwnProperty.call(this.attrs, k) ? this.attrs[k] : null; }
  addEventListener(type, fn) { (this._listeners[type] ||= []).push(fn); }
  dispatchEvent(type) { for (const fn of this._listeners[type] || []) fn.call(this); }
  get textContent() { return this.children.map(c => c.textContent || "").join(""); }
  querySelector(sel) { const a = queryAll(this, sel); return a[0] || null; }
  querySelectorAll(sel) { return queryAll(this, sel); }
}
function classHas(n, c) { return (" " + (n.attrs.class || "") + " ").includes(" " + c + " "); }
function matchesSimple(n, sel) {
  if (!n || !n.name) return false;
  if (sel.startsWith("#")) return n.attrs.id === sel.slice(1);
  if (sel.startsWith(".")) return classHas(n, sel.slice(1));
  return n.name === sel;
}
function descendants(n) {
  const out = [];
  for (const c of n.children || []) {
    if (!c.name) continue;
    out.push(c);
    out.push(...descendants(c));
  }
  return out;
}
function queryAll(root, sel) {
  sel = sel.trim();
  const gt = sel.indexOf(">");
  if (gt >= 0) {
    const left = sel.slice(0, gt).trim();
    const right = sel.slice(gt + 1).trim();
    const out = [];
    for (const p of queryAll(root, left)) {
      for (const c of p.children) if (c.name && matchesSimple(c, right)) out.push(c);
    }
    return out;
  }
  const parts = sel.split(/\s+/).filter(Boolean);
  let cur = [root];
  for (const part of parts) {
    const next = [];
    for (const n of cur) for (const d of descendants(n)) if (matchesSimple(d, part)) next.push(d);
    cur = next;
  }
  return cur;
}
const voidTags = new Set(["meta", "link", "br", "img", "input"]);
function parseHTML(src) {
  const root = new El("#document", {});
  const stack = [root];
  const re = /<!DOCTYPE[^>]*>|<!--[\s\S]*?-->|<\/([A-Za-z0-9]+)>|<([A-Za-z0-9]+)([^>]*)>|([^<]+)/g;
  let mm;
  while ((mm = re.exec(src))) {
    if (mm[1]) {
      const name = mm[1].toLowerCase();
      while (stack.length > 1 && stack[stack.length - 1].name !== name) stack.pop();
      if (stack.length > 1) stack.pop();
    } else if (mm[2]) {
      const name = mm[2].toLowerCase();
      const el = new El(name, parseAttrs(mm[3] || ""));
      const parent = stack[stack.length - 1];
      parent.children.push(el);
      el.parentNode = parent;
      const selfClose = /\/\s*$/.test(mm[3] || "") || voidTags.has(name);
      if (!selfClose) stack.push(el);
    } else if (mm[4]) {
      stack[stack.length - 1].children.push({ textContent: decode(mm[4]), name: "" });
    }
  }
  return root;
}
const root = parseHTML(html);
const document = {
  getElementById(id) {
    const hit = descendants(root).filter(n => n.attrs && n.attrs.id === id);
    return hit[0] || null;
  },
  querySelectorAll(sel) { return queryAll(root, sel); },
};
try { new Function("document", script)(document); }
catch (e) { console.error(String(e)); process.exit(1); }
const findings = document.querySelectorAll("#findings .finding");
const counts = {};
for (const id of ["p0", "p1", "p2", "p3"]) counts[id] = document.getElementById("count-" + id).textContent.trim();
const report = { findingCount: findings.length, counts, steps: [], opened: {} };
if (findings.length >= 2) {
  const sev = document.getElementById("severity");
  sev.value = findings[0].getAttribute("data-severity");
  sev.dispatchEvent("change");
  report.steps.push({ displays: findings.map(f => f.style.display) });
  sev.value = "";
  const dir = document.getElementById("directory");
  dir.value = findings[0].getAttribute("data-dir");
  dir.dispatchEvent("change");
  report.steps.push({ displays: findings.map(f => f.style.display) });
  const btn = findings[0].querySelector("button");
  btn.dispatchEvent("click");
  const detail = findings[0].querySelector(".detail");
  report.opened = {
    hidden: detail.hidden,
    issue: findings[0].querySelector(".issue").textContent.trim(),
    consequence: findings[0].querySelector(".consequence").textContent.trim(),
    quote: findings[0].querySelector(".quote").textContent.trim(),
  };
}
process.stdout.write(JSON.stringify(report));
`
