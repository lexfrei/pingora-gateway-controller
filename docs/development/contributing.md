# Contributing

Guidelines for contributing to Pingora Gateway Controller.

## Getting Started

1. Fork the repository
2. Clone your fork
3. Set up [development environment](setup.md)
4. Create a feature branch
5. Make changes and test
6. Submit a pull request

## Development Workflow

### Branch Naming

Use descriptive branch names:

- `feat/add-tls-support`
- `fix/route-sync-error`
- `docs/update-architecture`
- `refactor/simplify-builder`

### Commit Messages

Use semantic commit messages:

```text
type(scope): brief description

Optional longer explanation of what was changed and why.

Co-Authored-By: Claude <noreply@anthropic.com>
```

**Types**:

| Type | Description |
|------|-------------|
| `feat` | New features |
| `fix` | Bug fixes |
| `docs` | Documentation changes |
| `style` | Code style changes |
| `refactor` | Code refactoring |
| `test` | Test changes |
| `chore` | Maintenance tasks |
| `ci` | CI/CD changes |
| `perf` | Performance improvements |

**Examples**:

```text
feat(controller): add GRPCRoute support

Implement GRPCRouteReconciler to watch and sync GRPCRoute resources
to Pingora proxy. Includes service/method matching and header-based
routing.

Co-Authored-By: Claude <noreply@anthropic.com>
```

```text
fix(sync): handle connection timeout gracefully

Add retry logic with exponential backoff when gRPC connection
to Pingora proxy times out.

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Code Standards

### Go Code

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Pass `golangci-lint run` without errors
- Write tests for new functionality
- Use meaningful variable and function names

### Error Handling

```go
// Good
if err != nil {
    return errors.Wrap(err, "failed to sync routes")
}

// Bad
if err != nil {
    return err
}
```

### Logging

```go
// Good
logger.Info("syncing routes",
    "gateway", gateway.Name,
    "routeCount", len(routes),
)

// Bad
logger.Info(fmt.Sprintf("syncing %d routes for gateway %s", len(routes), gateway.Name))
```

## Testing Requirements

### Unit Tests

- All new code must have tests
- Maintain or improve coverage
- Use table-driven tests

### Test Patterns

```go
func TestFeature(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name     string
        input    InputType
        expected OutputType
    }{
        {name: "case 1", input: ..., expected: ...},
        {name: "case 2", input: ..., expected: ...},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // test logic
        })
    }
}
```

### Running Tests

```bash
# All tests
go test -race ./...

# With coverage
go test -race -coverprofile=coverage.out ./...
```

## Pull Request Process

### Before Submitting

1. Run tests: `go test -race ./...`
2. Run linter: `golangci-lint run`
3. Update documentation if needed
4. Add tests for new functionality

### PR Title

Use semantic format:

```text
feat(scope): add feature description
fix(scope): fix bug description
docs(scope): update documentation
```

### PR Description

Use the template from `.github/pull_request_template.md`:

- Summary of changes
- Related issues (if any)
- Testing performed
- Checklist completion

### Review Process

1. Automated checks must pass
2. Code review by maintainer
3. Address review feedback
4. Maintainer approval
5. Squash and merge

## Documentation

### Updating Docs

- Update docs alongside code changes
- Run `mkdocs build --strict` to verify
- Run `markdownlint-cli2 '**/*.md'` to lint

### Documentation Style

- Use clear, concise language
- Include code examples
- Use admonitions for notes/warnings
- Keep examples up to date

## Helm Chart Changes

### Testing Changes

```bash
# Run unit tests
helm unittest charts/pingora-gateway-controller

# Lint chart
helm lint charts/pingora-gateway-controller

# Update README
helm-docs charts/pingora-gateway-controller
```

### Values Changes

- Update `values.yaml` with comments
- Add tests for new values
- Update documentation

## Issue Reporting

### Bug Reports

Include:

- Controller version
- Kubernetes version
- Gateway API version
- Steps to reproduce
- Expected vs actual behavior
- Logs and error messages

### Feature Requests

Include:

- Use case description
- Proposed solution
- Alternative approaches considered

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Follow project guidelines

## Getting Help

- Check existing [issues](https://github.com/lexfrei/pingora-gateway-controller/issues)
- Ask questions in issue discussions
- Read the [documentation](../index.md)

## Next Steps

- Learn about [Testing](testing.md)
- Understand the [Architecture](architecture.md)
