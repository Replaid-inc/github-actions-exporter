# Contributing to GitHub Actions Exporter

Thank you for your interest in contributing to the GitHub Actions Exporter! We welcome contributions from the community.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. Please be respectful and constructive in all interactions.

## How to Contribute

### Reporting Issues

Before creating a new issue, please:
1. Search existing issues to avoid duplicates
2. Use the appropriate issue template
3. Provide clear reproduction steps
4. Include relevant logs and system information

### Submitting Changes

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following our coding standards
3. **Add tests** for any new functionality
4. **Update documentation** as needed
5. **Ensure all tests pass** by running `go test ./...`
6. **Submit a pull request** with a clear description

### Development Setup

```bash
# Clone your fork
git clone https://github.com/your-username/github-actions-exporter.git
cd github-actions-exporter

# Install dependencies
go mod download

# Run tests
go test ./...

# Build the application
go build ./cmd/gh-actions-exporter

# Run the application
./gh-actions-exporter
```

### Coding Standards

- Follow Go best practices and idioms
- Use `gofmt` to format your code
- Add comments for exported functions and types
- Write meaningful commit messages
- Keep functions small and focused
- Use descriptive variable names

### Testing

- Write unit tests for all new functionality
- Aim for high test coverage
- Use table-driven tests where appropriate
- Mock external dependencies
- Test error conditions

### Pull Request Process

1. Ensure your branch is up to date with `main`
2. Update the README.md if you change functionality
3. Add any new dependencies to go.mod
4. Your PR will be reviewed by maintainers
5. Address any feedback promptly
6. Once approved, your PR will be merged

### Release Process

Releases are automated when tags are pushed:
1. Create a new tag: `git tag v1.2.3`
2. Push the tag: `git push origin v1.2.3`
3. GitHub Actions will build and publish the release

## Getting Help

- Create an issue for bugs or feature requests
- Join discussions for general questions
- Check the README.md for documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
