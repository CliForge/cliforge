# ADR-000: Use Architecture Decision Records

## Status

Accepted

## Date

2025-01-11

## Context

As CliForge evolves, we've made numerous architectural and technology decisions:
- Choice of CLI framework (Cobra)
- Templating language selection (expr)
- Configuration override architecture
- Removal of features (commands, hooks, plugins)

These decisions are currently scattered across various documents with inconsistent formats. We need a structured way to:

1. **Document decisions**: Capture the context, rationale, and consequences of important architectural choices
2. **Track evolution**: Understand why decisions were made and when they might need revisiting
3. **Onboard contributors**: Help new team members understand the "why" behind the architecture
4. **Prevent re-litigation**: Avoid repeatedly discussing already-settled decisions

## Decision

We will use **Architecture Decision Records (ADRs)** following the lightweight format popularized by Michael Nygard.

### ADR Format

Each ADR will include:

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

### Numbering Convention

- **000-099**: Process and meta decisions (e.g., this ADR)
- **100-199**: Technology stack choices (frameworks, libraries)
- **200-299**: Architecture patterns and design decisions
- **300-399**: Configuration and data models
- **400+**: Feature-specific decisions

### File Location

All ADRs will live in `design/decisions/` with the naming format:
```
design/decisions/XXX-short-title.md
```

### When to Write an ADR

Write an ADR when:
- Choosing between multiple viable technology options
- Making architectural decisions with long-term impact
- Removing or significantly changing existing features
- Adopting new patterns or conventions

Don't write an ADR for:
- Bug fixes
- Minor refactoring
- Documentation updates
- Obvious choices with no alternatives

## Consequences

### Positive

- **Historical context**: Future developers can understand why decisions were made
- **Reduced discussion**: Decisions are documented and don't need re-explaining
- **Better onboarding**: New contributors can quickly get up to speed
- **Change tracking**: Can see evolution of architecture over time
- **Lightweight**: Simple markdown format, no tooling required

### Negative

- **Additional overhead**: Must write ADRs for significant decisions
- **Maintenance**: Need to keep ADR index up to date
- **Discipline required**: Team must commit to the practice

### Neutral

- **Not immutable**: ADRs can be superseded by new ADRs, which is expected
- **Living documents**: ADRs capture decisions at a point in time

## Notes

This ADR follows the format it proposes, serving as both documentation and example.

Inspired by:
- [Michael Nygard's ADR format](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- [ADR GitHub organization](https://adr.github.io/)
- [Joel Parker Henderson's ADR examples](https://github.com/joelparkerhenderson/architecture-decision-record)

---

**Version**: 1.0
**Last Updated**: 2025-01-11
