# Handoff

In a real deployment engagement, this repo’s artifacts are the “deploy evidence” bundle:

Render outputs:
- `deploy_manifest.json`
- `trigger_manifest.json`
- `iam_manifest.json`
- `manifest.sha256`

Verify output (optional):
- `verify_report.json` (or `error.txt` if mismatch / expected-fail)

These are meant to be diffable, hashable, and attachable to an audit pack or client handoff.
