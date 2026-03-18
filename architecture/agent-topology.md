# Agent Topology

## Active V1 Agent Packages

- `Field Companion`: handles check-ins, voice memos, field context, and Joel-facing capture
- `Ops Coordinator`: covers route flow, carryover, callback pressure, document readiness triage, and owner-facing operating summaries
- `Back Office`: covers invoice readiness, payment state, reconciliation, ledger cleanliness, and bookkeeping handoff
- `Product Planning`: combines prioritization, workflow shaping, packet authoring, and validation of operational improvements
- `Backend Architect`: stays on call for data contracts, event boundaries, and storage sequencing when a packet needs deeper design

## Deferred Package

- `Growth And Content`: stays deferred until roadmap priority 7 is active and the core operations loop is stable

## Coverage Rule

Department and PD&E responsibilities still live in `departments/` and `pde/`. The v1 agent package set is intentionally smaller than the business function map so the workspace does not carry unnecessary routing overhead.

## Orchestration Rule

Build-side agents can recommend and queue low-risk improvements, but coding work starts only when a work packet in `backlog/work-packets/` is specific enough for a coding agent to execute without inventing business rules.
