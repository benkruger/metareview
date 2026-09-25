# Grok speed

<!-- CODING-GOAL-BEGIN -->

/goal Make `metareview source-review --model grok` faster without losing
review quality, by testing six changes one at a time on the real code and
keeping only the ones that win.

This file is the contract. Do not edit it during the walk. Results go in
/Users/ben/code/metareview/metareview-grok-speed-progress.md. Create it at
the start.

## Authority

Repository /Users/ben/code/metareview, branch `whole-repo-source-review`,
starting at commit `0b05fa4`. This goal approves the code changes below and
one `flow-pro commit` per kept change. It does not approve anything else.

## Rules

- Every measurement runs the real command on the real `hh` checkout:
  `metareview source-review --model <m> --path <path>... --output <dir> /Users/ben/code/hh`,
  built from the working tree. No fixture repos, no scratch harnesses, no
  direct `grok` or `claude` calls outside `source-review`.
- Output directories go under `/tmp/grok-speed/`.
- Astra is out of scope. Do not run it. Keep `codexArgs` compiling and tested.
- No API keys. Everything goes through the logged-in CLIs.
- `--reasoning-effort low` is never kept. It loses findings (4.4 against 6.8).
- Every model keeps the same setup (system prompt, instruction, answer
  format, tools off). A change to shared setup is measured on Opus too.
- A change is kept only if `go test` passes with `internal/sourcereview` at
  100% coverage and `make cover` passes.

## The sample and the score

Sample S1: `--path web/src/routes/DirectDeposit.svelte` (one prompt).

Sample S2: a `--path` set of real `hh` source that packs into at least 4
prompts at the 120,000-byte limit. Choose it in checkpoint 1 and record it.
Use it unchanged for the rest of the walk.

The core bugs are three bugs Opus finds on S1 in every run:
1. `result.data.errors` read without checking `result.data` (lines ~294–306).
2. bank name, routing and account number only required when `checkPending`
   is true (lines ~87–134).
3. Cancel on a saved deposit leaves edited fields changed (lines ~202–208).

A finding covers a core bug when its cited lines overlap that range and its
issue text is about that bug. Check each finding by reading it, not by line
numbers alone.

An arm is 5 runs. Its score:
- time: mean seconds per run (S1), or mean sum of per-prompt seconds from the
  `prompt i/N done in Ns` lines (S2);
- findings: mean finding count;
- core: mean number of the 3 core bugs covered (S1 only).

A change wins when, against the current baseline arm:
- time is at least 10% lower, and
- findings is no more than 0.5 lower, and
- core is not lower.

Otherwise it loses. A loss is reverted from the working tree. A win is
committed with `flow-pro commit` and becomes the new baseline.

## Changes, in this order

1. Drop the forced answer format: remove `--json-schema` from `claudeArgs`
   and `grokArgs`. Measure Grok and Opus on S1.
2. Bigger prompts: raise `MaxPromptBytes` so S2 packs into about a quarter
   as many prompts. Measure Grok on S2 with `--jobs 1`, and Opus on S2.
   Also record findings per prompt. First check that Grok accepts a prompt
   that large; if it rejects it, record the CLI's error and pick the largest
   size it accepts.
3. `grok-4.7-build-fast`: change the Grok model id. Measure Grok on S1. The
   page's model id changes with it.
4. `--reasoning-effort medium` for Grok. Measure Grok on S1.
5. Sampling settings: look for a repetition-penalty, temperature, or similar
   setting the Grok CLI accepts (`grok --help`, `grok inspect`,
   `~/.grok/config.toml`, `~/.grok/docs`). If none exists, record where you
   looked and skip this change. If one exists, pass it per call from
   `grokArgs` (do not edit the user's config) and measure Grok on S1.
6. Duplicate a slow call: when a Grok call runs past 2× the median of the
   calls already finished in that run (at least 3 finished), start one
   duplicate of that prompt and use whichever finishes first; kill the other.
   Measure Grok on S2 with `--jobs 8` by wall-clock time, not the sum.

## Checkpoints

Record Pass or Fail in the progress file only. A later checkpoint starts only
after every earlier one is Pass. The first checkpoint is Unverified.

1. Baseline. Choose S2. Run 5 Grok and 5 Opus runs on S1, and 5 Grok (`--jobs 1`)
   and 5 Opus runs on S2, at commit `0b05fa4`. Record every run and the
   baseline scores.
2. through 7. One checkpoint per change, in the order above. Each records
   every run, the scores against the baseline, the decision (kept or
   dropped), and for a kept change the commit sha. A dropped change leaves
   `git status` clean.
8. Summary. One table: each change, Grok time before and after, findings,
   core, decision. Then one real `source-review --model grok --jobs 8` on the
   full `hh` checkout with the final code, recording wall-clock time, total
   findings, and whether it wrote `findings.json` and `review.html`.

<!-- CODING-GOAL-END -->
