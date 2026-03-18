# Work Packet 000 - Local CLI Runtime And Contract Foundation

## Problem

York ExAI needs one local software backbone before route, field, closeout, and billing slices can be implemented without drift. Right now the repo has workflow intent but no runtime contract, storage contract, or executable command surface for OpenClaw agents to call.

## Why It Matters To The Pest Business

If the first build starts from isolated feature code instead of one local runtime contract, the system will drift into ad hoc state management, fragile scripts, and integration-first behavior that raises setup burden for the business. The first release has to stay local, durable, and easy to replace on one machine.

## In Scope

- local runtime decision for v1
- `york` CLI contract
- local storage contract
- schema contract and event logging rule
- bootstrap, doctor, reporting, and backup foundations

## Out Of Scope

- iMessage transport implementation
- production transcription engine decision
- Google Calendar sync
- Stripe automation
- QuickBooks sync

## Inputs And Contracts

- `architecture/adrs/001-local-cli-runtime.md`
- `contracts/cli.md`
- `contracts/storage.md`
- `contracts/schema.md`
- `contracts/events.md`
- `architecture/system-overview.md`
- `operator/README.md`
- `operator/agent-writing-guide.md`
- `validation/README.md`
- `validation/evidence-matrix.md`

## Decision Rules And Ambiguity Handling

- Use one local binary instead of a service mesh or mandatory background server.
- SQLite is the required structured store for v1.
- Filesystem artifacts are required for voice memos, photos, documents, exports, and backups.
- OpenClaw agents call the CLI, not SQLite directly.
- Optional integrations may read from and write through the core, but they must not replace local truth.
- If a new feature needs storage or command behavior that is not covered by the shared contracts, update the contracts before feature implementation continues.

## Acceptance Criteria

- runtime decision is recorded in an ADR
- CLI, storage, and schema contracts exist and are referenced by downstream packets
- the codebase can initialize a local runtime without manual database setup
- mutating command paths write event rows
- backup captures both the database and artifact tree
- future agents have a validation trust section that maps Joel-style workflows to CLI evidence
- future agents have a Joel-facing expectations section that explains the local-first system in owner language without turning Joel into the CLI operator

## Risks And Compliance Notes

- local-first does not remove compliance sensitivity around document retention or regulated records
- SQLite concurrency must be handled explicitly through transactions and busy timeouts
- the stable CLI contract becomes an agent dependency and must not drift casually

## Non-Goals

- remote multi-user collaboration
- vendor-first architecture
- direct DB writes from OpenClaw agents
