# Stripe Integration

## Purpose

Use Stripe as the default payment collection layer.

## Use Cases

- invoice payment links
- payment confirmation
- reconciliation events

## Guardrails

- billing must still function if Stripe is temporarily unavailable
- payment state should map back to `InvoiceDraft`, `Payment`, and `LedgerEntry`
- do not embed business logic only inside Stripe objects
