# Conventions

These conventions exist so the repo can *prove* its outputs: deterministic artifacts + fixtures + goldens + a verification gate.

## Line endings
- **LF only** (enforced via `.gitattributes`).

## Determinism
Artifacts must be deterministic:
- stable ordering (sorted lists; no filesystem walk order dependence)
- stable formatting (JSON indentation + trailing newline)
- no timestamps, UUIDs, random IDs, or other entropy in outputs

## Output directory
- Commands that write artifacts (`render`, `verify`, `demo`) **clear the `--out` directory first**.
- A safety guard refuses unsafe paths (e.g., `.`, `..`, `/`, Windows volume roots).

## Expected-fail
Expected-fail cases emit **only**:
- `error.txt` (must end with a newline)
