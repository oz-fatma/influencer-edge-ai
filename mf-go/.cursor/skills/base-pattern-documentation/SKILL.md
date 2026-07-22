---
name: base-pattern-documentation
description: Document Go architectural patterns, base types, and conventions for the masterfabric_go project. Use when explaining architecture decisions, documenting base patterns, or creating architecture guides.
---

# Base Pattern Documentation Skill

Use this skill when documenting Go architectural patterns or explaining how base types and conventions work in the masterfabric_go project.

## When to Use

- Documenting a new architectural component (handler, use case, repository)
- Explaining Clean Architecture layers and dependency rules
- Creating guides for new contributors
- Documenting the Dynamic Handler System

## Documentation Template

### For Domain Models

```markdown
## {EntityName}

**Location**: `internal/domain/{context}/model/{entity}.go`

**Purpose**: {What this entity represents in the domain}

**Fields**:
| Field | Type | Description |
|-------|------|-------------|
| ID | uuid.UUID | Unique identifier |
| ... | ... | ... |

**Business Rules**:
- {Rule 1}
- {Rule 2}

**Related**: {Related entities, events, repositories}
```

### For Use Cases

```markdown
## {UseCaseName}

**Location**: `internal/application/{context}/usecase/{file}.go`

**Purpose**: {What business operation this performs}

**Input**: `dto.{InputDTO}`
**Output**: `*dto.{OutputDTO}, error`

**Flow**:
1. Validate input
2. Execute business logic
3. Persist changes
4. Publish domain event
5. Return result

**Events Published**: `{EventName}` on `{TopicName}`

**Error Cases**:
- `ErrNotFound`: {when}
- `ErrBadRequest`: {when}
```

### For Handlers

```markdown
## {HandlerName}

**Location**: `internal/infrastructure/http/handler/{context}/handler.go`

**Routes**:
| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| POST | /entities | CreateEntity | Create new entity |
| GET | /entities | ListEntities | List with pagination |

**Dependencies**: {Use cases, repositories}

**Headers Required**: `X-Organization-ID`, `Authorization`
```

## Output Format

- Use tables for structured data
- Include code examples from the actual codebase
- Reference specific file paths
- Document the WHY, not just the WHAT
- Keep concise (under 100 lines per component)
