---
name: tools
description: >
  Reference for the Oraculo Tools — the Trust Layer that all agents must use to
  interact with epics, stories, tasks, approvals, knowledge, and sessions.
  Covers both the CLI (`oraculo tools`) and MCP blocking tools (`request_approval`,
  `approval_status`). This skill is always loaded. Consult it whenever you need
  to call any `oraculo` command or MCP tool.
---

# Oraculo Tools Reference

The Oraculo CLI is the **only** interface agents use to read and write project state.
Never read `.oraculo/` files directly or query the SQLite database — always go through the CLI.

All agent commands live under `oraculo tools`. Output is always JSON unless noted.
Errors return `{"error": "...", "details": {...}}`.

---

## Sessions

Sessions track the lifecycle of a skill invocation. Every skill that persists state must open a session and close it.

```bash
# Open a session
oraculo tools session init --type <epic|story|plan|execute|validate> --epic <name>
# → {"id": "<uuid>", "type": "...", "epic": "...", "status": "active", "created": true}

# Check for active session
oraculo tools session status --type <type> --epic <name>
# → {"active": false} or {"active": true, "id": "...", "current_phase": "...", "completed_phases": [...]}

# Read full session state (all phases)
oraculo tools session state --session <id>
# → {"id": "...", "current_phase": "...", "phases": {"setup": {...}, ...}}

# Close a session
oraculo tools session close --session <id>
# Close abandoned:
oraculo tools session close --session <id> --reason abandoned
```

---

## Phase Lifecycle

Each session advances through phases. Persist outputs at every phase gate.

```bash
echo '{"key": "value"}' | oraculo tools phase complete <phase-name> --session <id>
# → {"phase": "setup", "completed": true, "next": "decomposition"}
# → {"phase": "artifact", "completed": true, "session_closed": true}  ← last phase
```

Valid phases by session type:
- **epic**: setup → reframing → divergence → codebase → convergence → assumptions → stress-test → exit-gate → artifact
- **story**: setup → reframing → assumptions → exit-gate → artifact → approval
- **plan**: setup → decomposition → dependencies → design → optimization → artifact
- **execute**: setup → team-assembly → monitoring → completion
- **validate**: setup → qa-dispatch → verdict

---

## Epics

```bash
# Create
oraculo tools epic init <name> --description "..."
# → {"name": "...", "path": "...", "created": true}

# Save requirements markdown (reads stdin)
echo "# Requirements..." | oraculo tools epic save <name>

# Read requirements (returns raw markdown, not JSON)
oraculo tools epic get <name>

# List all epics
oraculo tools epic list
# → [{"id": 1, "name": "...", "description": "...", "approval_status": "none"}, ...]

# Create a new version + design approval request
echo "# Requirements v2..." | oraculo tools epic version <name>
# → {"version_id": 3, "approval_id": "<uuid>"}

# List versions
oraculo tools epic versions <name>
# → [{"id": 1, "number": 1, "created_at": "..."}, ...]

# Update description
oraculo tools epic update <name> --description "..."

# Delete
oraculo tools epic delete <name>
```

---

## Stories

```bash
# Create
oraculo tools story init <name> --epic <epic-name> --description "..."

# Save requirements markdown (reads stdin)
echo "# Story requirements..." | oraculo tools story save <name> --epic <epic-name>

# Read requirements (returns raw markdown)
oraculo tools story get <name> --epic <epic-name>

# List stories for an epic
oraculo tools story list --epic <epic-name>

# Create a new version + design approval request
echo "# Story v2..." | oraculo tools story version <name> --epic <epic-name>
# → {"version_id": 2, "approval_id": "<uuid>"}

# Update approval status
oraculo tools story update-status <name> --epic <epic-name> \
  --status <none|pending|approved|rejected|needs_revision>

# List versions
oraculo tools story versions <name> --epic <epic-name>

# Delete
oraculo tools story delete <name> --epic <epic-name>
```

---

## Tasks

Tasks belong to a story. They follow a strict status machine: `pending → in_progress → completed | failed`.

```bash
# Create
oraculo tools task init <name> \
  --epic <epic-name> --story <story-name> \
  --description "..." \
  --depends-on <other-task-name>   # repeatable

# List tasks for a story
oraculo tools task list --epic <epic-name> --story <story-name>

# Read a task
oraculo tools task get <name> --epic <epic-name> --story <story-name>

# Start (pending → in_progress)
oraculo tools task start <name> --epic <epic-name> --story <story-name>

# Complete (in_progress → completed) — reads JSON from stdin
echo '{
  "summary": "Implemented login flow",
  "logs": "...",
  "skills_used": ["tdd"],
  "files_modified": ["src/auth.go", "src/auth_test.go"]
}' | oraculo tools task complete <name> --epic <epic-name> --story <story-name>

# Fail (in_progress → failed)
oraculo tools task fail <name> \
  --epic <epic-name> --story <story-name> \
  --reason "Dependency X is missing"

# Delete
oraculo tools task delete <name> --epic <epic-name> --story <story-name>
```

---

## Approvals

Approval gates are how humans review and authorize agent work before it proceeds.

```bash
# Request an approval (reads content from stdin)
echo "## Design proposal..." | oraculo tools approval request \
  --type <design|qa-escalation|execution-plan> \
  --epic <epic-name> \
  --story <story-name>    # optional, for story-scoped approvals
# → {"id": "<uuid>", "type": "...", "status": "pending", "requested_at": "..."}

# Check approval status
oraculo tools approval status <id>
# → {"id": "...", "type": "...", "status": "pending|approved|rejected|needs_revision", ...}

# List approvals
oraculo tools approval list
oraculo tools approval list --pending   # pending only

# Record a verdict (usually done by the human via UI, not agents)
oraculo tools approval verdict <id> \
  --verdict <approved|rejected|needs_revision> \
  --comment "Looks good"
```

---

## MCP Tools

MCP tools are blocking/interactive operations available through the Oraculo MCP server. They are used by agents that need to wait for human decisions.

### request_approval

Blocks until a human verdict is recorded. Use this instead of polling `approval status` in a loop.

**Input:**
- `type` — approval type (`qa-escalation`, `execution-plan`, `design`)
- `content` — the document or artifact awaiting review
- `epic_id` — optional numeric epic ID
- `story_id` — optional numeric story ID

**Output (when verdict is recorded):**
```json
{
  "id": "uuid",
  "type": "design",
  "epic_id": 1,
  "story_id": 3,
  "status": "rejected",
  "content": "## Original document...",
  "comment": "General rejection reason",
  "comments": [
    { "selected_text": "The system should...", "comment": "Contradicts requirement X" }
  ]
}
```

- `comment`: general comment from the rejection (may be empty if inline comments exist)
- `comments[]`: inline comments tied to specific text selections
- If `approved`, `comments` is an empty array

### approval_status

Non-blocking check of an approval's current state. Returns the same enriched format as `request_approval`.

---

## Design Artifacts

Design artifacts live alongside story requirements as `design.md`.

```bash
# Save design markdown (reads stdin)
echo "## Architecture Decision..." | oraculo tools design save <story-name> --epic <epic-name>

# Read design
oraculo tools design get <story-name> --epic <epic-name>
```

---

## Knowledge (Memory)

Persist codebase insights so future agents don't rediscover the same information.

```bash
# Store a finding
oraculo tools memory store \
  --domain "auth-service" \
  --category <pattern|convention|constraint|dependency|test|architecture> \
  --finding "JWT tokens expire in 24h; refresh tokens stored in HttpOnly cookies" \
  --source "src/auth/token.go" \
  --confidence <high|medium|low>

# Search findings
oraculo tools memory search "JWT expiry"
oraculo tools memory search "JWT expiry" --domain "auth-service" --limit 10

# List domains
oraculo tools memory domains
```

Categories:
- `pattern` — recurring code patterns
- `convention` — project standards (naming, structure)
- `constraint` — technical or architectural limits
- `dependency` — external or internal deps
- `test` — test coverage notes
- `architecture` — system design insights

---

## Validations (QA)

```bash
# Record a QA validation verdict
echo '{"critical": ["Login fails on Safari"]}' | oraculo tools validation save <story-name> \
  --epic <epic-name> \
  --verdict <approved|rejected>
```

---

## Document Reviews

```bash
# Create a review for a version
oraculo tools review create <version-id> \
  --type <epic|story> \
  --verdict <approved|rejected> \
  --comment "Requirements are clear"

# Read a review
oraculo tools review get <review-id>

# List reviews for a version
oraculo tools review list <version-id> --type <epic|story>
```

---

## Project Status

```bash
# Human-readable summary (not JSON)
oraculo status
```

---

## Key Patterns

**Piping markdown to stdin:**
```bash
cat <<'EOF' | oraculo tools epic save my-epic
# Requirements
...
EOF
```

**Reading the result of a command:**
```bash
SESSION=$(oraculo tools session init --type plan --epic my-epic | jq -r '.id')
```

**Checking if a story has requirements before proceeding:**
```bash
oraculo tools story get my-story --epic my-epic 2>/dev/null || echo "no requirements yet"
```

**File layout on disk** (for context — never read directly):
```
.oraculo/
├── oraculo.db
├── config.json
└── epics/<epic-name>/
    ├── requirements.md
    └── stories/<story-name>/
        ├── requirements.md
        └── design.md
```
