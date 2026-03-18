# System Overview

York ExAI is a future local-first operating system for an exterminator business. This repository is the planning and bootstrap layer that future agents will use to configure, design, and build the actual platform.

## Architectural Intent

- Local source of truth for business logic, workflows, and work packets
- OpenClaw-compatible skill packages for operational and build-side agents
- Machine-readable context index for stable repo truth and MCP-style resource loaders
- Codex-compatible implementation handoff for software build tasks
- Clear split between domain knowledge, operations playbooks, and software backlog

## High-Level Runtime Shape

- Business operations hub running on a dedicated Mac mini
- Phone-based field interaction through iMessage
- Voice memo ingestion and structured extraction
- Future software likely backed by SQLite plus filesystem documents, with selective vendor integrations
- External services limited to specialist functions like calendar, payment collection, and optional bookkeeping sync

## Design Constraints

- Must work for pest-control workflows, not generic ticketing
- Must preserve a direct owner-facing voice
- Must remain operable if optional integrations are unavailable
- Must keep regulated workflow changes behind approval
- Must keep the source package separate from the live OpenClaw workspace
