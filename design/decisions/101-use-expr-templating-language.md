# ADR-101: Use expr for Templating Language

## Status

Accepted

## Date

2025-01-11

## Context

CliForge's `x-cli-workflow` OpenAPI extension needs an expression language for:

1. **Conditional execution** - Only run steps when conditions are met
   - Example: `condition: "response.status == 200 && response.body.region != null"`

2. **Data transformation** - Transform API responses before passing to next step
   - Example: `transform: "filter(items, .status == 'active')"`

3. **Dynamic values** - Access previous step results in URLs, headers, parameters
   - Example: `url: "https://ec2.{get-objects.body.region}.amazonaws.com"`

4. **Array operations** - Filter, map, reduce over API response arrays
   - Example: `map(users, .name)`, `filter(orders, .total > 1000)`

### Requirements

**Must Have:**
- Safe execution (no side effects, no infinite loops, no memory exploits)
- Type safety (catch errors at compile time when possible)
- Performance (will execute many times per CLI invocation)
- Go integration (seamless library embedding)
- Familiar syntax (low learning curve for Go developers)
- Embeddable in YAML strings

**Nice to Have:**
- Optional chaining (handle null values gracefully)
- Rich standard library (array, string, math functions)
- Production usage (proven at scale)
- Good error messages

### Alternatives Considered

We evaluated three Go-based templating/expression languages:

#### 1. expr (expr-lang/expr) - 7,400+ stars
**What it is**: Expression language for dynamic configurations with type safety and performance

**Pros:**
- ✅ **Perfect for expressions** - Designed exactly for evaluating conditions and transformations
- ✅ **Safe by design** - Memory-safe, side-effect-free, always terminating
- ✅ **Type-safe** - Optional static type checking catches errors at compile time
- ✅ **Fast** - Bytecode compilation, can compile once and reuse
- ✅ **Go-like syntax** - Familiar to Go developers: `filter(arr, .field > 5)`
- ✅ **Optional chaining** - `user?.profile?.email` handles nulls gracefully
- ✅ **Production-proven** - Google Cloud Platform, Uber, ByteDance, Alibaba
- ✅ **Easy to embed** - Clean Go API, works well in YAML strings

**Cons:**
- ⚠️ **No native iteration** - Can't write `for` loops (by design for safety)
- ⚠️ **Expression-only** - Not a full programming language

**Example:**
```yaml
condition: "response.status == 200 && response.body.region != null"
transform: |
  {
    "total": len(items),
    "active": filter(items, .status == 'active'),
    "names": map(items, .name)
  }
```

#### 2. jsonnet (google/go-jsonnet) - 1,800+ stars
**What it is**: Data templating language that extends JSON with computation, designed by Google

**Pros:**
- ✅ **Powerful composition** - Object-oriented with mixins
- ✅ **Mature** - Battle-tested at Grafana Labs, Databricks
- ✅ **Rich standard library** - 60+ functions
- ✅ **Imports** - Can modularize complex configurations
- ✅ **Lazy evaluation** - Efficient for large configs

**Cons:**
- ❌ **Overkill** - Designed for generating large configuration files, not simple expressions
- ❌ **Wrong abstraction** - Config generation vs expression evaluation
- ❌ **Performance** - Slower than expr for simple evaluations
- ❌ **Learning curve** - New syntax to learn (not JSON, not Go)
- ❌ **Verbose** - More syntax than needed for simple conditions

**Example:**
```yaml
transform: |
  local data = get_objects.body.data;
  {
    total: std.length(data),
    active: std.filter(function(o) o.status == "active", data),
    names: std.map(function(o) o.name, data)
  }
```

#### 3. gomplate (hairyhenderson/gomplate) - 3,000+ stars
**What it is**: Template rendering tool with extensive datasource support, built on Go's `text/template`

**Pros:**
- ✅ **Datasource-driven** - Fetch from HTTP, AWS, Consul, Vault, etc.
- ✅ **200+ functions** - Extensive function library
- ✅ **Production-proven** - Popular in DevOps/SRE workflows

**Cons:**
- ❌ **Wrong abstraction** - Designed for CLI templating, not embedded expressions
- ❌ **Template syntax** - `{{ if eq .status 200 }}` is verbose for simple conditions
- ❌ **CLI-first** - Not optimized for library use
- ❌ **No compilation** - Templates parsed every time
- ❌ **No type safety** - Runtime errors only

**Example:**
```yaml
condition: |
  {{ if eq (index (datasource "response") "status") 200 }}true{{ else }}false{{ end }}
```

### Decision Matrix (Weighted Scoring)

| Criteria | Weight | expr | jsonnet | gomplate |
|----------|--------|------|---------|----------|
| **Syntax Simplicity** | 5 | 5 | 3 | 2 |
| **Expression Evaluation** | 5 | 5 | 4 | 2 |
| **Type Safety** | 4 | 5 | 2 | 2 |
| **Performance** | 4 | 5 | 3 | 3 |
| **Learning Curve** | 3 | 5 | 3 | 4 |
| **Go Integration** | 4 | 5 | 4 | 3 |
| **Data Transformation** | 5 | 4 | 5 | 3 |
| **Safety (sandboxing)** | 5 | 5 | 4 | 2 |
| **Ecosystem Maturity** | 3 | 4 | 5 | 4 |
| **Embeddable in YAML** | 4 | 5 | 3 | 2 |
| | | | | |
| **Weighted Score** | | **4.68** | **3.88** | **2.90** |

## Decision

**CliForge uses `expr-lang/expr` for the templating/expression language in `x-cli-workflow`.**

### Rationale

1. **Perfect Fit for Expression Evaluation**
   - Designed exactly for our use case: evaluating conditions and transformations
   - Not overkill (like jsonnet) or underkill (like gomplate templates)
   - Concise syntax for common operations

2. **Safety by Design**
   - Memory-safe, side-effect-free, always terminating
   - Perfect for user-provided expressions in OpenAPI specs
   - No risk of infinite loops, system calls, or resource exhaustion
   - Critical for security: API owners embed these expressions in OpenAPI specs

3. **Performance**
   - Two-phase compilation: compile to bytecode, execute in VM
   - Can compile once, execute many times (cache compiled programs)
   - Aggressive optimizations: constant folding, type inference, in-range check elimination
   - VM reuse for 4-40% additional performance improvement

4. **Type Safety**
   - Optional static type checking catches errors before execution
   - Aligns with Go's type system
   - Prevents runtime type errors

5. **Production-Proven**
   - Google Cloud Platform - Expression language
   - Uber - Customization for Uber Eats marketplace
   - ByteDance - Internal business rule engine
   - Alibaba - Web framework for recommendation services
   - Designed for exactly this use case: dynamic config evaluation

6. **Developer Experience**
   - Go-like syntax (familiar to target users)
   - Excellent error messages with line/column numbers
   - Online playground for testing: https://expr-lang.org/expr-editor
   - Clean integration with Go

7. **Simplicity**
   - Just expressions, not a full language
   - Easy to embed in YAML strings
   - Clear separation of concerns: structure in YAML, logic in expr

### Handling expr's Limitations

**Problem**: expr doesn't support `for` loops (by design)

**Solution**: Handle iteration at YAML level with `foreach` keyword

```yaml
steps:
  - id: get-objects
    request:
      method: GET
      url: "/api/objects"

  - id: process-each-object
    foreach: get-objects.body.data  # YAML structure handles iteration
    as: item                         # YAML defines loop variable
    request:
      url: "/api/objects/{item.id}/details"  # expr for interpolation
      condition: "item.status == 'active'"   # expr for filtering
```

**Benefits:**
- ✅ Clean separation: iteration in YAML, expressions in expr
- ✅ Type safety: We validate `foreach` points to array
- ✅ Performance: We control the loop, expr just evaluates per-item
- ✅ Simplicity: Users don't write complex loops in expressions

## Consequences

### Positive

✅ **Concise syntax** - Most common operations are one-liners
✅ **Safe execution** - Can't cause side effects or security issues
✅ **Fast performance** - Bytecode compilation and VM optimization
✅ **Type safety** - Catch errors early
✅ **Familiar syntax** - Go developers feel at home
✅ **Optional chaining** - `user?.email?.verified` handles nulls elegantly
✅ **Future-proof** - Active development, growing adoption
✅ **Easy onboarding** - Low learning curve

### Negative

⚠️ **No native iteration** - Must handle `foreach` separately in YAML (but this is actually a design benefit)
⚠️ **Limited to expressions** - Can't define complex multi-step logic (but we handle that in YAML structure)
⚠️ **Smaller ecosystem** - Fewer third-party libraries than jsonnet (but growing rapidly)

### Neutral

ℹ️ **New dependency** - Adds `github.com/expr-lang/expr` to project
ℹ️ **Two-phase processing** - YAML parsing + expr compilation (but enables caching)

### Implementation Strategy

**Phase 1: Basic Integration**
```go
import "github.com/expr-lang/expr"

func evaluateCondition(condition string, env map[string]interface{}) (bool, error) {
    program, err := expr.Compile(condition,
        expr.Env(env),
        expr.AsBool(),
    )
    if err != nil {
        return false, fmt.Errorf("compile condition: %w", err)
    }

    result, err := expr.Run(program, env)
    if err != nil {
        return false, fmt.Errorf("evaluate condition: %w", err)
    }

    return result.(bool), nil
}
```

**Phase 2: Template Interpolation**
```go
// Support {expr} syntax in strings
func interpolateString(template string, env map[string]interface{}) (string, error) {
    re := regexp.MustCompile(`\{([^}]+)\}`)
    return re.ReplaceAllStringFunc(template, func(match string) string {
        exprStr := match[1:len(match)-1]  // Remove { }
        program, _ := expr.Compile(exprStr, expr.Env(env))
        result, _ := expr.Run(program, env)
        return fmt.Sprint(result)
    }), nil
}
```

**Phase 3: Custom Functions**
```go
// Add custom functions for common tasks
expr.Compile(code, expr.Function(
    "unique",
    func(params ...any) (any, error) {
        arr := params[0].([]interface{})
        seen := make(map[interface{}]bool)
        result := []interface{}{}
        for _, item := range arr {
            if !seen[item] {
                seen[item] = true
                result = append(result, item)
            }
        }
        return result, nil
    },
    new(func([]interface{}) []interface{}),
))
```

### Migration Path

If we later need more power (complex object composition, imports, advanced transformations):
- Can add jsonnet as **optional** alternative syntax
- API owners choose: `language: expr` (default) or `language: jsonnet`
- Doesn't break existing expr-based workflows

### Future Enhancements

**Possible future feature** (deferred to Phase 2):
- `x-cli-functions` extension - Allow API owners to define custom functions
- Functions compiled from expr and registered at runtime
- Example:
  ```yaml
  x-cli-functions:
    - name: formatPrice
      params: [amount, currency]
      body: "currency + ' ' + string(round(amount, 2))"
  ```

---

**Version**: 1.0
**Last Updated**: 2025-01-11
**Supersedes**: templating-language-comparison.md (research document)
