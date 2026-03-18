# Work Packet 003 - Job Closeout And Document Packet

## Problem

Closed pest jobs are only valuable if they are documented well enough to bill, follow up, and defend later. The system needs a clean closeout model that understands inspection notes, treatment notes, prep notices, photos, and paperwork gaps.

## Why It Matters To The Pest Business

Closeout quality is the gate between field work and cash. If notes, prep evidence, photos, or paperwork are incomplete, invoice speed slows down, callbacks become harder to explain, and specialty work carries unnecessary downside.

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

## Inputs And Contracts

- domain entities: `Job`, `Inspection`, `Treatment`, `DocumentPacket`, `PrepNotice`, `Task`, `Callback`
- workflow contracts:
  - `Job Closeout And Documentation`
  - `Termite Or WDI Inspection`
  - `Bed Bug Workflow`
- event contracts touched in v1:
  - `job.closeout.blocked`
  - `job.closed`
  - `followup.scheduled`
  - `prep_notice.required`
- required closeout inputs:
  - service summary
  - inspection notes when applicable
  - treatment notes when applicable
  - photos or media when required by job type
  - prep status and follow-up plan when applicable
  - callback and blocker visibility

## Decision Rules And Ambiguity Handling

- Closeout state must distinguish `complete`, `blocked`, and `follow-up-needed`; blocked and follow-up-needed are not interchangeable.
- Termite and WDI work cannot close cleanly without paperwork-ready findings and required inspection detail.
- Bed bug work cannot close cleanly when prep status, treatment scope, or reinspection path is missing.
- Voice memos may append context to the packet, but they do not replace required structured records.
- If required evidence is missing or contradictory, move the job to blocked closeout instead of guessing completeness.

## Acceptance Criteria

- closeout can distinguish complete, blocked, and follow-up-needed states
- missing photos, notes, or paperwork are visible before invoicing
- different job types can enforce different closeout expectations without custom chaos

## Risks And Compliance Notes

- Termite and WDI paperwork is approval-sensitive and must not be watered down into generic closeout behavior.
- The system should separate missing-record blockers from optional nice-to-have additions so routine work can still close quickly.
- Closeout rules must protect invoice readiness without embedding accounting behavior inside the document packet itself.

## Non-Goals

- payment collection
- full accounting workflow
- customer portal document delivery
