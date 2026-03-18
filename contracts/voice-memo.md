# Voice Memo Contract

Voice memos are a first-class field input.

## Required Fields

- source channel
- received timestamp
- linked job if known
- audio file reference
- transcript
- short summary
- extracted pest or problem facts
- extracted treatment facts
- extracted follow-up needs
- extracted billing or scope changes
- extracted content ideas
- confidence flags

## Processing Rules

- Never discard the raw audio reference if transcript exists.
- Keep transcript and summary separate.
- Do not overwrite confirmed job facts with low-confidence extraction.
- If the linked job is unclear, open a review task instead of guessing.
- If the memo suggests callback, prep failure, billing scope change, or compliance risk, escalate it explicitly.

## Output Surfaces

- owner summary
- job timeline update
- document packet append
- task or callback creation
- feedback signal for PD&E when friction or repeated pain is present
