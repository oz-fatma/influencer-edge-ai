# Contributing to InfluencerEdge AI

Thank you for contributing. This monorepo includes:

- `frontend/` — Next.js app (Vercel)
- `backend/` — legacy Gin API
- `mf-go/` — Go API (Render)
- `llm-service/` — local Ollama + Caddy proxy for server-side inference

Open a pull request against the default branch. Use the [PR template](.github/PULL_REQUEST_TEMPLATE.md) when you submit.

---

## Commit Message Guidelines

All commit messages must be written in **English** and follow [Conventional Commits](https://www.conventionalcommits.org/):

```text
<type>: <description>
```

- Use the imperative mood in the description (e.g. “add handler”, not “added handler”).
- Keep the subject line concise (≤ 72 characters when possible).
- Optional body: explain **why**, not only what changed.

### Types

| Type | When to use |
|------|-------------|
| `feat` | New user-facing behavior or API capability |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `chore` | Tooling, deps, config, CI — no production logic change |
| `test` | Adding or updating tests |

### Examples from this project

```text
feat: read Render DATABASE_URL and REDIS_URL in mf-go config
```

```text
fix: listen on Render PORT env variable
```

```text
fix: avoid readiness panic when db or redis is unavailable
```

```text
feat: route Analyze through Render mf-go to local LLM proxy
```

```text
feat: fall back to WebLLM when server-side analyze times out
```

```text
refactor: replace mlc-llm Docker stack with Ollama in llm-service
```

```text
docs: add Ollama setup and update Caddy proxy demo for port 11434
```

```text
chore: add PR template and commit message guidelines
```

```text
test: add WebLLM engine singleton reload guard
```

### What to avoid

- Mixed-language subjects (Turkish + English in the same line)
- Vague messages: `update`, `fix stuff`, `WIP`
- Prefixes outside the allowed types unless agreed with maintainers

Future commits in this repository — including automated or assistant-generated commits — should follow this format.
