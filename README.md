# proof-first-deploy-gcp

**Repo B (Book 2): deterministic deploy evidence** — render a small, diffable set of deployment manifests from a config file, and verify a real-world snapshot against that config.

![ci](https://github.com/nicholaskarlson/proof-first-deploy-gcp/actions/workflows/ci.yml/badge.svg)
![license](https://img.shields.io/badge/license-MIT-blue.svg)

> **Book:** *Proof-First Pipelines in the Cloud* (Book 2)  
> This repo is **Repo B (Repo 3 of 4)**. The exact code referenced in the manuscript is tagged **[`book2-v1`](https://github.com/nicholaskarlson/proof-first-deploy-gcp/tree/book2-v1)**.

This repo is intentionally **not** a cloud-cert guide and it does **not** run `gcloud` for you. Instead it focuses on the “proof-first” layer:

- outputs are deterministic (stable ordering + stable formatting)
- artifacts are small, reviewable, and hashable
- expected-fail cases emit only `error.txt` (also deterministic)

**Go baseline:** 1.22.x (CI witnesses ubuntu/macos/windows on 1.22.x, plus ubuntu “stable”).

## Book 2 suite map

This repo is designed to be used alongside the other Book 2 repos:

- **[finance-pipeline-gcp](https://github.com/nicholaskarlson/finance-pipeline-gcp)** — anchor drop-folder workflow (trigger → run → artifacts → markers)
- **[proof-first-event-contracts](https://github.com/nicholaskarlson/proof-first-event-contracts)** — event parsing contract + fixtures/goldens + expected-fail
- **[proof-first-deploy-gcp](https://github.com/nicholaskarlson/proof-first-deploy-gcp)** — deterministic deploy evidence (render + verify) + fixtures/goldens
- **[proof-first-casefiles](https://github.com/nicholaskarlson/proof-first-casefiles)** — engagement kits you can hand to a client (or use in teaching)

## Quickstart

Run the proof gate:

```bash
make verify
# (optional) Equivalent, if you want to run it directly:
# go test -count=1 ./...
```

Run the deterministic fixture demo (recomputes outputs and diffs against `fixtures/expected/**`):

```bash
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

> Note: `--out` is cleared before writing (and refused if unsafe: `.`, `..`, `/`, or a Windows volume root), so reruns don’t leave stale files behind.

### Verify (compare config to a snapshot)

```bash
go run ./cmd/pfdeploy verify --config ./config.yaml --snapshot ./snapshot --out ./out/verify
```

To capture a snapshot with `gcloud` (optional, requires access to the target project):

```bash
mkdir -p ./snapshot
gcloud run services describe "$SERVICE" --project "$PROJECT" --region "$REGION" --format=json > ./snapshot/gcloud_service.json
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

Book 4 adds locked fixture cases for Chapters 4–6; Chapter 4 uses: `case01_deploy_render_smoke`.

Inputs:

- `fixtures/input/<case>/config.yaml`
- optional: `fixtures/input/<case>/gcloud_service.json` (if present, the case is a **verify** case)

Expected outputs:

- `fixtures/expected/<case>/` containing either:
  - the artifact set, or
  - `error.txt` only (expected-fail)

## Go + CI witness

- Go baseline: **1.22.x**
- CI witness: ubuntu/macos/windows on 1.22.x, plus a “stable” job on ubuntu.

## License

MIT (see `LICENSE`).
