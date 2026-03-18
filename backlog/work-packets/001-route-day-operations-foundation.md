# Work Packet 001 - Route-Day Operations Foundation

## Problem

The business needs a first software slice that understands a pest-control route day, not generic tickets. It must track ordered stops, urgent inserts, callbacks, and unresolved end-of-day work.

## Why It Matters To The Pest Business

Route flow is the daily operating backbone. If urgent yellowjacket work, callbacks, or incomplete specialty jobs disappear into a loose stop pile, the business loses billable time, misses follow-up, and buries paperwork-sensitive work that affects closeout and collections.

## In Scope

- `RouteDay`, `Job`, and route ordering model
- stop status updates
- urgent insert handling
- end-of-day unresolved queue
- summary outputs for owner and Ops Coordinator

## Out Of Scope

- payment processing
- customer-facing portal
- full accounting

## Pest-Control Use Cases

- squeeze in a same-day yellowjacket stop without losing termite inspection visibility
- detect when callbacks are crowding out recurring work
- carry incomplete bed bug follow-up cleanly into the next day

## Inputs And Contracts

- domain entities: `RouteDay`, `Job`, `Callback`, `Task`
- workflow contracts:
  - `Route-Day Execution`
  - `Callback Handling`
  - `Stinging Insect Same-Day Flow`
- event contracts touched in v1:
  - `callback.requested`
  - `callback.scheduled`
  - `followup.scheduled`
  - `feedback.detected` when repeated route friction should feed PD&E
- required route inputs:
  - ordered stops
  - scheduled window
  - job type
  - pest target
  - priority
  - assigned route day
  - exception list and unresolved carryover

## Decision Rules And Ambiguity Handling

- Urgent inserts may reorder the route, but the system must keep high-risk paperwork jobs visible instead of burying them behind same-day noise.
- A stop can be marked at risk, in progress, blocked, unresolved, or complete, but route state must not silently imply job closeout.
- If a callback or follow-up is detected mid-route, the route model must link the new work back to the origin job instead of treating it like an unrelated stop.
- If route completion is unclear at end of day, create an unresolved review item instead of marking the stop complete.
- Repeated route overruns, estimate failures, or callback crowding should surface as operating friction instead of disappearing inside schedule edits.

## Acceptance Criteria

- route summaries reflect ordered stops and disruptions
- urgent jobs can be inserted with explicit route impact
- unresolved jobs roll into a review queue instead of disappearing
- output language stays concise and owner-usable

## Risks And Compliance Notes

- Urgent route pressure cannot erase visibility into termite, WDI, or bed bug follow-up obligations.
- The route model may reorder work, but it must not redefine regulated documentation requirements.
- If actual drive-time or duration assumptions are weak, the system should surface schedule risk rather than pretending the route is clean.

## Non-Goals

- route optimization beyond simple ordered-stop management
- customer-facing scheduling promises
- invoice or ledger behavior
