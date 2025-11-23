# Gap Analysis: ROSA CLI Requirements vs CliForge Design

**Version**: 1.0.0
**Date**: 2025-11-23
**Status**: Analysis Complete

---

## Executive Summary

This document analyzes the gaps between CliForge's current design and the requirements identified from the ROSA CLI analysis (see `hack/rosa-cli/`). It categorizes missing features by priority and provides recommendations for addressing each gap.

## Methodology

Requirements were extracted from:
1. ROSA CLI v1.2.57 command mapping (`hack/rosa-cli/command-mapping.md`)
2. ROSA research summary (`hack/rosa-cli/rosa-research-summary.md`)
3. OpenAPI CLI extensions specification (`hack/rosa-cli/openapi-cli-extensions-spec.md`)

Gaps were identified by comparing against:
1. Technical Specification (`docs/technical-specification.md`)
2. Architecture documents (`design/architecture/`)
3. Architecture Decision Records (`design/decisions/`)

---

## Gap Categories

### 🔴 Critical Gaps (Architectural Blockers)

Features that require new architectural components or significant design changes.

#### 1. Plugin/External Tool Integration

**Requirement**: ROSA executes external tools (AWS CLI, CloudFormation) as part of workflows

**Current State**:
- Plugins explicitly marked as "too complex for v1.0" in configuration-override-matrix.md
- No plugin architecture designed
- No external tool execution framework

**Impact**: Cannot replicate ROSA's AWS integration patterns

**Recommendation**: Design plugin architecture (see proposed design doc)

**Related Files**:
- `design/architecture/configuration-override-matrix.md` (line 234: ❌ plugins)

---

#### 2. Multi-Provider Authentication

**Requirement**: ROSA uses OAuth2 for OCM API AND AWS credentials simultaneously

**Current State**:
- Single auth scheme per CLI assumed
- No design for multiple concurrent auth providers
- No credential chaining/fallback mechanism

**Impact**: Cannot handle APIs that require multiple auth mechanisms

**Recommendation**: Extend auth architecture to support multiple providers

**Related Files**:
- `docs/technical-specification.md` (Auth Manager component)

---

#### 3. File-Based Input Operations

**Requirement**: ROSA reads files for certificates, htpasswd, YAML configs

**Current State**:
- No file input handling design
- Request body mapping assumes primitive types
- No file validation/parsing framework

**Impact**: Cannot support `--from-file`, `--ca-file`, etc.

**Recommendation**: Design file operation framework

---

#### 4. Local State Management

**Requirement**: ROSA tracks recently used clusters, user preferences, command history

**Current State**:
- Cache design exists but only for API responses
- No context/state management design
- No cross-command state persistence

**Impact**: Cannot provide "current cluster" context or smart defaults

**Recommendation**: Design state management system

**Related Files**:
- `docs/technical-specification.md` (Cache Manager)

---

#### 5. Progress Indicators & Streaming

**Requirement**: ROSA shows spinners, progress bars, real-time log streaming

**Current State**:
- No UX framework design
- Async polling exists but no visual feedback
- No streaming/SSE/WebSocket support

**Impact**: Poor UX for long-running operations

**Recommendation**: Design progress/streaming framework

---

#### 6. Policy Validation Framework

**Requirement**: ROSA validates AWS SCPs, IAM boundaries, org policies

**Current State**:
- Pre-flight checks designed but limited to API calls
- No external policy engine integration
- No complex validation logic support

**Impact**: Cannot perform ROSA-style compliance checks

**Recommendation**: Extend pre-flight framework or create policy subsystem

---

### ⚠️ Important Gaps (Feature Parity)

Features needed for full rosa-like functionality but not architectural blockers.

#### 7. Multi-Step Workflow Orchestration

**Requirement**: ROSA executes complex workflows (pre-flight → IAM → cluster → post-config)

**Current State**:
- `x-cli-workflow` extension mentioned in ADR-101
- Expression language (expr) chosen
- **No workflow execution engine designed**
- No step dependencies, conditionals, rollback

**Impact**: Cannot replicate complex ROSA operations like cluster creation

**Recommendation**: Complete workflow orchestration design

**Related Files**:
- `design/decisions/101-use-expr-templating-language.md`

---

#### 8. Advanced Table Formatting

**Requirement**: ROSA has colored, multi-line, sortable tables

**Current State**:
- Basic output formatting mentioned
- No table library chosen
- No color/styling design
- No sorting/filtering design

**Impact**: Limited output quality

**Recommendation**: Design advanced output formatting system

---

#### 9. Watch/Follow Mode

**Requirement**: ROSA streams logs and status with `--watch`

**Current State**:
- Async polling designed (`x-cli-async`)
- **No streaming support** (SSE, WebSocket)
- No log tailing

**Impact**: Cannot provide real-time updates

**Recommendation**: Extend async framework with streaming protocols

**Related Files**:
- `hack/rosa-cli/openapi-cli-extensions-spec.md` (x-cli-async)

---

#### 10. Dry-Run Mode

**Requirement**: ROSA supports `--dry-run` for cost estimation and planning

**Current State**:
- No dry-run design
- No simulation framework

**Impact**: Cannot preview operations

**Recommendation**: Add dry-run extension and execution mode

---

#### 11. Dynamic Shell Completion

**Requirement**: ROSA fetches cluster names from API for tab completion

**Current State**:
- Shell completion mentioned in technical spec
- **No dynamic completion design**
- No API-driven suggestions

**Impact**: Limited autocomplete capabilities

**Recommendation**: Design dynamic completion system

**Related Files**:
- `docs/technical-specification.md` (Shell completion in CLI Framework Layer)

---

#### 12. Configuration Migration

**Requirement**: ROSA handles deprecated flags and config format changes

**Current State**:
- Deprecation strategy exists for APIs
- **No config migration framework**
- No version-to-version upgrade path

**Impact**: Breaking changes disrupt users

**Recommendation**: Extend deprecation strategy with migration tooling

**Related Files**:
- `design/architecture/deprecation-strategy.md`

---

### ✅ Well-Covered Areas

Features already designed or partially implemented in CliForge.

#### 13. Interactive Mode ✅

**Status**: Well-designed in `x-cli-interactive` extension

**Coverage**:
- Prompt types (text, select, confirm, etc.)
- Validation patterns
- Dynamic option loading from API
- Default values

**Related Files**:
- `hack/rosa-cli/openapi-cli-extensions-spec.md` (x-cli-interactive)

---

#### 14. OAuth2 Authentication ✅

**Status**: Single-provider OAuth2 designed

**Coverage**:
- Authorization Code flow
- Token storage (file, keyring)
- Auto-refresh
- `x-auth-config` extension

**Gap**: Multi-provider not supported (see Critical Gap #2)

**Related Files**:
- `hack/rosa-cli/openapi-cli-extensions-spec.md` (x-auth-config)

---

#### 15. Async Operation Polling ✅

**Status**: Well-designed in `x-cli-async`

**Coverage**:
- Status field tracking
- Terminal states
- Polling intervals
- Backoff strategies

**Gap**: No streaming support (see Important Gap #9)

**Related Files**:
- `hack/rosa-cli/openapi-cli-extensions-spec.md` (x-cli-async)

---

#### 16. Output Formatting ✅

**Status**: Basic support designed

**Coverage**:
- Multiple formats (table, JSON, YAML)
- Success/error messages
- Template interpolation

**Gap**: Advanced table features (see Important Gap #8)

**Related Files**:
- `hack/rosa-cli/openapi-cli-extensions-spec.md` (x-cli-output)

---

#### 17. Pre-flight Checks ✅

**Status**: Designed in `x-cli-preflight`

**Coverage**:
- API endpoint checks
- Required vs optional
- Skip flags

**Gap**: No external validation (see Critical Gap #6)

**Related Files**:
- `hack/rosa-cli/openapi-cli-extensions-spec.md` (x-cli-preflight)

---

#### 18. Error Mapping ✅

**Status**: Partially designed

**Coverage**:
- API error to user message mapping
- HTTP status code handling

**Gap**: Complex error recovery not designed

**Related Files**:
- `docs/technical-specification.md` (Error Handling section)

---

#### 19. Secrets Handling ✅

**Status**: Well-designed

**Coverage**:
- Pattern-based detection
- Multiple masking strategies
- `x-cli-secret` extension
- User controls

**Related Files**:
- `design/architecture/secrets-handling-design.md`

---

#### 20. Deprecation Strategy ✅

**Status**: Well-designed for APIs

**Coverage**:
- Warning levels
- Time-based escalation
- `x-cli-deprecation` extension
- Migration assistance

**Gap**: Config migration (see Important Gap #12)

**Related Files**:
- `design/architecture/deprecation-strategy.md`

---

## Priority Matrix

| Priority | Count | Percentage |
|----------|-------|------------|
| 🔴 Critical | 6 | 30% |
| ⚠️ Important | 6 | 30% |
| ✅ Covered | 8 | 40% |

---

## Recommended Actions

### Phase 1: Critical Architecture (v0.8.0)

**Goal**: Address architectural blockers

1. **Plugin Architecture** (Critical #1)
   - Design plugin interface
   - Security sandboxing
   - External tool execution framework
   - Create ADR: Use Plugin System for External Tools

2. **Multi-Provider Auth** (Critical #2)
   - Extend auth manager for multiple providers
   - Credential chaining design
   - Create ADR: Support Multiple Authentication Providers

3. **File Operations** (Critical #3)
   - File input handling framework
   - Validation/parsing system
   - `x-cli-file-input` extension design

4. **State Management** (Critical #4)
   - Context/state persistence
   - Cross-command state sharing
   - CLI context selection (current cluster, etc.)

### Phase 2: Workflow & UX (v0.9.0)

**Goal**: Complete workflow capabilities

1. **Workflow Orchestration** (Important #7)
   - Workflow execution engine
   - Step dependencies
   - Conditional execution
   - Rollback mechanisms

2. **Progress & Streaming** (Critical #5)
   - Progress indicator framework
   - SSE/WebSocket support
   - Log streaming

3. **Advanced Output** (Important #8)
   - Table library selection (ADR needed)
   - Color/styling system
   - Sorting/filtering

### Phase 3: Enhanced Features (v1.0.0)

**Goal**: Feature parity with ROSA

1. **Dynamic Completion** (Important #11)
2. **Watch Mode** (Important #9)
3. **Dry-Run** (Important #10)
4. **Config Migration** (Important #12)
5. **Policy Validation** (Critical #6)

---

## Design Documents Needed

### New Documents Required

1. **`plugin-architecture.md`**
   - Plugin interface design
   - Security model
   - External tool execution
   - Lifecycle management

2. **`workflow-orchestration.md`**
   - Workflow execution engine
   - Step types and dependencies
   - Conditional logic
   - Error handling and rollback

3. **`file-operations.md`**
   - File input handling
   - Validation and parsing
   - Supported file types
   - Security considerations

4. **`progress-and-streaming.md`**
   - Progress indicators (spinners, bars)
   - Streaming protocols (SSE, WebSocket)
   - Log tailing
   - UX patterns

5. **`state-management.md`**
   - State persistence
   - Context system
   - Cross-command state sharing
   - User defaults and preferences

6. **`multi-provider-auth.md`**
   - Multiple concurrent auth providers
   - Credential chaining
   - Provider selection logic
   - Token management

### Documents to Update

1. **`docs/technical-specification.md`**
   - Add plugin system component
   - Add workflow engine component
   - Add state manager component
   - Add streaming support

2. **`design/architecture/README.md`**
   - Link to new design documents
   - Update document relationships diagram

3. **`docs/configuration-dsl.md`**
   - Add plugin configuration
   - Add workflow configuration
   - Add state/context configuration

---

## ADRs Needed

### Technology Choices

1. **ADR-102**: Choose Plugin Architecture Pattern
   - Options: HashiCorp plugin, WebAssembly, gRPC, embedded scripts
   - Recommendation: TBD

2. **ADR-103**: Choose Table Rendering Library
   - Options: tablewriter, pterm, lipgloss
   - Recommendation: TBD

3. **ADR-104**: Choose Streaming Protocol
   - Options: SSE, WebSocket, gRPC streaming
   - Recommendation: SSE for simplicity, WebSocket as fallback

### Architecture Patterns

4. **ADR-200**: Support Multiple Authentication Providers
   - Context: APIs may require multiple auth schemes (OAuth2 + AWS)
   - Decision: TBD

5. **ADR-201**: Use Plugin System for External Tools
   - Context: Need to integrate AWS CLI, CloudFormation, etc.
   - Decision: TBD

6. **ADR-202**: Workflow Orchestration Engine Design
   - Context: Complex multi-step operations need orchestration
   - Decision: TBD

---

## Risks & Mitigations

### Risk 1: Scope Creep

**Risk**: Adding all these features delays v1.0

**Mitigation**:
- Phase implementation across versions
- Critical items in v0.8-0.9
- Nice-to-have in v1.x

### Risk 2: Plugin Security

**Risk**: External plugins introduce security vulnerabilities

**Mitigation**:
- Sandbox execution
- Permission model
- Signed plugins only
- Explicit user approval

### Risk 3: Complexity

**Risk**: Adding features makes CliForge too complex

**Mitigation**:
- Keep plugins optional
- Progressive disclosure (simple by default)
- Clear separation of concerns

---

## Conclusion

CliForge's current design covers **40% of ROSA requirements** well, but has **30% critical gaps** and **30% important gaps**.

**Key Insights**:

1. **Plugin architecture is essential** - Cannot replicate AWS integration without it
2. **Workflow orchestration needs completion** - Mentioned in ADR-101 but not designed
3. **Current design is solid** - Auth, output, deprecation, secrets are well-designed
4. **Phased approach recommended** - Address critical gaps first, iterate on enhancements

**Next Steps**:

1. Review this gap analysis
2. Prioritize which gaps to address for v0.8.0
3. Create new design documents (start with plugin architecture)
4. Write ADRs for major decisions
5. Update technical specification

---

## References

- ROSA CLI Research: `hack/rosa-cli/README.md`
- Command Mapping: `hack/rosa-cli/command-mapping.md`
- Research Summary: `hack/rosa-cli/rosa-research-summary.md`
- OpenAPI Extensions: `hack/rosa-cli/openapi-cli-extensions-spec.md`
- Technical Spec: `docs/technical-specification.md`
- Current Architecture: `design/architecture/`
