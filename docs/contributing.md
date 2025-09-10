# Contributing

Thank you for your interest in contributing to gns3util! This document provides guidelines for contributing to the project.

## Getting Started

### Prerequisites
- Go 1.19 or later
- Git
- Access to a GNS3v3 server for testing

### Development Setup
```bash
# Fork and clone the repository
git clone https://github.com/your-username/gns3-api-util.git
cd gns3-api-util

# Build the project
go build -o gns3util

# Run tests
go test ./...
```

## Contribution Guidelines

### Code Style
- Follow Go standard formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for exported functions
- Keep functions focused and small

### Commit Messages
- Use conventional commit format
- Start with a type: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`
- Keep the first line under 50 characters
- Provide detailed description if needed

Examples:
```
feat: add template-based exercise creation
fix: resolve project name resolution issue
docs: update CLI reference with new commands
```

### Testing
- Add tests for new functionality
- Ensure all tests pass before submitting
- Test with real GNS3 server when possible
- Include edge cases in test coverage

### Documentation
- Update relevant documentation files
- Add examples for new features
- Update CLI help text if needed
- Keep README.md current

## Development Workflow

### 1. Create a Feature Branch
```bash
git checkout -b feature/your-feature-name
```

### 2. Make Your Changes
- Implement the feature
- Add tests
- Update documentation
- Ensure code quality

### 3. Test Your Changes
```bash
# Run all tests
go test ./...

# Test with GNS3 server
./gns3util -s https://your-server:3080 --help
```

### 4. Commit Your Changes
```bash
git add .
git commit -m "feat: add your feature description"
```

### 5. Push and Create Pull Request
```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub.

## Areas for Contribution

### High Priority
- **Bug fixes**: Fix existing issues
- **Documentation**: Improve existing docs
- **Tests**: Add test coverage
- **Error handling**: Improve error messages

### Medium Priority
- **New commands**: Add useful CLI commands
- **Performance**: Optimize existing code
- **Features**: Add requested features
- **Examples**: Add more usage examples

### Low Priority
- **Refactoring**: Code cleanup
- **UI improvements**: Better output formatting
- **Advanced features**: Complex functionality

## Reporting Issues

### Bug Reports
When reporting bugs, please include:
- GNS3 server version
- gns3util version
- Steps to reproduce
- Expected vs actual behavior
- Error messages (if any)

### Feature Requests
For feature requests, please include:
- Use case description
- Proposed solution
- Alternative solutions considered
- Additional context

## Code Review Process

### For Contributors
- Address review feedback promptly
- Keep pull requests focused
- Respond to comments constructively
- Update documentation as needed

### For Maintainers
- Review code thoroughly
- Test changes when possible
- Provide constructive feedback
- Merge when ready

## Development Guidelines

### Adding New Commands
1. Create command file in `cmd/` directory
2. Add command to root command in `main.go`
3. Implement command logic
4. Add help text
5. Write tests
6. Update documentation

### Adding New API Endpoints
1. Add endpoint to `pkg/api/endpoints/`
2. Add request/response types to `pkg/api/schemas/`
3. Update client in `pkg/api/client.go`
4. Add tests
5. Update documentation

### Error Handling
- Use descriptive error messages
- Include context when possible
- Log errors appropriately
- Return meaningful exit codes

## Testing Guidelines

### Unit Tests
- Test individual functions
- Mock external dependencies
- Cover edge cases
- Aim for high coverage

### Integration Tests
- Test with real GNS3 server
- Test command combinations
- Test error conditions
- Test with different server versions

### Manual Testing
- Test new features thoroughly
- Test with different configurations
- Test error handling
- Verify documentation accuracy

## Release Process

### Versioning
- Follow semantic versioning (semver)
- Update version in `main.go`
- Update changelog
- Tag releases appropriately

### Release Checklist
- [ ] All tests pass
- [ ] Documentation updated
- [ ] Changelog updated
- [ ] Version bumped
- [ ] Release notes prepared

## Community Guidelines

### Be Respectful
- Use welcoming and inclusive language
- Be respectful of differing viewpoints
- Accept constructive criticism gracefully
- Focus on what's best for the community

### Be Collaborative
- Work together toward common goals
- Share knowledge and experience
- Help others when possible
- Build on others' work

## Getting Help

### Documentation
- Check existing documentation first
- Look at examples and tutorials
- Review CLI help text

### Community
- Open an issue for questions
- Use discussions for general topics
- Ask for help when needed

### Maintainers
- Contact maintainers for urgent issues
- Be patient with responses
- Provide clear information

## License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project (GNU General Public License v3.0).

## Thank You

Thank you for contributing to gns3util! Your contributions help make the tool better for everyone in the GNS3 community.
