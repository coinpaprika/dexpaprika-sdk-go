# Git Hooks for dexpaprika-sdk-go

This directory contains git hooks to maintain code quality and prevent broken code from being committed or pushed.

## Installation

To install the hooks, run:

```bash
make hooks
```

Or directly:

```bash
./hooks/setup.sh
```

## Available Hooks

### pre-commit

Runs before each commit to ensure code quality:

- **Code Formatting**: Checks that all Go files are properly formatted with `goimports`
- **Linting**: Runs `golangci-lint` if available

If the hook fails, you can:
1. Fix the issues (recommended): Run `make format` to auto-format code
2. Bypass the hook (not recommended): `git commit --no-verify`

### pre-push

Runs before pushing to remote to ensure tests pass:

- **Tests**: Runs the full test suite with race detection

If the hook fails, you can:
1. Fix the tests (recommended)
2. Bypass the hook (not recommended): `git push --no-verify`

## Bypassing Hooks

While not recommended, you can bypass hooks in special circumstances:

```bash
# Skip pre-commit hook
git commit --no-verify -m "commit message"

# Skip pre-push hook
git push --no-verify
```

## Troubleshooting

### Hook fails with "command not found"

Make sure you have the required tools installed:

```bash
# Install goimports
go install golang.org/x/tools/cmd/goimports@latest

# Install golangci-lint
make check
```

### Hook fails on formatting

Run the formatter to auto-fix formatting issues:

```bash
make format
```

### Hook fails on tests

Run tests locally to see what's failing:

```bash
make test
```
