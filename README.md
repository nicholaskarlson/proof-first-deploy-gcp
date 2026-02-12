# proof-first-deploy-gcp

**Repo B (Book 2): deterministic deploy evidence** — render a small, diffable set of deployment manifests from a config file, and verify a real-world snapshot against that config.

This repo is intentionally **not** a cloud-cert guide and it does **not** run `gcloud` for you. Instead it focuses on the “proof-first” layer:

- outputs are deterministic (stable ordering + stable formatting)
- artifacts are small, reviewable, and hashable
- expected-fail cases emit only `error.txt` (also deterministic)

## Quickstart

```bash
make verify
go test -count=1 ./...

# proof gate demo (recomputes outputs + byte-compares to fixtures/expected)
go run ./cmd/pfdeploy demo --out ./out
```

## Commands

### Render (create deterministic deploy artifacts)

```bash
go run ./cmd/pfdeploy render --config ./config.yaml --out ./out/render
```

Outputs on success (always with LF + trailing newline):

- `deploy_manifest.json`
- `trigger_manifest.json`
- `iam_manifest.json`
- `manifest.sha256` (sha256 lines over the three JSON files)

On failure, emits **only**:

- `error.txt`

> Note: `--out` is cleared before writing, so reruns don’t leave stale files behind.

### Verify (compare config to a snapshot)

```bash
go run ./cmd/pfdeploy verify --config ./config.yaml --snapshot ./snapshot --out ./out/verify
```

A snapshot is a small, “captured” JSON file (fixture-friendly) representing what’s deployed.
For this MVP the snapshot file is:

- `snapshot/gcloud_service.json`

Outputs:

- `verify_report.json` (match) **or**
- `error.txt` (mismatch / bad snapshot)

## Artifacts

This repo deliberately keeps the evidence bundle tiny:

```
out/
  render/
    deploy_manifest.json
    trigger_manifest.json
    iam_manifest.json
    manifest.sha256
```

These artifacts are designed to be:
- attached to a client handoff
- checked into an audit pack
- diffed in PR review
- re-rendered later and compared byte-for-byte

See:
- `docs/CONTRACT.md` for the precise rules
- `docs/HANDOFF.md` for how this fits a real engagement

## Fixture layout (proof gate)

Each case is a folder name (the book references only these names).

Inputs:

- `fixtures/input/<case>/config.yaml`
- optional: `fixtures/input/<case>/gcloud_service.json` (if present, the case is a **verify** case)

Expected outputs:

- `fixtures/expected/<case>/` containing either:
  - the artifact set, or
  - `error.txt` only (expected-fail)

## Where this fits (Book 2)

Book 2 is a 4-repo suite:

- **Anchor:** `finance-pipeline-gcp` — drop-folder pipeline (Eventarc → Cloud Run) with deterministic artifacts + replay safety
- **Repo A:** `proof-first-event-contracts` — event parsing/decision contract with fixtures/goldens + expected-fail
- **Repo B (this repo):** `proof-first-deploy-gcp` — deterministic deploy evidence (render + verify)
- **Repo C:** `proof-first-casefiles` — realistic “engagement kits” (inputs + expected outputs + handoff)

The intent: readers learn how to build **workflow-grade guarantees**, not just deterministic tools.

## Go + CI witness

- Go baseline: **1.22.x**
- CI witness: ubuntu/macos/windows on 1.22.x, plus a “stable” job on ubuntu.

## License

MIT (see `LICENSE`).
