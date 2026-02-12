# Contract (Repo B)

Render outputs (deterministic):
- deploy_manifest.json
- trigger_manifest.json
- iam_manifest.json
- manifest.sha256 (sha256 over the three JSON files)

Verify outputs:
- verify_report.json (match) OR error.txt (expected-fail)

No-secrets rule:
If any env key contains SECRET/TOKEN/PASSWORD/KEY (case-insensitive),
the case is expected-fail and must emit only error.txt.
