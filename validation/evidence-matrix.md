# Evidence Matrix

Use this matrix to decide whether a York CLI workflow is proven, degraded-but-acceptable, or not yet proven.

## Evidence Record Shape

Every validation run should capture:

- scenario ID
- workflow name
- current status: `proven`, `degraded_acceptable`, or `not_yet_proven`
- business story
- preconditions and synthetic fixtures
- CLI commands in order
- expected exit codes
- expected response `code` values
- expected events
- expected artifact files
- expected final business outcome
- observed result
- drift or follow-up needed

## Workflow Matrix

| Scenario ID | Workflow | Commands | Expected Exit / Response | Expected Events | Expected Artifacts | Pass / Fail Criteria | Default Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `VAL-001` | Bootstrap and runtime trust | `init`, `doctor` | `0`, `init.ok`, `doctor.ok` | none required beyond schema readiness | config file, DB file, runtime dirs | Pass if runtime is created locally and CLI is operable with `--json` only | proven |
| `VAL-002` | Morning route setup | `route create`, `route add-stop`, `report morning`, `route summary` | `0`, route/report success codes | `followup.scheduled` for stop creation | none required | Pass if route is ordered, specialty work stays visible, and morning risk counts are readable | proven |
| `VAL-003` | Normal recurring stop to invoice-ready | `field checkin`, `field photo`, `closeout note`, `closeout evaluate`, `billing draft` | `0`, closeout complete, billing draft success | `field_checkin.received`, `inspection.completed`, `treatment.completed`, `job.closed`, `invoice.drafted` | photo under artifact tree | Pass if recurring work closes cleanly and drafts without hold | proven |
| `VAL-004` | Same-day urgent insert | `route insert-urgent`, `route summary`, `report route-risk` | `0`, route/report success codes | `followup.scheduled` or callback event as applicable | none required | Pass if urgent insert is explicit and termite/WDI or callback-sensitive work remains visible | proven |
| `VAL-005` | Voice memo fallback | `field voice`, `field list-review` | `2` when review is required, `field.voice.ok` response | `voice_memo.received`, optional `voice_memo.transcribed`, `feedback.detected`, `followup.scheduled` if extracted | raw audio under `artifacts/audio/` | Pass if raw audio is durable and review is explicit; fail if the memo is lost or silently guessed | proven |
| `VAL-006` | Termite / WDI closeout blocker | `closeout note`, `closeout evaluate`, `billing draft`, `report blocked-closeouts` | `2` for blocked closeout or hold path | `inspection.completed`, `job.closeout.blocked`, optional `feedback.detected` from billing hold | optional inspection media if attached | Pass if incomplete paperwork blocks clean closeout and invoice-ready handoff | proven |
| `VAL-007` | Bed bug prep / follow-up path | `route add-stop`, `closeout prep`, `closeout note`, `closeout evaluate`, `closeout status` | `0` or `2` depending on follow-up vs blocked state | `prep_notice.required`, `prep_notice.cleared`, `followup.scheduled`, possible `job.closeout.blocked` | optional media | Pass if prep failure or missing follow-up prevents false clean closeout and the packet state stays explicit | proven |
| `VAL-008` | Callback billing hold | `route add-stop --callback-of`, `closeout note`, `closeout evaluate`, `billing draft`, `billing list-holds` | `2` for billing hold | `callback.requested`, `callback.scheduled`, `feedback.detected` or hold state evidence | none required | Pass if callback remains linked and billing becomes an explicit hold | proven |
| `VAL-009` | Missed check-in recovery | `route update-stop`, `report end-day`, `report events` | `0`, report success codes | `field_checkin.received` equivalent route-stop update event | none required | Pass if unresolved work is preserved and no silent closeout or billing occurs | degraded_acceptable |
| `VAL-010` | End-of-day close and backup trust | `report end-day`, `backup create`, `backup verify` | `0`, backup success codes | none required beyond prior workflow history | backup archive containing DB and artifacts | Pass if unresolved, blocked, and invoice-ready state are visible and backup includes state plus artifacts together | proven |

## Proven Behavior

- bootstrap and runtime path creation
- route ordering and urgent insert visibility
- recurring service closeout to invoice-ready handoff
- raw audio durability and review fallback
- termite closeout blocking
- bed bug prep and follow-up protection path
- callback hold path
- missed check-in unresolved carryover path
- backup and backup verification

These are backed by compiled-binary E2E tests and should still be recorded through the run-report process when future agents validate the full playbook.

## Degraded But Acceptable

- voice memo intake without transcript engine
- any scenario where the CLI returns review-required state but preserves route, artifact, closeout, or billing truth

## Not Yet Proven Or Still Manual

- customer-facing notices or scheduling promises tied to optional integrations
- downstream invoice delivery, payment reconciliation, and accounting sync behavior beyond the local hold/draft boundary

## Evidence Review Rule

If a scenario produces the right command output but the wrong business outcome, mark it failed.

Examples:

- route summary exists but specialty work vanished from practical visibility
- voice memo stored but no review task was created when linkage was weak
- closeout looked complete even though paperwork or follow-up was still missing
- backup archive exists but artifact and database paths do not align
