# Work Packet 004 - Invoice And Ledger Handoff

## Problem

The business needs closed work to become invoice drafts, payment state, and internal bookkeeping without manual reconstruction at the end of the week or month.

## In Scope

- `InvoiceDraft`, `Payment`, and `LedgerEntry` handoff contracts
- invoice-ready gating from closeout
- payment reconciliation rules
- internal ledger mapping

## Out Of Scope

- advanced tax automation
- full QuickBooks sync implementation

## Pest-Control Use Cases

- same-day yellowjacket stop closes and invoices quickly
- callback-related billing hold is visible before sending the invoice
- termite job revenue and payment state reconcile cleanly into the ledger

## Acceptance Criteria

- incomplete closeouts do not silently enter billable state
- payment events can update invoice and ledger state together
- finance can operate even if QuickBooks sync is absent
