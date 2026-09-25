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
exit 0 (`make-cover-c5.txt`). Commit `0600426` (pushed); it is the baseline for checkpoints 6 and 7.

## Checkpoint 6 — Sampling settings

Status: Pass (change dropped)

Where I looked: `grok --help` has no sampling flag. `~/.grok/config.toml` sets none, and `~/.grok/models_cache.json`
has `"temperature": null, "top_p": null` for every model, so xAI's server decides. The docs
(`~/.grok/docs/user-guide/26-config-reference.md`, `05-configuration.md`) list `models.temperature` and `models.top_p`,
and no repetition penalty. The `GROK_CONFIG` inline JSON overlay may set the `models` table per process without
touching the user's config; `grok inspect` with it set reports `env_overlay: $GROK_CONFIG (inline)`. Grok does not
print the effective sampling values, so whether the server applied them is not verified.

Change: run Grok as `env GROK_CONFIG={"models":{"temperature":0.7,"top_p":0.95}} grok …` (the docs' own example
values). Measured on S1 in one batch (`cp6.sh`): the change (label `c6`) side by side with a fresh baseline built
from `0600426` (label `c6base`, medium effort), 5 Grok runs each.

| Arm (Grok, S1) | Time | Findings | Core |
|---|---|---|---|
| Baseline (`0600426`) | 494 s (470–509) | 5.4 (4–7) | 2.8 |
| temperature 0.7, top_p 0.95 | 457 s (339–546) | 6.6 (5–8) | 3.0 |

Core: baseline run 3 has no finding on the unguarded `result.data.errors`; every other run covers all three.

Decision: dropped. Time is 7.5% lower, short of the 10% bar (findings and core did not get worse). The working
tree is reverted.

## Checkpoint 7 — Duplicate a slow call

Status: Pass (change dropped)

Change: in `review`, a call that runs past twice the median of the calls already finished in the run (once 3 have
finished) gets one duplicate; the first to finish without error wins and the other is cancelled
(`callHedged`, `durations`). Unit tests covered it at 100% (a slow first attempt is cancelled when the duplicate
wins; an early failure gets no duplicate; a failed first attempt is rescued by the duplicate; two failures return
the first error), passing 5 times under `-race`. Measured on S2 with `--jobs 8` in one batch (`cp7.sh`): the
change (label `c7`) side by side with a fresh baseline built from `0600426` (label `c7base`), 5 Grok runs each,
judged by wall-clock time.

| Arm (Grok, S2, `--jobs 8`) | Wall time | Findings |
|---|---|---|
| Baseline (`0600426`) | 1,133 s (983–1,248) | 103.6 (85–130) |
| Duplicate slow calls | 1,163 s (1,084–1,307) | 95.4 (88–101) |

No run started a duplicate (`grep "started a duplicate" logs/c7-*` finds nothing): all 5 S2 prompts run at once
under `--jobs 8`, so no call is still running after 3 have finished and it has passed twice their median.

Decision: dropped. No time gain (2.6% slower, within noise). `run.go` and `run_test.go` are restored to `0600426`;
`git diff` touches only this progress file.

## Checkpoint 8 — Summary

Status: Pass

Each change was measured side by side with a fresh baseline in the same batch (5 runs each); "before" is that
batch's baseline.

| Change | Sample | Grok time before → after | Findings before → after | Core before → after | Decision |
|---|---|---|---|---|---|
| No forced answer format | S1 | 717 → 530 s | 7.8 → 5.8 | 2.8 → 2.8 | dropped (findings −2.0) |
| 480 KB prompts | S2, `--jobs 1` | 5,745 → 2,087 s | 105.8 → 58.3 | — | dropped (findings −47.5) |
| `grok-4.7-build-fast` | S1 | 474 → 268 s (3 of 5 runs) | 6.4 → 6.3 | 3.0 → 3.0 | dropped (2 of 5 runs failed) |
| `--reasoning-effort medium` | S1 | 623 → 502 s | 6.4 → 7.8 | 2.8 → 3.0 | **kept** (`0600426`) |
| temperature 0.7 / top_p 0.95 | S1 | 494 → 457 s | 5.4 → 6.6 | 2.8 → 3.0 | dropped (7.5% < 10%) |
| Duplicate a slow call | S2, `--jobs 8` wall | 1,133 → 1,163 s | 103.6 → 95.4 | — | dropped (never fired) |

Final run with the final code (`0600426`), from `/Users/ben/code/metareview` (`final.sh`):

```
metareview source-review --model grok --jobs 8 --output /tmp/grok-speed/final /Users/ben/code/hh
```

Exit 0 after 4,188 s (69.8 minutes) of wall-clock time. 45 prompts, each finished (per-prompt time 172–1,174 s,
median 740 s). 671 findings. It wrote `/tmp/grok-speed/final/findings.json` (361,003 bytes) and
`/tmp/grok-speed/final/review.html` (1,264,512 bytes). stderr has two skipped findings (a path that is not kept, and
line numbers past the end of `data/reports/contractor.sql`); those are dropped by the parser, not failures.

For comparison: the run before this goal (`--jobs 8`, default effort) was measured at about 58 minutes from its
first 8 prompts, but was stopped there and never completed, so it is not a like-for-like total.
