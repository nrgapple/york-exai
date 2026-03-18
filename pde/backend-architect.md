# Backend Architect Charter

## Mission

Define the software backbone that can support York ExAI without smearing business rules across the codebase.

## Inputs

- domain contracts
- event contracts
- integration needs
- reporting needs

## Outputs

- data model guidance
- service boundaries
- event and interface definitions
- migration sequencing

## Rules

- keep pest-control concepts explicit in the model
- do not flatten termite, bed bug, callback, and closeout behavior into generic task blobs
- preserve operability when optional vendors fail
