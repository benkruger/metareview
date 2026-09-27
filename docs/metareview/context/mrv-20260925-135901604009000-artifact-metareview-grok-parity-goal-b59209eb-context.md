# metareview context: metareview-grok-parity-goal.md

Run ID: `mrv-20260925-135901604009000-artifact-metareview-grok-parity-goal-b59209eb`

## Target

- Path: `metareview-grok-parity-goal.md`
- Repository mode: `metaswarm-extension`
- Git branch: `whole-repo-source-review`
- Git head: `6eba389`

## Artifact Excerpt

```markdown
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
- Review `<reviewed-repository>` at commit
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
Record commands, build revision or diff hash, CLI versions, source commit,
prompt hashes, settings, wall time, per-call time, failures, and output paths.
Keep a durable summary in `metareview-grok-parity-progress.md`; commit the final
evaluation ledger and summaries without committing the private `hh` source.

1. Establish a fresh Opus baseline at the starting revision's medium effort,
   and a current Grok measurement. Earlier results are context, not final proof.
   Use small real `--path` samples to screen experiments, and full-repo runs to
   decide success. The known DirectDeposit issues are leads to verify, not
   automatically true bugs or the complete quality test.
2. Adjudicate every unique defect claim in the baseline and finalist outputs
   against the pinned source and relevant callers, tests, or dependency contracts.
   Record root-cause matches, evidence, actual impact/severity, and one of
   confirmed, false-positive, or unresolved. Adjudication is independent of the
   producing model and does not accept its explanation as proof. Deduplicate by
   root cause, not line overlap. Unresolved claims prevent a quality PASS.
3. Important means co
```

## Service Inventory

No service inventory found.

## Knowledge Facts

No Beads knowledge facts found.

## Suggested Reviewers

- Feasibility
- Completeness
- Scope and alignment
- Architecture
- Intent preservation
- Security
- Testing-quality
- Data-migration
- Runtime-reliability
- Mechanical-precision

Local checkout paths were redacted when archiving this context snapshot.
