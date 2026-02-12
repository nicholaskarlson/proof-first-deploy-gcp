# Conventions

## Line endings
- LF only (enforced via `.gitattributes`).

## Determinism
Artifacts must be deterministic:
- stable ordering (sorted lists)
- stable formatting (indentation + trailing newline)
- no timestamps, UUIDs, random IDs, or entropy in output files

## Expected-fail
Expected-fail cases emit **only**:
- `error.txt` (must end with a newline)
