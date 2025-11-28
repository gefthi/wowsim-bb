# Session Playbook
Lightweight flow to keep continuity between sessions.

## Start of Session
- Read `PROJECT_STATE.md` and the latest brief in `doc/sessions/`.
- Capture today’s goal, references, and success criteria using the template below.
- Load only the snippets you need (interfaces/signatures) into the LLM.

### Session Start Template (scratch)
```
# Session <NN> - <Date>
Goal: <what you will ship today>
Relevant systems: <files/interfaces>
Success criteria: <checks/tests/output>
Notes: <any setup, seed, config>
```

## During Session
- Update code against existing interfaces; avoid re-creating systems already present.
- If new interfaces are added, note signatures for the brief.

## End of Session
- Write a brief in `doc/sessions/session_<NN>_integration.md`:
```
# Session <NN> Brief
What changed:
- <files touched + 1-liner>

Key interfaces:
- <type/function signature + short usage note>

Tests/status:
- <what ran, what is missing>

Next:
- <next logical step or open question>
```
- Update `PROJECT_STATE.md` (Completed/In Progress/Planned, decisions, TODOs).

## File Locations
- Current state: `doc/PROJECT_STATE.md`
- Session briefs: `doc/sessions/`
- Archived legacy docs: `doc/old_doc/`
