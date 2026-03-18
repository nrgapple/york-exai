# Schema Contract

This contract defines the York ExAI v1 structured data boundaries.

## Core Tables

- `route_days`
- `jobs`
- `route_stops`
- `field_checkins`
- `voice_memos`
- `document_packets`
- `callbacks`
- `tasks`
- `invoice_drafts`
- `payments`
- `ledger_entries`
- `artifacts`
- `events`

## Key Foreign-Key Rules

- `route_stops.route_day_id -> route_days.id`
- `route_stops.job_id -> jobs.id`
- `field_checkins.linked_job_id -> jobs.id`
- `voice_memos.field_checkin_id -> field_checkins.id`
- `voice_memos.linked_job_id -> jobs.id`
- `document_packets.job_id -> jobs.id`
- `callbacks.origin_job_id -> jobs.id`
- `callbacks.callback_job_id -> jobs.id`
- `invoice_drafts.job_id -> jobs.id`
- `payments.invoice_draft_id -> invoice_drafts.id`
- `payments.job_id -> jobs.id`
- `ledger_entries.job_id -> jobs.id`
- `artifacts.linked_entity_id` must pair with `linked_entity_type`
- `events.entity_id` must pair with `entity_type`

## Modeling Rules

- keep termite, WDI, bed bug, callback, and closeout behavior explicit
- do not use a generic notes table as the primary workflow model
- transcript, summary, extracted facts, and raw audio reference remain separate fields
- closeout state and billing hold reason remain queryable columns
- event history is append-only

## Event Log Rules

- every mutating command writes one or more rows to `events`
- `event_name` must use existing names from `contracts/events.md`
- `payload_json` may extend detail, but not invent a new event type when an approved name already fits

## Initial Index Expectations

- route date lookup on `route_days.route_date`
- route ordering on `route_stops(route_day_id, position)`
- job status lookup on `jobs(status, closeout_state, billing_hold_reason)`
- callback pressure lookup on `callbacks(status, urgency)`
- report queries on `events(event_name, occurred_at)`
