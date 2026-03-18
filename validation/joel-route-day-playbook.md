# Joel Route-Day E2E Playbook

This playbook explains York CLI validation in Joel terms. The point is not to test commands in a vacuum. The point is to prove the CLI can support an actual exterminator day without burying route pressure, callback risk, paperwork gaps, or billing exposure.

## Shared Validation Setup

- use a fresh synthetic runtime home such as `/tmp/york-e2e`
- invoke every command with `--json`
- do not write SQLite directly
- use synthetic customer names, addresses, and artifact files
- capture returned IDs from CLI responses instead of inventing them

Example bootstrap:

```bash
york --json --home /tmp/york-e2e init
york --json --home /tmp/york-e2e doctor
```

Expected trust outcome:

- runtime home exists
- database exists
- artifact directories exist
- backup directory exists
- the CLI can be operated entirely through structured responses

## Scenario 1: Morning Startup And Ordered Route

Joel starts the morning needing one clean picture of the day. He does not need a loose pile of stops. He needs the ordered route, the callbacks that are still hanging around, and the paperwork or billing weirdness that could trip the day up.

Validation flow:

```bash
york --json --home /tmp/york-e2e route create --date 2026-03-18
york --json --home /tmp/york-e2e route add-stop --date 2026-03-18 --customer "Smith Residence" --address "101 Queen St, York PA" --job-type "general pest" --pest "ants" --priority medium --window am
york --json --home /tmp/york-e2e route add-stop --date 2026-03-18 --customer "Keystone Realty" --address "455 Market St, Harrisburg PA" --job-type "termite inspection" --pest "termite" --priority high --window midday --paperwork-required=true
york --json --home /tmp/york-e2e report morning --date 2026-03-18
york --json --home /tmp/york-e2e route summary --date 2026-03-18
```

Trust outcome:

- route exists and returns ordered stops
- termite work is visible as specialty/high-risk work
- morning report surfaces callback, blocked-closeout, and invoice-risk counts if present
- no agent needed to touch the DB to understand the route

## Scenario 2: Normal Recurring Service Stop

Joel gets to a regular ant service stop. He does the work, sends a quick update, throws in a photo, finishes the notes, and the job should move cleanly toward billing without nonsense.

Validation flow:

```bash
york --json --home /tmp/york-e2e field checkin --job <general_pest_job_id> --status "on site" --text "Kitchen ants active. Exterior needs hit too."
york --json --home /tmp/york-e2e field photo --job <general_pest_job_id> --file /tmp/york-e2e-fixtures/ants-kitchen.jpg
york --json --home /tmp/york-e2e closeout note --job <general_pest_job_id> --service-summary "Interior crack and crevice plus exterior perimeter treatment complete." --treatment-notes "Kitchen baiting added. Exterior foundation treated."
york --json --home /tmp/york-e2e closeout evaluate --job <general_pest_job_id>
york --json --home /tmp/york-e2e billing draft --job <general_pest_job_id>
```

Trust outcome:

- field evidence attaches to the right job
- closeout becomes `complete` when the recurring service documentation is sufficient
- invoice draft is created without a hold
- Joel does not lose time re-entering what he already said and photographed

## Scenario 3: Same-Day Yellowjacket Insert

Mid-route, Joel gets an urgent yellowjacket call. The route has to absorb it without making the termite inspection disappear from view.

Validation flow:

```bash
york --json --home /tmp/york-e2e route insert-urgent --route-id <route_day_id> --customer "Brown Residence" --address "22 Hill Rd, Lancaster PA" --job-type "stinging insect response" --pest "yellowjacket" --priority high --position 1
york --json --home /tmp/york-e2e route summary --date 2026-03-18
york --json --home /tmp/york-e2e report route-risk --date 2026-03-18
```

Trust outcome:

- urgent work appears explicitly in route order
- existing termite / WDI / callback-sensitive jobs remain visible
- route-risk output reflects the disruption instead of pretending the route is still clean

## Scenario 4: Voice Memo With Review Fallback

Joel does not want to type a novel in the truck. He sends a voice memo. If transcription or job matching is weak, the system still has to keep the raw audio and route the ambiguity into review instead of guessing.

Validation flow:

```bash
york --json --home /tmp/york-e2e field voice --job <general_pest_job_id> --file /tmp/york-e2e-fixtures/ants-update.m4a --summary "Ants still active in kitchen." --confidence 0.50
york --json --home /tmp/york-e2e field list-review
```

Trust outcome:

- raw audio is stored under the runtime artifact tree
- review tasks are created for weak job linkage and/or missing transcript
- confirmed job truth is not overwritten by low-confidence extraction
- degraded mode is explicit and durable

## Scenario 5: Termite / WDI Closeout Blocker

Joel finishes a termite inspection, but the paperwork-ready detail is not fully buttoned up yet. The CLI has to protect the business by blocking clean closeout and billing handoff.

Validation flow:

```bash
york --json --home /tmp/york-e2e closeout note --job <termite_job_id> --service-summary "Evidence found at rear sill plate." --inspection-notes "Mud tubes visible at foundation seam."
york --json --home /tmp/york-e2e closeout evaluate --job <termite_job_id>
york --json --home /tmp/york-e2e billing draft --job <termite_job_id>
york --json --home /tmp/york-e2e report blocked-closeouts
```

Trust outcome:

- closeout returns `blocked`
- missing-item output names the specific paperwork gap
- billing does not quietly turn the job into a draft invoice
- blocked-closeouts reporting makes the risk visible to the office

## Scenario 6: Bed Bug Prep Failure Or Follow-Up Gap

Bed bug work falls apart fast if prep is missed or reinspection is undefined. Joel needs the CLI to protect follow-through instead of letting a bad closeout look clean.

Validation flow:

```bash
york --json --home /tmp/york-e2e route add-stop --date 2026-03-18 --customer "Miller Apartment" --address "88 Locust St, York PA" --job-type "bed bug treatment" --pest "bed bugs" --priority high --requires-prep=true --requires-follow-up=true
york --json --home /tmp/york-e2e closeout prep --job <bed_bug_job_id> --required=true --completed=false --follow-up-plan "Return after prep is completed."
york --json --home /tmp/york-e2e closeout note --job <bed_bug_job_id> --service-summary "Prep incomplete on arrival." --treatment-notes "No clean treatment path today."
york --json --home /tmp/york-e2e closeout evaluate --job <bed_bug_job_id>
```

Trust outcome:

- prep state is explicit
- closeout is not falsely marked `complete`
- follow-up-needed or blocked state is visible
- future agents can see exactly why the bed bug workflow did not close cleanly

## Scenario 7: Callback Handling And Billing Hold

Joel goes back to a callback-linked ant job. The system has to keep the callback tied to the origin work and stop billing from getting sloppy.

Validation flow:

```bash
york --json --home /tmp/york-e2e route add-stop --date 2026-03-18 --customer "Smith Residence Callback" --address "101 Queen St, York PA" --job-type "general pest callback" --pest "ants" --callback-of <general_pest_job_id>
york --json --home /tmp/york-e2e closeout note --job <callback_job_id> --service-summary "Returned for persistent ant activity."
york --json --home /tmp/york-e2e closeout evaluate --job <callback_job_id>
york --json --home /tmp/york-e2e billing draft --job <callback_job_id>
york --json --home /tmp/york-e2e billing list-holds
```

Trust outcome:

- callback remains linked to the origin job
- billing returns a hold instead of a clean draft
- callback pressure remains visible to PD&E and the office

## Scenario 8: Missed Check-In Recovery

Joel misses a check-in on a stop. The CLI has to support the runbook behavior of review and recovery instead of silently treating the job as done.

Validation flow:

```bash
york --json --home /tmp/york-e2e route update-stop --stop-id <stop_id> --status unresolved --unresolved-reason "missed check-in"
york --json --home /tmp/york-e2e report end-day --date 2026-03-18
```

Trust outcome:

- the stop appears in unresolved or blocked lists
- no closeout or invoice handoff happens by implication
- the route preserves tomorrow carryover visibility

## Scenario 9: End-Of-Day Close And Backup Trust

At the end of the day, Joel and the office need to know what is unresolved, what is blocked, what is billable, and whether the day can be backed up safely.

Validation flow:

```bash
york --json --home /tmp/york-e2e report end-day --date 2026-03-18
york --json --home /tmp/york-e2e backup create
york --json --home /tmp/york-e2e backup verify --file <backup_archive_path>
```

Trust outcome:

- end-of-day report lists unresolved stops, blocked closeouts, invoice-ready jobs, and billing holds
- backup archive contains the SQLite state and artifact files together
- backup verification proves restore inputs exist, not just archive creation

## What A Future Agent Should Conclude From A Good Run

- Joel can work the route without the CLI flattening route reality into generic ticket churn.
- Field evidence remains durable even when enrichment is weak.
- Compliance-sensitive documentation is protected by blocking logic instead of optimistic guessing.
- Billing moves fast only when the underlying closeout is actually clean.
- Route-day, closeout, and backup trust can be checked from the CLI itself.
