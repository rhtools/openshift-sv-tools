# Contributing

## Reporting Issues

- Check existing issues before creating a new one
- Include reproduction steps
- Provide cluster version and environment details
- Share relevant error messages or logs

## Submitting Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linters (`make fix && make test`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Development Guidelines

- Follow Go best practices and idioms
- Add tests for new functionality
- Update documentation as needed
- Use the test data in `testdata/` for testing
- Run `make fix` before committing to ensure code quality

## Build Commands

```bash
# Build the binary
make build

# Build for multiple platforms (linux, darwin, windows)
make build-all

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code and run linters
make fix

# Format code (go fmt)
make fmt

# Organize imports (goimports)
make imports

# Run linter (golangci-lint)
make lint

# Type checking (go vet + staticcheck)
make typecheck

# Clean build artifacts
make clean

# Install development dependencies
make install-deps
```

## Further Reading

- [Developer Setup Guide](vm-scanner/DEV_README.md) — local development, running without compilation, quick examples
- [Refactoring Guide](vm-scanner/REFACTORING_GUIDE.md) — code organization, architecture decisions, best practices
- [Test Data](vm-scanner/testdata/README.md) — test fixtures for development without a live cluster
