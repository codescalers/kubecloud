# Contributing to Mycelium Cloud

We welcome contributions to Mycelium Cloud! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

Please be respectful and constructive in all interactions with the community. We're committed to providing a welcoming and inclusive environment.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/kubecloud.git
   cd kubecloud
   ```
3. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Setup

Follow the [Getting Started Guide](../GETTING_STARTED.md) to set up your development environment.

## Making Changes

### Code Style

#### Backend (Go)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Run `gofmt` before committing
- Run `golint` to check code style
- Write clear, descriptive comments
- Use meaningful variable and function names

#### Frontend (Vue.js/TypeScript)
- Use ESLint configuration in the project
- Format code with Prettier
- Follow Vue 3 Composition API conventions
- Write clear component descriptions
- Use TypeScript for type safety

### Testing

#### Backend
- Write unit tests for new functions
- Run tests: `go test ./...`
- Aim for >80% code coverage on new code

#### Frontend
- Write tests for components and utilities
- Run tests: `npm run test`
- Test user interactions and edge cases

### Documentation

- Update relevant README files with new features
- Add JSDoc comments for JavaScript/TypeScript code
- Add godoc comments for Go code
- Update the main README if changing project structure

## Commit Guidelines

Write clear, descriptive commit messages:

```bash
git commit -m "feat: add support for cluster auto-scaling"
```

Use conventional commit format:
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `style:` for code style changes
- `refactor:` for code refactoring
- `test:` for test additions/changes
- `chore:` for maintenance tasks

## Pull Request Process

1. **Ensure your branch is up to date**:
   ```bash
   git fetch origin
   git rebase origin/master
   ```

2. **Push your changes**:
   ```bash
   git push origin feature/your-feature-name
   ```

3. **Create a Pull Request** on GitHub with:
   - Clear title describing the changes
   - Detailed description of what changed and why
   - Reference to related issues (if any)
   - Screenshots for UI changes
   - Test results

4. **Address review comments** promptly

5. **Ensure CI passes** (tests, linting, etc.)

6. **Rebase and squash commits** if requested

## Pull Request Checklist

- [ ] Code follows project style guidelines
- [ ] Tests written for new functionality
- [ ] All tests pass locally
- [ ] Documentation updated (README, comments, etc.)
- [ ] Commit messages are clear and descriptive
- [ ] No breaking changes (or documented if necessary)
- [ ] Docker builds successfully (if applicable)

## Reporting Issues

When reporting bugs or suggesting features:

1. **Check existing issues** to avoid duplicates
2. **Use clear, descriptive titles**
3. **Provide detailed information**:
   - System information (OS, versions)
   - Steps to reproduce
   - Expected vs actual behavior
   - Screenshots/logs if applicable
4. **Use labels** to categorize the issue

## Areas for Contribution

### Backend
- Feature development
- Performance optimization
- Bug fixes
- API improvements
- Database optimizations

### Frontend
- UI/UX improvements
- New features
- Bug fixes
- Accessibility improvements
- Performance optimization

### Documentation
- Guides and tutorials
- API documentation
- Architecture diagrams
- Troubleshooting guides
- Example configurations

### DevOps
- CI/CD improvements
- Docker/Kubernetes configurations
- Deployment automation
- Monitoring enhancements

## Project Structure

```
kubecloud/
├── backend/              # Go backend API
├── frontend/kubecloud/   # Vue.js frontend
├── crd/                  # Kubernetes CRDs
├── ingress-controller/   # Ingress controller
├── mycelium-cni/         # CNI plugin
├── mycelium-peer/        # Peer networking
├── k3s/                  # K3s deployment
├── docs/                 # Documentation
└── README.md             # Main README
```

## Review Process

1. **Automated Checks**: CI pipeline validates code quality
2. **Code Review**: Maintainers review your changes
3. **Feedback**: Address any comments or suggestions
4. **Approval**: Once approved, your PR will be merged

## Merge Criteria

- All CI checks pass
- At least one approval from maintainers
- Documentation updated
- No breaking changes (or clearly documented)

## Building and Testing Locally

### Backend
```bash
cd backend
go test ./...          # Run tests
go build              # Build binary
make run              # Run locally
```

### Frontend
```bash
cd frontend/kubecloud
npm test              # Run tests
npm run build         # Build for production
npm run dev           # Run development server
```

### Full Stack
```bash
docker-compose up    # Run all services
```

## Development Tools

### Recommended IDE Extensions (VS Code)
- **Go**: golang.go
- **Vue**: Vue - Official
- **ESLint**: dbaeumer.vscode-eslint
- **Prettier**: esbenp.prettier-vscode
- **Docker**: ms-azuretools.vscode-docker

### CLI Tools
- `go fmt` - Format Go code
- `gofmt` - Official Go formatter
- `golint` - Go linter
- `prettier` - Format JS/TS code
- `eslint` - Lint JS/TS code

## Questions?

- Check existing documentation in `docs/`
- Open a discussion on GitHub
- Create an issue for bugs
- Contact maintainers

## License

By contributing to Mycelium Cloud, you agree that your contributions will be licensed under the Apache License 2.0.

## Acknowledgments

Thank you for contributing to making Mycelium Cloud better!

---

For more information, see the [main README](../../README.md) and [Architecture Overview](../architecture/OVERVIEW.md).
