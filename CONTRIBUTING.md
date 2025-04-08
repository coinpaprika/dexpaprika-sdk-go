# Contributing to DexPaprika Go SDK

Thank you for your interest in contributing to the DexPaprika Go SDK! This document provides guidelines and instructions for contributing.

## Development Setup

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR-USERNAME/dexpaprika-sdk-go.git`
3. Create a branch: `git checkout -b your-feature-branch`

## Running Tests

Use the Makefile to run tests:

```bash
# Run comprehensive tests
make run-comprehensive

# Run basic tests
make run-basic

# Run the production example
make run-example

# Run unit tests
make test
```

## Code Style

- Follow standard Go conventions and best practices
- Use `gofmt` to format your code
- Comments should be complete sentences
- Add tests for new functionality

## Pull Request Process

1. Ensure your code passes all tests
2. Update documentation if necessary
3. Submit a pull request with a clear description of the changes
4. Reference any related issues

## Reporting Issues

Please use the GitHub issue tracker to report bugs or request features. When reporting a bug, include:

- A clear description of the issue
- Steps to reproduce
- Expected behavior
- SDK version and Go version

## License

By contributing, you agree that your contributions will be licensed under the project's MIT license. 