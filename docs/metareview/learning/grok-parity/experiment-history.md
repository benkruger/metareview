# Grok parity progress

## Starting point

- Goal: `metareview-grok-parity-goal.md`.
- Branch: `whole-repo-source-review`; starting revision `6eba389`.
- Target: `hh` at `7d257e0f8efcfcfdd5d31ef71346a3fa48cf192b`.
- Status: user-approved goal reset written; revised artifact review pending; fresh baseline recorded; no candidate has passed acceptance.
- Historical final Grok run: 4,188 seconds, 45 prompts, 671 findings. The saved
  outputs and logs still exist. Historical raw finding counts do not prove parity.

## Goal review

Artifact review `mrv-20260925-140350512115000-artifact-metareview-grok-parity-goal-b59209eb`: PASS, zero blockers. Ten independent lenses reviewed the goal; Completeness and Security re-reviewed their resolved findings.

## Investigation: tools were still exposed

The two historical fast-model failures were `max turns reached`, not malformed
review answers. Their Grok session records expose 25 tool definitions and actual
`grep` tool calls despite `--tools ""`. Both had high reasoning effort. Current
installed CLIs: Grok 1.0.41, Claude Code 2.1.281. This is evidence for fixing
the shared no-tools contract before retesting the fast variant. No speed or
quality improvement has yet been demonstrated.

## Fresh Opus baseline

At `6eba389`, medium effort and jobs 8, full `hh` runs completed in 295.99s (293 findings), 290.37s (283 findings), 266.13s (288 findings). All exit 0; source SHA unchanged. Median 290.37s; the 3x budget is 871.11s (14.52 minutes). Actual model reported: `claude-opus-5-5`. Raw responses, prompt hashes and full logs: `/tmp/metareview-grok-parity/baseline-full-opus-(1, 2, 3)/`. Quality adjudication is still in progress; raw counts are not scores.

## Experiments in progress

- Explicit nonempty allowlist plus denial removes Grok tools: live session `01a0d8e6-81f8-7172-b2f7-82458c2387d4` records an empty tool list, one turn, successful output. The diagnostic on `app/models/application_record.rb` took 131.76s.
- Fast model, medium, no tools, original cwd: DirectDeposit sample took 262.99s and returned six findings (unadjudicated). Session `01a0d8e9-620a-7f42-87ba-44e0798d5412` records medium effort, but its actual system message is the 7,513-character coding-agent prompt, despite the requested 115-character reviewer prompt saved separately.
- Isolated working directories and per-process customization suppression are implemented for the shared runner. They reduce ambient context, but the fast variant still changes the system message. An explicit reviewer agent profile is the next diagnostic.
- `go test -cover ./internal/sourcereview` passes at 100.0% after the runner changes. Full coverage and task review remain owed.

## Corrected setup and low-effort screening

- Setting the requested Grok model as the per-process `models.default` prevents model selection from replacing the system prompt. Live fast-model session `01a0d8f4-ea2c-7f73-83bf-5d855ecc4c1d` records the exact shared 115-character system prompt and zero tools. An explicit agent file alone did not fix the reset and was dropped.
- Fast model at low effort, corrected setup: S1 55.02s vs Opus 19.96s (2.76x), but Grok omitted the verified Cancel state defect. Rejected as a quality pass.
- Shared systematic, evidence-grounded instructions at low effort: S1 Grok 64.75s vs Opus 21.97s (2.95x); Grok recovered Cancel. This is a screening result only. A full-repo pilot runs sequentially, Opus then Grok, under `/tmp/metareview-grok-parity/pilot-grounded-low-{opus,grok}-1/`.
- The live benchmark exposed a parser defect: a real kept directory named `.. -t=silently` is rejected as traversal. A regression test first failed, and the parser now rejects only complete `..` path components. These historical harness rejections must not be scored as model false positives.

## Full pilot: grounded instructions, low effort, fast Grok

- Opus completed all 45 prompts in 150.04s, with 126 reported findings. This sets a matched 450.13s Grok limit; the initial baseline limit also remains binding.
- Grok completed all 45 prompts in 902.07s (15.03 minutes), with 375 reported findings. The ratio is 6.01x, so this candidate fails timing. Its DirectDeposit prompt also omitted the confirmed Cancel defect, despite finding it in the small sample. It cannot pass quality either. Both runs exited 0 with the source SHA unchanged; their 45 prompt hashes match exactly.
- Live session metadata confirms both a 13s call and a 297s call actually used `grok-4.7-build-fast`, low effort, and zero tools. Long calls still generate tens of thousands of reasoning tokens; the latency difference is not explained by accidentally retaining tools or high effort.
- The pilot binary predates the literal-double-dot parser fix. Valid citations rejected for that reason are harness defects, not model false positives.
- Independent adjudication has confirmed date-filter and invoice-cent truncation bugs and rejected claims contradicted by actual inherited scopes, soft-delete behavior, helper return values, and the database schema. The initial Opus reference is not yet frozen; adjudication remains incomplete.

## Next experiment: smaller related groups

The first-fit packer puts unrelated files together and separates relevant helpers. The next experimental build sorts by path and prefers 48 KB groups, retaining whole files up to the existing 120 KB hard limit. Selection, instructions, model variants, and effort remain the same as the failed pilot. This is a shared packing change, not a provider-specific prompt. Existing partition, UTF-8, line-number, and complete-file coverage tests pass with the grouping expectation adjusted. The implementation remains in a private Go overlay until measurements justify keeping it.

Screening paths: `app/controllers/admin`, `app/models/reporting`, `app/models/concerns`, and `web/src/routes/DirectDeposit.svelte`. Paired sequential runs use jobs 8 and label `screen-grouped-low-s2plus`.

Opus completed 14 prompts in 56.62s; Grok finished in 407.19s (7.19x). Both exited 0 with unchanged source SHA. Grouping alone fails the speed screen. Its first completed Grok calls took 134–193s, showing that smaller prompts do not reliably make the reasoning shorter. Grok recovered Cancel in this sample; that does not establish full quality parity.

Prepared, not yet measured: a common concise-analysis instruction with a requested 1,500-token reasoning budget (a prompt request, not an enforced cap), and the supported Grok 4.6 variant at the same low effort. Both preserve the supplied review task and interface; neither is an accepted implementation.

The concise-analysis screen ran sequentially under `screen-brief-critical`, after the grouped run finished. It supplied both models the same 12 files covering the verified seed defects and relevant API/validation/login helpers. No findings or expected answers were included in their prompts. Opus took 22.44s; Grok took 83.30s (3.71x), exceeding both the speed limit and the requested reasoning budget (5,777 and 8,231 reasoning tokens). Grok still missed verified defects and made a false claim that literal browser-side `ENV.fetch` exposed a server secret, so this is not a passing candidate.

Executable adjudication checks now reproduce the date-filter boundary, scheduler missing-code acceptance, decimal subtotal truncation, and an actual rejected response-body read that leaves DirectDeposit saving indefinitely. The latter differs from the models' frequent false claim that ordinary fetch failures or invalid JSON omit `data`: the real API helper supplies `{}` or an error object for those cases.

Next screen: `screen-cap4096-critical` applies the same 4,096-token response limit to both adapters, retaining the concise prompt and low effort. Grok uses the documented per-process `models.max_completion_tokens` overlay; Claude uses [`CLAUDE_CODE_MAX_OUTPUT_TOKENS`](https://code.claude.com/docs/en/env-vars). A cap-induced truncation is a failed review, not an empty success.

Result: Opus 23.08s, Grok 82.85s (3.59x), both exit 0. Grok reported 8,266 and 4,394 total output tokens, including reasoning, despite the 4,096 setting. It is not an enforceable total-generation cap in this observed setup; no causal explanation is assumed. The candidate fails the speed screen.

That documentation also shows `CLAUDE_CODE_EFFORT_LEVEL` overrides the CLI flag. The shared runner now explicitly sets the selected effort in the child environment. The regression test first failed with an inherited `high` value, then passed after the change. This workspace had no inherited effort or output-budget overrides, so prior measurements are unaffected. The source-review package remains at 100% coverage.

The adapter now also rejects an explicit `max_tokens` stop from either CLI and Grok's `structuredOutputError`, even if its text happens to contain valid findings JSON. Regression tests demonstrated those incomplete responses were previously accepted. Package tests pass at 100% after the fix.

Next: `screen-grok46-critical`, the same real-code sample with the supported Grok 4.6 variant, shared low effort and grounded instructions (without the ineffective response cap or concise-analysis request). Both adapters use a three-minute per-call timeout for this screen.

Result: Opus 26.96s, Grok 61.24s (2.27x), both successful. Grok still missed the verified scheduler missing-code defect, Cancel, and the future-day boundary, so this does not pass quality. `screen-grok46-medium-critical` now tests the same variant/sample with medium effort for both models, preserving the same timeout.

Medium result: Opus completed in 66.36s; Grok hit the three-minute per-call timeout and the command failed cleanly in 180.36s without publishing a partial review. This is a failed screening run. The timeout is shorter than 3x this new Opus time (199.08s), so it does not by itself prove the variant cannot meet the ratio with a longer timeout. The failed result is retained.

## Focused targets with supporting source

`screen-focused-context-critical` tests a documented combination: preferred 20 KB primary groups, up to 40 KB of deterministic supporting definitions/callers, and a shared system prompt that explicitly acknowledges omitted code. It uses Grok 4.6 and shared low effort. Retrieval uses only source identifiers and declarations, never benchmark findings or another model's answers. Every selected file remains a primary target; supporting files do not replace coverage. Primary grouping/UTF-8/line-number tests and provider/caller retrieval checks pass in the private overlay. Production packing remains unchanged while this is measured.

Result: Opus 34.58s, Grok 120.61s (3.49x), both successful across seven prompts. Six Grok calls took 26–34s; DirectDeposit took 120s. Grok recovered the future-day boundary but still missed Cancel, the rejected-body-read defect, and scheduler missing-code acceptance. Its actual session confirms the new system text, low effort, and zero tools. This is not a quality pass.

Next, `screen-focused-lenses-critical` keeps that source/context setup but performs three common discovery passes: security; state/correctness; and errors/recovery. This tests whether narrower tasks recover the missed defect classes. There is no model-to-model sharing and no expected findings in the prompts. All passes count toward elapsed time; this experiment has no automatic adjudication stage yet.

Result: Opus 53.30s, Grok 116.84s (2.19x), both successful across 21 prompts. Grok recovered Cancel but still missed scheduler missing-code acceptance and the rejected response-body read. Opus also missed some seed defects in this run. The speed screen passes; quality does not. Next screen holds the three-pass pipeline and low effort constant and swaps Grok 4.6 for standard Grok 4.7.

The standard Grok 4.7 three-pass screen failed its three-minute call timeout at 180.28s; Opus completed in 77.66s. No partial review was published. This cutoff is less than 3x the paired Opus time, so it is a failed bounded screen, not proof that the ratio is impossible with a longer timeout.

Goal criterion 4 has been corrected from each Grok run reproducing the union of six Opus runs to equal three-run discovery frequencies per root cause. Grok must equal or exceed both initial and finalist Opus; finalist Opus must preserve its initial frequency. Every confirmed important reference bug remains mandatory. Six affected independent artifact-review lenses returned PASS; four unaffected results carry forward. Review: `mrv-20260925-154013886105000-artifact-metareview-grok-parity-goal-b59209eb`. No experiment is retroactively accepted.

Next screen, `screen-trace-lenses-critical`, retains Grok 4.6, low effort, context retrieval, and three passes, but asks each shared pass to evaluate concrete guard values, state transitions, and deferred error propagation. These are general review procedures; no expected findings or other-model outputs are supplied. It tests the classes missed in prior screens.

Result: Opus 62.41s, Grok 123.00s (1.97x), both successful. Grok covered all eight independently confirmed seed roots, including scheduler nil-equality acceptance and uncaught response-body rejection. Raw findings were 56 and 50; false-positive adjudication is not complete and this is not a full quality PASS.

Full-source shape check: 1,346 selected files, 291 primary groups, 873 discovery calls, 51.8 MB of total prompt text across three passes, 58.8 KB median group including context. Packing took 11.57s. `pilot-trace-lenses-full` now measures the full repo sequentially with common jobs 48 and a five-minute per-call timeout. The machine has 128 GiB RAM; concurrency is explicitly recorded and shared. No verification stage or automatic semantic deduplication is present yet. This is a pilot, not a final validation batch.

## User-approved reset from the side conversation

The user explicitly approved revising the goal and resuming with cheap subset screening before bounded full-repo validation. The shared goal now freezes the earliest fresh full-repo Opus run (`baseline-full-opus-1`, 295.9862s), replaces the multi-run union/frequency requirements, and requires one-factor experiments. Grok must cover the frozen reference's verified important bugs, with no more false positives, and finish within 3x both the reference and a matched Opus run. Keep optimizing below that ceiling.

Next steps for the main run: review the revised artifact; finish adjudicating this single reference and freeze the representative screening paths; implement and verify a whole-run timeout before another full-repo benchmark. The Grok full-run cap is 887.9586s or 3x the matched Opus time, whichever is lower. Timeout must cancel all model subprocesses, preserve logs, and fail the review without publishing partial coverage. Only candidates that pass the cheap quality/speed screen advance to a full-repo scorecard. Existing per-call timeouts do not implement this overall deadline.

The side conversation edited only the goal and this handoff. It could not resume or change the persistent goal controller (`Goal tools require a persistent thread`). No benchmark was launched, cancelled, or claimed successful here. Prior artifact PASS verdicts do not cover this amendment; its review remains pending.

The interrupted 873-call pilot stopped during Opus at 211.76s, exit 1, with no completed review; Grok never started. Observed CLI elapsed time far exceeded reported provider time at jobs 48. This failed pilot is retained and is not timing evidence for parity.

The user-approved revised goal passed artifact review `mrv-20260925-155346658258000-artifact-metareview-grok-parity-goal-b59209eb` with zero blockers. Seven affected lenses re-reviewed; three unchanged lens results carry forward. Work is continuing manually following the user's resume instruction; the persistent controller still reports paused and exposes no agent-side resume operation.

The benchmark harness now requires an overall deadline and launches each review in a private process session. Timeout kills all session process groups, including the separate groups created for model calls; premature output moves to failed-output. Three process tests passed (success/error exit, nested-group timeout preserving unrelated processes, and early-parent failure cleanup). End-to-end fake runs returned timeout exit124 at 0.89s for a0.8s cap and0.69s for the matched cap0.6s, retaining diagnostics and no completed output. Cleanup time is included; timeout runs always fail.

The frozen screening manifest is `/tmp/metareview-grok-parity/frozen-screening.json`: 24 exact paths, retaining all prior critical targets and adding Swift, Ruby parsing and source helpers needed to reject known false positives. Screening cap:240s per Opus run; Grok is bounded by min(240s,3x matched Opus). No new model experiment has started under this reset. Frozen baseline identity is the earliest run; its source adjudication is still incomplete.

The frozen subset's 27 original Opus claims have now been checked against source: 16 confirmed (12 important roots, four P3 items), 11 false positives. Evidence and root IDs are in `/tmp/metareview-grok-parity/screening-reference.json`. Checking the rest of the full reference continues; no full-repo quality PASS is claimed.

Next hypothesis: the same three tracing checklists can run in one model call per source group, retaining coverage while removing repeated startup/context. `screen-combined-trace-fixed` changes only this call grouping from the prior trace-lenses candidate. Model/effort, primary grouping, retrieval and jobs8 stay fixed. It uses the frozen24 paths and verified overall cap240s (Grok also <=3x matched Opus). It cannot advance unless required subset roots are covered and false positives meet the comparison.

Combined-call result: Opus56.97s, Grok92.65s (1.63x), both completed11prompts without timeout. Grok covers8/12 required subset roots; misses scheduler missing-code acceptance, iOS Cancel discarding edits, the weekend dashboard gap, and the ban parser extended-regex spaces. Candidate rejected on recall before further false-positive scoring.

Next screen `screen-combined-medium-fixed` changes only shared effort from low to medium. Earlier medium screening lacked the focused context and concrete tracing instructions, so this is a new configuration. Frozen24paths, jobs8, cap240s, Grokalso<=3x matched Opus.

Medium effort: Opus82.26s; Grok exceeded240s and failed. Cleanup raised PermissionError before saving the result; an immediate follow-up verified the private session82280 had no live children. Exact total elapsed was lost, so the reconstructed failure records elapsed>=240s, not an invented time. No completed output was accepted. The harness now deduplicates process-group signals, tolerates an exited-group race only after verifying no live members, and saves cleanup failures as failures. Five timeout tests pass after reproducing both missing cases.

Next `screen-focused-fixed` changes the combined low-effort candidate into three focused calls using the same checklist content and exact packed source bodies. This controls the earlier promising three-pass result on the full frozen24-file sample; no full run follows until recall and false positives pass. Other settings and overall limits are unchanged.

Focused fixed-sample result: Opus84.34s/79raw findings; Grok190.17s/81raw findings (2.25x),33calls each. Grok recovers iOS Cancel but still misses scheduler missing-code acceptance, weekend filtering and extended-regex spaces:9/12required roots. Recall fails; no full run and no claimed false-positive pass.

Next `screen-focused-wide-fixed` changes only preferred primary grouping20KB->80KB, retaining40KB supporting-context ceiling within120KB total, three tracing passes, low effort, Grok4.6 and jobs8. This tests fewer repeated calls and more coherent local source; the earlier larger-group trial lacked this model and focused procedure. Frozen sample/deadlines stay fixed.

115of293 frozen full-reference claims have now been independently source-adjudicated. Reference verification remains incomplete; that does not change the fixed sample or mandatory bugs.

Wide-group result: Opus43.18s/39raw findings, Grok83.10s/43raw findings (1.92x),9calls each. Grok finds7/12required roots; misses scheduler MFA, iOS Cancel, weekend gap and both ban-parser roots. Speed improves but quality regresses, so no full run.

Next `screen-focused-wide-fast-fixed` changes only Grok model4.6->4.7-build-fast. The common three-pass procedure, larger groups, low effort, frozen sample and limits are identical. This revisits the fast variant with the newer review procedure, not the earlier failed whole-repo configuration.

Fast-model result: matched Opus38.62s; Grok4.7-build-fast timed out at115.996s with deadline115.868s (3x matched). Cleanup succeeded and no partial result was published. No quality pass.

Next `screen-verified-fixed` returns to the combined low-effort control and changes one factor: adds a same-model verification pass for each source group. It receives its own provisional draft and the same source, checks unsupported claims, and searches for omissions; only its final replacement findings are published. This addresses measured false assertions, duplicate claims and missing roots, without cross-model answer sharing. Raw drafts are retained separately as provisional evidence. Every stage counts toward elapsed time; failed/malformed verification fails the run. Other settings and frozen24paths remain unchanged.

Verification-pass result: Opus85.84s, Grok169.92s (1.98x),22calls each. Grok still misses scheduler MFA, iOS Cancel, weekend filtering and extended-regex spaces. Recall fails; no full run. The additional pass does not justify its cost.

Reference correction: B1-65 (Kaiser fractional adjustment) is false-positive. Earlier arithmetic reproduction did not establish reachable input: the production builder converts to Integer, the preview uses70, and draft edits permit integers only. Evidence and superseded adjudication are retained. The fixed subset now has11 important roots, four P3 claims and12 false positives. All prior candidates still fail the same recall gaps; each candidate that reported the Kaiser claim loses one true match and gains a false claim. The verification candidate finds7/11. Paths and reference run are unchanged.

Next screen `screen-focused-wide-medium-fixed` changes only shared effort low->medium from the wide three-pass Grok4.6 control. Fewer concurrent groups may allow deeper reasoning within the bounded screen where medium with11 groups timed out. Same frozen24paths, jobs8,240s overall cap and Grok<=3x matched Opus.

Wide-group medium result: Opus87.24s, Grok timed out at240.10s, cleanup successful. Three completed Grok calls took197-208s; not all9 completed. No output accepted and no full run.

Next `screen-focused-grok45-fixed` changes only Grok4.6->supportedGrok4.5 from the20KB,three-pass,low-effort control. That control had the best fixed-sample recall so far (8/11 after the evidence-based correction), but missed three required roots. Same frozen paths, jobs8 and bounded240s/3x screen; no model gets expected answers.

Grok4.5 result: Opus101.99s; Grok timed out at240.12s with successful cleanup. First completed calls were91-145s; not all33calls finished. This bounded screen fails and provides no completed quality result.

Next `screen-focused-audit-fixed` changes the shared response procedure from the low-effortGrok4.6 three-pass control: each call must provide concise concrete check results per primary file before final findings. It adds no calls or models. Audit notes remain private; published findings retain the existing contract. Schema and instructions are identical between providers, all primary files must have nonempty check notes or the call fails, and grouping uses the prior header length to preserve primary membership. Five experimental validation cases passed after missing/empty/malformed audit cases failed against the initial stub. Frozen24paths, jobs8 and240s/3x limits remain unchanged.

The audit Opus run completed in115.95s. A cosmetic pair-harness edit introduced an indentation error before Grok could start; no Grok subprocess launched. Two mock orchestration regressions reproduced the missing matched-time handoff and failed-run stop defect, then passed after correction. The already completed Opus measurement remains valid; Grok is being launched sequentially with its exact matched time and the original240s cap. No model inputs or candidate code changed between the pair.

Audit result: Opus115.95s/48raw findings, Grok231.50s/72raw findings (2.00x),33calls each. Grok now covers9/11required roots, recovering MFA and weekend filtering but missing iOS Cancel and extended-regex spaces. False positives remain; no full run.

Next screen `screen-focused-function-audit-fixed` changes only audit granularity: every source-detected declaration plus each file's top-level executable code gets an explicit audit entry, generated from source without expected findings. Same model, low effort, grouping, three passes, context, jobs and frozen paths. Validator rejects missing items. Experimental packing is not yet suitable for full-repo promotion.

Function-audit result: Opus failed at65.53s because one required function audit was absent; validation rejected the result, cleanup succeeded and Grok never started. Next `screen-focused-function-map-fixed` changes only the audit representation to an object whose source-generated keys are individually required by the common schema. The procedure, source, model and caps remain the same. This tests reliable checklist completion without retries.

All293 frozen reference findings have now received an initial source adjudication. Earlier counts included two separately tracked false components; the actual item count before this batch was277/293. Root deduplication and mixed-claim FP accounting still need consolidation before a final scorecard. No benchmark pass is inferred from completion of this ledger.

Required-map audit result: Opus150.25s/50raw findings; Grok timed out at240.11s after28of33prompts, with successful cleanup. Required schema keys prevented the earlier omission failure in completed calls, but this screen is a time failure and partial findings are not accepted. Next `screen-combined-function-map-fixed` changes only call grouping: the same three checklists run together in one call per primary group, retaining mandatory per-declaration audits. Frozen24paths, model4.6,loweffort,context,grouping,jobs8 and240s/3x limits remain unchanged.

Raw-response reconciliation found four additional baseline claims rejected by the original parser because valid directory names contained literal double dots. All four describe the same verified P3 generated-coverage clutter, not important functional bugs. Baseline:297raw claims,293published findings; initial root ledger has77important and80P3roots. The77important-root target is unchanged by these parser-skipped claims. The existing parser correction accepts these literal names while still rejecting parent-directory traversal.

Combined mandatory-audit result: Opus87.04s/29findings; Grok112.28s/34findings (1.29x),11calls each. Grok covers7/11required roots and misses scheduler MFA, iOS Cancel, weekend filtering and extended-regex spaces. The extra function-level output does not improve recall over ordinary combined calls. Candidate rejected; no full run.

Next `screen-focused-semantics-fixed` changes only the correctness checklist of the best9/11 per-file-audit control. It explicitly asks for concrete parsing/pattern examples under actual language semantics and UI-event state comparisons. These are general procedures, with no benchmark answers, source-specific names or model-produced findings in the prompt. Same three passes,20KBprimary/40KBcontext,model4.6,loweffort,jobs8 and frozen screening limits.

Adjudication consistency correction: B1-52 is a real configuration-dependent SQL day-boundary bug. ActiveRecord with this app's :local setting serializes Friday18:00Pacific to Saturday01:00 on a UTC host; :local alone did not disprove the claim. The important reference now contains78roots. Separately retained three disproved factual components from mixed claims (weekday parser example, unchanged fractional input validity, and alleged deletion of merged bytes). These corrections are source-driven and do not change screening paths or its11required roots.

Semantics-checklist run: Opus213.66s/45findings; Grok stopped at206.69s when the experimental audit validator rejected malformed JSON text. Inspection shows a restarted, complete trailing JSON object exactly equal to the CLI structuredOutput, with the required file audit and an actual empty security-pass findings array. The existing production extractFindings parser already handles this case; the new audit validator incorrectly bypassed it. Next `screen-focused-semantics-normalized-fixed` changes only audit validation to use the same extraction before checking file coverage. Complete-restart, unfinished-restart and missing-coverage regressions cover this repair. Model inputs/settings/deadlines remain unchanged; the failed run is not relabeled a success.

Normalized semantics result: Opus123.05s/46findings; Grok234.00s/75findings (1.90x),33completed calls. Grok covers8/11required roots: Cancel is recovered, but MFA, weekend filtering and extended-regex spaces are missed. No quality pass or full run.

Next `screen-focused-wide-medium-jobs16-fixed` changes only concurrency8->16 from the earlier wide-group,medium-effort,Grok4.6 trial. The frozen subset has9calls, so all can start in one wave. That earlier configuration timed out before completing its queued ninth call; higher concurrency tests whether medium reasoning fits the bound without changing either model's task. Reuses the exact existing binary, same24paths and240s/3x cap.

Wide-medium jobs16 result: Opus97.52s/61findings; Grok timed out at240.11s with clean cancellation, only4of9calls completed (157–231s). Removing the queued ninth call did not make medium effort fit this bound.

New packing hypothesis: source-review's120KB byte cap is much smaller than the documented model context windows: [Grok4.6,500Ktokens](https://docs.x.ai/developers/models/grok-4.6) and [Opus5.5,1Mtokens](https://platform.claude.com/docs/en/models/opus-5-5/overview). These API capabilities do not guarantee the CLI accepts every large prompt; errors remain failures. `screen-focused-large-fixed` changes only source packing capacity from the normalized-semantics low-effort control:1MBtotal,960KBpreferred primary instead of20KB, same40KBsupport limit. The fixed24files fit one shared source batch and retain all three passes. Model4.6,loweffort,jobs8,schema,checklists and240s/3x cap stay fixed. This is an isolated source-review experiment; diff-gate120KBsharding is unchanged. A packing check verifies complete original numbered bytes beyond120KB and the new hard byte bound.

Large-batch low-effort result: Opus52.84s/14findings; Grok108.96s/19findings (2.06x),3calls. Both CLIs accepted the complete sample (Grok about60.6Kinput tokens/call), but Grok found only5/11required roots. Opus also missed several. Large context at low effort loses recall; no full run.

Next `screen-focused-large-medium-fixed` changes only shared model reasoning effort to medium from that large-batch control. The3-call shape tests whether deeper reasoning can use the broader source context effectively. The declared measurement safety bound is300s per run with5m call timeout; Grok remains capped at3x matched Opus. This extra minute cannot turn a >3x result into a pass, and it does not change the completed control measurements. Source, prompts, stages, schema, jobs8 and frozen24paths are unchanged.

Large-batch medium result: Opus149.35s/25findings; Grok timed out at300.10s with clean cancellation and no completed review. Deeper reasoning still exceeds the cheap-test bound.

Next `screen-focused-large-fast-fixed` changes only Grok model4.6->4.7-build-fast from the completed large-batch low-effort control. Larger batches now need just3calls, so this tests the stronger fast variant without the earlier queued calls. Shared loweffort,source,checklists,per-file audits,schema,jobs8 and3x comparison stay fixed. Bound240s,calltimeout4m as in the completed loweffort control.

Large-batch fast result: Opus47.22s; Grok timed out at141.77s (3x matched deadline141.66s), cleanup successful. No partial result accepted.

Next `screen-focused-file-results-fixed` returns to the large-batch Grok4.6 low-effort control and changes only response organization: each primary file has its concrete checks immediately followed by its own findings. The same existing eight-field findings are flattened deterministically for publication. This tests an observed omission: the prior Grok audit described the inactive-alias crash but left it out of its global findings array. No expected findings or other-model answers enter the prompt. Source, three checks, effort, jobs8 and240s/3x limits stay fixed. Missing file coverage, malformed entries, or incomplete JSON still fail.

Per-file findings result: Opus56.55s/15findings; Grok failed at75.35s because its audit changed a primary path (`direct_deposit_controller.rb` to `direct_deposit/controller.rb`). Coverage validation correctly rejected the run; cleanup succeeded. Next `screen-focused-file-map-fixed` changes only the response representation to a map with exact primary paths as individually required schema keys. It retains per-file checks followed immediately by findings, without extra calls. Same source, model4.6,low effort, three passes and bounded comparison.

Exact-file-map result: Opus83.65s/19findings; Grok117.16s/33findings (1.40x),3calls. Required file keys worked, but Grok covers only6/11important roots; misses both dashboard bugs, both ban-parser bugs and iOS Cancel. False claims remain; no quality pass or full run.

Next `screen-focused-wide-medium-extended-fixed` reuses the exact prior wide-medium binary at jobs16. This is a measurement-bound change, not a claimed optimization: the prior run stopped at the240s screening safety cap although 3x matched Opus allowed292.57s, with calls completing at157–231s. The new declared cap is360s/6m, always further limited to3x the freshly matched Opus time. This tests whether the cheap-test safety cap censored a configuration that can satisfy the actual goal. No source, model4.6, medium effort, three checks, grouping, schema, or concurrency changes; no additional stages. All failures retained.

Reference mixed-claim audit: split two independent important causes already present in the frozen output (legacy BG waiver-copy omission and course-load stale error) from their neighboring claims. Required whole-repo roots80; the fixed sample remains11. Also retained the dashboard-test timezone overgeneralization as a separate false component (UTC+2 disproves it), and corrected the adjudicator's mistaken claim that ComplianceSign submit checks canSubmit. Reference unique false roots109. These are source-driven accounting corrections, not discoveries added from later model runs. Final audit remains in progress.

Extended medium result: Opus85.34s/52findings; Grok timed out at256.14s against256.03s matched limit, with2of9calls complete. Cleanup succeeded. The actual3x requirement, not the screening safety cap, now rejects this configuration.

Next `screen-focused-all-context-fixed` returns to the normalized-semantics Grok4.6 low-effort control (33calls) and changes supporting-context retrieval. Frozen24primary paths/grouping stay fixed; helpers can now come from all selected first-party source at the same pinned commit, and explicit Ruby partial names match underscored template filenames. Source-derived dependencies only, no handpicked expected answers. This addresses observed false claims about omitted caller validation and shared UI handlers. Same three passes, per-file audit,40KBsupport bound and jobs8. Screening cap360s/6m and Grok<=3x matched, following the declared extended bound; no full run without quality.

Supporting-context preflight caught another retrieval problem: raw identifier overlap crowded out an explicitly rendered helper even after making it available. Explicit quoted source references now outrank incidental matches, including Rails partial path normalization; no benchmark-specific path is in the retrieval implementation. Synthetic dependency/primary-scope checks and pinned-source manifest verification pass with24primary files and11source groups.

Dashboard adjudication reproduced the actual view script plus shared table-search script in jsdom with the actual jQuery3.7.1. Initial `mode=future` hides the boundary day; clicking Future after ready shows it because the shared keyup handler replaces the filter. Clear shows weekend rows, and a Saturday text search works. The midnight root remains confirmed for reload/bookmarked future mode; its trigger is narrower than earlier isolated-expression evidence. The weekend preset gap remains present, with the Clear/search workaround explicitly recorded. [jQuery3 ready callbacks are asynchronous](https://jquery.com/upgrade-guide/3.0/#breaking-change-document-ready-handlers-are-now-asynchronous).

Retained the original Future-button overstatement as a separate false component after the full DOM reproduction, rather than quietly narrowing the claim without FP accounting. Whole-repo reference remains80important roots; unique false roots110. Future-midnight matching must preserve the real initial/reload trigger, and analogous overstatements in finalist claims must be counted consistently.

All-source supporting-context result: Opus137.25s/47findings; Grok227.15s/83findings (1.65x),33calls. Grok covers10/11required roots, the best recall so far; only extended-regex spaces remains missing. Cancel is identified through retained change flags. Multiple false/mixed claims remain, so no quality PASS or full run.

Next `screen-focused-mixed-effort-fixed` changes one shared policy from this10/11control: correctness calls use medium effort; security and reliability keep low. Both models receive the identical per-stage effort policy. The remaining recall failure concerns exact language semantics, which deeper reasoning has recovered in other comparisons; increasing only that existing pass limits the extra time. No added stages or expected answers. Same source, supporting-context retrieval,33calls,per-file audit and jobs8. Declared360s/6m cap, further limited to3x matched Opus.

Prepared, not yet benchmarked: stronger priority for directly referenced type definitions over generic method-name overlap. The10/11control's auth group omitted AuditTrail despite a direct AuditTrail call, while including unrelated methods with matching names; Grok consequently repeated a false claim that the audit stores full SSNs. A synthetic tight-budget regression reproduced the omission, then passed with type-reference weighting. The pinned-source manifest now includes AuditTrail; explicit rendered partials and all24primary targets remain covered. This preparation does not change the running mixed-effort benchmark or count as a measured improvement.

Mixed-effort result: Opus169.02s/60findings; Grok timed out at the360s screening cap with clean cancellation. This cap is lower than3x matched Opus (507.07s), so it proves a failed bounded screen, not an intrinsic >3x ratio. One completed medium call took329s; several queued correctness calls remained unfinished. A reversed citation953..328 was rejected in the partial output and retained in diagnostics. No complete Grok output or quality score is accepted.

Next `screen-focused-type-context-fixed` tests the prepared direct-type context priority from the completed10/11 all-context low-effort control. Only retrieval weighting changes; same33calls,primary groups,support byte limit,audit,model4.6,low effort,jobs8 and360s/3x cap. This specifically addresses the observed omission of called class definitions behind false claims; no handpicked source file or benchmark answer is inserted.

Direct-type context result: Opus178.71s/50findings; Grok269.30s/73findings (1.51x),33completed calls. Grok covers8/11required roots, missing scheduler MFA, weekend filtering and extended-regex spaces. The full-SSN audit claim is absent, but other false claims remain and recall regressed. No full run.

Next `screen-focused-declaration-audit-fixed` changes audit granularity from the best all-context low-effort control: every source-detected declaration, constant and file top level receives a required concise concrete-check entry. Short generated IDs avoid repeating long paths in output. Both models get identical source-generated inventory and schema; no expected findings enter prompts. Same33calls,24primary paths,context retrieval,model4.6,loweffort,jobs8 and360s/3x cap. Synthetic inventory/normalization checks passed; this experimental packing needs hardening before any full-repo promotion.

Source-adjudication correction: the frozen reference already reported repeated invoice subtotals in Kaiser, Generic and Sutter templates (B1-66/72/86). Earlier dismissal traced production batch sorting but missed the separately routed preview actions, which order only by start_at. Rendering each actual ERB template with synthetic A,B,A chronological rows reproduces two full A subtotals; A,A,B renders one. All three have the same P2 financial-display impact (including B1-86 originally labeled P3). Required reference now83important roots; unique false roots107. The unchanged24-file sample now requires14roots. The best all-context result covers13/14; direct-type context covers11/14. Both still miss extended-regex spaces; no past failed candidate becomes a pass. Superseded adjudication and scorecards are retained.

Production adapter/parser verification repeated in the current workspace: `go test ./internal/sourcereview -coverprofile=...` passed with100%statement coverage; `git diff --check` passed. These checks do not establish benchmark quality or task completion.

Declaration-audit result: Opus142.68s/59findings; Grok247.78s/82findings (1.74x),33calls. Grok covers10/14important roots; misses scheduler MFA, iOS Cancel, weekend filtering and extended-regex spaces. Required audit entries do not guarantee accurate evaluations: both models explicitly mis-evaluate the same regex despite noting its modifier. More checklist output regressed recall. No full run.

Next `screen-focused-literal-normalization-fixed` returns to the best13/14 all-context control and changes only one correctness instruction: write the interpreted literal/pattern after escapes and modifiers, then evaluate examples against that form instead of its visual spelling. This addresses the observed incorrect audit evaluation, without source-specific examples or expected findings. No extra pass or schema expansion. Same24paths,33calls,model4.6,loweffort,jobs8 and360s/3x cap.

Prepared (not benchmarked) selective-effort fallback: retain the best all-context prompts and use medium only for correctness groups whose primary source has an extended-mode pattern literal; all other calls stay low. This is a source heuristic shared by both adapters, with no target paths or expected findings. Tests first rejected the old all-correctness-medium policy, then passed for primary-only classification and CLI/environment agreement. A model-free full-source packing preflight selected all1346files into291groups/873calls, with only3correctness calls classified medium and no missing primary headers. No performance or quality gain is inferred from this preflight.

Literal-normalization result: Opus141.93s/45findings; Grok229.34s/71findings (1.62x),33calls. Opus recovered the regex defect at low effort, but Grok still mis-evaluated the interpreted form. Grok covers11/14required roots, missing scheduler MFA, weekend filtering and extended-regex spaces. No quality pass or full run.

Next `screen-focused-pattern-effort-fixed` runs the prepared selective-effort policy from the best all-context control, without the unsuccessful literal-normalization instruction or declaration audit. Only correctness groups with an extended-mode pattern literal in primary source use medium; other calls use low. Identical policy for both adapters. Same24paths,33calls,context,jobs8. Declared bounded screen480s/8m, still capped at3x freshly matched Opus; this larger safety bound avoids prematurely censoring the single medium call like earlier360s mixed-effort measurement. It cannot turn a >3x measurement into a pass.

Selective-pattern-effort result: Opus161.31s/47findings; Grok451.39s/72findings (2.80x),33calls. The single medium call took318s and still missed extended-regex spaces. Grok covers11/14required roots, also missing weekend filtering and iOS Cancel. Extra reasoning at this granularity did not recover the target. No quality pass or full run.

Next `screen-focused-all-context-fast-fixed` changes only Grok4.6 to4.7-build-fast from the best all-context low-effort control. Earlier fast-variant trials lacked the improved context/audit setup, or used large batches with poor recall. This tests the stronger variant with the best current shared setup and enough bounded time for all33calls. Same24paths,prompts,schema,loweffort,context,jobs8. Cap480s/8m and3x matched Opus. Actual fast-variant IDs and failures remain recorded; no change to production default.

All-context fast-variant result: Opus129.44s/52findings; Grok timed out at388.41s against388.31s (3x matched), with clean cancellation. No complete Grok output or accepted quality score. Diagnostic calls retained.

Fast-variant diagnostics: the completed correctness call correctly reports extended-regex spaces and the iOS call reports Cancel state. Those are observations from a failed run, not a quality score; the completed dashboard calls still miss the weekend preset gap. Next `screen-focused-all-context-fast-jobs16-fixed` changes only concurrency8->16, retaining the exact binary and shared inputs. The previous run had started all33calls and completed25at323s, leaving queued work before its388s limit. This tests whether additional parallelism makes the stronger variant practical. Same480s/8m safety cap, always3x freshly matched Opus.

Prepared, not benchmarked: linked supporting-context retrieval replaces generic method-name overlap with explicit source references, named/qualified types, reverse filename references to callers, and up to three dependency steps. Templates contribute executable ERB/script references, not human headings: the old picker treated the heading Invoice as a reference to the unrelated Invoice model, feeding the private-method false claim. Synthetic tests reproduced missing receiver context and this heading mistake before the repair, and now pass along with byte-budget/primary-source checks. Dotfiles no longer all alias the literal dot. The pinned sample now includes AuditTrail for auth, preview callers for all invoice templates, and the actual public-method parent for Sutter. Same40KBsupport cap and unchanged24primary files/11groups. This is a private experimental overlay; no measured quality gain is claimed.

Fast-variant jobs16 result: Opus124.25s/48findings; Grok303.04s/80findings (2.44x),33completed calls. First complete screen covering14/14required roots. Several false/mixed claims remain, including omitted audit redaction, a wrongly resolved private method, and a disproved alias example. Recall and timing pass alone do not establish quality; source FP adjudication and a complete full-repo pair are still required. No full run yet.

Next `screen-focused-linked-context-fast-fixed` changes only supporting-context retrieval from the completed14/14 fast/jobs16 control. It uses the tested source-reference/type/caller graph, template-aware reference extraction, qualified Ruby type resolution and a bounded inheritance preference instead of incidental method-name overlap. Same24primary files,11groups/33calls,40KBsupport cap,prompts,audit,schema,model4.7-build-fast,loweffort and jobs16. Both adapters receive identical selected context. Declared480s/8m cap, further limited to3x matched Opus. The hypothesis is fewer false claims about omitted guards or wrongly resolved receiver methods while retaining recall.

Mixed-claim accounting correction: an additional actual-script Saturday reproduction shows Same Day includes Saturday. B1-34 correctly identifies the preset gap (both weekend days on Friday; Sunday on Saturday), but its consequence overstates Saturday self-exclusion. Keep the real P2 root and count the false component separately, consistently with candidate claims. Reference remains83important roots and14sample roots; unique false roots108. This does not turn an earlier recall failure into a pass.

Linked-context result: Opus107.64s/52findings; Grok249.69s/92findings (2.32x),33completed calls. Grok retains14/14required roots, and the earlier bank-redaction/private-Sutter-method/raw-date-injection false claims disappear. Other false claims remain, including misresolved inherited API authentication and a hypothetical truthy :expired caller despite the actual caller using == true. Larger raw count is not a quality improvement. No quality PASS or full run; per-claim FP adjudication remains necessary.

Next `screen-focused-linked-verified-fast-fixed` adds one same-model verification stage to the completed linked-context14/14 control. False claims persisted after improving source context, so each model now checks its own draft against the identical original source, traces reachable callers and narrows mixed claims before publishing. No reference answers or other-model output enter prompts. Empty drafts need no second call; drafts with parser-skipped claims are still verified. Final audit/parse failures fail the run. Same24paths,primary packing,context,schema,model4.7-build-fast,loweffort and jobs16. All stages count in elapsed time; bounded screen600s/10m, further limited to3x freshly matched Opus. The larger safety bound accommodates the measured extra stage but cannot turn >3x into a pass. Focused tests cover both adapters, removing/retaining findings, empty drafts, invalid draft citations, final coverage failure and bounded source-plus-draft prompts; passed. No production promotion or full run yet.

Mixed-claim audit correction: the actual MfaAuthentication concern sets both client_id and mfa_code cookies with the same ten-minute lifetime. A fixture loading that concern confirms that deleting only the code cookie permits nil==nil with a resolved identity, while waiting for both to expire leaves no login_object and is rejected. The important scheduler-MFA root remains; the original frozen claim's alternative expiry-only bypass is separately false. Reference remains83important roots/14sample roots; false roots109. Apply this distinction to finalist claims as well. This correction does not turn any failed candidate into a pass.

Per-draft verification result: Opus226.35s/41final findings (33discovery+25verification calls); Grok378.64s/68final findings (33+28calls),1.67x. Shared discovery prompt hashes match exactly. Grok covers11/14required roots, missing browser ENV.fetch, weekend filtering and extended-regex spaces; those were already absent from its discovery drafts, so this trial does not establish that verification removed them. False claims remain, including a nonexistent truthy :expired caller and external e-badge URL, despite verification. Mixed factual errors are retained in the partial ledger; all remaining claims are unresolved. Reject on recall, no full run. The extra step is not promoted.

Prepared next context repair from the faster completed14/14 linked-context control: resolve named types relative to Rails namespaces, honor explicit root ::, retain acronym directories, and retrieve actual consumers of named classes/modules. The current auth pack omitted LoginController and Api::V1::ApplicationController, directly causing false statements about login result checks and inherited authentication. Synthetic tests cover both distinctions and ensure context caching cannot change later prompts. The unchanged40KBsupport budget now includes both real callers/parents in the pinned auth group. This is source-derived retrieval, with no expected answers or path-specific exception. No extra model stage.

Next `screen-focused-scoped-context-fast-fixed` tests that namespace/consumer context repair as the sole changed factor from `screen-focused-linked-context-fast-fixed`. It returns to33discovery calls without the unproven verification stage. Same24primary files,source groups,40KBsupport limit,instructions,audit,schema,Grok4.7-build-fast,sharedloweffort and jobs16. Declared480s/8m safety bound and Grok<=3x freshly matched Opus. Focused retrieval/audit tests and pinned auth-context checks pass; no measured quality improvement yet.

Namespace/consumer context result: Opus75.93s/46findings; Grok timed out at227.90s against227.80s (3x matched Opus), with25of33calls complete and clean cancellation. No complete Grok output or accepted recall/FP score. Completed auth diagnostics no longer invent the truthy-expired caller or miss the inherited API guard. Slow calls are still provider reasoning: completed calls reach207–225seconds despite low effort, with thousands of reasoning tokens. Increasing jobs alone would not shorten those already-running calls, and would also tighten the matched Opus deadline. No full run.

Next `screen-focused-small-batches-fast-fixed` changes preferred primary batch size20KB->10KB from the namespace-aware control. The longest completed Grok calls spend207–225seconds/21K–23Kreasoning tokens on groups containing several primary files, versus Opus under48seconds. Smaller primary groups test whether reducing the number of operations per call limits this reasoning cost and preserves recall. Whole files are retained when larger than the preference; the hard prompt budget120KB and support limit40KB remain. Same24paths,context selection,three passes,audit,schema,Grok4.7-build-fast,sharedloweffort,jobs16; no new review stage. Declared480s/8m safety bound, always capped at3x fresh matched Opus. Tests verify each frozen primary path appears exactly once in the source groups, along with scoped caller/parent retrieval and prompt bounds.

Smaller-batch result: Opus206.96s/63findings; Grok429.22s/106findings (2.07x),48calls each. Grok covers12/14required roots, missing the weekend gap and iOS Cancel's unsaved-field/flag restoration (the different pending-check banner claim does not substitute). Many false claims remain. Dashboard/layout separation also loses the global keyup helper from the dashboard context, causing false Clear-button reports. Reject: more calls and longer absolute Grok time without sufficient quality. Opus startup/waiting varied substantially: some calls spent about100seconds outside the reported API duration; the ratio alone is not an optimization result. Aggregate host inspection showed substantial existing memory use from other processes, so high-concurrency full runs need capacity checks rather than assuming all128GB is free.

Next `screen-focused-scoped-context-main-fixed` returns to the20KB namespace-aware control and changes only the Grok variant from4.7-build-fast to standard4.7. The earlier standard4.7 test used the old smaller sample without per-file audits/current caller context and stopped at180seconds, below its233second matched limit, so it did not establish impossibility. This tests the stronger standard variant with corrected source links and the current bounded protocol. Same24paths,33calls,three passes,audit,schema,sharedloweffort,jobs16 and480s/8m safety bound, further limited to3x freshly matched Opus. Model-ID and focused retrieval/audit tests pass. No API keys, tools, extra stages or production-default change.

StandardGrok4.7 result: Opus146.92s/48findings; Grok438.81s/91findings (2.987x),33completed calls. It narrowly meets the paired time ceiling but covers13/14required roots, still missing the weekend preset gap. Multiple false/mixed claims remain, including Kaiser active-scope omission, unsupported fractional adjustments and treating literal ENV.fetch as ERB interpolation. Reject on recall; no quality PASS or full run. Standard4.7 offers no demonstrated improvement over the faster14/14 control.

Next `screen-focused-split-correctness-fast-fixed` changes review-pass granularity from the20KB namespace-aware fast control. The existing correctness checklist is split into two independent passes: value/language semantics and state transitions; its substantive checks are unchanged. Security and reliability stay as before. Across completed runs the combined pass inconsistently recovers arithmetic/pattern/time issues and Cancel state, while reducing primary batch size lost global context. A separate focus tests that measured failure without shrinking source context or adding reference answers. This makes44calls over the same11groups instead of33. Both models use the identical four-pass procedure, per-file audit, source/context, schema, sharedloweffort and jobs16. ModelGrok4.7-build-fast. Declared480s/8m safety bound and Grok<=3x freshly matched Opus. Focused context/audit checks pass; no production promotion.

Split-correctness result: Opus106.37s/68findings; Grok timed out at319.19s against319.10s (3x matched), with41of44calls complete and clean cancellation. No completed quality score or full run. A first-wave reliability call still took307seconds, so more focus alone did not remove the long reasoning tail.

The supporting-context picker had two concrete ranking problems: a doubled inheritance preference exactly cancelled distance decay, allowing an ancestor to displace the actual receiver; and a file mentioning its own filename diluted the weight of its real callers. Synthetic regressions fail under the old picker and pass after stronger distance decay and excluding already-primary files from caller normalization. The actual dashboard context now includes Assignment::Interpretation, whose query was omitted while its ancestor was included. No target path or expected finding is special-cased in the implementation.

Next `screen-focused-near-context-fast-fixed` tests that context-priority repair as the sole changed component from the four-pass control. Same24primary files,11groups/44calls,40KBsupport limit,120KBhard bound,existing check content,audit,schema,Grok4.7-build-fast,sharedloweffort and jobs16. Bound480s/8m and Grok<=3x fresh matched Opus. The hypothesis is fewer unsupported assumptions and less wasted reasoning when the actual receiver and callers outrank distant inherited code. Focused retrieval/audit tests pass. No production promotion.

Near-context result: Opus194.10s/71findings; Grok393.01s/114findings (2.025x),44calls each. Grok covers13/14required roots, missing the gap across all three weekend presets; reporting the separate ignored-date query bug does not substitute. False claims remain, including attr_reader failing to read its instance variable, Rails before_action halting on false, and unreachable fractional Kaiser adjustments. No quality PASS or full run.

Next `screen-focused-group-verified-fast-fixed` changes one stage from the near-context four-pass control: a same-model final verification per source group sees all four provisional passes together, checks claims and contradictions, and checks omissions. The earlier per-draft verifier could not compare independent passes and explicitly prohibited adding missing issues. This targets measured false claims and the missing condition-set gap, using only generic source-review procedures; no reference answers or other-model findings are supplied. Both models use the same procedure. Same24paths,11groups,44discovery+11final calls,source/context,model4.7-build-fast,loweffort,jobs16. Source-plus-drafts hard bound240KB; invalid or missing audit fails the run. Raw parser-skipped draft claims remain available to verification. Declared600s/10m safety bound, always capped at3x fresh matched Opus; all55calls and both stages count. Focused tests passed for both adapters, preservation of all raw drafts/source, supported retention, removal, omission recovery, invalid citations, empty drafts, malformed final audits and bounded inputs. No production promotion or full run.

Group-verification result: Opus260.30s/36findings; Grok506.13s/58findings (1.94x),55calls each. Discovery prompt hashes match the previous near-context control. Grok covers13/14required roots: the weekend gap was already present in discovery and survives, but verification removes the real future-day-midnight defect. False language/API claims survive, including locale-sensitive default Swift uppercasing and successful iOS save despite errors in an HTTP200 body. Reject on recall; no quality PASS or full run. Opus final claims have33confirmed,1false and2unresolved UI-lifecycle claims; all unresolved claims still block a final quality verdict. An actual-source JavaScript fixture confirms the rejected response-body read escapes the API helpers and leaves calling UI state stuck.

Next `screen-focused-wide-verified-fast-fixed` changes preferred primary batch size20KB->80KB from group-verification control. The current55-call sample is too costly to scale comfortably, and repeated passes still miss guards spread across source groups. Larger batches now retain per-file audits, corrected source-reference context and group verification, which the earlier wide fast-model trial lacked. This tests fewer calls plus more co-located source, with no new check or stage. All24frozen primary files appear exactly once in3source groups:12discovery+3verification calls instead of55. Same120KBdiscovery hard bound,40KBsupport cap,240KBsource-plus-drafts bound,Grok4.7-build-fast,sharedloweffort,jobs16 and600s/10m safety bound; Grok is always capped at3x fresh matched Opus. Context/audit/verification tests and exact sample coverage/bounds checks pass. No production promotion or full run.

Wide-batch verification run: Opus failed at82.20s when the coverage validator found omitted primary-file audits; cleanup succeeded and Grok never started. The affected call returned only the first file audit and omitted five required files. No partial review is accepted. Next repair changes only audit representation from a free array to an object with each actual primary path required in the shared JSON schema. This is the same coverage procedure and source; the schema now enforces the expected membership rather than merely checking it after a model can forget entries.

Next `screen-focused-wide-map-verified-fast-fixed` tests the required-path audit object repair. Same24primary paths,3source groups,12discovery+3verification calls,80KBpreferred/120KBhard primary prompts,40KBsupport,240KBverification bound,model4.7-build-fast,sharedloweffort,jobs16 and600s/10m screen cap, further limited to3x fresh matched Opus. Both adapters receive the same source-derived required keys; findings/page contract remains unchanged. Missing, extra, wrong or empty audits still fail closed; restarted complete JSON uses the existing parser. Focused audit/coverage/verification tests pass. No retry of the failed result, no partial publication, and no production promotion.

Required-audit wide result: Opus111.90s/23findings; Grok timed out at335.79s against335.69s (3x matched), after all12discovery calls but only1of3verification calls completed. Cleanup succeeded. Exact audit keys fixed the coverage failure; the large group's236second discovery call plus verification still misses the elapsed bound. No accepted Grok quality score or full run.

Source adjudication of the earlier linked-context14/14 result now establishes at least18unique false Grok roots, versus9proven false Opus roots. The Opus inventory covers all52published findings, with two SwiftUI lifecycle roots and one global Honeybadger-reporting component still unresolved; Grok's remaining claims are not treated as true. Several errors directly trace to omitted actual callers/guards, while others misstate language/runtime behavior. No quality PASS is claimed from its successful recall/time score. Ledgers and source evidence are retained.

Prepared40KBmidpoint did not reach a model benchmark: its context preflight failed because named domain types crowded out the explicitly rendered shared table_search partial, despite layout and dashboard being primary files. Stop before measuring a known context regression. Explicit file/partial references currently have the same weight as inferred named-type links; the planned repair gives exact source references a bounded2x preference, still decaying with graph distance. To keep measured factors separate, test that priority change first against the existing80KBcontrol, then evaluate a smaller primary preference if needed.

Exact-reference regression fails under the old picker (inferred type displaces the explicitly rendered helper) and passes with the2x exact-reference preference. The formerly failing40KBpinned-source preflight now passes, with every primary file once and the real shared table_search helper present. Existing context/audit/verification checks also pass.

Next `screen-focused-wide-import-map-verified-fast-fixed` measures only that exact-reference priority repair against the80KBrequired-map control. Same24paths,3groups/15calls,allfourdiscoverychecks,same-modelverification,40KBsupport limit,120KBdiscovery/240KBverification bounds,Grok4.7-build-fast,sharedloweffort,jobs16,600s/10m cap further limited to3x fresh matched Opus. The40KBpacking change remains prepared but unmeasured, so its effects are not conflated with this retrieval repair. No production promotion or full run.

The linked-context14/14 candidate now has a decisive false-positive rejection: the matched Opus inventory has9proven false roots and only3unresolved unique components, so its worst-case count is12. Grok already has18independently proven false roots. Resolving the remaining claims cannot make that candidate meet the no-more-false-positives criterion. This failure bound does not label unreviewed Grok claims true or pretend finalist adjudication is complete.

Exact-reference80KB result: Opus123.1315s/23findings; Grok313.9569s/26findings (2.55x), all15calls complete. Grok recovers8/14required roots: CSRF, scheduler MFA, upload-after-save, both truncated subtotals, future midnight, browser ENV and inactive alias. It misses weekend presets, extended-regex spaces, iOS Cancel and all three repeated-preview subtotals. The future finding mixes a valid midnight predicate with an incorrect all-interactions consequence; multiple other false claims remain. Reject on recall; no full run.

Next `screen-focused-medium-batch-map-verified-fast-fixed` changes only preferred primary batch size80KB->40KB from that completed control. Same exact-reference priority,40KBsupport,120KBdiscovery/240KBverification bounds,four discovery passes and one combined verification per group,required per-file audit map,Grok4.7-build-fast,sharedloweffort,jobs16. The fixed24paths form6groups/30calls; preflight checks all primaries once and the actual shared table-search helper. Rebuilt binary after the priority repair. Declared600s/10m safety bound, Grok additionally capped at3x freshly matched Opus. Smaller groups test the observed loss of findings in the large-batch control.

The other earlier14/14 candidate (`screen-focused-all-context-fast-jobs16-fixed`,2.44x) also now fails the false-positive criterion: at least17proven unique false Grok roots versus8proven plus6unresolved Opus roots (worst-case14). All48Opus raw claims match its published count, with no unparsed calls or skipped findings. The remaining Grok claims stay unresolved; this is a decisive failure bound, not a completed positive adjudication. Both earlier14/14speed successes are now rejected on quality.

40KBresult: Opus177.0164s/30findings; Grok319.0852s/48findings (1.80x),30calls each with identical discovery prompts. Grok recovers12/14required roots, missing weekend presets and extended-regex spaces. The weekend defect was present in a discovery draft and dropped in final verification; regex was not found in discovery. Invoice preview order and iOS Cancel recovered. False claims remain, including hypothetical clients treating HTTP200 as unconditional success and DateFormatter thread-unsafety. Reject on recall; no full run.

Next `screen-focused-medium-verify-effort-fast-fixed` changes only final-verification effort low->medium, identically for both providers. Discovery stays low. Same40KBprimary batches,40KBsupport,four passes,required audit map and6combined verifications (30calls),Grok4.7-build-fast,jobs16,24frozenpaths. This tests the measured loss of a real draft and retained false semantics; final verification already checks omissions, without any benchmark answers supplied. Focused context/audit/verification tests plus a stage-to-CLI/environment effort agreement check pass. Declared600s/10m safety bound, Grok additionally capped at3x fresh matched Opus. No production default changes.

Further source checking corrects one claim in the older all-context inventories: the Sutter empty preview fails in its controller before entering the template; a claimed empty-template failure is not reachable through that path or through the guarded production batch. The frozen controller-level empty-preview root remains real and unchanged. Conservative all-context false-positive bounds are now Grok>=18, Opus9proven+6unresolved<=15, still a rejection. A standalone macOS Foundation probe also confirms default DateFormatter.isLenient=false and rejects2024-02-31; explicit lenient=true accepts it. This corroborates the official Foundation implementation and refutes the default-leniency component in the recent Grok report; no iOS simulator was used.

Medium-verification result: Opus115.1491s/23findings; Grok timed out at345.5452s against345.4474s (3x matched Opus), with clean cancellation. All24discovery calls completed, but only1of6verification calls finished (98seconds); the longest discovery took224seconds. No accepted complete Grok quality score. Captured arguments confirm low discovery and medium verification for both models.

Next `screen-focused-medium-verify-effort-46-fixed` changes only the Grok variant to4.6. Earlier4.6low discovery was faster than4.7-build-fast, but was not tested with this exact caller context, four-pass split, required audit map or combined medium-effort verification. Same24paths,6groups/30calls,40KBprimary/40KBsupport budgets,sharedlow-discovery/medium-verification effort,jobs16,600s/10m safety and <=3x fresh matched Opus. The previous timeout stays a failure; this is not a repeat of the same candidate.

Grok4.6medium-verification result: Opus110.0868s/26findings; Grok timed out at330.3753s against330.2603s (3x matched Opus), clean cancellation. All24discovery calls completed (longest99seconds), and5of6verification calls finished at147–220seconds, leaving the auth group incomplete. No complete quality score. The provider variant improved discovery time but did not meet the end-to-end ceiling.

Prepared pipelined scheduling passes its race-enabled context/audit/verification and scheduling tests. A blocking regression demonstrates the old stage barrier cannot verify a finished group while another discovery is waiting; the new scheduler succeeds, never exceeds the shared job limit, retains draft/output order, and cancels unfinished work after failures or timeouts. This changes scheduling only; source prompts, schema, stage effort, group draft membership, model calls and final parser stay unchanged.

Next `screen-focused-pipelined-verify-fast-fixed` measures that scheduler against the failed `screen-focused-medium-verify-effort-fast-fixed` control, returning toGrok4.7-build-fast. Ready verification work uses free shared slots before more discovery; there is no extra concurrency or added stage. Same24paths,6groups/30calls,40KBprimary/40KBsupport,low discovery and medium verification,jobs16,600s/10m safety and <=3x fresh matched Opus. The separate4.6trial remains recorded. This tests the observed idle slots before the all-discovery barrier released; no quality improvement is assumed from scheduling alone.

The initial pipelined benchmark launch was not executed because the automatic approval service returned a401authentication error. On the next goal continuation, the normal approval path recovered and the same declared benchmark started. No sandbox bypass, credential change or alternate model access was used.

Pipelined result: Opus114.0332s/25findings; Grok timed out at342.1903s against342.0995s (3x matched), with clean cancellation. All24discovery calls completed;2of6medium verifications completed,4were cancelled. Discovery prompt hashes match the barrier control. Verification began32.21seconds before the last Opus discovery ended, and108.86seconds before the last Grok discovery ended: overlap is real, but individual medium-effort calls still exceed the available end-to-end time. No complete Grok quality score and no production promotion.

Next `screen-focused-decision-verified-fast-fixed` returns to the completed low-effort40KBcontrol (12/14roots,319.09seconds) and changes only the final-verification response contract. Every own-model draft finding gets a required decision ID; the model must retain, reject or deduplicate it with a source-based rationale. Retained/deduplicated claims must reference an actual final finding index; missing decisions, missing rationale and invalid indices fail the run. Discovery prompts, source/context, four passes, per-file audit, six combined verification calls, modelGrok4.7-build-fast, low effort throughout, jobs16 and30total calls stay unchanged. This addresses observed silent loss of the true weekend draft without adding a stage or injecting reference answers. The old validator accepts a final answer that accounts for none of its candidates; a regression now rejects it. Focused decision/schema/context/verification tests pass. Declared600s/10m safety and <=3x fresh matched Opus. The pipelined scheduler is kept separate from this quality test.

Investigated Grok CLI shutdown overhead after finding its documented optional trace-upload drain. Existing captured stdout files are written directly by the live CLI, so their final modification time can be compared with capture start+elapsed. Across the completed low-effort40KBcontrol, pipelined run's completed calls, and current decision run's completed discovery, the median exit tail is about2.6seconds; observed largest tails are7.5,15.1and5.6seconds. The cancelled long calls had emitted no JSON. Thus a150second exit-drain budget is not the main observed delay. No telemetry or global settings were changed; diagnostic measurements are retained.


Explicit-decision result: Opus147.8942s/27findings; Grok406.9520s/46findings (2.75x), all30calls complete. Grok covers10/14required roots, missing future midnight, weekend presets, extended-regex spaces and Sutter repeated preview subtotals. Required decision coverage prevents silent disappearance but does not make a rejection correct: Grok wrongly assumes the shared keyup handler always supersedes the initial date filter, and rejects the Sutter order defect because the actual preview controller is absent from its context. Multiple false claims remain. Reject on recall; no full model run or promotion.

A model-free whole-repository preflight exercised the current decision prototype on all1346selected files, forming149source groups and745calls. Every primary source byte was verified, including reconstruction of the sliced SQL structure file; no model calls were made. The initial reconstruction check incorrectly used concurrent runner-start order and was corrected to use deterministic prompt-emission order. The repaired test passes. This establishes source coverage, not quality or acceptable full-run speed.


Next `screen-focused-view-action-context-fast-fixed` returns to the completed40KB low-effort control (12/14roots,319.09seconds) and changes only supporting-source links for Rails views. An actual controller/concern definition of the view action now outranks incidental mentions of its filename. No target paths, source literals or expected findings are special-cased. A synthetic tight-budget regression fails before the change and passes after it; pinned-source preflight confirms the previously omitted Sutter preview controller is now present, all24primary files remain covered and existing context/audit/verification tests pass. Same6groups/30calls,four discovery passes,one group verification,required audit map,Grok4.7-build-fast,sharedloweffort,jobs16,40KBprimary/40KBsupport,120KBdiscovery/240KBverification bounds. Declared600s/10m safety and Grok<=3x fresh matched Opus. This trial does not include the additional per-candidate decision contract or pipelined scheduler.


While the view-action trial runs, source inspection identifies another concrete context limit: its invoice-model group omits SutterHealthRateGuard (which performs the cleanup) and its Kaiser-template group omits the integer-converting BatchBuilder. Matched Opus consequently repeats both false claims. Prepared a separate supporting-context capacity experiment40KB->80KB, retaining the120KBtotal discovery bound,40KBprimary preference and same retrieval rules. A synthetic two-direct-dependency fixture fails at40KB and passes at80KB; pinned-source preflight confirms the Sutter cleanup guard is included and all existing context/audit/verification tests pass. Not yet benchmarked or promoted; the current view-action trial remains unchanged.


Provider capability check: the logged-in Grok model cache still offers4.7,4.7-build-fast,4.6and4.5. Current [official reasoning documentation](https://docs.x.ai/developers/model-capabilities/text/reasoning) confirms reasoning cannot be disabled for these model families; low is the minimum. No unsupported effort value, authentication method or global configuration was changed. Current measured long calls generate tens of thousands of reasoning tokens even at low effort.


View-action context result: Opus140.3275s/28findings; Grok timed out at421.0811s against420.9826s (3x matched Opus), with clean cancellation. Completed calls: {'discovery': 24, 'verification': 2}. The longest discovery took354.29seconds and generated30,788reasoning tokens on the13-file auth/invoice-model group. No completed Grok quality score or full run. The matched Opus inventory contains2proven false roots plus2unresolved components (upper bound4); all28raw final claims match the published count.

Next `screen-focused-support-budget-fast-fixed` measures the prepared supporting-context capacity change against the view-action control:40KB->80KB, still bounded by120KBtotal discovery and240KBfinal input. Primary preference40KB,6groups/30calls,four discovery passes,one verification,required audit map,retrieval rules,Grok4.7-build-fast,sharedloweffort,jobs16 and24frozenpaths stay unchanged. The actual missing-rate cleanup guard is now present; no expected bug answers are added. The Kaiser builder remains absent from its template group because it is not reached by the current graph, so this is not claimed as a complete context repair. Focused tests pass. Declared600s/10m safety, Grok capped at3x fresh matched Opus.


Prepared packing hardening separately from the live benchmark: source classification and bin admission now account for the required primary-target header. An empty file with an impossibly small header budget returns an error rather than silently disappearing. Both boundary regressions fail before the repair and pass afterward, with context/audit/verification and UTF-8 slicing checks. Model-free complete Run preflights on both versions verify all1346source files,149groups and745calls. Every generated prompt hash is identical on the pinned whole repository, proving the boundary repair does not alter this benchmark input. No model calls or production promotion were involved.


Independent nil-association adjudication now confirms the older matched Opus Kaiser/Sutter deleted-department claims: root-admin deletion permits the soft deletion, assignment preview queries retain those assignments, and the actual templates crash at Kaiser line5/Sutter line214 when the scoped association is nil. The Nana address case is analogous via its unguarded root-admin address deletion and was reproduced at template line65. A fixture also reproduces a missing Kaiser routing-code crash, but that fixture alone does not settle its upstream reachability, so that older claim stays unresolved. Resolving the two department claims lowers the older all-context Opus false-positive upper bound from15to13 (9proven+4unresolved), strengthening its existing rejection against Grok>=18; it does not change the frozen required roots.


Reference audit found an adjudication error in B1-245 (ignored invoice-stamp validation failure): the prior deleted-department trigger assumed required associations are always revalidated. hh uses Rails8.1 defaults, which disable that check for unchanged nonnil foreign keys. Executing the actual ActiveRecord8.1.3.1 validator condition confirms unchanged=false, changed=true, nil=true. B1-245 is now unresolved while an alternative reachable validation-failure trigger is investigated; its required root is retained provisionally, not dropped to let a candidate pass. The frozen set remains83required roots (82confirmed+1provisional), with no change to the14screen roots. Full quality PASS remains blocked by this unresolved reference claim. Evidence and the superseded reason are preserved.


80KBsupport result: Opus180.2242s/25findings; Grok420.7579s/40findings (2.3346x), all30calls complete. Grok recovers12/14required roots, now including extended-regex spaces and all three repeated preview subtotals. It misses both future-midnight and weekend filtering; midnight was in discovery but dropped in verification, while the weekend root was not discovered. Known false claims include Rails before_action halting on false, hypothetical status-only clients, invoice CSRF-meta leakage and successful blank-bank submissions. Reject on recall; no full run or production promotion. Matched Opus has1proven false root plus4unresolved components after correcting an erroneous assumption that web pending-check Cancel behaves like Swift (the web version navigates away). All25raw Opus final claims match the published count.


Next `screen-focused-runtime-notes-fast-fixed` changes only the shared system prompt from the completed80KBsupport control (12/14,420.76seconds). It adds2,599bytes of general Ruby, Rails8.1, JavaScript/jQuery3 and Swift/Foundation reference semantics, checked against primary language/framework documentation and local runtime probes. These contain no hh paths, business names, expected findings or model-produced answers. Same note goes to both providers at discovery and verification. This addresses observed false framework facts and incorrect initialization-order rejection. Source prompts, grouping,context,required audit schema,6groups/30calls,four discovery passes,one verification,Grok4.7-build-fast,sharedloweffort,jobs16 stay unchanged. Focused context/audit/verification tests pass; only model.go differs from the control. Declared600s/10m safety and Grok<=3x fresh matched Opus. The independent packing-boundary repair remains separate. References include [Ruby regex syntax](https://docs.ruby-lang.org/en/3.4/Regexp.html), [Rails callback termination](https://raw.githubusercontent.com/rails/rails/v8.1.3.1/actionpack/lib/abstract_controller/callbacks.rb), [jQuery3 ready scheduling](https://jquery.com/upgrade-guide/3.0/), [ECMAScript Date](https://tc39.es/ecma262/multipage/numbers-and-dates.html), and [Foundation DateFormatter](https://developer.apple.com/documentation/foundation/dateformatter).


Runtime-reference result: Opus152.2060s/31findings; Grok447.0467s/42findings (2.9371x), all30calls complete with clean exit and pinned source unchanged. Grok covers13/14required roots: future midnight now survives verification, but weekend preset coverage is absent from all discovery passes and final output. False/mixed claims remain. Reject on recall; no full run or production promotion.

Prepared next predicate-coverage trial changes only the values/language discovery instruction: enumerate small finite input domains and boundary representatives, compare related predicates together, and trace gaps/overlaps through actual callers to observable consequences. Both models currently note that the date offsets match the server query but do not check whether their combined visible subsets cover that query. The new instruction is generic and contains no repository paths, business facts, weekdays or expected findings. Runtime notes and all other settings remain the same.


Predicate preflight initially detected a context change because the longer instruction also reduced the packer budget. Kept the existing reservation (which already exceeds every new per-pass prefix) and added the new text only after calculating it. The direct pack comparison passes. The first full model-free Run comparison incorrectly read Grok stdin instead of its prompt file and failed; that fixture is being corrected before any model benchmark. Existing context/audit/verification tests pass.

Next `screen-focused-predicate-coverage-fast-fixed` measures that sole discovery-instruction change. Same24paths,6groups/30calls,four passes,one same-model verification,required audit map,40KBprimary/80KBsupport,shared runtime notes,Grok4.7-build-fast,low effort throughout,jobs16. Declared600s/10m safety bound, Grok capped at3x fresh matched Opus. No production promotion.

Corrected model-free Run comparison passes:24discovery calls match the captured control byte-for-byte except six intended values-pass additions;6final calls and prompt bounds pass. Build exit0 (benign module stat-cache warning). The new benchmark is now ready to start.


Resolved the global Honeybadger component in three older Opus ledgers: the local API catch is bypassed, but the absolute never-reported claim is false. The actual web entry configures Honeybadger with default unhandled-rejection capture; its installed6.16.3browser plugin reports the rejection. A no-network fixture uses the actual API source and installed browser bundle, dispatches the actual body-read error to the handler, and observes a notice with transport suppressed. It explicitly does not emulate native browser event scheduling. The underlying API/UI failure remains real. Moving one unique component from unresolved to false leaves each previous worst-case bound unchanged. [Honeybadger configuration](https://docs.honeybadger.io/lib/javascript/reference/configuration/) corroborates the default.


Reference B1-245 is reconfirmed with a different reachable trigger, preserving all83required roots. Scheduler creation uses Assignment::Create, where visit_type is optional, and its accepted columns include billing status. A valid UCSF assignment with no visit_type can therefore persist and be selected as approved/unbilled. UcsfBatch reloads that row as Assignment::Ucsf, whose stricter visit_type validators fail during stamping. An executable fixture runs the pinned scalar validation declarations, actual ActiveRecord8.1.3.1 update/save validation, and actual HH stamping/record_run methods: update fails, ClientInvoice has already been recorded, and record_run returns normally; a valid-visit-type control stamps. Persistence/transactions are in-memory, with reachability traced separately through actual controller/query source. The earlier deleted-association trigger stays recorded as disproved. No required reference root was removed or softened.


Predicate-coverage benchmark completed: Opus175.8631s/28findings; Grok410.2788s/41findings (2.3329x), all calls complete, clean exit and pinned source unchanged. Quality adjudication has not yet been completed; no PASS or full run claimed. User requested a candid status checkpoint after the long experiment loop. No benchmark is currently running.


## September 29 bounded autonomous campaign

The campaign began at 13:59:36 UTC with an absolute experiment deadline of
15:39:36 UTC and total deadline of 15:59:36 UTC. The existing private handoff
records these limits and counters across continuations. The existing run harness
now refuses an unregistered candidate, missing state, an expired deadline, an
exhausted attempt limit or a stopped campaign before invoking a provider. Its
process-group timeout is capped by the remaining experiment budget. Seven guard
cases passed without model calls.

Scored the previously completed `screen-focused-predicate-coverage-fast-fixed`:
Opus 175.8631s, Grok 410.2788s (2.3329x), **11/14 required roots**. The final
findings omit the weekend preset gap, extended-regex whitespace and inactive-alias
nil dereference. The department/town nil lookup is a different path and does not
cover the alias defect. Reused the existing source-backed reference adjudications;
no reference acceptance was changed. Reject on recall without completing an
unnecessary false-positive inventory. Its scorecard remains private.

Candidate 1, `screen-focused-findings-only-fast-fixed`, changes only the final
verification payload: retain every raw draft finding and source line, but discard
the draft audit narratives. This tests false assurances in those narratives being
repeated by verification. A regression reproduced that forwarding before the
change; isolation, context, audit and verification preflights pass after it. All
24 discovery prompts match the prior control byte for byte. Opus finished in
121.6276s with 25 findings. Grok hit its 364.8829s deadline and stopped in 364.9931s;
all 24 discovery calls completed, six verification calls started and none completed.
Cleanup succeeded. No final Grok review exists, so no quality score is accepted.
This is the first non-improvement; nothing is promoted.

Candidate 2, `screen-focused-findings-only-46-fixed`, changes only the Grok model
from 4.7-build-fast to 4.6. Earlier evidence showed faster 4.6 discovery, but its
medium-verification experiment timed out and lacked the current runtime notes.
Both providers retain the same source, context, prompts, audit schema, tools,
low effort, concurrency and review stages. Preflights pass. Opus finished in
118.6537s with 26 findings; Grok finished in 257.1117s with 48 findings (2.1669x).
All 30 calls completed for each model; source remained pinned and cleanup succeeded.
The final Grok findings cover **11/14 required roots**, missing the same three as
the existing predicate result. Reject on recall; false and mixed claims remain
unresolved and no false-positive total is asserted.

At 14:22:41 UTC the two consecutive non-improvements ended the campaign. The
persisted state is `STOPPED_NO_IMPROVEMENT`; the harness refuses another launch.
No third candidate or full-repository model run was launched, and no experimental
speed change entered product code. Raw results, prototypes, scorecards and the
harness stay in the existing private artifact directory. The whole-repository
quality/speed goal remains unfinished.

## September 29 user-authorized accuracy diagnostics

After the prior campaign stopped, the user explicitly authorized testing isolated
verification and smaller discovery contexts. A separate bounded diagnostic record
preserves the old campaign; no acceptance criterion changed. Product code stayed
unchanged and all private inputs and prototypes stayed outside the repository.

`isolated-claim-diagnostic` replayed each provider's own saved reliability findings
from the problematic source group: five Opus and seven Grok candidates, each judged
separately against the original source context. Both used the same system prompt,
decision schema, low effort, disabled tools and eight-worker limit. Code assembles
accepted findings; an unresolved decision prevents a successful report. All twelve
calls completed. Opus took 47.7650s; Grok took 45.1729s. Grok retained the verified
inactive-alias crash that its combined verifier had discarded, but also retained
questionable claims and returned one unresolved decision about nil billed amounts.
Its report therefore failed. No precision or complete-pipeline improvement is claimed.

`small-discovery-diagnostic` used the existing values-focused instructions and source
dependency selection, one target at a time, with a 48,000-byte prompt budget.
Preflight rejected a 32,000-byte version that omitted the dashboard query dependency;
no model call used it. Both models received identical final prompts (47,849 and
47,905 bytes), system prompts, schemas and settings. Opus completed both calls in
17.4234s. Grok completed the dashboard call but missed the weekend gap and claimed
that no text-filter keyup handler exists. That claim is contradicted by the layout's
included `_table_search` partial, which the reduced context omitted. The ban-parser
call timed out: deadline 52.2702s, elapsed including cleanup 52.3753s. Its regex
detection is unscored because no completed response exists. Descendant cleanup
succeeded and the pinned `hh` HEAD was unchanged.

Both diagnostics are rejected for promotion. Their timings cover isolated stages,
not source-review end to end. No new 24-file or full-repository pair was launched.
The handoff records two non-improvements and refuses further launches; the limited
retention signal is saved without claiming parity or weakening accuracy requirements.

## September 29 evidence-obligation and runtime-fact diagnostics

The user authorized another attempt. It began at 20:22:39 UTC with an experiment
deadline of 21:02:39 UTC and a closure deadline of 21:22:39 UTC. Previous campaigns
remain preserved. These are bounded replay/discovery diagnostics, not fresh complete
review pairs; each provider stage had a 300-second safety cap. Whole-review quality
and timing acceptance remain unchanged.

`proof-obligations-diagnostic` replayed each model's own linked-context drafts for
the authentication, dashboard/layout and ban-parser groups. Each claim required
reachability, behavior and counterevidence before its verdict; accepted findings
were collected directly rather than rewritten. Source contexts were identical
between providers. Both completed three calls, Opus in 54.9398s and Grok in
126.5931s. Grok retained the regex and alias bugs, but rejected the real midnight
and weekend dashboard roots and retained the incorrect global-controller inheritance
claim. Reject on recall; the remaining eight groups were not run.

Reviewing that control exposed a scorecard error: finding G56 alleges Honeybadger
key exposure, not the actual browser `ENV` ReferenceError. It was already recorded
as false in the claim ledger but incorrectly counted toward recall. Correct the
linked-context result to 13/14, retaining the previous scorecard and correction
evidence. The other historical all-context candidate explicitly reports the `ENV`
failure in finding53 and still covers14/14 with mixed false components; it remains
rejected on its false-positive bound. No source or required reference root changed.

`literal-facts-diagnostic` then changed only the existing Grok4.6 values-discovery
input: append static-regex match observations generated by Ruby4.0.0 and Prism,
matching the repository's Ruby version. The generator parses source without executing
application code; it tests literal word runs already in each pattern and their
space-removed forms. Both models receive identical actual match results, with no
expected outcomes or bug labels. This follows the parser's documented
[static regex node and flags](https://ruby.github.io/prism/rb/Prism/RegularExpressionNode.html)
and Ruby's [Regexp construction and matching](https://docs.ruby-lang.org/en/4.0/Regexp.html).
Eight synthetic literal-equivalence cases, interpolation exclusion and parse-error
rejection passed. The source contains34static patterns and5dynamic patterns, which
are explicitly skipped. The original117,315-byte prompt is unchanged; the augmented
prompt is119,921bytes, within the120,000-byte bound.

Opus completed in28.9169s and Grok in69.6861s (2.4099x), including0.1300s and0.1323s
of preprocessing respectively. Both completed normally with clean process cleanup
and the pinned source unchanged. Grok now reports the extended-regex whitespace
bug, increasing required-root coverage in this group from4/7 to5/7. The weekend
and inactive-alias roots remain absent. It omits the previously reported CPMC alias
precedence defect and adds a false component to the real subtotal bug: it says
10.9+0.2 becomes10 after truncating the accumulator, whereas executing the actual
BigDecimal operation gives10.2. All six published findings in the old Grok stage
and both new stages were inspected: the new Grok output has this one false component;
the old Grok stage and new Opus stage have none. Audit notes are not accepted findings.

This is a useful local recovery signal, not an accuracy pass. Two diagnostics failed
the no-regression requirement, so the batch stopped without expanding to the24-file
sample or whole repository. All prototypes remain private; no product change or
benchmark acceptance was promoted.

## September 29: five requested options on untuned primary files

The user explicitly authorized trying all five proposed options and keeping only
improvements. This batch started at 21:05:28 UTC, with an experiment deadline of
22:45:28 and closure deadline of 23:05:28. That authorization superseded the old
three-candidate/two-failure stopping limits for this batch. The source revision,
reference, quality requirements and whole-review speed requirements stayed fixed.
All requested trials finished before the deadline; no implementation qualified.

The fifth option, untuned-code evaluation, applied to every trial. Before new prompt
design, a deterministic hash selection chose two primary files each from controller,
model and frontend strata, excluding previous narrowed-review primary targets.
Each selected file has a verified important reference root. These six files are
untuned primary targets within the same repository, not a new unseen repository;
some may previously have appeared as supporting context. The hypotheses, selection
and initial scripts were frozen before reading their reference labels or outputs.
Labels never entered provider prompts.

The private harness used the logged-in Opus and Grok 4.6 CLIs, identical initial
source/prompts/system/schema, low effort and a shared 16-worker limit. Models had
no tools. Source-dependent follow-ups used each model's own findings and requests,
with the same policy and accessible committed source. Review stages, per-run tool checks,
startup and process cleanup were timed. Initial prompt packs were precomputed;
these timings do not establish full-command performance. Grok was capped at three times fresh Opus
elapsed time. The pinned source was checked before and after each provider; every
check passed and descendant cleanup succeeded. Actual discovery prompts and shared
settings were compared after execution. No full-repository run was launched.

| Configuration | Opus seconds | Grok seconds | Grok recall screen | Decision |
| --- | ---: | ---: | --- | --- |
| Fresh control | 26.1739 | 60.6847 | 3 confirmed, at most 4 of 6 | Reject: missing date-default and radio-selection roots |
| Source-linked grouping | 33.7866 | 68.6963 | 3 confirmed, at most 4 of 6 | Reject: missing timezone and radio-selection roots |
| One bounded lookup | 49.3406 | 97.7715 | 4 confirmed, at most 5 of 6 | Reject: missing timezone root |
| Standard tool evidence | 28.0442 | 84.2532 including cancellation | Incomplete; unscored | Reject: deadline 84.1326 exceeded |
| Selective medium reasoning | 56.8381, failed | 97.9614, diagnostic | 4 of 6 | Reject: Opus unresolved; Grok misses concurrent-draft and date-default roots |

For the first three Grok reports, possible additional recall credit for overlapping
draft includes was left unresolved: they describe duplicate billing without tracing
concurrency past the activation guard. Their other missing roots already establish
failure even with that credit. Precision was not exhaustively adjudicated after
these decisive failures. Each completed Grok report contains at least one verified
false component about PDF widget values: the pinned and installed pdfjs-dist 6.3.289
normalizes stored strings and uses boolean checkbox/radio widget state. The lookup
and selective-effort outputs also identify the genuine loss of radio-option identity;
they receive recall credit for it while retaining the false-component penalty.

Source-linked grouping separated the six unrelated primary targets, using existing
source/type/caller links and excluding high-fanout shared ancestors. This is an
approximation to execution-flow grouping, not a compiler-resolved call graph. The
existing supporting-context picker remained in use. Every primary was present once,
all primary source was preserved, and all six initial prompts fit within 120,000 bytes.
The control used two groups. Grok recovered the date-default root but lost the
system-timezone root; the false widget-value claim persisted. Reject the prototype.

Bounded lookup allowed at most three relative paths or symbols per group, four
matches per request and 40 KB of additional committed source, followed by one completion
call. Both providers returned empty request lists for both groups. Thus the code
path was offered but retrieval itself was not exercised; this run does not establish
whether useful requested context would help. Grok still missed the timezone root
and retained a false library premise. Reject this configuration without another run.

Standard evidence came from six syntax parses, the repository's RuboCop lint rules
on four Ruby files, and ten existing companion JavaScript tests. All passed. The
isolated snapshot used committed source and privately cloned installed dependencies,
so compiler/test caches did not write into `hh`. Preflight repairs routed RuboCop
JSON to its output file and fixed Vitest's external setup-path problem; no custom
bug probes or tests were introduced. Both providers received byte-identical evidence,
including existing test names and outcomes. Grok timed out before a complete report;
partial responses were not scored as a successful review. Reject the prototype.

Selective effort added a low-effort candidate checker, escalating at most two
rejected or uncertain claims per group to medium reasoning. Opus made two medium
calls: one narrowed a date-formatting claim, while the catalog-association crash
remained uncertain, correctly preventing a successful report. Grok's arm was then
run once diagnostically, bounded to 170.5142 seconds (three times that failed Opus
elapsed time), to finish the requested option. Grok retained every candidate at low
effort and therefore made zero medium calls. It missed two required roots and kept
a false PDF-library premise. This is not evidence of medium-effort performance on
Grok, nor a passing matched comparison. Reject the escalation policy as tested.

All five requested options are now attempted, the launch guard is stopped, and
no new product code, full-sample run or full-repository run is being promoted.
This records failures of these prototypes, not proof that the ideas cannot work.
Private manifests, scripts, prompts, outputs, timings and claim mappings remain
under the existing private benchmark directory; only this results checkpoint and
its required review artifacts are committed.
