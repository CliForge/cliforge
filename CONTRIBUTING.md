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

### Commit Body Best Practices

The commit body should explain **why** the change was made, not **what** changed (the diff shows that).

**Too Verbose (BAD)**:
```
docs(design): add comprehensive architecture designs for ROSA requirements

Add six new architecture design documents addressing gaps identified
in ROSA CLI analysis:

1. gap-analysis-rosa-requirements.md
   - Comprehensive gap analysis (6 critical, 6 important gaps)
   - Priority matrix and phased implementation plan
   - 40% requirements covered, 60% needs work

2. plugin-architecture.md
   - Plugin system for external tool integration
   - Three plugin types: built-in, binary, WASM
   - Security model with permissions and sandboxing
   - x-cli-plugin extension specification

3. workflow-orchestration.md
   - Multi-step workflow execution engine
   - Six step types: api-call, plugin, conditional, loop, wait, parallel
   - DAG-based execution with dependencies
   - Rollback support and error handling

(Lists every file and details - this is what the diff is for!)
```

**Concise (GOOD)**:
```
docs(design): complete architecture for ROSA-like functionality

Add comprehensive design documents addressing gaps identified in ROSA
CLI analysis: plugin architecture, workflow orchestration, file
operations, progress/streaming, and state management.

Gap analysis shows 40% requirements already covered, 60% needs new
architectural components. Phased implementation plan targets v0.8-1.0.
```

**Why the second is better:**
- Explains the impact (complete architecture, identifies gaps)
- Provides context (40/60 split, phased plan)
- 3-4 lines, not 20+ lines
- Doesn't enumerate files (visible in diff)

### Guidelines

**DO:**
- Use present tense ("add feature" not "added feature")
- Be concise in description (50 chars or less)
- Use body for detailed explanation if needed
- Reference issues: "Closes #123"
- Focus on WHY, not WHAT (the diff shows what changed)
- Keep body to 3-4 lines maximum

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
- Write multi-paragraph essays in commit bodies

### Breaking Changes

For breaking changes, add `!` after type/scope and explain in footer:

```
feat(config)!: change DSL syntax for authentication

BREAKING CHANGE: auth.strategy is now auth.method in configuration
DSL. Update all config files accordingly.
```

## Development Workflow

### Fork Setup

CliForge uses a fork-based workflow:

```bash
# One-time setup: Fork the repository on GitHub, then:
git clone git@github.com:YOUR_USERNAME/cliforge.git
cd cliforge
git remote add upstream git@github.com:CliForge/cliforge.git

# Verify remotes
git remote -v
# origin    git@github.com:YOUR_USERNAME/cliforge.git
# upstream  git@github.com:CliForge/cliforge.git
```

### Standard Go Development

```bash
# Run tests
go test ./...

# Build
go build ./...

# Run linter
golangci-lint run

# Install pre-commit hooks
pip install pre-commit
pre-commit install
```

### Feature Development

1. **Sync with upstream:**
   ```bash
   git checkout main
   git pull upstream main
   git push origin main
   ```

2. **Create feature branch:**
   ```bash
   git checkout -b feat/my-feature
   ```

3. **Make changes with atomic commits:**
   - Follow commit message guidelines
   - One logical change per commit
   - Test each commit

4. **Push to your fork:**
   ```bash
   git push origin feat/my-feature
   ```

5. **Create PR to upstream:**
   ```bash
   gh pr create --repo CliForge/cliforge \
     --base main \
     --head YOUR_USERNAME:feat/my-feature \
     --title "feat(scope): description" \
     --body "PR description"
   ```

   Or visit: https://github.com/CliForge/cliforge/compare

6. **After PR is merged:**
   ```bash
   git checkout main
   git pull upstream main
   git push origin main
   git branch -D feat/my-feature
   ```

### Direct Commits (Maintainers Only)

For documentation-only changes or minor fixes, maintainers may commit directly:
```bash
# Make changes on main
git checkout main
git add [files]
git commit -m "docs: description"
git push upstream main
```

## Code Style

- Go 1.21+
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting
- Use `golangci-lint` for linting
- Write tests for new features
- Document exported types and functions

See project documentation for detailed guidelines.
