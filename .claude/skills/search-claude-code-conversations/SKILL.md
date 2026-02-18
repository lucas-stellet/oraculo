---
name: search-claude-code-conversations
description: Search through Claude Code conversation history across all projects.
allowed-tools:
  - Bash
---

# Search Claude Code Conversations

Search through Claude Code conversation history across all projects.

## When to Apply

Use this skill when the user says:
- "search conversations", "find where we discussed", "which session"
- "conversation history", "buscar conversa", "achar conversa"
- "find that conversation", "em qual conversa", "procurar sessao"

## Data Layout

- **Quick index**: `~/.claude/history.jsonl` — one JSON line per user message
  - Fields: `display` (message text), `project` (absolute path), `sessionId`, `timestamp` (epoch ms)
- **Full transcripts**: `~/.claude/projects/-{url-encoded-path}/{sessionId}.jsonl`
  - Each line: `{type, message, sessionId, timestamp}`
  - Types: `user`, `assistant`, `progress`, `file-history-snapshot`

## Process

### 1. Parse the user's query

Extract from the arguments:
- **keyword**: the search term (required)
- **project filter**: if the user mentions a project name, use it to filter the `project` field
- **date filter**: if the user mentions a time range ("today", "last week", "last N days"), calculate the epoch ms cutoff
- **deep**: if the user says "deep search" or "search assistant messages", search full transcripts instead of just history

### 2. Quick Search (default)

Search `history.jsonl` — this is fast because it's a single file with user message previews.

```bash
python3 -c "
import json, sys, os
from datetime import datetime

keyword = sys.argv[1].lower()
project_filter = sys.argv[2].lower() if len(sys.argv) > 2 and sys.argv[2] else ''
days_back = int(sys.argv[3]) if len(sys.argv) > 3 and sys.argv[3] else 0
cutoff = (datetime.now().timestamp() - days_back * 86400) * 1000 if days_back else 0

seen = set()
results = []
for line in open(os.path.expanduser('~/.claude/history.jsonl')):
    entry = json.loads(line)
    display = entry.get('display', '')
    project = entry.get('project', '')
    ts = entry.get('timestamp', 0)
    sid = entry.get('sessionId', '')

    if not display or display.startswith('/') or not keyword in display.lower():
        continue
    if project_filter and project_filter not in project.lower():
        continue
    if cutoff and ts < cutoff:
        continue

    key = (sid, display[:80])
    if key in seen:
        continue
    seen.add(key)

    parts = [p for p in project.rstrip('/').split('/') if p] if project else []
    proj_short = '/'.join(parts[-2:]) if len(parts) >= 2 else (parts[0] if parts else '?')
    date = datetime.fromtimestamp(ts / 1000).strftime('%Y-%m-%d %H:%M')
    preview = display[:100].replace('\n', ' ')
    results.append((date, proj_short, preview, sid))

results.sort(key=lambda x: x[0], reverse=True)
for date, proj, preview, sid in results[:25]:
    print(f'{date}  {proj[:25]:<25}  {preview[:60]:<60}  {sid[:8]}')

if not results:
    print('No results found.')
" "KEYWORD" "PROJECT_FILTER" "DAYS_BACK"
```

Replace `KEYWORD`, `PROJECT_FILTER` (empty string if none), and `DAYS_BACK` (0 if none) with actual values.

### 3. Deep Search (when quick search is insufficient)

Search full session transcripts for matches in both user AND assistant messages. Use this when:
- The keyword appears in Claude's responses but not in user messages
- The quick search returned no results but the user is sure the conversation exists

```bash
python3 -c "
import json, sys, os, glob
from datetime import datetime

keyword = sys.argv[1].lower()
project_filter = sys.argv[2].lower() if len(sys.argv) > 2 and sys.argv[2] else ''
base = os.path.expanduser('~/.claude/projects')

results = []
for proj_dir in sorted(glob.glob(os.path.join(base, '*'))):
    proj_name = os.path.basename(proj_dir).replace('-', '/')
    if project_filter and project_filter not in proj_name.lower():
        continue
    for f in sorted(glob.glob(os.path.join(proj_dir, '*.jsonl')), key=os.path.getmtime, reverse=True):
        sid = os.path.splitext(os.path.basename(f))[0]
        found_lines = []
        try:
            for line in open(f):
                entry = json.loads(line)
                if entry.get('type') not in ('user', 'assistant'):
                    continue
                msg = entry.get('message', {})
                content = msg.get('content', '')
                if isinstance(content, list):
                    content = ' '.join(c.get('text', '') for c in content if isinstance(c, dict))
                if keyword in content.lower():
                    role = entry['type']
                    ts = entry.get('timestamp', '')
                    preview = content[:120].replace('\n', ' ')
                    found_lines.append((ts, role, preview))
        except: continue
        if found_lines:
            parts = [p for p in proj_name.rstrip('/').split('/') if p]
            proj_short = '/'.join(parts[-2:]) if len(parts) >= 2 else (parts[0] if parts else '?')
            results.append((proj_short, sid, found_lines[:3]))
        if len(results) >= 15:
            break
    if len(results) >= 15:
        break

for proj, sid, matches in results:
    print(f'\\n--- {proj[:30]} | session: {sid[:8]}')
    for ts, role, preview in matches:
        date = ''
        if ts:
            try: date = datetime.fromisoformat(ts.replace('Z','+00:00')).strftime('%Y-%m-%d %H:%M')
            except: pass
        print(f'  [{role:9}] {date}  {preview[:80]}')

if not results:
    print('No results found in full transcripts.')
" "KEYWORD" "PROJECT_FILTER"
```

### 4. Present Results

After running the search:
1. Show the results table to the user
2. If the user wants to resume a session, provide: `claude --resume <sessionId>`
3. If too many results, suggest narrowing with a project filter or date range

### 5. List Projects

If the user asks "which projects" or "list projects", show available projects:

```bash
python3 -c "
import os, glob
base = os.path.expanduser('~/.claude/projects')
for d in sorted(glob.glob(os.path.join(base, '*'))):
    name = os.path.basename(d).replace('-', '/')
    count = len(glob.glob(os.path.join(d, '*.jsonl')))
    if count > 0:
        parts = [p for p in name.rstrip('/').split('/') if p]
        short = '/'.join(parts[-2:]) if len(parts) >= 2 else (parts[0] if parts else '?')
        print(f'{count:3} sessions  {short}')
"
```

## Rules

- **Always start with quick search** — it's fast and covers most cases
- **Fall back to deep search** only if the user explicitly asks or quick search returned no results
- **Limit output to 25 results** — suggest filters if more exist
- **Never modify** conversation files — read-only access
- **Decode project paths** — convert URL-encoded directory names to readable paths for display
