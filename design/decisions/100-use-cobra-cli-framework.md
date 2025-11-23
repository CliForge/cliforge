# ADR-100: Use Cobra for CLI Framework

## Status

Accepted

## Date

2025-01-11

## Context

CliForge generates command-line interfaces dynamically from OpenAPI specifications. We need a Go CLI framework that supports:

1. **Nested subcommands** - API endpoints map to hierarchical commands (`mycli users list`, `mycli config set`)
2. **Global and local flags** - Common flags like `--verbose` and `--output` should work everywhere
3. **Auto-generated help** - Users discover commands and flags via `--help`
4. **Shell completion** - Tab completion for commands, flags, and values
5. **Flexible flag parsing** - Support POSIX and GNU-style flags
6. **Programmatic command building** - Commands must be generated at runtime from OpenAPI spec
7. **Production-ready** - Stable, maintained, documented, proven at scale

### Alternatives Considered

We evaluated four major Go CLI frameworks:

#### 1. spf13/cobra (37K+ stars)
**Pros:**
- Industry standard - Used by kubectl, Docker, GitHub CLI, Hugo, Helm
- Perfect for nested subcommands and deep hierarchies
- Persistent flags (global flags inherited by all subcommands)
- Excellent auto-generated help and completion
- Viper integration for config files
- Programmatic API for dynamic command building
- Mature, battle-tested, active maintenance

**Cons:**
- Slightly more complex than simpler alternatives
- Opinionated structure (but matches our needs)

#### 2. urfave/cli (22K+ stars)
**Pros:**
- Simpler API than Cobra
- Good for single-command or flat CLIs
- Clean, minimal design

**Cons:**
- **Nested subcommands are clunky** - Not designed for deep hierarchies
- Global flags have issues with subcommand inheritance
- Less powerful command organization
- Weaker ecosystem for complex CLIs

#### 3. alecthomas/kong (2K+ stars)
**Pros:**
- Struct-tag based - Define CLI via Go structs
- Type-safe, minimal boilerplate
- Great for CLIs with known structure

**Cons:**
- **Struct-based definition doesn't fit dynamic generation** from OpenAPI
- Less flexible for programmatic command creation
- Smaller ecosystem

#### 4. mitchellh/cli (1.7K+ stars)
**Pros:**
- Simple, lightweight
- Used by Terraform, Vault, Consul

**Cons:**
- **No auto-generated help** (must write manually)
- **No shell completion** (must implement yourself)
- More boilerplate required

### Decision Criteria

| Feature | Cobra | urfave/cli | kong | mitchellh/cli |
|---------|-------|------------|------|---------------|
| Nested subcommands | ⭐⭐⭐ | ⭐ | ⭐⭐ | ⭐⭐ |
| Global flags | ⭐⭐⭐ | ⭐ | ⭐⭐ | ⭐ |
| Auto-generated help | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ❌ |
| Shell completion | ⭐⭐⭐ | ⭐ | ⭐⭐ | ❌ |
| Dynamic command building | ⭐⭐⭐ | ⭐⭐ | ⭐ | ⭐⭐ |
| Ecosystem/docs | ⭐⭐⭐ | ⭐⭐ | ⭐ | ⭐ |
| Industry adoption | ⭐⭐⭐ | ⭐⭐ | ⭐ | ⭐⭐ |

**Legend**: ⭐⭐⭐ = Excellent, ⭐⭐ = Good, ⭐ = Adequate, ❌ = Poor/Missing

### Real-World Validation

Major API-driven CLIs using Cobra:
- **kubectl** - Kubernetes CLI with massive complex hierarchy and deep nesting
- **docker** - Multi-level subcommands, global flags
- **gh** - GitHub CLI, API-driven (similar to CliForge!)
- **hugo** - Complex static site generator
- **helm** - Kubernetes package manager

**Key insight**: These are all API-driven or complex domain-specific CLIs - exactly like what CliForge generates.

## Decision

**CliForge uses `spf13/cobra` as the CLI framework.**

### Rationale

1. **Perfect for API-driven CLIs**
   - API endpoints naturally map to nested subcommands
   - Cobra excels at deep hierarchies (e.g., `mycli resources subresources action`)

2. **Programmatic Command Building**
   - Cobra's API supports dynamic command generation:
   ```go
   for path, pathItem := range spec.Paths.Map() {
       for method, operation := range pathItem.Operations() {
           cmd := &cobra.Command{
               Use:   generateCommandName(operation),
               Short: operation.Summary,
               Long:  operation.Description,
               Run:   generateHandler(operation),
           }
           // Add flags from parameters
           for _, param := range operation.Parameters {
               addFlag(cmd, param)
           }
           parentCmd.AddCommand(cmd)
       }
   }
   ```

3. **Persistent Flags for Global Options**
   ```go
   rootCmd.PersistentFlags().StringP("output", "o", "json", "Output format")
   rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
   ```
   These flags automatically work on ALL subcommands - critical for CliForge.

4. **Auto-Generated Help Matches OpenAPI**
   - Cobra generates help text from command metadata
   - Maps perfectly to OpenAPI's `summary`, `description`, and `operationId`
   ```bash
   $ mycli users --help
   Manage users

   Usage:
     mycli users [command]

   Available Commands:
     list        List all users
     create      Create a new user
     get         Get user by ID

   Flags:
     -h, --help   help for users

   Global Flags:
     -o, --output string   Output format (default "json")
     -v, --verbose         Verbose output
   ```

5. **Shell Completion**
   - Built-in support for bash, zsh, fish, powershell
   - Understands subcommands, flags, enum values, file paths

6. **Viper Integration**
   - Seamless config file support
   - Automatically handles: CLI flags > env vars > config file precedence
   ```go
   viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
   ```

7. **Industry-Proven**
   - Used by kubectl (as long as Kubernetes exists, Cobra will be maintained)
   - 37K+ stars, massive community
   - Even if abandoned (unlikely), it's stable and feature-complete

## Consequences

### Positive

✅ **Perfect feature match** - Cobra was designed for exactly our use case
✅ **Reduced implementation time** - Don't need to build command parsing, help generation, or completion
✅ **Familiar to users** - If they've used kubectl/docker/gh, they know the patterns
✅ **Excellent documentation** - Tutorials, guides, examples readily available
✅ **Future-proof** - Active maintenance guaranteed (kubectl dependency)
✅ **Type safety** - Cobra's programmatic API is type-safe
✅ **Extensible** - Can add custom validators, preprocessors, middleware

### Negative

⚠️ **Learning curve** - Team must learn Cobra's API (but excellent docs)
⚠️ **Opinionated** - Forces specific patterns (but they align with our needs)
⚠️ **Dependency** - Tied to Cobra's release cycle (but very stable)

### Neutral

ℹ️ **Slightly heavier** than minimalist frameworks - But features justify the weight
ℹ️ **Cobra + Viper** - Two dependencies instead of one (but official pairing)

### Migration Path

If we ever needed to switch frameworks (unlikely):
- Command structure is abstracted via our OpenAPI → Command generator
- Only the generator implementation would change
- User-facing CLI interface remains identical

### Future Flexibility

For simple use cases, we could optionally generate `urfave/cli`-based CLIs in a "simple mode", but for the primary use case (complex API-driven CLIs), Cobra is the clear choice.

---

**Version**: 1.0
**Last Updated**: 2025-01-11
**Supersedes**: cobra-framework-decision.md (informal decision document)
