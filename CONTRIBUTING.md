# Contributing to Tailscale Ingress Operator

Thank you for your interest in contributing to the Tailscale Ingress Operator! This document provides guidelines and instructions for contributing to this project.

## Code of Conduct

By participating in this project, you agree to abide by the [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

1. Fork the repository
2. Create a new branch for your feature or bugfix
3. Make your changes
4. Ensure your code follows the project's style guidelines
5. Write or update tests as needed
6. Submit a pull request

## Development Setup

1. Ensure you have Go 1.24+ installed
2. Clone your fork of the repository
3. Build the project:
   ```bash
   go build -o tailscale-ingress-operator main.go
   ```
4. Run tests (if available):
   ```bash
   go test ./...
   ```

## Pull Request Process

1. Update the CHANGELOG.md with details of your changes
2. Ensure your PR description clearly describes the problem and solution
3. Include relevant tests and documentation updates
4. The PR must pass all CI checks before it can be merged

## Style Guidelines

- Follow standard Go formatting (`go fmt`)
- Write clear, concise commit messages
- Include comments for complex logic
- Keep functions focused and small
- Write tests for new functionality

## Reporting Issues

When reporting issues, please include:
- The version of the operator you're using
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Any relevant logs or error messages

## Feature Requests

For feature requests, please:
- Describe the feature in detail
- Explain why it would be valuable
- Provide any relevant use cases

## Questions?

Feel free to open an issue if you have any questions about contributing to the project. 