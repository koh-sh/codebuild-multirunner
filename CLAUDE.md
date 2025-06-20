# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**codebuild-multirunner** is a Go CLI tool and GitHub Action for orchestrating multiple AWS CodeBuild projects simultaneously. It enables parallel execution of CodeBuild projects with configuration overrides via YAML configuration files.

## Development Commands

### Primary Development Workflow

```bash
make ci             # Complete CI pipeline: format + modernize + lint + test
```

Use `make ci` for all testing and linting - it runs the complete pipeline including formatting, modernization checks, linting, and testing.

### Individual Commands (for specific needs)

```bash
make test           # Run tests with tparse formatting
make cov            # Generate HTML coverage report
make blackboxtest   # Run black-box integration tests
make lint           # Run golangci-lint
make fmt            # Format code with gofumpt
make tidy           # Tidy go.mod dependencies
make modernize      # Check for Go modernization opportunities
make modernize-fix  # Apply Go modernization fixes
```

### Build and Run

```bash
go build -o codebuild-multirunner
make dockerbuild    # Build Docker image
make dockerrun      # Run containerized version
```

## Architecture

### Core Structure

- **cmd/**: CLI commands using Cobra framework
  - `root.go`: Root command and CLI setup
  - `run.go`: Main parallel execution logic
  - `retry.go`: Build retry functionality
  - `cwlog.go`: CloudWatch log retrieval
  - `dump.go`: Configuration dump functionality
- **internal/cb/**: Core business logic with AWS CodeBuild API interactions
- **internal/cwlog/**: CloudWatch logs processing and retrieval
- **internal/types/**: Auto-generated AWS API type definitions

### Key Patterns

- **Interface-based design**: `CodeBuildAPI` interface enables mocking
- **Concurrent processing**: Goroutines with WaitGroup for parallel builds
- **Clean error handling**: Structured errors with context, graceful degradation
- **Configuration flexibility**: Supports both legacy list format and modern map format

### Configuration Format

The tool processes YAML configurations with environment variable substitution:

```yaml
builds:
  group1:
    - projectName: project-name
      environmentVariablesOverride:
        - name: BRANCH_NAME
          value: ${GITHUB_REF_NAME}
```

## AWS Integration

Requires AWS credentials with permissions for:

- `codebuild:StartBuild`
- `codebuild:BatchGetBuilds`
- `codebuild:RetryBuild`
- `logs:GetLogEvents`

## Testing Strategy

- **Unit tests**: Comprehensive mocking of AWS APIs in `internal/cb/cb_test.go` and `internal/cwlog/cwlog_test.go`
- **Integration tests**: Black-box testing in `_testscripts/`
- **Coverage reporting**: Use `make cov` for HTML coverage reports

## Dependencies

- Go 1.24.0+ required
- Uses AWS SDK v2 for all AWS interactions
- Cobra for CLI framework
- goccy/go-yaml for YAML processing with environment variable expansion

## Git Workflow

- **NEVER commit directly to the `main` branch** - always create feature branches
- Create feature branches with descriptive names (e.g., `feat/deprecate-list-format`)
- Always run `make ci` before committing to ensure all tests and lints pass
- Use conventional commit format with detailed commit messages
- Commit changes only after all CI checks pass successfully

## Claude Code Instructions

### Key Development Practices

- **ALWAYS run `make ci`** before making any commits - this ensures code formatting, modernization, linting, and testing all pass
- **Interface-first development**: Use the existing `CodeBuildAPI` interface pattern for new AWS integrations
- **Concurrent-safe patterns**: Follow existing goroutine + WaitGroup patterns for parallel operations
- **Error handling**: Use structured errors with context, maintain graceful degradation
- **Configuration**: Support both legacy and modern YAML formats, preserve backward compatibility

### Available Commands

The tool provides these main commands:

- `run`: Execute multiple CodeBuild projects in parallel
- `retry`: Retry a specific CodeBuild build by ID
- `log`: Retrieve CloudWatch logs for a build
- `dump`: Display parsed configuration
- `completion`: Generate shell completion scripts

### Testing Requirements

- Run `make ci` to execute the full test suite
- Unit tests use comprehensive AWS API mocking
- Integration tests validate end-to-end functionality
- All code must pass linting and formatting checks
