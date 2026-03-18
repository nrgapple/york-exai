# PD&E Triage System

## Purpose

This document defines how York ExAI turns field reality into build priorities.

## Signal Sources

- callbacks
- missed prep
- unresolved closeouts
- invoice lag
- missing paperwork
- route overruns
- owner complaints in check-ins or voice memos

## Triage Categories

- `route friction`
- `documentation friction`
- `billing friction`
- `compliance risk`
- `customer communication friction`
- `data quality gap`

## Auto-Queue Rule

PD&E may auto-queue low-risk changes such as:

- summary wording improvements
- reminder timing changes
- dashboard ordering changes
- tagging and classification refinements

## Approval Rule

Approval is required before changing:

- accounting logic
- customer promises or regulated notices
- document retention behavior
- compliance-sensitive workflows

## Expected Output

Each meaningful signal cluster should create or update:

- an `ImprovementCandidate`
- a roadmap priority decision
- a work packet request if the problem is ready for engineering
