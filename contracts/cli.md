# CLI Contract

This contract defines the stable local command surface for the York ExAI v1 binary.

## Purpose

The `york` CLI is the only supported machine-write interface for OpenClaw agents in v1. Agents should call the CLI and read structured output. They should not write SQLite directly.

## Global Rules

- every command supports `--json`
- every command supports `--home` to override the runtime home directory
- mutating commands must write an event row
- text output is for humans and debugging
- JSON output is the stable automation surface

## Command Groups

### `york init`

- create the runtime home layout
- create the SQLite database if missing
- apply schema migrations
- write the minimal config file if missing

### `york doctor`

- validate home paths
- validate database access
- validate artifact directories
- validate backup directory
- report optional integration readiness from config

### `york route`

- `create`
- `add-stop`
- `insert-urgent`
- `update-stop`
- `summary`

### `york field`

- `checkin`
- `photo`
- `voice`
- `list-review`

### `york closeout`

- `note`
- `photo`
- `prep`
- `evaluate`

### `york billing`

- `draft`
- `list-holds`

### `york report`

- `morning`
- `route-risk`
- `blocked-closeouts`
- `callback-pressure`
- `end-day`

### `york backup`

- `create`
- `verify`

## JSON Envelope

All JSON responses use this shape:

```json
{
  "ok": true,
  "code": "route.summary.ok",
  "message": "Route summary ready.",
  "data": {},
  "warnings": []
}
```

On failure:

```json
{
  "ok": false,
  "code": "job.not_found",
  "message": "Job was not found.",
  "errors": [
    {
      "code": "job.not_found",
      "detail": "job_id=job_123"
    }
  ]
}
```

## Exit Codes

- `0`: success
- `1`: usage error or validation error
- `2`: not found or review-required state
- `3`: storage or migration failure
- `4`: command execution failed after validation

## Stable Enums

### Stop Status

- `scheduled`
- `at_risk`
- `in_progress`
- `blocked`
- `unresolved`
- `complete`

### Closeout State

- `pending`
- `complete`
- `blocked`
- `follow_up_needed`

### Billing Hold Reason

- `none`
- `closeout_incomplete`
- `callback_under_review`
- `scope_disputed`
- `missing_records`

### Task Reason

- `review_required`
- `job_link_unclear`
- `transcription_unavailable`
- `closeout_missing_records`
- `callback_followup`
- `billing_reconciliation_review`

## Command Behavior Rules

- `route insert-urgent` must preserve visibility into termite, WDI, bed bug, and callback-sensitive work.
- `field voice` must keep raw audio, transcript, summary, and extracted facts separate.
- `field voice` must create a review task instead of guessing when job linkage is unclear.
- `closeout evaluate` must use job-type-aware rules.
- `billing draft` may auto-create drafts only for `complete` closeout state.
- `backup create` must package the database and artifact storage together.

## Validation References

Future agents should use the validation trust packet alongside this contract:

- `validation/README.md`
- `validation/joel-route-day-playbook.md`
- `validation/evidence-matrix.md`

The CLI contract defines what the interface is. The validation section defines what has been proven safe enough to trust while supporting Joel's business.
