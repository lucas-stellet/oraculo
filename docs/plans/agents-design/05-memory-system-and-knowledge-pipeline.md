# Research Prompt 5: Memory System and Knowledge Pipeline

## Context: Oraculo Agent Philosophy

Memory has three layers: working (ephemeral, per-task), episodic (immutable event log), and semantic (validated knowledge, versioned). Nothing enters memory without a contract. Wisdom is curated through a pipeline, not accumulated raw. Writes are serialized through the CLI; reads are parallel.

Key beliefs from the philosophy:
- Three distinct memory tiers with different rules; mixing them corrupts all three
- CLI validates every write (schema mandatory, provenance mandatory)
- Contradictions stored as versioned proposals, not forced into premature consensus
- Only QA-validated knowledge promoted to operational truth
- Curation pipeline: episode -> score -> reflection -> validated promotion -> relational linking

Architecture constraints:
- SQLite is the storage engine (WAL mode for concurrent reads)
- CLI (Go binary) is the single write gateway
- FTS5 for full-text search
- Event sourcing: append-only, immutable event log
- All agents access memory exclusively through CLI commands

## Research Mission

Design the memory system and knowledge pipeline for Oraculo. I need to understand how to implement the three-tier architecture in SQLite, how to build the curation pipeline, and how to make memory useful for agents without overwhelming their context windows.

## Guiding Questions

1. **SQLite schema design**: What's the concrete schema for three-tier memory? Tables for events (episodic), knowledge items (semantic), and how working memory is assembled? How to model provenance, versioning (valid_from, valid_to, supersedes_id), and status (proposed/validated/deprecated)? Foreign key relationships between tiers?

2. **Event sourcing implementation**: How to implement append-only event sourcing in SQLite? What's the event schema (event_type, agent_id, run_id, phase, payload, timestamp)? How to derive current state from event stream? How to handle event stream growth over long-lived projects? Compaction strategies?

3. **Working set builder**: How does the CLI assemble working memory for a specific task? What query combines FTS5 search, recency, relevance scoring, and importance to select the right context? How to enforce size limits (MVI: concepts < 100 lines, guides < 150 lines)? How to test that the working set is actually useful?

4. **Curation pipeline**: How to implement the episode-to-wisdom pipeline? What scores episodes (impact, risk, recurrence)? Who triggers consolidation -- the orchestrator, a dedicated curation agent, or a scheduled process? How to generate "reflections" (consolidated insights) from raw episodes? How to link reflections relationally (depends_on, conflicts_with, supersedes)?

5. **Concurrency and write serialization**: How to implement the "N readers + 1 writer" pattern with SQLite WAL? What timeout/retry strategy for SQLITE_BUSY? How to keep write transactions short? How to handle the case where multiple agents want to write simultaneously (queue? reject? batch)?

6. **Search strategy evolution**: How to start with FTS5/BM25 and evolve to hybrid search? What's the migration path to add sqlite-vec for embeddings? How to implement Reciprocal Rank Fusion (RRF) for combining lexical and semantic results? When does the system need graph-based search (entity-relation queries)?

## Expected Output

For each finding, provide:
- **Pattern/Concept**: Name
- **Source**: URL, paper, repository, or documentation
- **Summary**: 2-3 sentences describing the implementation approach
- **Applicability to Oraculo**: How it maps to SQLite + CLI + FTS5 architecture
- **Key design decision**: The one implementation choice this informs

Focus on practical SQLite-based implementations (2024-2026). Include patterns from knowledge management systems, event sourcing frameworks, and any agent memory implementations that use structured databases over vector stores.

## Reference Implementations to Study

Study these projects for concrete memory and state patterns:

- **Gastown** (https://github.com/steveyegge/gastown) -- Multi-agent workspace manager with git-backed state persistence. Work state stored as structured data with typed IDs (prefix + alphanumeric). Agent identity persists across ephemeral sessions. Relevant for: git-backed state management, structured work state schemas, persistent agent context across sessions.

- **GSD / Get Shit Done** (https://github.com/gsd-build/get-shit-done) -- Features state management behind the scenes and codebase mapping that persists project patterns. `/gsd:map-codebase` creates persistent knowledge about stack, architecture, and conventions. Relevant for: project knowledge extraction and persistence, codebase-aware context that survives sessions.
