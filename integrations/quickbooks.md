# QuickBooks Integration

## Purpose

QuickBooks is optional downstream sync or export for bookkeeping and tax prep.

## Current Status

- status: deferred optional integration
- current role: downstream sync or export target only after the internal ledger is stable
- blocking gap: no validated account mapping or sync contract has been accepted yet

## Required Account And Credentials

- QuickBooks company file or tenant access
- approved auth method for sync or export
- chart-of-accounts mapping and reconciliation review path

## Owner

- primary owner: Adam for approval and accountant coordination
- operational consumer: back-office workflow after internal ledger state is dependable

## Current Stance

- internal bookkeeping is required
- QuickBooks is helpful only if its API and mapping fit cleanly
- York ExAI must remain usable if QuickBooks is deferred or failing

## Event Ownership

- no dedicated QuickBooks sync event contract in v1
- consumes internal invoice, payment, and ledger state after York ExAI bookkeeping is already clean

## Degraded Mode

- if QuickBooks is unavailable, the internal ledger remains the bookkeeping source of truth
- failed sync or export should create a visible reconciliation follow-up, not block day-to-day operations

## Approval Boundary

- approval required before changing accounting mappings, export shape, or any automation that could alter book treatment

## Guardrails

- do not let QuickBooks become the only finance source of truth
- sync must be explicit and auditable
- accounting rule changes require approval
