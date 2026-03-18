# Workflow Contracts

## Recurring General Pest Service

- verify stop and service plan
- inspect active areas and conducive conditions
- perform treatment
- record notes and customer-facing summary
- flag callback risk or upsell potential
- close out for invoicing

## Termite Or WDI Inspection

- confirm inspection type and access scope
- record evidence, conducive conditions, and inaccessible areas
- capture paperwork-ready findings
- schedule treatment or reporting follow-up
- enforce documentation completeness before closeout

## Bed Bug Workflow

- assess evidence and room spread
- verify prep status
- if prep is incomplete, trigger `PrepNotice` and route follow-up
- document treatment and reinspection plan
- do not treat closeout as complete without follow-up path

## Stinging Insect Same-Day Flow

- classify urgency and access
- fit into route with travel impact acknowledged
- document nest location and treatment outcome
- issue customer safety note
- close quickly for rapid billing

## Callback Handling

- tie callback to original job and service line
- classify cause: incomplete treatment, prep failure, persistent conditions, customer perception, or unknown
- schedule return work
- log margin and learning signal
- feed repeated patterns to PD&E

## Route-Day Execution

- morning brief
- ordered route with time risk flags
- live field check-ins
- urgent insert management
- end-of-day unresolved list

## Field Check-In And Voice Memo Ingestion

- accept text, photo, and voice input
- match to active job or create review task
- extract pest facts, treatment facts, blockers, and follow-up needs
- summarize for owner and attach to document packet

## Job Closeout And Documentation

- verify service summary
- verify document packet completeness
- create invoice draft inputs
- create follow-up tasks if needed
- block clean closeout when required records are missing

## Invoice Drafting And Reconciliation

- turn billable closeouts into invoice drafts
- deliver or queue delivery
- apply payment updates
- reconcile payment state back to job and ledger
