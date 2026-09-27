# Grok quality and speed parity

## Goal

Make `metareview source-review --model grok` review the whole `hh` repository
as well as or better than Opus, in no more than three times Opus's elapsed time.
Quality means finding at least the important real bugs Opus finds, with no more
false positives. More reported findings alone is not an improvement.

Keep one review interface and one shared review pipeline. Switching between
Opus and Grok changes the provider/model adapter, not the review task. A GPT
adapter must fit that interface later; evaluating or tuning GPT is outside this goal.

This is the active scope, superseding the earlier fixed lists of experiments.
The agent owns measurement details and may keep testing new approaches and
combinations until the acceptance criteria are met. Do not stop because the
first list of ideas is exhausted. Never report success from partial evidence.

## Scope and invariants

- Work on `whole-repo-source-review`, starting at `6eba389`.
- Review `/Users/ben/code/hh` at commit
  `7d257e0f8efcfcfdd5d31ef71346a3fa48cf192b`; do not modify that repository.
  Check its HEAD before and after each measurement and reject a run if it changes.
- Use the logged-in CLIs. Do not add API keys, direct authenticated HTTP calls,
  or changes to the user's global CLI configuration.
- Preserve the existing command and findings/page contract. Source selection,
  prompts, system instructions, output schema, review stages, validation, retry
  policy, concurrency, and reporting are shared. Translate their settings to
  provider-specific CLI flags in the adapters. Record unavoidable differences.
- Both models get the same source, context, instructions, allowed tools, and
  effort setting for a comparison. No model receives benchmark answers or
  findings from the other model. Keep tools disabled for review calls.
- Shared prompt, packing, scheduling, validation, and recovery improvements are
  allowed. Supported Grok model variants may be explored, with the actual model
  ID recorded. A recovery must produce a real review; never turn a failed call,
  truncation, missing coverage, or malformed answer into an empty successful review.
- Preserve cancellation and timeouts. Count all retries and review stages in
  elapsed time. Do not introduce unbounded retries or silently omit source.
- Leave diff review, its anchor gate, and its FSM semantics unchanged.
- Keep the existing GPT adapter compiling and tested; do not run GPT benchmarks.
- Local implementation, experiments, tests, and commits are authorized by this
  goal. Do not push, open a PR, or publish anything as part of this work.

## Evidence and acceptance

Use the actual `source-review` command built from the measured source revision.
Keep run outputs outside the reviewed tree under `/tmp/metareview-grok-parity/`.
Verify that directory is owned by the current user with mode `0700`, and use
`umask 077` for benchmark artifacts containing private prompts or source excerpts.
Record commands, build revision or diff hash, CLI versions, source commit,
prompt hashes, settings, wall time, per-call time, failures, and output paths.
Keep a durable summary in `metareview-grok-parity-progress.md`; commit the final
evaluation ledger and summaries without committing the private `hh` source.

1. Freeze one full-repo Opus run as the reference, before further experiments.
   Use the earliest fresh baseline already recorded at the starting revision's
   medium effort: `baseline-full-opus-1` (295.9862 seconds). Keep the other runs
   as historical evidence, not additional mandatory discoveries. Verify this
   run's real issues against the pinned source; raw finding counts are not scores.
   Do not select a different reference because a candidate fails.
2. Adjudicate every unique defect claim in the baseline and finalist outputs
   against the pinned source and relevant callers, tests, or dependency contracts.
   Record root-cause matches, evidence, actual impact/severity, and one of
   confirmed, false-positive, or unresolved. Adjudication is independent of the
   producing model and does not accept its explanation as proof. Deduplicate by
   root cause, not line overlap. Unresolved claims prevent a quality PASS.
3. Important means confirmed P0, P1, or P2 defects, classified by actual impact
   rather than the producing model's label. The required reference contains
   these defects from the single frozen Opus run. Do not expand it into a union
   of discoveries across repeated Opus runs. Do not remove a real reference bug
   to make a candidate pass; evidence-based corrections are recorded with reasons.
4. A successful full-repo Grok review must find every important reference bug,
   and may find additional real issues. Extra findings do not compensate for
   missing a required bug. Count each root cause once. Report P3 defects and
   advisories separately, and report additional verified discoveries from either
   model without silently changing the frozen target.
   Count unique false defect claims, including invalid citations skipped by the
   parser, for each run. Grok's false-positive count must be no greater than
   either the frozen reference's or the matched Opus run's. Incorrect factual
   advisory claims also count as false
   positives, so relabeling cannot hide them. Record duplicates separately.
5. Validate a promising candidate with a complete, sequential Opus/Grok pair
   using the same source, shared pipeline, settings, and concurrency. Neither
   run may overlap another benchmark. Count startup, every review stage, retries,
   and output generation in elapsed time. Grok must finish in at most 3x the
   matched Opus time and 3x the frozen reference time. A full run is required
   to demonstrate success; subset results and extrapolations cannot establish it.
   Do not require a six-run batch or discovery-frequency accounting before the
   first decisive comparison. Retain every failure and label changed candidates.
6. Preserve Opus's usefulness: do not meet the ratio by slowing Opus or reducing
   its review quality. Compare the final shared pipeline with the original
   Opus reference as well. Three times its elapsed time is a ceiling, not the
   optimization target: keep reducing time while preserving quality.
   Re-establish a baseline only for demonstrated environment changes,
   with the old and new measurements both retained.
7. Tests appropriate to the changed code pass, changed packages retain 100%
   statement coverage, `make cover` passes, and `git diff --check` passes.
   Run the required artifact and task-done reviews, resolve blockers, and keep
   durable review artifacts. Record what was actually tested.

## Execution

Review this revised goal and finish verifying the single frozen reference.
Keep a fixed, representative subset for cheap screening, including verified
reference bugs, difficult code, and the supporting context needed to distinguish
real defects from false positives. Freeze its paths before the next experiment;
do not shrink it to hide a failure or put expected findings in model prompts.

State one hypothesis and change one factor per experiment. Reject candidates
that miss required subset bugs, add false positives beyond the comparison, or
remain clearly too slow. Only promising candidates advance to full-repo runs.
Do not add more review stages or broaden the experiment plan without evidence
that the added work addresses a measured failure.

Before another full-repo benchmark, implement and verify an overall wall-clock
deadline in the benchmark harness, separate from the existing per-call timeout.
Cap a full Grok run at 3x the frozen Opus reference time (887.9586 seconds), or
3x its matched Opus time if that is lower. On expiry, cancel all review subprocesses,
retain diagnostic logs, and record the run as a timeout failure. Never publish
partial coverage as a completed review or leave child model calls running.
Use an explicit bounded screening deadline as well, recorded before each test.

The next milestone is one full-repo scorecard: both elapsed times, verified real
issues found, reference issues missed, and false positives. Keep improvements
that preserve quality and reduce elapsed time. Do not repeat a known loss
without a new reason it could improve the result.

Update progress after each meaningful result. Continue autonomously through
ordinary implementation choices. If external access prevents progress, record
the exact failure. If the target remains unachieved, say so; an improvement,
passing unit tests, or exhausted experiments is not completion.

## User-approved measurement reset

The original goal required one Grok run to reproduce the union of multiple
Opus runs; an intermediate revision substituted three-run discovery frequencies.
The user subsequently approved a simpler target: one frozen full-repo Opus
reference, its verified important issues or more with no more false positives,
and the fastest Grok run possible within the 3x ceiling. The user also approved
cheap subset screening and a hard overall limit on full-repo Grok runs.
These criteria supersede both earlier formulations. Preserve their review and
experiment history; no earlier failed experiment is retroactively a success.
