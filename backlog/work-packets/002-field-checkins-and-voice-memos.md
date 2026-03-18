# Work Packet 002 - Field Check-Ins And Voice Memos

## Problem

The owner needs to log job reality from the field without long typing. Voice memos and quick check-ins have to become structured job state.

## Why It Matters To The Pest Business

Joel should be able to hand over field truth without stopping the day to type. If field capture is weak, callbacks, prep failures, scope changes, and paperwork gaps do not reach closeout, billing, or PD&E in time to matter.

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

## Inputs And Contracts

- domain entities: `FieldCheckIn`, `VoiceMemo`, `Job`, `DocumentPacket`, `Task`, `FeedbackSignal`
- workflow contracts:
  - `Field Check-In And Voice Memo Ingestion`
  - `Bed Bug Workflow`
  - `Callback Handling`
- contract docs:
  - `contracts/voice-memo.md`
  - `contracts/events.md`
- event contracts touched in v1:
  - `field_checkin.received`
  - `voice_memo.received`
  - `voice_memo.transcribed`
  - `callback.requested`
  - `prep_notice.required`
  - `feedback.detected`
- required captured outputs:
  - raw audio reference
  - transcript
  - short summary
  - extracted pest facts
  - extracted treatment facts
  - extracted follow-up needs
  - extracted billing or scope changes
  - confidence flags

## Decision Rules And Ambiguity Handling

- Raw audio, transcript, summary, and extracted facts are separate artifacts and must remain separately visible.
- If a memo cannot be tied confidently to a job, create a review task instead of guessing.
- Low-confidence extraction may append candidate facts, but it must not overwrite confirmed job truth.
- If a memo implies prep failure, callback risk, billing scope change, or compliance-sensitive content, the system must surface the flag explicitly even if the rest of the memo is low confidence.
- Voice memos can enrich the document packet, but they do not replace required structured records for closeout.

## Acceptance Criteria

- raw audio is retained
- summary and structured extraction stay separate
- low-confidence extraction never overwrites confirmed facts
- compliance and callback signals are surfaced explicitly

## Risks And Compliance Notes

- Termite, WDI, and other compliance-sensitive findings must be surfaced, not normalized into vague summaries.
- Joel-facing flows must stay terse and companion-like; the capture system cannot feel like software intake.
- Engine choice is intentionally deferred, so the contract must stay stable whether transcription is local or API-backed.

## Non-Goals

- choosing the production speech-to-text vendor
- outbound customer messaging
- long-form field form design
