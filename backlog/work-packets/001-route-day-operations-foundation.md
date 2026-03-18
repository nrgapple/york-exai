# Work Packet 001 - Route-Day Operations Foundation

## Problem

The business needs a first software slice that understands a pest-control route day, not generic tickets. It must track ordered stops, urgent inserts, callbacks, and unresolved end-of-day work.

## In Scope

- `RouteDay`, `Job`, and route ordering model
- stop status updates
- urgent insert handling
- end-of-day unresolved queue
- summary outputs for owner and Chief of Staff

## Out Of Scope

- payment processing
- customer-facing portal
- full accounting

## Pest-Control Use Cases

- squeeze in a same-day yellowjacket stop without losing termite inspection visibility
- detect when callbacks are crowding out recurring work
- carry incomplete bed bug follow-up cleanly into the next day

## Acceptance Criteria

- route summaries reflect ordered stops and disruptions
- urgent jobs can be inserted with explicit route impact
- unresolved jobs roll into a review queue instead of disappearing
- output language stays concise and owner-usable
