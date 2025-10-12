# Local CI testing script
# Usage: ./scripts/ci-local.sh

#!/bin/bash
set -e

echo "🔧 Starting local CI pipeline simulation..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Validate stage
echo "📋 VALIDATION STAGE"
echo "Checking Go formatting..."
if ! gofmt -l . | grep -q .; then
    log_info "Code formatting is correct"
else
    log_error "Code formatting issues found. Run 'go fmt ./...'"
    exit 1
fi

echo "Running go vet..."
if go vet ./...; then
    log_info "Static analysis passed"
else
    log_error "Static analysis failed"
    exit 1
fi

echo "Verifying Go modules..."
if go mod verify && go mod tidy; then
    log_info "Go modules are valid"
else
    log_error "Go module issues found"
    exit 1
fi

# Test stage
echo ""
echo "🧪 TEST STAGE"
echo "Running unit tests..."
if go test -v ./...; then
    log_info "Unit tests passed"
else
    log_error "Unit tests failed"
    exit 1
fi

echo "Running race condition tests..."
if go test -race ./...; then
    log_info "Race condition tests passed"
else
    log_error "Race condition tests failed"
    exit 1
fi

echo "Checking test coverage..."
go test -coverprofile=coverage.out ./...
COVERAGE=$(go tool cover -func=coverage.out | grep total: | awk '{gsub(/%/,"",$$3); split($$3,a,"."); print a[1]}')
THRESHOLD=70

if [ "$COVERAGE" -ge "$THRESHOLD" ]; then
    log_info "Coverage ${COVERAGE}% meets threshold ${THRESHOLD}%"
else
    log_error "Coverage ${COVERAGE}% below threshold ${THRESHOLD}%"
    exit 1
fi

# Build stage
echo ""
echo "🏗️  BUILD STAGE"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

echo "Building application..."
LDFLAGS="-X git.tyss.io/cj3636/dman/internal/buildinfo.Version=$VERSION \
         -X git.tyss.io/cj3636/dman/internal/buildinfo.Commit=$COMMIT \
         -X git.tyss.io/cj3636/dman/internal/buildinfo.BuildTime=$BUILD_TIME"

mkdir -p bin/
if go build -ldflags "$LDFLAGS" -o bin/dman ./cmd/dman; then
    log_info "Build completed successfully"
    echo "Version info:"
    ./bin/dman version
else
    log_error "Build failed"
    exit 1
fi

# Docker build test (optional)
echo ""
echo "🐳 DOCKER BUILD TEST"
if command -v docker &> /dev/null; then
    echo "Building Docker image..."
    if docker build -t dman:test \
        --build-arg VERSION="$VERSION" \
        --build-arg COMMIT="$COMMIT" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        . ; then
        log_info "Docker build successful"
        
        # Test the Docker image
        echo "Testing Docker image..."
        if docker run --rm dman:test version; then
            log_info "Docker image test passed"
        else
            log_warn "Docker image test failed"
        fi
    else
        log_warn "Docker build failed"
    fi
else
    log_warn "Docker not available, skipping Docker build test"
fi

echo ""
log_info "🎉 All CI stages completed successfully!"
echo "Ready for GitLab CI/CD pipeline"