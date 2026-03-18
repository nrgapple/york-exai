# York ExAI

York ExAI is meant to take the weight of running a bug business off the humans who carry it without getting in the way of the actual work. For an exterminator outfit in central Pennsylvania, that means tighter route days, fewer missed callbacks, faster job closeout, quicker invoicing, cleaner records, and better follow-through on termite, bed bug, and other paperwork-heavy work. The point is not flashy software. The point is to help the business stay buttoned up, get paid faster, and stop losing time to scattered admin.

This repo is the starting package for building that system. It gives future humans and agents the business context, operating rules, skill packages, and implementation backlog needed to turn York ExAI into a real operating system for a pest-control company.

## What This Repo Is

This is the source package and launchpad for York ExAI. It is not the finished software product.

It exists so future collaborators, OpenClaw agents, and Codex implementation agents can:

- understand the exterminator business they are serving
- use the right pest-control terminology and workflows
- work from a shared source of truth
- build the platform in the right order

## Business Context

York ExAI is built for an exterminator business in central Pennsylvania, with York, Harrisburg, Lancaster, and nearby territory as the default operating assumption.

The business model behind this repo is not generic field service. It is pest-control work with real-world pressure around:

- recurring general pest routes
- callbacks and follow-up work
- urgent stinging insect jobs
- termite and wood-destroying insect inspections
- bed bug prep, treatment, and reinspection
- job closeout, invoicing, collections, and compliance-sensitive records

This repo also models three human roles explicitly:

- Adam: CEO, technical, and final authority on company and system changes
- Matt: Adam's delegate and operational contact
- Joel: owner and end user in the field, non-technical, blue-collar central PA, with 35+ years in pest control

If you need the deeper business framing, start with [domain/business-model.md](domain/business-model.md), [domain/pest-catalog.md](domain/pest-catalog.md), and [domain/central-pa-operations.md](domain/central-pa-operations.md).

## What Is Already In The Repo

- `domain/`: exterminator business truth, pest catalog, terminology, regional assumptions, compliance scope, and communication rules
- `contracts/`: system entities, workflow contracts, voice memo behavior, and event definitions
- `departments/`: operating charters for field ops, dispatch, billing, finance, compliance, growth, and chief-of-staff work
- `pde/`: product, design, and engineering planning charters plus triage rules
- `skills/`: OpenClaw-ready skill packages for business-side and build-side agents
- `runbooks/`: recurring procedures for daily operations, finance review, closeout, and roadmap review
- `integrations/`: Google Calendar, Stripe, QuickBooks, iMessage, and transcription boundaries
- `backlog/`: build order and implementation-ready work packets
- `templates/`: repeatable templates for ADRs, experiments, post-job reviews, workflow specs, and work packets

## If You Are...

### The Owner Or Operator

Start here:

1. [domain/business-model.md](domain/business-model.md)
2. [backlog/roadmap.md](backlog/roadmap.md)
3. [context/voice-and-tone.md](context/voice-and-tone.md)

That will show you what this system is trying to do for the business, what gets built first, and how the agents are expected to talk and report back.

Joel-facing agents are expected to respect trade experience, feel like dependable coworkers, and quietly route Joel's feedback back into the product team and Adam's review path.

### A Planner Or PD&E Collaborator

Start here:

1. [BOOTSTRAP.md](BOOTSTRAP.md)
2. [contracts/domain.md](contracts/domain.md)
3. [contracts/workflows.md](contracts/workflows.md)
4. [pde/triage-system.md](pde/triage-system.md)
5. [backlog/roadmap.md](backlog/roadmap.md)

That gives you the business operating model, the workflow contracts, and the rules for deciding what should be built next.

### An Implementation Engineer

Start here:

1. [BOOTSTRAP.md](BOOTSTRAP.md)
2. [contracts/domain.md](contracts/domain.md)
3. [contracts/events.md](contracts/events.md)
4. [backlog/work-packets](backlog/work-packets/)
5. [skills/york-implementation-orchestrator/SKILL.md](skills/york-implementation-orchestrator/SKILL.md)

Use the work packets as the entrypoint for build work. If the packet is not decision-complete, fix the packet before you start coding.

## How Humans Should Start

If you are new to the repo and want the shortest useful read order, use this:

1. [BOOTSTRAP.md](BOOTSTRAP.md)
2. [domain/business-model.md](domain/business-model.md)
3. [domain/pest-catalog.md](domain/pest-catalog.md)
4. [contracts/domain.md](contracts/domain.md)
5. [backlog/roadmap.md](backlog/roadmap.md)

That is enough to understand what the business is, what the repo is doing, and what work comes first.

## How Agents Fit In

This repo is structured so future agents can work without guessing.

- OpenClaw agents use the local skill packages in `skills/` to take on specific roles like field companion, dispatch, billing, finance, product planning, and implementation orchestration.
- Codex is meant to pick up decision-complete implementation work from `backlog/work-packets/`.
- The split between `domain/`, `contracts/`, `departments/`, `pde/`, and `backlog/` is intentional. It keeps business truth, operating behavior, and build work from getting mixed together.

If you need the exact agent startup rules, use [AGENTS.md](AGENTS.md).

## Current Build Order

The current build priority is practical and operations-first:

1. route-day operations
2. check-ins and voice memos
3. job closeout and document packet handling
4. invoice drafting and collections
5. internal ledger and accounting handoff
6. product iteration loop
7. customer messaging and growth

The reasoning is simple: the route has to run clean, field input has to be easy, closeout has to support billing, and the books have to stay straight before the business layers on more automation.

## Repo Map

If you just need to know where truth lives:

- business and exterminator domain truth: [domain/](domain/)
- system behavior and contracts: [contracts/](contracts/)
- owner and voice preferences: [context/](context/)
- operating team charters: [departments/](departments/)
- product, design, and engineering planning: [pde/](pde/)
- agent skill packages: [skills/](skills/)
- recurring procedures: [runbooks/](runbooks/)
- roadmap and implementation backlog: [backlog/](backlog/)
- vendor boundaries and setup notes: [integrations/](integrations/)

## Working Rules

- Do not treat this repo like generic field service software.
- Use pest-control terminology correctly.
- Keep compliance-sensitive changes behind approval.
- Use work packets before coding.
- Do not let the README replace the deeper source-of-truth docs.

For the agent-specific operating rules, use [AGENTS.md](AGENTS.md).
