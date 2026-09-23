# Whole-repo source review progress

Session: 01a0ca43-3878-7771-aeec-ddc5f7ae9106

Evidence: /var/folders/88/6w9jsm3s1v757_nzs4ybffj40000gn/T/grok-goal-2c246734b978/implementer

Goal file `metareview-whole-repo-source-review-goal.md` is not edited. No commit, push, or pull request.

## Checkpoint 1 — Seams

Status: Pass

Started Unverified. Named the seams below after reading `docs/ARCHITECTURE.md` and the cited files. No product change in this checkpoint.

Selection (new code will compose these; `Classify` stays unchanged and still returns Code for an unrecognised path):

- `classify.Classify` — `internal/classify/classify.go:72`
- `configBase` (lockfile basenames, reached only through `Classify`) — `internal/classify/classify.go:65`
- `claimcheck.IsTestPath` — `internal/claimcheck/claimcheck.go:69`

The new selection function will be `sourcereview.Select` in `internal/sourcereview/select.go`. It is not in the tree at this checkpoint.

The one function that will call `claude`, `codex`, or `grok` will be `sourcereview.Call` in `internal/sourcereview/model.go`. It is not in the tree at this checkpoint. It will read model text the way these do:

- Claude `result` and `is_error` — `claudeResult` at `internal/fsm/judge/claude.go:156`, read in `parseClaudeResult` at `internal/fsm/judge/claude.go:201`
- Codex last `item.completed` / `agent_message` `item.text` — `parseCodexEvents` at `internal/fsm/judge/codex.go:160` (the assignment is `internal/fsm/judge/codex.go:171`)
- Grok `text` — no existing helper; `Call` will read that JSON string

Findings:

- `wireEntry` — `internal/lensoutput/lensoutput.go:320`
- `TypedFinding.Validate` — `internal/lensoutput/lensoutput.go:128`

`ValidatePayload` is not used. Diff review, `gitcontext.generatedExcludedFiles`, `internal/shardpack`, and `review-lenses` stay as they are.

## Checkpoint 2 — Source cut

Status: Pass

`sourcereview.Select` is `internal/sourcereview/select.go:27`. `keepPath` drops Config and Docs from `Classify`, `IsTestPath`, a `vendor` / `node_modules` / `third_party` segment, invalid UTF-8, and a generated line `^// Code generated .* DO NOT EDIT\.$` whose preceding lines are blank or comments. A line that only says `DO NOT EDIT` stays kept. `Classify` is unchanged.

`go test ./internal/classify` exited 0. Evidence: `cover.txt` in the evidence directory.

Fixture `fixture-kept` commit `1720b406e24e6f15e069f8aeb63a723582ce3968`. Kept `pkg/app.go`, `pkg/note.go`. Dropped the other seven. Fixture `fixture-empty` commit `17a0f35d9dd710c9e860d2d7a69b1644afc1db6d` keeps nothing, calls no model, and writes `{"findings":[]}` plus a zero-count page (`empty-kept-1/`, `empty-kept-2/`). The two captures match.

## Checkpoint 3 — Model choice

Status: Pass

`sourcereview.Call` is the one function, `internal/sourcereview/model.go:66`. It runs `claude -p --model opus --output-format json` (prompt on stdin, text in `result`), `codex exec --json -m gpt-6-astra -` (prompt on stdin, text in the last `agent_message` `item.text`), and `grok -p <prompt> -m grok-4.7 --output-format json` (text in `text`). Fake-exec tests record argv, stdin versus `-p`, the instruction, kept paths immediately before file text, no dropped file, and prompts at most 120000 bytes. Whole-file packing and UTF-8 slice partitions are tested. The parser accepts `{"findings":[]}` and a complete entry that passes `TypedFinding.Validate` (`internal/lensoutput/lensoutput.go:128`), and rejects null findings, a missing field, a CLI error, missing model text, a bad entry, a path that is not kept, and line numbers that are not in the file. Local `wireEntry` is `internal/sourcereview/parse.go:19`, same pointer fields as `internal/lensoutput/lensoutput.go:320`.

## Checkpoint 4 — Findings

Status: Pass

Success writes `<dir>/findings.json` as one `{"findings":[...]}` of accepted findings in prompt order, and only after every prompt succeeds. The record has no model field. An output directory inside the reviewed tree is refused (`refuse-inside.txt`); that directory stayed empty. A non-zero exit, Claude `is_error`, a process that cannot start, and a bad payload each leave the output directory empty.

## Checkpoint 5 — Page

Status: Pass

`sourcereview.Render` is `internal/sourcereview/page.go:102`. One self-contained `review.html`. Severity and directory filters are the shipped page script. A repo-root file's parent is `.`. Opening a finding shows issue and consequence beside the quoted lines. Headless Chrome (`browser.json`, `page.png`): zero page exceptions; the severity filter and the directory filter each hide the non-matching finding; the quote is `package main`.

## Checkpoint 6 — Proof

Status: Fail

Opus and Grok passed the structural bar. Astra did not. Codex accepted `-m gpt-6-astra`, then exited 1. Stdout on the retry:

```
{"type":"error","message":"You've hit your usage limit. Visit https://chatgpt.com/codex/settings/usage to purchase more credits or try again at Sep 23rd, 2026 1:41 PM."}
{"type":"turn.failed","error":{"message":"You've hit your usage limit. Visit https://chatgpt.com/codex/settings/usage to purchase more credits or try again at Sep 23rd, 2026 1:41 PM."}}
```

`live-astra/` is empty. That is a failed proof. No fake CLI was substituted.

Commands, from `/Users/ben/code/metareview`:

```
go run ./cmd/metareview source-review --model opus --output <evidence>/live-opus <evidence>/fixture-kept
go run ./cmd/metareview source-review --model astra --output <evidence>/live-astra <evidence>/fixture-kept
go run ./cmd/metareview source-review --model grok --output <evidence>/live-grok <evidence>/fixture-kept
```

Fixture commit `1720b406e24e6f15e069f8aeb63a723582ce3968`. Opus wrote `live-opus/findings.json` (`{"findings":[]}`) and `live-opus/review.html` (model `opus`, that commit, both filters, P0 0, P1 0, P2 0, P3 0). Grok wrote `live-grok/findings.json` and `live-grok/review.html` (model `grok-4.7`, same commit and zero counts). Logs: `live-opus.log`, `live-grok.log`, `live-astra.log`. Each log prints the kept pair and the seven dropped paths before the prompt. The prompt's only source text is `pkg/app.go` and `pkg/note.go`, each path immediately before its text, and the prompt asks for `{"findings":[...]}`.

`go test` for `internal/sourcereview`, `internal/classify`, and `cmd/metareview` exited 0 (`go-test.txt`). Coverage is 100.0% of statements for all three (`cover.txt`). `internal/sourcereview` was not added to `tests/coverage-exclude.txt`.

`git log -1` is `02b0527`. No commit, push, or pull request. The goal file was not edited.

User action for the failed proof: wait until the Codex usage limit resets (2026-09-23 13:41) and re-run the astra command above. The shell also logs `codex_core::shell_snapshot` line 20716 `syntax error near unexpected token '('`; the turn still started, and the exit reason on stdout is the usage limit.
