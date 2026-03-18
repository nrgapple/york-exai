# Work Packet 002 - Field Check-Ins And Voice Memos

## Problem

The owner needs to log job reality from the field without long typing. Voice memos and quick check-ins have to become structured job state.

## In Scope

- `FieldCheckIn` and `VoiceMemo` ingestion
- transcript, summary, and extraction pipeline contract
- job matching rules
- escalation paths for unclear or risky memos

## Out Of Scope

- production-grade speech engine choice
- customer message automation

## Pest-Control Use Cases

- memo says ants were active in the kitchen and exterior baiting was added
- memo says bed bug prep was not done and follow-up has to be rescheduled
- memo says termite evidence was found and paperwork still needs cleaned up

## Acceptance Criteria

- raw audio is retained
- summary and structured extraction stay separate
- low-confidence extraction never overwrites confirmed facts
- compliance and callback signals are surfaced explicitly
