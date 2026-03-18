# Google Calendar Integration

## Purpose

Use Google Calendar as the scheduling calendar for route and appointment visibility.

## Current Status

- status: planned, not configured in this source repo
- current role: downstream calendar visibility for route and appointment timing
- blocking gap: no workspace credential inventory or calendar mapping has been committed yet

## Required Account And Credentials

- Google account that owns the operating calendar
- OAuth client or other approved calendar auth method for the live workspace
- calendar ID for the York ExAI route calendar

## Owner

- primary owner: Adam for account setup and integration approval
- operational consumer: Joel for route visibility

## Use Cases

- route-day time blocks
- inspection appointments
- follow-up visits
- owner visibility from phone and desktop

## Event Ownership

- no dedicated sync event contract in v1
- consumes route and job scheduling state from the York ExAI platform
- mirrors scheduling state outward without becoming the source of truth

## Degraded Mode

- if calendar sync is unavailable, route order and appointment truth remain in York ExAI
- missed sync should create an operator-visible setup or review task, not silent drift

## Approval Boundary

- approval required before calendar sync changes customer-facing scheduling promises or auto-reschedule behavior

## Guardrails

- calendar sync should not become the only source of operational truth
- route state and closeout state still belong to the York ExAI platform
- reschedules should preserve callback and document context
