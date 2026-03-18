# AGENTS.md - York ExAI Repo Rules

This repository is the source package for York ExAI, an exterminator operations system for central Pennsylvania. It is not the software product itself. It exists so future OpenClaw and Codex agents can understand the business, configure a surrounding workspace, and build the right system without guessing.

## Read Order

When you enter this repo, read these first:

1. `BOOTSTRAP.md`
2. `MEMORY-SEEDS.md`
3. `context/human-roles.md`
4. `context/voice-and-tone.md`
5. `domain/business-model.md`
6. `domain/pest-catalog.md`
7. `contracts/domain.md`
8. `backlog/roadmap.md`

Then load the department, PD&E, integration, runbook, and skill-specific material needed for the task at hand.

## What This Repo Is For

- Pest-control business operations in York, Harrisburg, Lancaster, and nearby central Pennsylvania territory
- General pest plus wood-destroying insect workflows
- Owner-run field operations with heavy emphasis on route work, callbacks, inspections, closeout, invoicing, collections, and compliance
- Future OpenClaw and Codex agents that will build and operate the real system

## Working Rules

- Do not treat this like generic field service. It is for extermination work.
- Use pest-control terminology correctly. If unsure, check `domain/terminology.md`.
- Owner-facing outputs must follow `context/voice-and-tone.md`.
- Internal planning, design, and engineering artifacts must stay technical and explicit.
- Compliance-sensitive automation changes need approval. Check `domain/compliance-scope.md`.
- Prefer updating source-of-truth docs and work packets before proposing code changes.
- Keep skill packages lean. Put detail in `references/`.

## Where Truth Lives

- Business context: `domain/`
- System rules and entities: `contracts/`
- Human roles and voice preferences: `context/`
- Department responsibilities: `departments/`
- Product, design, and engineering planning: `pde/`
- Future build queue: `backlog/`
- Agent-loadable packages: `skills/`
- Recurring procedures: `runbooks/`
- Vendor setup and boundaries: `integrations/`

## Planning Vs Implementation

- If the task is about scope, operations, priorities, data models, tone, or team behavior, update docs first.
- If the task is about building software, do not start coding until the relevant work packet is decision-complete.
- If there is no suitable work packet, create or update one in `backlog/work-packets/`.

## Memory Discipline

- Durable York ExAI business facts belong in `MEMORY-SEEDS.md`.
- Repo-specific decisions belong in the relevant doc plus a work packet or ADR when needed.
- Avoid burying critical assumptions in chat-only responses.
