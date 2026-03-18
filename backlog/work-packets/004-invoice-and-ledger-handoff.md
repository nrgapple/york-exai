# Work Packet 004 - Invoice And Ledger Handoff

## Problem

The business needs closed work to become invoice drafts, payment state, and internal bookkeeping without manual reconstruction at the end of the week or month.

## Why It Matters To The Pest Business

Cash flow depends on clean handoff from field closeout into invoicing and reconciliation. If that handoff is weak, same-day work sits unbilled, callbacks hide billability questions, and month-end becomes reconstruction instead of review.

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

## Inputs And Contracts

- foundation packet:
  - `backlog/work-packets/000-local-cli-runtime-and-contract-foundation.md`
- shared runtime contracts:
  - `contracts/cli.md`
  - `contracts/storage.md`
  - `contracts/schema.md`
- domain entities: `InvoiceDraft`, `Payment`, `LedgerEntry`, `Job`, `DocumentPacket`, `Callback`
- workflow contracts:
  - `Invoice Drafting And Reconciliation`
  - `Job Closeout And Documentation`
- integration boundaries:
  - `integrations/stripe.md`
  - `integrations/quickbooks.md`
- event contracts touched in v1:
  - `job.closed`
  - `invoice.drafted`
  - `invoice.sent`
  - `payment.received`
  - `payment.reconciled`
- required handoff inputs:
  - invoice-ready closeout state
  - billable items
  - callback or billing-hold visibility
  - payment method and received date
  - ledger mapping target

## Decision Rules And Ambiguity Handling

- Only jobs with complete closeout may enter invoice-ready state automatically.
- Callback activity, disputed scope, or missing records must create an explicit billing hold instead of silently delaying the invoice.
- If Stripe is unavailable, invoice and payment state must still be tracked internally for later reconciliation.
- If QuickBooks is unavailable or deferred, the internal ledger remains the active source of bookkeeping truth.
- If a payment arrives without a clean invoice link, route it to reconciliation review instead of forcing a ledger match.

## Acceptance Criteria

- incomplete closeouts do not silently enter billable state
- payment events can update invoice and ledger state together
- finance can operate even if QuickBooks sync is absent

## Risks And Compliance Notes

- Accounting rule changes require approval and should not be invented inside the first billing handoff build.
- Stripe and QuickBooks are downstream tools; they cannot become the only record of billing truth.
- Payment and ledger state should reconcile together, but bookkeeping edge cases should still surface for review instead of being auto-forced clean.

## Non-Goals

- advanced tax automation
- full QuickBooks sync implementation
- customer collections strategy beyond basic invoice and payment state

## Handoff Status

- implementation surface: `york billing`
- depends on: packets 000-003
- implementation note: v1 billing should focus on invoice-ready gating and explicit holds; deeper payment and ledger flows can remain skeletal until the ops core is proven

## Open Risks

- accounting treatment changes remain approval-sensitive
- invoice automation can drift into hidden callback write-offs if billing holds are not explicit and queryable
