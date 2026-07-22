# Contributing to masterfabric-go

Thank you for your interest in contributing to masterfabric-go! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

### Reporting Bugs

- Use the [Bug Report template](.github/ISSUE_TEMPLATE/bug_report.md)
- Include steps to reproduce, expected vs actual behavior
- Provide environment details (OS, Go version, Docker version)
- Include relevant logs and configuration

### Suggesting Features

- Use the [Feature Request template](.github/ISSUE_TEMPLATE/feature_request.md)
- Clearly describe the problem and proposed solution
- Consider alternatives and implementation approach

### Pull Requests

1. **Fork the repository** and create a feature branch
   ```bash
   git checkout -b feature/amazing-feature
   ```

2. **Follow coding standards**
   - Run `./scripts/lint.sh` before committing
   - Follow Go best practices ([Effective Go](https://go.dev/doc/effective_go))
   - Write clear commit messages (see [Conventional Commits](https://www.conventionalcommits.org/))

3. **Write tests**
   - Add unit tests for new functionality
   - Ensure all tests pass: `./scripts/test.sh`
   - Aim for good test coverage

4. **Update documentation**
   - Update README.md if needed
   - Add/update code comments
   - Update CHANGELOG.md for user-facing changes

5. **Submit PR**
   - Use the [PR template](.github/pull_request_template.md)
   - Link related issues
   - Request reviews from maintainers

## Development Setup

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make (optional, for Makefile targets)

### Quick Start

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/masterfabric-go.git
cd masterfabric-go

# Start infrastructure
./dev.sh infra

# Run with hot-reload
./dev.sh server
```

### Running Tests

```bash
# Run all tests
./scripts/test.sh

# Run with coverage
./scripts/test.sh -cover

# Run specific package
./scripts/test.sh ./internal/shared/events
```

### Code Quality

```bash
# Format code
gofmt -w .

# Run linters
./scripts/lint.sh

# Auto-fix issues
./scripts/lint.sh -fix
```

## Project Structure

```
cmd/server/              - Application entry point
internal/
  shared/                - Cross-cutting concerns
  domain/                - Domain layer (entities, interfaces)
  application/          - Use cases and DTOs
  infrastructure/        - External implementations
  gateway/               - API Gateway pipeline
scripts/                 - Utility scripts
deployments/             - Docker and deployment configs
```

## Architecture Guidelines

- **Domain Layer**: Zero external dependencies, pure business logic
- **Application Layer**: Orchestrates domain logic, defines use cases
- **Infrastructure Layer**: Implements domain interfaces (repositories, services)
- **Dependency Rule**: Dependencies point inward (infrastructure → application → domain)

## Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:
- `feat(iam): add user registration endpoint`
- `fix(gateway): correct rate limiting calculation`
- `docs(readme): update quick start guide`

## License

By contributing, you agree that your contributions will be licensed under the AGPL v3.0 license.

## Questions?

- Open a [Discussion](https://github.com/masterfabric-go/masterfabric-go/discussions)
- Check existing [Issues](https://github.com/masterfabric-go/masterfabric-go/issues)

Thank you for contributing! 🎉
