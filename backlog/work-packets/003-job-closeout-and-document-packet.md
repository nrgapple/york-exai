# Work Packet 003 - Job Closeout And Document Packet

## Problem

Closed pest jobs are only valuable if they are documented well enough to bill, follow up, and defend later. The system needs a clean closeout model that understands inspection notes, treatment notes, prep notices, photos, and paperwork gaps.

## In Scope

- `DocumentPacket` state
- closeout readiness checks
- missing-record blockers
- follow-up task creation from incomplete closeouts

## Out Of Scope

- payment processing
- full bookkeeping

## Pest-Control Use Cases

- termite inspection cannot close without findings and paperwork-ready notes
- bed bug treatment cannot close cleanly if prep or follow-up state is missing
- recurring ant service can close quickly if notes and treatment summary are complete

## Acceptance Criteria

- closeout can distinguish complete, blocked, and follow-up-needed states
- missing photos, notes, or paperwork are visible before invoicing
- different job types can enforce different closeout expectations without custom chaos
