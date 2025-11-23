# Design Documentation

This directory contains **internal design and architecture documentation** for CliForge contributors and maintainers.

## Structure

```
design/
├── architecture/     # System architecture and design docs
└── decisions/        # Architecture Decision Records (ADRs)
```

## Architecture Documents

Located in `architecture/`:

- **builtin-commands-design.md** - Built-in commands and global flags system
- **secrets-handling-design.md** - Sensitive data detection and masking
- **deprecation-strategy.md** - API and CLI deprecation handling
- **configuration-override-matrix.md** - Detailed configuration override rules

## Architecture Decision Records (ADRs)

Located in `decisions/`:

ADRs document significant architectural and technology decisions with context, rationale, and consequences.

### Meta Decisions (000-099)
- **ADR-000**: Use Architecture Decision Records

### Technology Stack (100-199)
- **ADR-100**: Use Cobra for CLI Framework
- **ADR-101**: Use expr for Templating Language

### Future ADRs

When documenting new decisions, follow the ADR format defined in ADR-000:

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

- **000-099**: Process and meta decisions
- **100-199**: Technology stack choices (frameworks, libraries)
- **200-299**: Architecture patterns and design decisions
- **300-399**: Configuration and data models
- **400+**: Feature-specific decisions

### When to Write an ADR

**Write an ADR when:**
- Choosing between multiple viable technology options
- Making architectural decisions with long-term impact
- Removing or significantly changing existing features
- Adopting new patterns or conventions

**Don't write an ADR for:**
- Bug fixes
- Minor refactoring
- Documentation updates
- Obvious choices with no alternatives

## Contributing

When adding new design documents:

1. **Architecture docs** - Add to `architecture/` and update this README
2. **ADRs** - Add to `decisions/` with appropriate number and update ADR index
3. **Cross-references** - Update related documents with links
4. **Version** - Add version number and last updated date

## Related Documentation

- **External docs**: `../docs/` - User-facing documentation
- **Research**: `../research/` - Research and analysis documents
- **Branding**: `../branding/` - Brand guidelines and assets

---

*⚒️ Forged with ❤️ by the CliForge team*
