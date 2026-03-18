# Domain Contracts

## Core Entities

### Customer

- billing contact
- service contact
- communication preferences
- outstanding balance state
- linked properties and jobs

### Property

- service address
- property type
- access notes
- pest history
- conducive conditions history

### ServicePlan

- plan type
- covered pests or visit scope
- service interval
- renewal state

### Job

- job type
- pest target
- priority
- scheduled window
- assigned route day
- closeout state

### Inspection

- findings
- evidence level
- conducive conditions
- inaccessible areas
- approval or next-step recommendation

### Treatment

- treatment scope
- areas serviced
- follow-up requirement
- completion timestamp

### RouteDay

- date
- ordered stops
- travel considerations
- exception list

### FieldCheckIn

- source channel
- timestamp
- linked job
- short status
- blockers

### VoiceMemo

- audio reference
- transcript
- summary
- extracted facts
- confidence flags

### Callback

- origin job
- reason
- urgency
- margin impact note

### PrepNotice

- required prep steps
- due date
- completion status

### DocumentPacket

- notes
- media
- forms
- follow-up instructions
- completeness status

### InvoiceDraft

- billable items
- taxes or fees if applicable
- delivery state
- payment status

### Payment

- amount
- method
- received date
- reconciliation state

### Expense

- category
- amount
- vendor
- linked period

### LedgerEntry

- account bucket
- amount
- source event
- period

### Task

- owner
- due date
- linked entity
- resolution state

### FeedbackSignal

- source
- friction theme
- severity
- supporting evidence

### ImprovementCandidate

- problem statement
- affected workflow
- expected benefit
- approval state

### RoadmapItem

- priority
- linked signals
- acceptance criteria
- implementation packet reference
