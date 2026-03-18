# Voice Transcription Integration

## Purpose

Turn field voice memos into searchable and structured operating records.

## Current Status

- status: planned contract, engine intentionally undecided
- current role: transcript, summary, and extraction service behind field memo handling
- blocking gap: no engine choice, credential inventory, or deployment mode has been approved yet

## Required Account And Credentials

- local engine install path or approved API credentials
- storage path for raw audio references
- approved runtime location for transcription processing

## Owner

- primary owner: Adam for engine approval and deployment choice
- operational consumer: field companion and product planning workflows

## Requirements

- retain raw audio reference
- generate transcript
- generate short owner summary
- extract pest facts, treatment facts, follow-up needs, billing changes, and content ideas
- flag low-confidence extraction

## Deployment Flexibility

- allow local transcription or API-backed transcription
- keep the contract stable even if the engine changes later

## Event Ownership

- primary inbound events:
  - `voice_memo.received`
  - `voice_memo.transcribed`
- downstream effects:
  - structured extraction into field, document, callback, billing, and feedback flows

## Degraded Mode

- if transcription is unavailable, raw audio stays attached and the memo is routed to review instead of being discarded
- summary and extraction may be delayed, but raw capture must remain durable

## Approval Boundary

- approval required before changing retention behavior, transcription provider, or any normalization that could flatten compliance-sensitive content

## Guardrails

- transcript and summary are separate artifacts
- unclear job linkage opens a review task
- compliance-sensitive content should be surfaced, not silently normalized
