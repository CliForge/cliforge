# Contributing to CliForge

## Commit Message Guidelines

CliForge follows the [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **chore**: Maintenance tasks (dependencies, configs)
- **refactor**: Code refactoring
- **test**: Test additions or modifications
- **perf**: Performance improvements

### Scopes

Project-specific scopes for CliForge:

- **cli**: CLI generation and runtime
- **config**: Configuration DSL and parsing
- **openapi**: OpenAPI spec processing
- **template**: Template engine and rendering
- **branding**: Branding and customization
- **auth**: Authentication strategies
- **update**: Self-update mechanism
- **cache**: Caching and XDG compliance
- **secrets**: Secrets detection and masking
- **deprecation**: Deprecation handling
- **repo**: Repository structure changes

### Examples

**Good:**
```
feat(cli): add hybrid command style support

Implements subcommand, flag, and hybrid CLI styles based on
OpenAPI spec structure and user configuration.
```

```
fix(openapi): handle missing x-cli-* extensions gracefully

Default values now applied when OpenAPI spec lacks custom
CLI extensions to prevent generation failures.
```

```
chore(deps): update cobra to v1.8.0
```

**Bad:**
```
🎉 ADDED COOL NEW FEATURE!!! (with emojis and caps)
```

```
Updated files - Changed 50 files, 200 lines added, files:
src/foo.go, src/bar.go, ... (file listings)
```

```
chore(repo): rename module to cliforge

Changes:
- go.mod: Update module name
- Documentation: Explain installation methods
- CLI help text updated

Benefits:
- Clear branding
- Better discoverability
- Module name matches project

(Too verbose - lists files, explains changes, documents benefits)
```

**Better:**
```
chore(repo): rename module to cliforge

Module name now matches project branding for consistency
with documentation and user expectations.

(Concise - explains why, brief technical detail, no file listing)
```

### Guidelines

**DO:**
- Use present tense ("add feature" not "added feature")
- Be concise in description (50 chars or less)
- Use body for detailed explanation if needed
- Reference issues: "Closes #123"

**DON'T:**
- Use emojis in commit messages
- Use ALL CAPS
- List changed files (git does this)
- Include detailed statistics (lines changed, etc.)
- Add meta-commentary ("Generated with...", "Co-Authored-By...")
- Document "benefits" or justifications (focus on the technical change)
- Include irrelevant context (stars, contributors, popularity metrics)
- Enumerate what changed (the diff shows this)
- Explain file-by-file changes (git diff does this)

### Breaking Changes

For breaking changes, add `!` after type/scope and explain in footer:

```
feat(config)!: change DSL syntax for authentication

BREAKING CHANGE: auth.strategy is now auth.method in configuration
DSL. Update all config files accordingly.
```

## Development Workflow

### Standard Go Development

```bash
# Clone the repository
git clone https://github.com/cliforge/cliforge.git
cd cliforge

# Run tests
go test ./...

# Build
go build ./...

# Run linter
golangci-lint run
```

### Feature Development

1. **Create feature branch:**
   ```bash
   git checkout -b feat/my-feature
   ```

2. **Make changes with atomic commits:**
   - Follow commit message guidelines
   - One logical change per commit
   - Test each commit

3. **Push and create PR:**
   ```bash
   git push origin feat/my-feature
   ```

4. **After PR is merged:**
   ```bash
   git checkout main
   git pull origin main
   git branch -D feat/my-feature
   ```

## Code Style

- Go 1.21+
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting
- Use `golangci-lint` for linting
- Write tests for new features
- Document exported types and functions

See project documentation for detailed guidelines.
