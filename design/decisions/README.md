# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records for CliForge.

## What are ADRs?

Architecture Decision Records (ADRs) document significant architectural and technology decisions. Each ADR captures:

- **Context**: Why we're making this decision
- **Decision**: What we decided to do
- **Consequences**: What becomes easier or harder

ADRs are **lightweight**, **immutable** (except for status), and **numbered** for easy reference.

## Format

See **ADR-000** for the complete format specification.

**Template**:

```markdown
# ADR-XXX: Title (Short present tense imperative phrase)

## Status
[Proposed | Accepted | Deprecated | Superseded by ADR-YYY]

## Date
YYYY-MM-DD

## Context
What is the issue that we're seeing that is motivating this decision or change?

## Decision
What is the change that we're proposing and/or doing?

## Consequences
What becomes easier or more difficult to do because of this change?
```

## Index

### Meta Decisions (000-099)

- **[ADR-000](000-use-architecture-decision-records.md)**: Use Architecture Decision Records
  - **Status**: Accepted
  - **Date**: 2025-01-11
  - **Summary**: Adopt ADR format for documenting architectural decisions

### Technology Stack (100-199)

- **[ADR-100](100-use-cobra-cli-framework.md)**: Use Cobra for CLI Framework
  - **Status**: Accepted
  - **Date**: 2025-01-11
  - **Summary**: Use `spf13/cobra` as the CLI framework for generated CLIs
  - **Alternatives**: urfave/cli, kong, mitchellh/cli

- **[ADR-101](101-use-expr-templating-language.md)**: Use expr for Templating Language
  - **Status**: Accepted
  - **Date**: 2025-01-11
  - **Summary**: Use `expr-lang/expr` for `x-cli-workflow` expressions
  - **Alternatives**: jsonnet, gomplate

### Architecture Patterns (200-299)

*Reserved for future ADRs*

### Configuration & Data Models (300-399)

*Reserved for future ADRs*

### Feature-Specific (400+)

*Reserved for future ADRs*

## Numbering Convention

- **000-099**: Process and meta decisions (e.g., ADR-000)
- **100-199**: Technology stack choices (frameworks, libraries)
- **200-299**: Architecture patterns and design decisions
- **300-399**: Configuration and data models
- **400+**: Feature-specific decisions

## When to Write an ADR

**Write an ADR when:**

✅ Choosing between multiple viable technology options
✅ Making architectural decisions with long-term impact
✅ Removing or significantly changing existing features
✅ Adopting new patterns or conventions

**Don't write an ADR for:**

❌ Bug fixes
❌ Minor refactoring
❌ Documentation updates
❌ Obvious choices with no alternatives

## Creating a New ADR

1. **Choose a number** following the numbering convention
2. **Copy the template** from ADR-000
3. **Fill in all sections** (Context, Decision, Consequences)
4. **Create file**: `design/decisions/XXX-short-title.md`
5. **Update this index** with the new ADR
6. **Submit for review** via pull request

## ADR Lifecycle

```
Proposed → Accepted → [Deprecated | Superseded by ADR-YYY]
```

- **Proposed**: Under discussion
- **Accepted**: Decision made, active
- **Deprecated**: No longer recommended
- **Superseded**: Replaced by newer ADR

## Related Documentation

- **Architecture Docs**: `../architecture/` - Detailed design documents
- **External Docs**: `../../docs/` - User-facing documentation
- **Research**: `../../research/` - Research and analysis

## References

- [Michael Nygard's ADR format](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- [ADR GitHub organization](https://adr.github.io/)
- [Joel Parker Henderson's ADR examples](https://github.com/joelparkerhenderson/architecture-decision-record)

---

*⚒️ Forged with ❤️ by the CliForge team*
