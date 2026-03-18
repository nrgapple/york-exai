# Voice Transcription Integration

## Purpose

Turn field voice memos into searchable and structured operating records.

## Requirements

- retain raw audio reference
- generate transcript
- generate short owner summary
- extract pest facts, treatment facts, follow-up needs, billing changes, and content ideas
- flag low-confidence extraction

## Deployment Flexibility

- allow local transcription or API-backed transcription
- keep the contract stable even if the engine changes later

## Guardrails

- transcript and summary are separate artifacts
- unclear job linkage opens a review task
- compliance-sensitive content should be surfaced, not silently normalized
