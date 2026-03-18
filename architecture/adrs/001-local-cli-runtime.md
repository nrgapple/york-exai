# ADR

## Title

Go local CLI runtime for York ExAI v1

## Status

Accepted

## Context

York ExAI needs a first software release that OpenClaw agents can call locally on the same Mac without asking the operator to install and manage infrastructure such as Postgres, Redis, Docker, or a long-running app server. The first release needs to support route-day operations, field check-ins and voice memos, closeout readiness, invoice-ready handoff, and local backup while keeping optional integrations outside the system core.

The repo already points toward a local-first operating model:

- business operations hub on a dedicated Mac mini
- iMessage adjacency for field intake
- SQLite plus filesystem documents as the likely v1 runtime
- optional downstream integrations for calendar, payments, bookkeeping, and transcription

The first release also needs a stable machine-callable interface for OpenClaw agents. That interface should not require agents to read and write SQLite directly.

## Decision

York ExAI v1 will use:

- Go for the implementation language
- a single local `york` binary as the primary execution surface
- SQLite as the structured source of truth
- local filesystem artifact storage for raw audio, photos, document packet media, exports, and backups
- a minimal config file with path overrides and optional integration sections only
- an agent-first CLI contract with structured JSON output for all commands
- append-only event rows using the existing event names in `contracts/events.md`

Default runtime shape:

- default home directory under macOS application support
- overridable via `YORK_HOME` and a CLI flag
- no always-on server required in v1
- no mandatory external services in v1
- optional integrations treated as adapters around the core, not the source of truth

## Consequences

Positive:

- low setup burden on the Mac that runs York ExAI
- simple portability for backup, restore, and machine replacement
- easier local validation and operational inspection
- direct OpenClaw integration through one stable CLI instead of ad hoc DB access
- degraded modes stay explicit when optional services are unavailable

Tradeoffs:

- richer multi-user and remote-access behavior is deferred
- SQLite concurrency needs disciplined transactions and busy timeouts
- the CLI contract must stay stable because agents will depend on it
- downstream integrations need clear boundaries so they do not leak into the local source of truth
