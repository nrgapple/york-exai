# Validation And Trust

This section is the trust layer for the York CLI. Future OpenClaw and Codex agents should use it to understand what has been validated, how validation should be run, and what evidence is strong enough to trust the CLI while supporting Joel's day-to-day exterminator work.

## Purpose

The York CLI is meant to run real route work, field capture, closeout, invoice handoff, and backup without direct database writes from agents. That means future agents need more than command docs. They need:

- realistic Joel-facing workflow context
- scenario-by-scenario command expectations
- proof that route, documentation, and billing protection hold under normal and degraded conditions
- a standard way to record validation runs, drift, and unresolved risk

The primary trust mechanism is now a compiled-binary E2E suite that runs the built `york` binary as a subprocess against a fresh temp runtime with a real SQLite file and real artifact files.

## Read Order

1. `contracts/cli.md`
2. `contracts/storage.md`
3. `contracts/schema.md`
4. `validation/README.md`
5. `validation/joel-route-day-playbook.md`
6. `validation/evidence-matrix.md`
7. `validation/templates/run-report.md`

## Trust Rules

- Treat the CLI as the only supported machine-write interface in v1.
- Treat compiled-binary E2E validation as stronger evidence than in-process helper coverage.
- Do not trust unvalidated behavior just because a command exists.
- Trust is strongest when command output, event history, artifact persistence, and business outcome all line up.
- If a workflow affects compliance-sensitive documentation, closeout, or bookkeeping treatment, validation must prove the CLI blocks risky shortcuts instead of smoothing them over.
- If a workflow degrades into review-required state, that is acceptable only when the review path is explicit and durable.

## Status Categories

### Proven Behavior

Use this label when a workflow has:

- a documented scenario in the playbook
- a matching evidence row
- a compiled-binary E2E test that drives the real `york` binary
- clear command sequence and expected outputs
- durable artifact and event expectations
- at least one recorded validation run with no unresolved blocking drift

### Degraded But Acceptable

Use this label when:

- the core system preserves truth
- the CLI creates a review task or hold instead of guessing
- route, closeout, or billing safety is preserved
- an optional integration or enrichment layer is unavailable

Examples:

- missing transcript engine but raw audio is retained
- unclear job match but a review task is created
- callback-linked invoice draft becomes a hold instead of a clean billable draft

### Not Yet Proven Or Still Manual

Use this label when:

- the workflow is documented but no validation run has been recorded
- the scenario depends on behavior not yet implemented
- the workflow still requires manual operational follow-through outside the CLI

## How Future Agents Should Use This Section

1. Start with the Joel narrative playbook to understand what the operator is trying to get done.
2. Use the evidence matrix to choose the exact scenario you are validating.
3. Run the scenario through the CLI using `--json`.
4. Verify exit codes, response codes, events, artifact files, and final business state.
5. Record the run using the run-report template.
6. If the CLI behavior diverges from the expected business outcome, log the drift and feed the signal into PD&E instead of silently redefining the workflow.

## Validation Boundaries

- Use synthetic customers, route days, and artifact files in validation runs.
- Do not validate by writing SQLite directly.
- Optional integrations should be mocked, simulated, or marked out of scope unless the scenario specifically validates degraded behavior without them.
- Compliance-sensitive workflows must be validated conservatively. Blocking incomplete closeout is acceptable. Pretending incomplete paperwork is complete is not acceptable.

## Relationship To Other Repo Sections

- `contracts/` defines the stable machine-facing interface.
- `runbooks/` describe the operating loop Joel and the office need to execute.
- `validation/` proves that the CLI can support those runbooks without breaking route flow, documentation quality, or billing protection.
- `pde/` consumes repeated drift and validation failures as product signals.
