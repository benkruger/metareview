# Grok parity checkpoint

The quality/speed goal is **not achieved**. No benchmark is running. Cleanup is
committed and full validation passed. The revised goal limits each future work
batch to one paired benchmark and 30 minutes, followed by a user checkpoint.
The next batch only scores the completed run; it makes no new model calls.

Local code commit `6775da0` fixes model-call isolation, Grok tool disabling,
shared effort/system-prompt handling, incomplete-response rejection, and valid
filenames containing double dots. It also adds tests and updates the architecture
reference. The source-review package tests pass at 100% statement coverage.
These are correctness fixes; no experimental speed candidate has been promoted.

- [Goal](goal.md): current acceptance criteria and bounded execution rules. The
  previous definition remains in git history (`08ce3cf`).
- [Experiment history](experiment-history.md): preserved detailed log, originally
  `metareview-grok-parity-progress.md`. Historical paths and observations are retained.
- The four historical goal reviews and context snapshots remain in the existing
  `docs/metareview/reviews/` and `docs/metareview/context/` directories. They review
  the goal, not completion of the implementation or successful benchmark quality.

Latest completed sample: Opus 175.86 seconds; Grok 410.28 seconds (2.33x), across
24 fixed primary files. Its quality score is unfinished. The previous fully
scored sample covered 13/14 required roots and retained false claims.
A successful whole-repository comparison has not been demonstrated.

Private benchmark inputs, raw outputs, prototypes and harnesses remain under
`/tmp/metareview-grok-parity/`; they have not been copied into this repository.
Nothing from this cleanup has been pushed or published.
