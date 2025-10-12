# CI/CD Pipeline Documentation

This document describes the GitLab CI/CD pipeline configuration for the dman dotfile synchronization tool.

## Pipeline Overview

The pipeline consists of 6 main stages:

1. **Validate** - Code quality and format checks
2. **Test** - Unit tests, race conditions, and coverage analysis
3. **Build** - Compile binaries for multiple architectures
4. **Security** - Security scanning and vulnerability detection
5. **Package** - Docker image creation and release packaging
6. **Deploy** - Deployment to staging and production environments

## Pipeline Stages

### 1. Validation Stage

- **format-check**: Ensures Go code follows standard formatting (`gofmt`)
- **vet-check**: Runs `go vet` static analysis
- **lint-check**: Runs `staticcheck` linter (allows failure)
- **mod-verify**: Verifies Go modules are up-to-date

### 2. Test Stage

- **unit-tests**: Runs all unit tests with race detection
- **test-with-coverage**: Generates coverage reports (minimum 70% required)
- **race-condition-tests**: Stress tests for concurrency issues
- **benchmark-tests**: Performance benchmarks (allows failure)

### 3. Build Stage

- **build**: Creates Linux binary with version injection
- **build-multi-arch**: Creates binaries for multiple OS/architecture combinations

### 4. Security Stage

- **security-scan**: Uses `gosec` for security vulnerability scanning
- **dependency-scan**: Uses `govulncheck` for dependency vulnerabilities

### 5. Package Stage

- **docker-build**: Creates Docker images and pushes to registry
- **create-release**: Packages binaries for distribution

### 6. Deploy Stage

- **deploy-staging**: Manual deployment to staging environment
- **deploy-production**: Manual deployment to production (tags only)

## Triggered Events

### Merge Requests

- Validation stages
- Test stages
- Build stage
- Security stages

### Main Branch Commits

- All stages except production deployment

### Git Tags

- All stages including production deployment and release creation

## Environment Variables

The pipeline uses several GitLab CI predefined variables:

- `CI_COMMIT_TAG`: Git tag for releases
- `CI_COMMIT_SHORT_SHA`: Short commit hash
- `CI_REGISTRY_IMAGE`: Docker registry image path
- `CI_DEFAULT_BRANCH`: Main branch name

## Artifacts

### Coverage Reports

- `coverage.out`: Go coverage profile
- `coverage.html`: HTML coverage report
- Available for 30 days

### Build Artifacts

- `bin/dman`: Linux binary
- `dist/`: Multi-architecture binaries
- Available for 1 week to 1 month

### Security Reports

- `gosec-report.json`: Security scan results
- Available for 30 days

### Release Archives

- Platform-specific archives in `release/`
- Available for 1 year

## Docker Images

Docker images are built with:

- Alpine Linux base for minimal size
- Non-root user for security
- Health checks included
- Multi-stage build for optimization

### Image Tags

- `latest`: Latest main branch build
- `<commit-sha>`: Specific commit builds
- `<git-tag>`: Release builds

## Local Testing

Use the provided script for local CI simulation:

```bash
chmod +x scripts/ci-local.sh
./scripts/ci-local.sh
```

Or run individual commands:

```bash
# Format check
gofmt -l .

# Tests with coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Build with version info
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS="-X git.tyss.io/cj3636/dman/internal/buildinfo.Version=$VERSION \\
         -X git.tyss.io/cj3636/dman/internal/buildinfo.Commit=$COMMIT \\
         -X git.tyss.io/cj3636/dman/internal/buildinfo.BuildTime=$BUILD_TIME"
go build -ldflags "$LDFLAGS" -o bin/dman ./cmd/dman
```

## Development Workflow

1. **Feature Development**
   - Create feature branch from main
   - Make changes and commit
   - Push to trigger validation and test stages

2. **Merge Request**
   - Pipeline runs validation, tests, build, and security checks
   - Must pass before merge is allowed

3. **Main Branch**
   - Full pipeline runs including Docker build
   - Staging deployment available manually

4. **Release**
   - Create git tag (`git tag v1.0.0`)
   - Push tag to trigger full pipeline
   - Multi-arch binaries and Docker images created
   - Production deployment available manually

## Troubleshooting

### Common Issues

1. **Coverage Below Threshold**
   - Add more tests to reach 70% minimum coverage
   - Check coverage report for missed areas

2. **Format Issues**  
   - Run `go fmt ./...` to fix formatting
   - Commit the changes

3. **Docker Build Failures**
   - Check Dockerfile syntax
   - Verify base image availability
   - Check build context (.dockerignore)

4. **Security Scan Failures**
   - Review `gosec-report.json` artifact
   - Address high-severity issues
   - Use `// #nosec` comments sparingly for false positives

### Pipeline Debugging

1. Check job logs in GitLab CI interface
2. Use artifacts to download reports and binaries
3. Run local CI script to reproduce issues
4. Test Docker builds locally before pushing

## Configuration Files

- `.gitlab-ci.yml`: Main pipeline configuration  
- `Dockerfile`: Multi-stage Docker build
- `.dockerignore`: Docker build context exclusions
- `docker-compose.yml`: Local development setup
- `scripts/ci-local.sh`: Local testing script

## Security Considerations

