# BOOTSTRAP.md - First Agent Startup

You are the first OpenClaw or Codex agent pointed at the York ExAI repository. Your job is to turn this repo into an active build-and-operations launchpad inside an OpenClaw workspace.

## Mission

Set up the surrounding workspace so York ExAI agents can:

- understand the exterminator business they serve
- understand Adam, Matt, and Joel correctly
- speak to each human in the correct tone
- load the right local skills
- start planning and building the software in the right order

## First Actions

1. Read `AGENTS.md`, `MEMORY-SEEDS.md`, `context/human-roles.md`, `context/voice-and-tone.md`, `domain/`, `contracts/`, and `backlog/roadmap.md`.
2. If the surrounding OpenClaw workspace is missing `IDENTITY.md`, `USER.md`, `MEMORY.md`, daily memory files, or `HEARTBEAT.md`, create them using the workspace conventions.
3. Copy or register the local skills from `skills/` into the workspace skill path that OpenClaw will use.
4. Seed workspace memory with the facts in `MEMORY-SEEDS.md`.
5. Confirm the field channel defaults to iMessage unless Joel overrides it.
6. Review `integrations/` and identify what is configured, what is blocked, and what secrets or vendor accounts are still missing.
7. Open or refresh implementation-ready work packets under `backlog/work-packets/` before any software build begins.

## Non-Negotiables

- Do not frame this as generic field service software.
- Treat York ExAI as a pest-control operating system for central Pennsylvania.
- Preserve human-specific tone and authority rules, especially for Joel-facing operation.
- Do not automate regulated workflow changes without explicit approval.
- Keep QuickBooks optional and downstream until its fit is validated.

## Initial Build Order

1. Route-day operations
2. Check-ins and voice memos
3. Job closeout and document packet handling
4. Invoice drafting and collections
5. Internal ledger and accounting handoff
6. Product iteration loop
7. Customer messaging and growth automation

## Success Check

You are done with bootstrap when:

- the workspace knows what York ExAI is
- the workspace understands Adam, Matt, and Joel
- the right skills are available
- the core docs are indexed in memory
- the next implementation work is queued as decision-complete packets
