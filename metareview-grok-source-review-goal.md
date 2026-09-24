# Grok source review

<!-- CODING-GOAL-BEGIN -->

/goal Make `metareview source-review --model grok` finish on a real
repository and write its findings file and page.

This file is the contract. Do not edit it during the walk. Scores, the
session id, and evidence paths go in
/Users/ben/code/metareview/metareview-grok-source-review-progress.md.
Create that progress file at the start. A new session starts at the first
checkpoint.

## Authority

Repository: /Users/ben/code/metareview, branch
`whole-repo-source-review`. This goal narrows
`metareview-whole-repo-source-review-goal.md`. Every rule in that file
still holds except where this file changes the Grok call or adds to
`Run`. Do not edit that file.

This goal is the approval to implement this scope. Do not commit, push,
or open a pull request.

## What went wrong

The command was:

```
go run ./cmd/metareview source-review --model grok --output /tmp/hh-grok /Users/ben/code/hh
```

It was killed after 10 hours. `/tmp/hh-grok` is empty. It did not hang.

- `hh` at commit `7d257e0f8efcfcfdd5d31ef71346a3fa48cf192b` keeps 1,346
  files, 4,616,754 bytes. `Pack` turns that into 40 prompts.
- `grok -p <prompt>` runs Grok's full agent, not a single completion.
  Grok cuts the ~120 KB argument down to an excerpt and saves the full
  text to `~/.grok/sessions/<cwd>/<id>/prompts/prompt_0.txt`. The model
  then reads that file back with `read_file` and `grep`. Session
  `01a0cb4a-3519-7f11-8ffd-628326f49543` made 31 tool calls and 11 model
  calls and read 844,756 input tokens for one prompt.
- In 10 hours Grok finished 35 of the 40 prompts, at 10 to 25 minutes
  each. The kill came during prompt 36. The Grok session folders under
  `~/.grok/sessions/%2FUsers%2Fben%2Fcode%2Fmetareview/` are the record,
  from `01a0cb4a…` at 15:43 to `01a0cd6b…` at 01:39.
- `Run` sends the prompts one at a time. It writes nothing until every
  prompt succeeds. No call has a time limit.

## Measured call shapes

These were run by hand against prompt 1 of `hh`, the 119,998-byte
`prompt_0.txt` above. Each was run once. The goal must re-measure.

1. `grok -p "<prompt>" -m grok-4.7 --output-format json --tools ""
   --max-turns 1 --no-subagents --disable-web-search`: exit 1 after
   14 s with `max turns reached`. The model text was "The omitted source
   is required before I can finish the review." The argument was still
   cut to an excerpt.
2. `grok --prompt-file <f> --verbatim` plus the flags in 1: exit 1 after
   170 s with `max turns reached`. The model still tried to call a tool.
3. The flags in 2 plus `--json-schema <findings schema>` and
   `--system-prompt-override "You are a code reviewer. You have no tools.
   Everything you need is in the user message. Answer in a single
   message."`: exit 0 after 451 s. One turn, `stopReason` `end_turn`,
   49,803 input tokens, 30,712 output tokens (29,374 of them reasoning).
   `text` held `{"findings":[...]}` with 11 findings that cite real
   lines in `data/dumps/schema.sql`.

`-m grok-4.7` shows up as `grok-4.7-build` in `modelUsage`. That is the
CLI's own mapping. Leave the model id as `grok-4.7`.

## Outcome

1. Grok call shape. `Call` runs Grok as one turn with no tools:

   ```
   grok --prompt-file <file> --verbatim -m grok-4.7 --output-format json
        --json-schema <schema> --tools "" --max-turns 1 --no-subagents
        --disable-web-search --system-prompt-override <text>
   ```

   `<file>` holds exactly the prompt bytes. It is created in
   `os.TempDir()`, never inside the reviewed tree or the output
   directory, and removed after the call whether the call worked or not.
   `<schema>` is a JSON Schema for `{"findings":[...]}`. It requires all
   eight fields, limits `tag` to `bug`/`advisory` and `severity` to
   `P0`–`P3`, and makes `start_line`, `end_line`, and `confidence`
   integers. The prompt text, the parser, and `Validate` are the same as
   before. The model text is still `text`. Opus and Astra call shapes do
   not change. Use only flags that `grok --help` (1.0.40) lists. If a
   flag in this list turns out not to be needed, keep it anyway. It
   stops Grok from acting as an agent. If one is rejected, stop and
   record the CLI's error.

2. Per-call time limit. Each CLI call has a time limit, set with
   `--call-timeout <duration>` and defaulting to 30m. When the limit is
   reached, the process and its children are killed, the run fails with
   an error that names the prompt number, and neither output file is
   written. This applies to all three models.

3. Concurrency. `--jobs <n>` runs up to n prompts at once. It defaults
   to 8 and must be at least 1. The findings array stays in prompt order
   whatever order the calls finish in. When one prompt fails, no new
   prompts start, calls already running are cancelled, and nothing is
   written. The `prompt-begin`/`prompt-end` blocks on stdout are still
   printed in prompt order, and each block is printed whole. This
   applies to all three models.

4. Progress. After each prompt finishes, stderr gets one line:
   `prompt <i>/<N> done in <seconds>s (<k> findings)`. A long run is
   never silent.

## Checkpoints

Record Pass or Fail in the progress file only. A later checkpoint starts
only after every earlier one is Pass. The first checkpoint is Unverified.

1. Reproduce. Re-run call shape 3 above on one `hh` prompt, built with
   `Pack` from `Select` at the `hh` commit and not taken from the Grok
   session folder. Record the exit code, time, turns, and finding count.
   Then run call shape 1 or 2 once and record that it fails. No product
   change in this checkpoint.
2. Grok call. The fake exec shows the exact argv in Outcome 1. The
   `--prompt-file` file holds the prompt bytes during the call, is
   outside the tree and the output directory, and is gone after a
   successful call, a non-zero exit, and a process that cannot start.
   `--json-schema` parses as JSON and requires the eight fields. Every
   existing parser and failure test for Grok still passes, updated only
   for the new argv.
3. Time limit and concurrency. Fake-exec tests show: a call that runs
   past `--call-timeout` fails the run, names the prompt, and leaves the
   output directory empty; with `--jobs 3` and replies that finish in
   reverse order, the findings array is in prompt order; at most `n`
   calls run at the same time (checked with a counter in the fake);
   after a failure no new call starts and the output directory is
   empty; `--jobs 0` and a bad `--call-timeout` exit 2 with usage. The
   progress line appears once per prompt.
4. Fixture proof. The live Grok command from the first goal, run against
   that goal's `fixture-kept`, writes `findings.json` and `review.html`
   with model `grok-4.7`, the fixture commit, and counts that match the
   findings. Record the time.
5. Real proof. From `/Users/ben/code/metareview`:

   ```
   go run ./cmd/metareview source-review --model grok --output /tmp/hh-grok /Users/ben/code/hh
   ```

   Empty `/tmp/hh-grok` first. It exits 0 within 1 hour of wall-clock
   time, runs all 40 prompts, and writes `/tmp/hh-grok/findings.json` and
   `/tmp/hh-grok/review.html`. The page shows model `grok-4.7`, commit
   `7d257e0f8efcfcfdd5d31ef71346a3fa48cf192b`, both filters, and P0–P3
   counts that match the findings. Record the wall-clock time, the
   `--jobs` value, the time of each prompt from the progress lines, and
   the finding count. If Grok rate-limits at the default `--jobs`,
   record the CLI's error and the `--jobs` value that worked. A
   `--call-timeout` that fires is recorded along with its prompt number.
   It counts as a Fail, not as a reason to raise the limit silently.
6. Gate. `go test` for every package this goal touched exits 0.
   `internal/sourcereview` stays at 100% statement coverage (`make
   cover`). The progress file names the commands, the commits, the
   output paths, and the Grok session ids of the live runs.

<!-- CODING-GOAL-END -->
