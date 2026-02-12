# proof-first-deploy-gcp

Repo B scaffold: deterministic deploy evidence (**render + verify**) with fixtures/goldens.

Quickstart:
- `make verify`
- `go run ./cmd/pfdeploy demo --out ./out`

Artifacts (render):
- deploy_manifest.json
- trigger_manifest.json
- iam_manifest.json
- manifest.sha256

Artifacts (verify):
- verify_report.json OR error.txt
