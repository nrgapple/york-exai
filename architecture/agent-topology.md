# Agent Topology

## Operating Agents

- `Chief of Staff`: runs the owner's day, summarizes risk, assigns work
- `Field Companion`: handles check-ins, voice memos, and field context
- `Dispatch And Route`: manages route flow, follow-ups, and callbacks
- `Billing And Collections`: moves closed work into invoice and payment state
- `Finance And Tax`: keeps books, categories, export packs, and close procedures straight
- `Document And Compliance`: guards paperwork quality and regulated records
- `Growth And Content`: turns field work into education, seasonal content, and review prompts

## PD&E Agents

- `Product Manager`: detects friction and prioritizes improvements
- `Service Designer`: shapes workflows and interfaces around field reality
- `Engineering Planner`: turns approved needs into decision-complete build packets
- `QA Ops Analyst`: validates that changes reduced pain instead of moving it
- `Backend Architect`: defines system boundaries, data contracts, and sequencing

## Orchestration Rule

Build-side agents can recommend and queue low-risk improvements, but coding work starts only when a work packet in `backlog/work-packets/` is specific enough for a coding agent to execute without inventing business rules.
