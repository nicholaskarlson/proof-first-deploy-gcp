# Contract (Repo B)

This document is the **behavioral contract** for `pfdeploy` (render + verify).
The book references only:
- fixture folder names, and
- artifact file names

…not raw code excerpts.

## Inputs

### Config file (YAML)

`pfdeploy` reads a single YAML config (examples in `fixtures/input/**/config.yaml`).

Required fields:

- `project_id` (string)
- `region` (string)
- `service_name` (string) — must match: `^[a-z0-9][a-z0-9-]{0,62}$`
- `image_digest` (string) — must end with: `@sha256:<64 hex>`
- `eventarc.trigger_name` (string)
- `eventarc.bucket` (string)
- `eventarc.input_prefix` (string)
- `eventarc.ce_type_allowlist` (list of strings, non-empty)

Optional runtime fields:

- `cpu` (string)
- `memory` (string)
- `concurrency` (int)
- `timeout_seconds` (int)
- `max_instances` (int)

Optional environment variables:

- `env` (map of string → string)

Optional IAM config:

- `iam.service_account` (string)
- `iam.roles` (list of strings)

### No-secrets rule (hard fail)

To keep configs safe to commit and share, **any env key containing** one of:

- `SECRET`, `TOKEN`, `PASSWORD`, `KEY` (case-insensitive substring match)

…is forbidden and causes a failure.

In that case, the output must be **only**:

- `error.txt` (must end with a newline)

## Outputs

All successful outputs are deterministic and end with a trailing newline.

### Render outputs

`pfdeploy render --config <config.yaml> --out <outdir>`

On success, emits:

- `deploy_manifest.json` — Cloud Run service intent (project/region/name/image + runtime + env)
- `trigger_manifest.json` — Eventarc trigger intent (trigger/bucket/prefix + CE allowlist)
- `iam_manifest.json` — service account + roles (roles default to `roles/run.invoker` if empty)
- `manifest.sha256` — sha256 lines over the three JSON files

`manifest.sha256` format:

- stable order (sorted by filename)
- each line is: `<hex>␠␠<filename>`
- file ends with a newline

On failure, emits **only**:

- `error.txt`

### Verify outputs

`pfdeploy verify --config <config.yaml> --snapshot <dir> --out <outdir>`

The snapshot is an intentionally tiny, fixture-friendly representation of what’s deployed.

For this MVP, the snapshot file is:

- `<snapshot>/gcloud_service.json`

Required JSON keys:

- `service_name`
- `region`
- `image_digest`

On match, emits:

- `verify_report.json` (pretty JSON + trailing newline)

On mismatch or invalid snapshot, emits **only**:

- `error.txt`

## Out dir clearing

For all commands that write output (`render`, `verify`, and `demo`):

- `--out` is cleared before writing outputs
- safety guard refuses unsafe out dirs (e.g. `.`, `..`, `/`, and Windows volume roots)

This prevents “stale file” confusion and makes reruns reliable.

## Determinism rules

Artifacts must be deterministic:

- stable ordering: sorted allowlists / roles, sorted env pairs
- stable formatting: `json.MarshalIndent` + trailing newline
- no timestamps, UUIDs, random IDs, or other entropy
- LF line endings (`.gitattributes` enforces LF across the repo)

## Proof gate (fixtures + goldens)

`pfdeploy demo --out <outdir>` recomputes every fixture case and byte-compares the result tree to `fixtures/expected/**`.

Fixture convention:

- Input case folder: `fixtures/input/<case>/`
- Expected outputs: `fixtures/expected/<case>/`

Heuristic used by the demo runner:

- if `fixtures/input/<case>/gcloud_service.json` exists → run **verify**
- otherwise → run **render**

Expected-fail cases are represented by an expected tree that contains **only** `error.txt`.
