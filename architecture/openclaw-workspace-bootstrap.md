# OpenClaw Workspace Bootstrap

This repo stays the source package for York ExAI. The live OpenClaw workspace should live outside this repo so workspace state, memories, bindings, and local secrets can change without turning this repo into the running system.

## Recommended Layout

- source package repo: `/path/to/york-exai-source`
- OpenClaw workspace root: `~/.openclaw/workspaces/york-exai`
- workspace git repo: optional but recommended

## Sample OpenClaw Binding

Use one primary York ExAI agent workspace in v1. Keep the binding simple until the route, closeout, and billing loops are stable.

```json5
{
  agents: {
    defaults: {
      workspace: "~/.openclaw/workspaces/york-exai",
    },
    list: [
      {
        id: "york-exai",
        default: true,
        workspace: "~/.openclaw/workspaces/york-exai",
      },
    ],
  },
  bindings: [
    {
      agentId: "york-exai",
      match: {
        channel: "imessage",
        accountId: "joel-field",
      },
    },
  ],
}
```

## Expected Workspace Files

The workspace should contain at least:

- `AGENTS.md`
- `IDENTITY.md`
- `USER.md`
- `MEMORY.md`
- `HEARTBEAT.md`
- daily memory files under the workspace memory convention
- a skill directory or links to the active York ExAI skills

## Skill Install Strategy

Prefer symlinks from the workspace skill path into this repo so skill updates stay centralized.

Active v1 skills:

- `york-bootstrap`
- `york-field-companion`
- `york-ops-coordinator`
- `york-back-office`
- `york-product-planning`
- `york-backend-architect`
- `york-implementation-orchestrator`

Deferred until roadmap priority 7 is active:

- `york-growth-content`

## Memory Seeding Procedure

1. Seed durable facts from `MEMORY-SEEDS.md`.
2. Preserve the three-human model from `context/human-roles.md`.
3. Preserve voice rules from `context/voice-and-tone.md`.
4. Load the stable resource map from `resources/context-index.json`.
5. Record integration readiness from `integrations/`.
6. Record which work packets are decision-complete and which are blocked.

## Repo Vs Workspace Boundary

Keep these in the source repo:

- business truth
- contracts
- work packets
- skill definitions
- integration readiness docs
- templates

Keep these in the live workspace:

- message history
- memory state
- channel bindings
- local secrets
- local identities
- machine-specific configuration
- operational logs

## First Bootstrap Review

Before calling bootstrap complete, verify:

- the iMessage default is bound or explicitly marked blocked
- active v1 skills are installed and callable
- `resources/context-index.json` is available to the workspace
- every integration file has a status, owner, required credential list, degraded mode, and approval boundary
- the next build packet is decision-complete enough for Codex to execute without inventing business rules
