# Stripe Integration

## Purpose

Use Stripe as the default payment collection layer.

## Current Status

- status: planned, not configured in this source repo
- current role: downstream payment collection and confirmation layer
- blocking gap: no account credential inventory or invoice-link pattern has been committed yet

## Required Account And Credentials

- Stripe account access
- API credentials or approved payment-link setup for the live workspace
- webhook or reconciliation access for payment confirmation events

## Owner

- primary owner: Adam for account setup and approval
- operational consumer: back office for payment tracking and reconciliation

## Use Cases

- invoice payment links
- payment confirmation
- reconciliation events

## Event Ownership

- primary inbound events:
  - `payment.received`
  - `payment.reconciled`
- downstream effects:
  - invoice payment state updates
  - ledger reconciliation updates

## Degraded Mode

- if Stripe is unavailable, invoice state and manual payment tracking continue internally
- unlinked or delayed payment confirmations should go to reconciliation review instead of being auto-assigned

## Approval Boundary

- approval required before changing payment-link behavior, fee handling assumptions, or other billing logic that affects customer-facing payment flow

## Guardrails

- billing must still function if Stripe is temporarily unavailable
- payment state should map back to `InvoiceDraft`, `Payment`, and `LedgerEntry`
- do not embed business logic only inside Stripe objects
