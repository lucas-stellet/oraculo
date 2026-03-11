# Plan Design Patterns

## Design Document Template

Use this structure. Omit sections not relevant to the current story.

### 1. Context
- Parent epic and story reference
- Tech stack summary
- Key existing patterns discovered by codebase research agent

### 2. Data Model
- Schema definitions (tables, columns, types, constraints)
- Relationships and indexes
- Migration strategy

### 3. Contracts
- Server actions / API endpoints (name, inputs, outputs, errors)
- Validation rules (schema shape, business constraints)
- Error handling strategy

### 4. Component Architecture
- Component tree and hierarchy
- State management approach
- Data flow (server → client, form submission, mutations)

### 5. Technical Decisions
- Libraries or tools chosen and rationale (from web research)
- Patterns to follow (from codebase research)
- Patterns to avoid and why

### 6. File Map
- Files to be created
- Files to be modified
- Where each piece lives in the project structure

### 7. Testing Strategy
- Critical paths to test
- Test boundaries (unit vs integration)
- Test data approach

## Rules

- Omit sections that don't apply (backend-only task → skip Component Architecture)
- Scale each section to complexity (2 lines or 2 paragraphs, both acceptable)
- No implementation code — decisions, contracts, and shapes only
- Findings from research agents must be visible in the relevant sections
