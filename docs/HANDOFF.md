# Handoff (Repo B)

This repo exists to produce a small, deterministic “deploy evidence” bundle you can hand to:

- a client
- a reviewer
- yourself, six months later

The output is designed to be **diffable**, **hashable**, and **stable**.

## What to hand off

### Render evidence (the core bundle)

From `pfdeploy render ...`:

- `deploy_manifest.json`
- `trigger_manifest.json`
- `iam_manifest.json`
- `manifest.sha256`

These files are intentionally small so they can live inside:
- a PR review
- an audit pack
- a casefile kit (Repo C)

### Verify evidence (optional)

From `pfdeploy verify ...`:

- `verify_report.json` (match)
- OR `error.txt` (mismatch)

This is the “trust check” that the deployed snapshot still matches the intended config.

## Suggested client-facing bundle

```
deploy-evidence/
  config.yaml
  deploy_manifest.json
  trigger_manifest.json
  iam_manifest.json
  manifest.sha256
  verify_report.json   # optional
```

If verification fails, include `error.txt` instead of `verify_report.json`.

## How this connects to Book 2

- `finance-pipeline-gcp` (anchor) shows the runtime workflow (Eventarc → Cloud Run) + deterministic artifacts + replay safety.
- `proof-first-event-contracts` defines what events to accept/ignore and produces deterministic decisions.
- `proof-first-deploy-gcp` (this repo) produces deterministic **deployment intent** and a small verification step.
- `proof-first-casefiles` packages realistic “engagement kits” that can include deploy evidence + runtime packs.

This is how Book 2 avoids “toy code”: the outputs are shaped to survive real review and real handoff.
