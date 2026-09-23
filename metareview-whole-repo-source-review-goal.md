# Whole-repo source review

<!-- CODING-GOAL-BEGIN -->

/goal Add a whole-repo source review to /Users/ben/code/metareview.

This file is the contract. Do not edit it during the walk. Scores, the
session id, and evidence paths go in
/Users/ben/code/metareview/metareview-whole-repo-source-review-progress.md.
Create that progress file at the start. A new session starts at the first
checkpoint. Do not resume an older progress file unless Ben points at it.

## Authority

Repository: /Users/ben/code/metareview. Origin is
https://github.com/benkruger/metareview. Upstream is
https://github.com/dsifry/metareview.

This goal is the approval to implement this scope. Do not commit, push, or
open a pull request.

Read docs/ARCHITECTURE.md before writing code. Leave the diff review and
its anchor gate alone. This run is a separate command.

## Outcome

A run reviews the first-party source of a repository. It keeps the files the
project builds or runs as the product. It drops tests, vendored and
third-party dependencies, lockfiles, generated code, documentation, and
configuration. It prints the kept list and the dropped list before any model
call.

`internal/classify` frames a path as code, config, or docs and still sends
all three to review (`internal/classify/classify.go`). Leave that behavior
in place. This cut drops a path when `Classify` returns Config or Docs.
That drops lockfiles, configuration, and documentation, including the
lockfile names in `configBase`.

It drops a path when `claimcheck.IsTestPath` is true
(`internal/claimcheck/claimcheck.go`). `internal/foo_test.go` is Code to
`Classify` and a test to `IsTestPath`.

It drops a path whose segments include `vendor`, `node_modules`, or
`third_party`.

It drops a file with a line matching
`^// Code generated .* DO NOT EDIT\.$` when every line before that
match is blank or a comment. A blank line is a comment line. A line whose first non-whitespace
characters are `//` is a comment line, and the rest of that line is the
comment. A `/*` starts a block and `*/` ends it. A line with no code
outside that block is a comment line. A line with code before `/*` or
code after `*/` is not a comment line. The block still runs from `/*`
to `*/` on the following lines.
A line that only says `DO NOT EDIT` does not match and the file is
kept. It drops a file that is not valid UTF-8. `Classify` calls an
unrecognised path Code.

`gitcontext.generatedExcludedFiles` stays the path exclude for
metareview's own artifacts. The whole-repo cut is a new selection in
front of this run.

The user picks Astra, Opus, or Grok. One function runs all three. It
reads the kept files, sends one or more prompts, and parses
`lensoutput.TypedFinding`. The prompt text and the parser stay the
same for every choice. The choice changes only the CLI and the model
id. Each CLI is the subscription already logged in. The run never sees
a token. Do not add an HTTP provider or an API key.

Each prompt contains the instruction, then each kept path on its own
line immediately before that file's text. A dropped file never appears.
The instruction tells the model to return only `{"findings":[...]}`,
with `tag` of `bug` or `advisory`, `severity` of `P0`, `P1`, `P2`, or
`P3`, and every field present: `tag`, `file`, `start_line`, `end_line`,
`issue`, `consequence`, `confidence`, `severity`. `confidence` is an
integer from 0 to 100. The whole prompt, which is the instruction, the
path lines, and the file text, is at most 120,000 bytes. `AGENTS.md` uses 120,000 bytes as the size that
cannot be held in one review context. Whole kept files are packed into
the smallest number of prompts that stay under that size. A kept file
that cannot fit in one prompt is cut into
consecutive byte slices that partition the file in order. Each slice
is non-empty and starts and ends on a UTF-8 character boundary. Each
slice prompt names the original path and the line
that contains the slice's first byte, tells the model that citations
use original line numbers, and the whole slice prompt is at most
120,000 bytes. A finding cites the original path and the original line
numbers. The findings file is one `{"findings":[...]}` whose array is
every accepted finding from every prompt, in order. It is not the raw
answers glued together. The page is still one page. The counts are the
number of findings for each of `P0`, `P1`, `P2`, and `P3`.

- Opus is `claude` with model id `opus`.
- Astra is `codex` with model id `gpt-6-astra`. The live Astra command
  passes it to `-m`. If the CLI rejects it, stop and record the CLI's
  error.
- Grok is `grok` with model id `grok-4.7`. That is the default
  `grok models` prints.

`claude` takes `-p`, `--model`, and `--output-format json`, and reads
the prompt on stdin. Its stdout is one JSON document. The model text
is the `result` string (`claudeResult` in
`internal/fsm/judge/claude.go`). `codex` takes `exec`, `--json`, and
`-m`, reads the prompt on stdin, and the argv includes `-`. Its stdout
is a JSON event stream. The model text is `item.text` of the last
`item.completed` event whose `item.type` is `agent_message`
(`parseCodexEvents` in `internal/fsm/judge/codex.go`). `grok` takes
`-p`, `-m`, and `--output-format json`, and `-p` takes the prompt as
its argument. Its stdout is one JSON document. The model text is the
`text` string in its JSON output. Use those flags.
Do not invent flags. The parser reads that model text as `{"findings":[...]}`. Surrounding
whitespace is ignored. A surrounding fence is a first line of three
backticks, optionally followed by one word of letters, and a last line
of three backticks. When the body of that fence is the object, the
parser uses the body. Any other extra text fails the run. Each
element is unmarshaled as `wireEntry`
(`internal/lensoutput/lensoutput.go`). An element that does not
unmarshal fails the run. A missing field is a nil pointer and fails
the run. A present element must then pass
`TypedFinding.Validate`. The path must be a kept file, and the line
numbers must exist in that file. This run does not call
`ValidatePayload`. An empty answer is `{"findings":[]}`.
`{"findings":null}`, a CLI error, missing model text, a missing field,
or text that is not that object fails the run. A CLI error is a
process that cannot start, a non-zero exit status, or `is_error` true
on the Claude document. It is not stored as an empty list. One failed prompt fails the run.
That run writes neither the findings file nor the page. When no file
is kept, the run still prints the kept list and the dropped list,
calls no model, and writes `{"findings":[]}` and the page. The page
shows that run's commit, the chosen model id, both filters, and four
zero counts.

`internal/shardpack` and `review-lenses` stay on diff reviews. A finding
in this run is a `lensoutput.TypedFinding`: tag, file, start_line,
end_line, issue, consequence, confidence, severity. The cited range
must fall inside a kept file. The page quotes those lines from the
file. That quote is the excerpt. It is not a schema field. The page
shows which model ran. The finding record does not carry a model field.

The output directory is an argument. The proof points it outside the
fixture. The run refuses an output directory inside the tree under review.

- A findings file.
- One self-contained HTML file. It opens in a browser with no server and no
  sibling assets. It shows the repo, commit, model, and counts. Findings
  filter by severity and by the directory component of `file`. A file
  in the repo root has parent `.`. Opening a
  finding shows the
  issue and consequence beside the quoted lines. The page does not modify
  the repo, post anywhere, or start another review.

A fake exec proves the call shape for all three. For each one it
records the binary, the model id, where the prompt was passed, and that
every prompt contains the instruction and names each kept path before
that file's text, and contains no dropped file. The whole prompt is at
most 120,000 bytes. One test packs kept files that do not fit in one
prompt and shows more than one prompt, each file whole, each prompt at
most 120,000 bytes. Another test feeds one kept file that does not fit
in one prompt and shows consecutive non-empty byte slices that
partition the file on UTF-8 character boundaries, with findings citing
the original path and original line numbers. A parser test accepts `{"findings":[]}`, accepts an entry with all
eight fields present that passes `Validate`, and rejects
`{"findings":null}`, a missing `confidence`, a CLI error, missing model
text, and an entry that fails `Validate`. A non-zero exit, Claude `is_error` true, and a process that cannot
start each write neither file. The output directory is empty before
the failure and still empty after it.
A renderer test loads two findings with different severities and
directories, the directory being the directory component of `file`,
and shows the
severity filter and the directory filter each
hide the finding that does not match. The same page shows the `P0`,
`P1`, `P2`, and `P3` counts, and those counts match the findings. No
API key is read.

The goal is not done on a fake exec. Three live commands run Astra,
Opus, and Grok against the fixture. Each command writes to its own
output directory. Each prompt asks for
`{"findings":[...]}`. The only source text in each prompt is
`pkg/app.go` and `pkg/note.go`, and each path is printed before that
file's text. Each command prints the kept list and the dropped list
before the model call. Each command writes a findings file and an HTML
page from that model's answer. Each page shows that model id, the
fixture commit, both filters, and the `P0`, `P1`, `P2`, and `P3`
counts, and those counts match the findings. An empty `findings` list
shows no findings and four zero counts. A mock run does not satisfy
this.

## Checkpoints

Record Pass or Fail in the progress file only. A later checkpoint starts
only after every earlier one is Pass. The first checkpoint is Unverified.

1. Seams. Name the files and symbols for the new file selection, the one
   function that calls `claude`, `codex`, or `grok`, `wireEntry`, and
   `TypedFinding.Validate`. Cite file:line in the progress file. No
   product change in this checkpoint.
2. Source cut. A fixture is a git repository with one commit containing
   `pkg/app.go`, `pkg/note.go`, `pkg/app_test.go`, `vendor/lib.go`,
   `go.sum`, `pkg/gen.go`, `pkg/blob.dat`, `README.md`, and
   `config.yaml`. `pkg/blob.dat` is not valid UTF-8. `pkg/gen.go`
   has a first line matching `^// Code generated .* DO NOT EDIT\.$`.
   `pkg/note.go`'s first line is `DO NOT EDIT`.
   The run prints `pkg/app.go` and `pkg/note.go` as kept, and prints
   `pkg/app_test.go`, `vendor/lib.go`, `go.sum`, `pkg/gen.go`,
   `pkg/blob.dat`, `README.md`, and `config.yaml` as dropped. A second
   fixture is a git repository with one commit of only those dropped
   files. It prints an empty kept list and that dropped list, calls no
   model, and writes `{"findings":[]}` and a page. The page shows that
   commit, the chosen model id, both filters, and four zero counts.
   Existing
   classify tests still pass unchanged.
3. Model choice. Astra, Opus, and Grok go through the one function.
   The fake exec shows `claude` with `opus`, `codex` with `--json` and
   `-m gpt-6-astra`, and `grok` with `grok-4.7`. The model text is read
   from `result`, the last `item.text`, and `text` respectively, then
   read as `{"findings":[...]}`. Each element is a `wireEntry` with all
   eight fields present and then passes `TypedFinding.Validate`. The
   path is a kept file and the line numbers exist in that file.
   `{"findings":[]}` is empty. `{"findings":null}`, a missing field, a
   CLI error, missing model text, or an entry that fails `Validate`
   fails the run. A non-zero exit, Claude `is_error` true, and a
   process that cannot start each write neither file. The output
   directory is empty before the failure and still empty after it.
   Every prompt contains the instruction and asks for that object,
   names each kept path before that file's text, and is at most
   120,000 bytes. Kept files that do not fit in one prompt become more
   than one prompt, each file whole, each prompt at most 120,000
   bytes. One kept file that does not fit
   in one prompt becomes consecutive non-empty byte slices that
   partition the file on UTF-8 character boundaries, and a finding
   cites the original path and
   original line numbers. No dropped file appears
   in any prompt. No API key is read.
4. Findings. The findings file is one `{"findings":[...]}` written
   outside the fixture. Its array is every accepted finding from every
   prompt, in order. A record is a complete `wireEntry` that passes
   `Validate`, its path is a kept file, and its line numbers exist in
   that file. The page shows the model id. An empty answer is
   `{"findings":[]}` from the model. A failed call writes neither the
   findings file nor the page.
5. Page. The HTML file renders from that findings file, is self-contained,
   and filters by severity and by the directory component of `file`.
   A file in the repo root has parent `.`. It
   shows the fixture commit, the model id, and the `P0`, `P1`, `P2`,
   and `P3` counts. Opening a finding shows the issue and consequence
   beside the quoted lines. It is outside the fixture. There is no
   existing findings-page seam. The page is new.
6. Proof. The fake-exec tests cover Astra, Opus, and Grok. Input that
   does not fit in 120,000 bytes is split into whole-file prompts and
   into non-empty byte slices that partition one large file on UTF-8
   character boundaries, and every prompt is at most 120,000 bytes. Three live commands against the fixture, one for each model,
   each writing to its own output directory,
   each ask for `{"findings":[...]}`. The only source text in each
   prompt is `pkg/app.go` and `pkg/note.go`, with each path printed
   before that file's text. Each writes a findings file and an HTML
   page. Each command prints the kept list and the dropped list from
   checkpoint 2 before the model call. Each page shows that model id,
   the fixture commit, both filters, and the `P0`, `P1`, `P2`, and `P3`
   counts, and those counts match the findings. An empty `findings`
   list shows no findings and four zero counts.
   `go test` for
   every package this goal touched exits 0. A package this goal adds is
   at 100% statement coverage, which is what `make cover` enforces. The
   progress file names the three commands, the commit, and the output
   paths.

<!-- CODING-GOAL-END -->
