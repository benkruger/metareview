# Grok speed progress

Session: Claude Code ccb3697d-57f7-43c7-afbe-b80f430f5235

Evidence: /private/tmp/claude-501/-Users-ben-code-metareview/ccb3697d-57f7-43c7-afbe-b80f430f5235/scratchpad/speed
(`arm.sh` runs the real source-review; `results.tsv` has one row per run; `logs/` holds each run's stdout and stderr;
outputs are under `/tmp/grok-speed/`).

Goal file `metareview-grok-speed-goal.md` is not edited.

## Checkpoint 1 — Baseline

Status: Pass

Commit `0b05fa4`. S2: `--path app/controllers/admin --path app/models/reporting --path app/models/concerns`
(488,772 raw bytes across those directories before the source cut). It packs into 5 prompts in every run.

All 20 runs were started together (`baseline.sh`), so up to 10 Grok and 10 Opus calls ran at once. Rows are in
`results.tsv` with label `base`. An earlier attempt exited 2 on every run because `arm.sh` passed the run number
as an extra argument; that was a script bug, fixed, and those rows were discarded.

| Arm (5 runs) | Time | Findings | Core |
|---|---|---|---|
| Grok S1 | 539 s (381–615) | 6.2 (4–9) | 3.0 |
| Opus S1 | 63 s (52–76) | 5.8 (5–6) | 3.0 |
| Grok S2, `--jobs 1`, sum of 5 prompts | 5,644 s (5,183–5,934) | 113 (103–119) | — |
| Opus S2, `--jobs 1`, sum of 5 prompts | 265 s (234–289) | 33.2 (27–41) | — |

Core: every run of both models covers all three core bugs (read with `show.sh`): the required-field check
(lines ~87–134), the unguarded `result.data.errors` (~294–306), and Cancel not restoring fields (~202–208).

Method note for later checkpoints: run times depend on how many calls run at once (Opus S1 took ~30 s alone
earlier and ~63 s here). So each change is measured side by side with a fresh baseline arm in the same batch,
and the change is judged against that fresh baseline, which is stricter than comparing with this table.

## Checkpoint 2 — No forced answer format

Status: Pass (change dropped)

Change: remove `--json-schema` from `claudeArgs` and `grokArgs`. Measured on S1 in one batch (`cp2.sh`): the
change (label `c2`) side by side with a fresh baseline built from `0b05fa4` (label `c2base`), 5 runs each,
for Grok and Opus.

| Arm (5 runs, S1) | Time | Findings | Core |
|---|---|---|---|
| Grok, baseline | 717 s (546–914) | 7.8 (4–13) | 2.8 |
| Grok, no answer format | 530 s (458–589) | 5.8 (5–8) | 2.8 |
| Opus, baseline | 87 s (62–116) | 4.8 (4–5) | 3.0 |
| Opus, no answer format | 85 s (64–101) | 5.8 (5–7) | 3.0 |

Core, read finding by finding (`show.sh`): Grok baseline run 3 and Grok change run 1 have no finding on the
unguarded `result.data.errors`; every other run covers all three.

Decision: dropped. Grok time is 26% lower, but Grok findings are 2.0 lower, past the 0.5 limit. The working
tree is reverted; `git status` shows only the untracked goal and progress files.

## Checkpoint 3 — Bigger prompts

Status: Pass (change dropped)

Change: `MaxPromptBytes` 120,000 → 480,000, so S2 packs into 2 prompts instead of 5. Grok accepted the larger
prompts (no size error). Measured on S2 with `--jobs 1` in one batch (`cp3.sh`): the change (label `c3`) side by
side with a fresh baseline from `0b05fa4` (label `c3base`), 5 runs each, Grok and Opus.

| Arm (S2, `--jobs 1`) | Time (sum of prompts) | Findings | Findings per prompt |
|---|---|---|---|
| Grok, 120 KB (4 of 5 runs) | 5,745 s (5,429–6,011) | 105.8 (99–111) | 21.2 |
| Grok, 480 KB (4 of 5 runs) | 2,087 s (1,886–2,475) | 58.3 (52–64) | 29.1 |
| Opus, 120 KB | 320 s (290–389) | 36.0 (34–38) | 7.2 |
| Opus, 480 KB | 168 s (144–186) | 13.2 (12–15) | 6.6 |

One Grok run in each arm failed on the 30-minute call timeout and wrote nothing: `c3base` run 2 at prompt 4/5,
`c3` run 2 at prompt 1/2. The Grok means cover the 4 runs that finished.

Decision: dropped. Grok time is 64% lower, but Grok findings are 47.5 lower (Opus 22.8 lower), far past the
0.5 limit. Grok's time per call barely changes at four times the size (about 1,000 s against about 1,100 s),
but both models report far fewer findings for the same code. The working tree is reverted.

## Checkpoint 4 — grok-4.7-build-fast

Status: Pass (change dropped)

Change: Grok model id `grok-4.7` → `grok-4.7-build-fast` (`ModelID` and `grokArgs`). Measured on S1 in one batch
(`cp4.sh`): the change (label `c4`) side by side with a fresh baseline from `0b05fa4` (label `c4base`), 5 Grok runs each.

| Arm (Grok, S1) | Runs that wrote output | Time | Findings | Core |
|---|---|---|---|---|
| Baseline `grok-4.7` | 5 of 5 | 474 s (344–566) | 6.4 (5–8) | 3.0 |
| `grok-4.7-build-fast` | 3 of 5 | 268 s (90–358) | 6.3 (6–7) | 3.0 |

Runs 1 and 2 of the change exited 1 after 143 s and 69 s: the Grok CLI returned
`"structuredOutputError": "model did not produce structured output"` (`logs/c4-grok-S1-1.err`). The three runs that
finished cover all three core bugs.

Decision: dropped. On the runs that finished it is 43% faster with the same findings and core, but 2 of 5 runs
produced no review at all, and one failed prompt fails a whole `source-review` run. The working tree is reverted.

## Checkpoint 5 — Medium reasoning effort

Status: Pass (change kept)

Change: `grokArgs` adds `--reasoning-effort medium`. Measured on S1 in one batch (`cp5.sh`): the change (label `c5`)
side by side with a fresh baseline from `0b05fa4` (label `c5base`), 5 Grok runs each.

| Arm (Grok, S1) | Time | Findings | Core |
|---|---|---|---|
| Baseline (default effort) | 623 s (511–727) | 6.4 (4–9) | 2.8 |
| `--reasoning-effort medium` | 502 s (406–576) | 7.8 (5–10) | 3.0 |

Core, read finding by finding: baseline run 3 has no Cancel finding; every other baseline run and every medium
run covers all three core bugs.

Decision: kept. Time is 19% lower, findings are 1.4 higher, and core is not lower. `grokWant` in `model_test.go`
names the new flag; `go test` for `internal/sourcereview` and `cmd/metareview` passes at 100.0%, and `make cover`
exit 0 (`make-cover-c5.txt`). Commit: see below; this is the new baseline for later checkpoints.

## Checkpoint 6 — Sampling settings

Status: Unverified

## Checkpoint 7 — Duplicate a slow call

Status: Unverified

## Checkpoint 8 — Summary

Status: Unverified
