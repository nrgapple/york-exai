# Local Storage Contract

This contract defines the York ExAI v1 local runtime layout.

## Default Home

- macOS default: `~/Library/Application Support/YorkExAI`
- override order:
  1. `--home`
  2. `YORK_HOME`
  3. default macOS path

## Directory Layout

```text
YORK_HOME/
  config.json
  state/
    york.db
  artifacts/
    audio/
    photos/
    documents/
    exports/
  backups/
  tmp/
```

## Storage Rules

- SQLite is the source of truth for structured records and artifact references.
- Artifact files are durable local files referenced from SQLite.
- Raw audio must be retained even if transcription is unavailable.
- Backup must include both the database and artifacts so references remain valid after restore.
- Optional integrations may cache local state, but they must not become the source of truth.

## Artifact Path Conventions

- audio: `artifacts/audio/<voice_memo_id>/<original_filename>`
- photos: `artifacts/photos/<job_id>/<timestamp>-<original_filename>`
- documents: `artifacts/documents/<job_id>/<artifact_id>-<name>`
- exports: `artifacts/exports/<date>/<filename>`

## Backup Bundle Contents

- `manifest.json`
- `state/york.db`
- `artifacts/` subtree

The manifest must include:

- creation timestamp
- database relative path
- artifact root relative path
- artifact file count
- schema version

## Migration Rules

- migrations are owned by the application code, not manual operator steps
- `york init` and `york doctor` may apply or verify migrations
- schema version lives inside SQLite
- migrations must be forward-only in v1
- destructive retention or deletion changes remain approval-sensitive
