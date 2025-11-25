# CliForge Tutorials

Welcome to the CliForge tutorial series! These comprehensive, hands-on tutorials will guide you through building production-ready CLIs for common use cases.

## Tutorial Series Overview

Each tutorial is designed to be completed independently, though they build on each other in complexity. All tutorials include complete working examples, troubleshooting sections, and best practices.

### 📚 Available Tutorials

#### 1. [Building a Simple REST API CLI](tutorial-rest-api.md)

**Difficulty**: Beginner
**Time**: 45-60 minutes
**Lines**: ~1,500

Learn the fundamentals of CliForge by building a GitHub API CLI from scratch.

**What You'll Learn**:
- Creating OpenAPI specifications for REST APIs
- Implementing authentication with bearer tokens
- Building CRUD operations (Create, Read, Update, Delete)
- Handling API errors gracefully
- Formatting output (table, JSON, YAML)
- Adding confirmation prompts for destructive actions
- Enabling shell completion

**Perfect For**:
- Developers new to CliForge
- Teams wanting to provide CLIs for their APIs
- Anyone building simple API client tools

**Example Commands You'll Build**:
```bash
github-cli repos list --user octocat
github-cli repos create --name my-repo --description "My project"
github-cli issues create --title "Bug report" --body "Description"
```

---

#### 2. [Cloud Infrastructure Management CLI](tutorial-cloud-management.md)

**Difficulty**: Intermediate
**Time**: 90-120 minutes
**Lines**: ~2,000

Build an enterprise-grade CLI for managing cloud infrastructure with async operations and workflows.

**What You'll Learn**:
- Handling long-running asynchronous operations
- Implementing workflow orchestration
- Managing infrastructure state
- Providing progress feedback and streaming output
- Implementing rollback strategies
- Using circuit breakers for resilience
- Building multi-step deployments

**Perfect For**:
- DevOps engineers
- SRE teams
- Platform engineers
- Infrastructure automation specialists

**Example Commands You'll Build**:
```bash
cloud-cli compute instances create --name web-1 --size large --wait
cloud-cli deploy full-stack --template infra.yaml --env production
cloud-cli storage sync --source ./data --destination s3://bucket --watch
cloud-cli state drift-detect
```

---

#### 3. [Integrating CliForge CLI in CI/CD Pipelines](tutorial-ci-cd-integration.md)

**Difficulty**: Intermediate
**Time**: 60-90 minutes
**Lines**: ~1,800

Master CI/CD integration with working examples for GitHub Actions, GitLab CI, and Jenkins.

**What You'll Learn**:
- Running CLIs in non-interactive/headless mode
- Managing secrets and credentials securely
- Implementing robust error handling and retries
- Parsing and validating CLI output
- Creating reusable workflow templates
- Implementing blue-green deployments
- Setting up monitoring and notifications

**Perfect For**:
- Platform engineers
- DevOps teams
- CI/CD pipeline maintainers
- Release engineers

**Example Integrations You'll Build**:
```yaml
# GitHub Actions
- Deploy on pull request
- Run integration tests
- Auto-cleanup resources

# GitLab CI
- Multi-stage pipelines
- Parallel deployments
- Approval gates

# Jenkins
- Parameterized builds
- Blue-green deployments
- Slack notifications
```

---

## Tutorial Path Recommendations

### For Beginners

**Start Here**: [REST API CLI Tutorial](tutorial-rest-api.md)

This tutorial teaches you the fundamentals:
1. Complete the REST API tutorial first
2. Experiment with your own APIs
3. Then move to Cloud Management or CI/CD based on your needs

### For DevOps/SRE Teams

**Recommended Path**:
1. [REST API CLI](tutorial-rest-api.md) - Learn basics (30 mins)
2. [Cloud Management CLI](tutorial-cloud-management.md) - Master infrastructure automation (90 mins)
3. [CI/CD Integration](tutorial-ci-cd-integration.md) - Automate deployments (60 mins)

### For Platform Engineers

**Recommended Path**:
1. [REST API CLI](tutorial-rest-api.md) - Quick foundation (30 mins)
2. [CI/CD Integration](tutorial-ci-cd-integration.md) - Pipeline integration (60 mins)
3. [Cloud Management CLI](tutorial-cloud-management.md) - Advanced workflows (90 mins)

---

## Prerequisites

### Required for All Tutorials

- **Go**: Version 1.21 or later
- **CliForge**: Latest version installed
- **Git**: For version control
- **Text Editor**: VS Code, Vim, or your preference
- **Command Line**: Basic proficiency

### Tutorial-Specific Requirements

**REST API Tutorial**:
- GitHub account (free tier)
- Personal Access Token

**Cloud Management Tutorial**:
- Docker (optional, for local testing)
- Understanding of cloud infrastructure concepts

**CI/CD Tutorial**:
- Access to GitHub, GitLab, or Jenkins
- Basic CI/CD knowledge
- jq (JSON processor)

---

## Tutorial Features

Each tutorial includes:

### ✅ Complete Working Examples
All code examples are tested and working. You can copy-paste and run them.

### ✅ Step-by-Step Instructions
Clear, detailed steps with explanations of what each part does and why.

### ✅ Expected Output
See what the output should look like at each step to verify you're on track.

### ✅ Troubleshooting Section
Common issues and their solutions, plus debugging tips.

### ✅ Best Practices
Production-ready patterns and anti-patterns to avoid.

### ✅ Next Steps
Links to related documentation and suggestions for extending what you've built.

---

## Getting Help

### Stuck on a Tutorial?

1. **Check the Troubleshooting Section**: Each tutorial has a comprehensive troubleshooting guide
2. **Review Prerequisites**: Ensure all required software is installed and configured
3. **Check the FAQ**: [CliForge FAQ](../faq.md)
4. **Search Documentation**: Use the search feature in the docs
5. **GitHub Discussions**: Ask questions in the community

### Found an Issue?

If you find errors or have suggestions for improving these tutorials:
- Open an issue on GitHub
- Submit a pull request with improvements
- Share feedback in GitHub Discussions

---

## Additional Resources

### Documentation

- [Getting Started Guide](../getting-started.md) - Quick introduction to CliForge
- [User Guide - Configuration](../user-guide-configuration.md) - Configuration details
- [User Guide - Authentication](../user-guide-authentication.md) - Authentication patterns
- [User Guide - Workflows](../user-guide-workflows.md) - Workflow orchestration
- [OpenAPI Extensions Reference](../openapi-extensions-reference.md) - Complete extension reference

### Examples

- [Examples & Recipes](../examples-and-recipes.md) - Quick patterns and snippets
- [GitHub Repository](https://github.com/CliForge/cliforge) - Source code and examples

### Community

- [GitHub Discussions](https://github.com/CliForge/cliforge/discussions) - Ask questions and share ideas
- [Issue Tracker](https://github.com/CliForge/cliforge/issues) - Report bugs and request features

---

## Tutorial Statistics

| Tutorial | Difficulty | Time | Lines | Focus Area |
|----------|-----------|------|-------|------------|
| REST API CLI | Beginner | 45-60 min | ~1,500 | Basics & CRUD |
| Cloud Management | Intermediate | 90-120 min | ~2,000 | Async & Workflows |
| CI/CD Integration | Intermediate | 60-90 min | ~1,800 | Automation |
| **Total** | - | **3-4 hours** | **~5,300** | - |

---

## Contributing

We welcome contributions to improve these tutorials! Please see our [Contributing Guide](https://github.com/CliForge/cliforge/blob/main/CONTRIBUTING.md) for details.

### Ways to Contribute

- **Fix Errors**: Typos, broken links, incorrect commands
- **Add Examples**: Real-world use cases and patterns
- **Improve Explanations**: Clarify confusing sections
- **Add Diagrams**: Visual aids for complex concepts
- **Translate**: Help make tutorials available in other languages

---

**Version**: 1.0.0
**Last Updated**: 2025-11-25
**Maintainer**: CliForge Team
