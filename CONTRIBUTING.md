# Contributing to DNT-Vault

Thank you for your interest in contributing to DNT-Vault!

## Development Setup

1. **Prerequisites**
   - Go 1.22 or higher
   - Git

2. **Clone and Build**
   ```bash
   git clone <repository-url>
   cd dnt-vault
   ./build.sh
   ```

3. **Run Tests**
   ```bash
   ./test.sh
   ```

## Project Structure

See [STRUCTURE.md](STRUCTURE.md) for detailed project structure.

## Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Add comments for exported functions
- Keep functions small and focused

## Making Changes

1. **Create a branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, readable code
   - Add tests if applicable
   - Update documentation

3. **Test your changes**
   ```bash
   ./build.sh
   ./test.sh
   ```

4. **Commit**
   ```bash
   git add .
   git commit -m "Add: your feature description"
   ```

## Commit Message Format

Use conventional commits:
- `Add:` - New feature
- `Fix:` - Bug fix
- `Update:` - Update existing feature
- `Refactor:` - Code refactoring
- `Docs:` - Documentation changes
- `Test:` - Test changes

## Pull Request Process

1. Update documentation if needed
2. Add tests for new features
3. Ensure all tests pass
4. Update CHANGELOG.md
5. Submit PR with clear description

## Areas for Contribution

### High Priority
- TLS/HTTPS support
- Unit tests
- Profile versioning
- Config validation

### Medium Priority
- Web UI
- Additional storage backends
- Profile templates
- Better error messages

### Documentation
- More examples
- Video tutorials
- API documentation
- Deployment guides

## Questions?

Feel free to open an issue for:
- Bug reports
- Feature requests
- Questions
- Documentation improvements

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Help others learn and grow
- Focus on what's best for the project

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
