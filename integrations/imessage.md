# iMessage Integration

## Purpose

iMessage is the default owner field channel for check-ins, photos, and voice memos.

## Current Status

- status: required default field channel, but workspace binding is not committed in this source repo
- current role: primary Joel-facing intake path for live route updates
- blocking gap: the live OpenClaw workspace still needs channel binding and account confirmation

## Required Account And Credentials

- access to the field iMessage account used for Joel-facing operations
- any local automation, relay, or device authorization the live workspace requires

## Owner

- primary owner: Adam for workspace and device approval
- field operator: Joel as the active route user

## Use Cases

- mid-route check-ins
- quick status updates
- voice memo capture
- reminder nudges

## Event Ownership

- primary inbound events:
  - `field_checkin.received`
  - `voice_memo.received`
  - `feedback.detected` when Joel friction should route to PD&E
- downstream effects:
  - review task creation when job linkage is unclear
  - route, closeout, callback, and feedback updates based on parsed field input

## Degraded Mode

- if automated intake is unavailable, Joel can still communicate through iMessage, but the workspace must flag the channel as degraded and require manual review
- unclear or partial messages should open review tasks instead of being forced into job state

## Approval Boundary

- approval required before changing Joel-facing wording patterns materially or routing the default field channel away from iMessage

## Guardrails

- if a message cannot be confidently tied to a job, create a review task
- keep owner-facing responses terse and practical
- do not bury urgent route or compliance issues in chatty replies
