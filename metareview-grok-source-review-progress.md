# Grok source review progress

Session: Claude Code ccb3697d-57f7-43c7-afbe-b80f430f5235

Evidence: /private/tmp/claude-501/-Users-ben-code-metareview/ccb3697d-57f7-43c7-afbe-b80f430f5235/scratchpad/evidence

Goal file `metareview-grok-source-review-goal.md` is not edited. No commit, push, or pull request.

## Checkpoint 1 — Reproduce

Status: Pass

The 40 `hh` prompts were built with `Select` + `Pack` at `hh` commit
`7d257e0f8efcfcfdd5d31ef71346a3fa48cf192b` through a scratch test file overlaid with `go test -overlay`
(nothing written to the repo) into `hh-prompt-01.txt` … `hh-prompt-40.txt`. `hh-prompt-01.txt` is
byte-identical to the Grok session's `prompt_0.txt` (`cmp` exit 0). Measured on `hh-prompt-02.txt` (`cp1.sh`):

- Call shape 3: exit 0, 795 s, 1 turn, `stopReason` `end_turn`, 5 findings, 47,994 input tokens,
  55,233 output tokens (54,532 reasoning). Grok session `01a0cede-8b2d-7ef2-9f24-a4ada8e47544`.
  Files: `cp1-shape3.json`, `cp1-shape3.meta`.
- Call shape 1: exit 1, 13 s, `stopReason` `cancelled`, stderr `Error: max turns reached`, text "The
  omitted middle of the review request is the part I still need." Grok session
  `01a0cede-8b95-7033-8813-5a2f8b5206df`. Files: `cp1-shape1.json`, `cp1-shape1.err`.

No product change in this checkpoint.

## Checkpoint 2 — Grok call

Status: Pass

`Call` (`internal/sourcereview/model.go`) writes the prompt with `writePromptFile` (`os.CreateTemp("",
"metareview-grok-prompt-*.txt")`), runs `grok` with `grokArgs(path)`, and removes the file in a `defer`.
`grokSchema` and `grokSystem` are the schema and system prompt. `TestCallShapes` checks the exact argv
against `grokWant`, with nil stdin, and checks that the file holds the prompt. `TestGrokPromptFile` checks
during the call that the file holds the prompt bytes, sits in `os.TempDir()`, and is outside a tree dir and an
output dir. It checks the file is gone after success, a non-zero exit, and a runner that cannot start, and after
a real `OSRunner` with an empty `PATH`. `TestWritePromptFileErrors` covers the create and write failures.
`TestGrokSchema` parses the schema and checks the eight required fields, the integer types, and the enums. The
existing Grok parser/failure tests in `TestCallErrors` are unchanged apart from the added `ctx` argument.
`TestRunSpecFixtureThreeModels` now reads the Grok prompt from `--prompt-file`.

## Checkpoint 3 — Time limit and concurrency

Status: Pass

`Runner.Run` takes a `context.Context`. `OSRunner` uses `exec.CommandContext` with the `internal/fsm/cmdexec`
process-group pattern (`Setpgid`, `Cancel` = `kill(-pgid, SIGKILL)`, `WaitDelay`). `Run` → `review` runs up
to `Jobs` prompts (default 8) under a semaphore, each call has `context.WithTimeout(CallTimeout)` (default
30m), and the results come back in prompt order. CLI: `--jobs`, `--call-timeout`. Tests:

- `TestOSRunnerKillsProcessGroup`: a 300 ms deadline kills `sh` and its backgrounded `sleep 30` child
  (the child pid is gone).
- `TestRunCallTimeout`: an error with `prompt 1/1` and `timed out after 50ms`, and an empty output directory.
- `TestRunJobsKeepPromptOrder`: `--jobs 3`, five prompts that finish in reverse. Findings are in prompt
  order, max concurrent calls is exactly 3, prompt blocks print whole and in order, and each
  `prompt i/5 done in Ns (1 findings)` line appears once.
- `TestRunFailureStopsNewCalls`: with 1 job, a failure means 1 call and no more. With 2 jobs, prompt 1 fails
  while prompt 2 blocks: prompt 2 is cancelled, exactly 2 calls are made, prompt 1's error is returned, and
  the output directory is empty.
- `TestCLI`: `--jobs 0`, `--jobs two`, `--call-timeout 0`, `-5m`, `soon`, and a missing value each exit 2 with
  usage. The defaults are 8 and 30m.
- `TestRunBadJobsAndTimeout`: a negative `Jobs`/`CallTimeout` in `Options` fails before any call.

`go test -count=5 -race ./internal/sourcereview` exit 0.

## Checkpoint 4 — Fixture proof

Status: Pass

`live.sh live-grok-fixture-2 <fixture-kept>` ran the built binary with `source-review --model grok --output
<evidence>/live-grok-fixture-2 <first goal's fixture-kept>`. Fixture commit
`1720b406e24e6f15e069f8aeb63a723582ce3968`. Exit 0 in 32 s. stderr `prompt 1/1 done in 31s (1 findings)`.
Kept `pkg/app.go`, `pkg/note.go`, and the seven dropped paths printed before `prompt-begin`. `findings.json`
has 1 finding (P1). The page shows `grok-4.7`, the commit, `id="severity"`, `id="directory"`, P0 0 / P1 1 /
P2 0 / P3 0, and 1 finding block (`check.sh`). An earlier run (`live-grok-fixture`, 9 s, 0 findings, four zero
counts) was made before checkpoint 1 finished. It is kept only as a second sample.

## Checkpoint 5 — Real proof

Status: Fail

This is the exact goal command, run from `/Users/ben/code/metareview` by `real.sh`, which emptied `/tmp/hh-grok`
first (0 entries, `real.before`), with the default `--jobs` 8 and `--call-timeout` 30m. It started 08:38:54.
Result: exit 1 after 3,631 s. stderr: `prompt 19/40: call timed out after 30m0s: grok: signal: killed`.
`/tmp/hh-grok` is empty afterwards. No `grok --prompt-file` process was left behind (`pgrep`).

22 of 40 prompts finished before the timeout. Their times in seconds (sorted): 107 624 855 896 951 955 974
991 1083 1087 1095 1099 1123 1142 1170 1218 1242 1259 1279 1291 1434 1701. The median is about 1,100 s. The
same prompt shape took 451 s and 795 s alone (checkpoint 1). With 8 at once, each call is slower, and one
took longer than 30 minutes. At that rate 40 prompts take about 40 × 1,100 / 8 ≈ 92 minutes even without a
timeout. Neither the 1-hour bound nor the 30-minute call limit is met. The limit was not raised. Log:
`real.log`, `real.err`, `real.exit`.

Cause: Grok's per-prompt latency is reasoning. Checkpoint 1 was 54,532 reasoning tokens out of 55,233
output tokens. It is not the call shape (1 turn, no tools). Meeting the bound needs a change this contract
does not allow, such as `--reasoning-effort` (listed by `grok --help`, not measured) or a different `--jobs`
/ prompt budget. Those are Ben's decisions.

## Checkpoint 6 — Gate

Status: Unverified. It does not start while checkpoint 5 is Fail. For the record: `make cover` exit 0 with
`coverage gate passed` and `internal/sourcereview` at 100.0% (`make-cover.txt`). `go test` for
`internal/sourcereview`, `internal/classify`, and `cmd/metareview` exit 0 at 100.0% each.
