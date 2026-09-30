# metareview context: .metareview/reports/full-comparison-20260930T133301Z/comparison.json

Run ID: `mrv-20260930-134049972166000-artifact-comparison-ae9780dc`

## Target

- Path: `.metareview/reports/full-comparison-20260930T133301Z/comparison.json`
- Repository mode: `metaswarm-extension`
- Git branch: `whole-repo-source-review`
- Git head: `57c824b`

## Artifact Reviewed

Final artifact SHA256: 1274c65bbb498e3bf8bacd7c7231a4833aed52608922a6100ef03e9224d7d9d3

The complete comparison metrics follow. Source excerpts, prompts, full findings, and HTML reports stay in the ignored local report directory.

```json
{
  "status": "complete",
  "source": {
    "repository": "hh",
    "commit": "7d257e0f8efcfcfdd5d31ef71346a3fa48cf192b",
    "selected_files": 1346,
    "source_bytes": 4616754,
    "prompts": 45,
    "identical_prompt_hashes": true,
    "coverage": "Every selected source byte and original line number checked against pinned git blobs."
  },
  "product_commit": "e936e76e701d708ad8e4f99d4ee6afcf68aa7cc2",
  "settings": {
    "jobs": 8,
    "call_timeout": "30m",
    "effort": "medium",
    "tools": false,
    "shared_system_prompt": true,
    "shared_schema": true,
    "path_filter": null
  },
  "opus": {
    "status": "standard command completed",
    "elapsed_seconds": 297.916461834684,
    "findings": 222,
    "report": "opus/review.html",
    "findings_json": "opus/findings.json",
    "reference_roots_found": 58,
    "reference_root_count": 83,
    "precision": {
      "sample_size": 30,
      "counts": {
        "supported": 18,
        "mixed": 3,
        "unresolved": 2,
        "false": 7
      }
    }
  },
  "grok": {
    "status": "standard command failed; complete report recovered",
    "standard_failure": "Invalid JSON: trailing characters at line 1 column 14542 in prompt 30/45. 22/45 prompts had completed; no partial report was published.",
    "original_elapsed_seconds": 1801.6777763329446,
    "stopped_recovery_seconds": 84.63092816714197,
    "successful_recovery_seconds": 2523.0929506663233,
    "total_active_command_seconds": 4409.40165516641,
    "recovery_gaps_seconds": 339.2947338335898,
    "total_wall_seconds": 4748.696389,
    "wall_ratio_to_opus": 15.93969114615454,
    "active_ratio_to_opus": 14.80079894884499,
    "recovery": "Reused 22 exact successful assistant texts; ran the remaining 23 calls through the unchanged production pipeline and OSRunner. A stopped cache-matching attempt is included in the elapsed totals. Cached outer JSON envelopes were reconstructed; assistant texts were unchanged. No failed response was repaired.",
    "findings": 625,
    "report": "grok-recovered/review.html",
    "findings_json": "grok-recovered/findings.json",
    "reference_roots_found": 60,
    "reference_root_count": 83,
    "precision": {
      "sample_size": 30,
      "counts": {
        "unresolved": 2,
        "supported": 11,
        "false": 14,
        "mixed": 3
      }
    }
  },
  "reference_overlap": {
    "both": 47,
    "grok_only": 13,
    "opus_only": 11,
    "neither": 12
  },
  "accuracy_method": {
    "reference_panel": "83 previously source-confirmed important root causes at the same source revision, originating in an older Opus review. All published findings were examined for root/trigger/consequence matches.",
    "precision_sample": "30 published findings per model, including bugs and advisories, selected using a fixed SHA256 seed set before inspection. Two reviewers adjudicated the samples against source and callers.",
    "verdicts": {
      "supported": "Factual claim and trigger supported; does not certify severity.",
      "false": "False or unsupported current-impact claim.",
      "mixed": "Supported core plus a concrete false assertion.",
      "unresolved": "Not established; no supported credit."
    },
    "protocol": "accuracy-protocol.json",
    "reference": "reference.json",
    "evidence": [
      "opus/recall-review.json",
      "grok-recovered/recall-review.json",
      "opus/precision-reviewed.json",
      "grok-recovered/precision-reviewed.json"
    ]
  },
  "performance_details": "performance.json",
  "speed_requirement": {
    "maximum_ratio": 3,
    "maximum_seconds": 893.749385504052,
    "met": false
  },
  "conclusion": "Grok recovered two more known reference roots, but produced more false or unsupported claims in the checked samples and required 15.94 times the elapsed time to obtain a full report. The speed requirement is not met; overall accuracy parity is not demonstrated.",
  "limitations": [
    "One successful Opus run and one recovered Grok report provide no repeatability estimate.",
    "The standard Grok command failed. Its recovered report combines multiple attempts; this is not a successful single standard run.",
    "Grok wall time includes all attempts and manual recovery gaps. Active command time includes the stopped cache-matching attempt. Neither is a clean single-run model-latency estimate.",
    "The reference panel originated in Opus findings and is not an unbiased or exhaustive bug inventory.",
    "The precision samples contain 30 findings per model, including bugs and advisories; results do not establish whole-repository accuracy. Mixed and unresolved findings are shown separately.",
    "Supported source behavior does not establish calibrated severity or observed production impact."
  ]
}
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
