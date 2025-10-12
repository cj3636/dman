# dman

[![Pipeline Status](https://git.tyss.io/cj3636/dman/badges/main/pipeline.svg?style=for-the-badge)](https://git.tyss.io/cj3636/dman/-/pipelines)
[![Coverage](https://git.tyss.io/cj3636/dman/badges/main/coverage.svg?style=for-the-badge)](https://git.tyss.io/cj3636/dman/-/jobs)
[![Go Report Card](https://goreportcard.com/badge/github.com/cj3636/dman?style=for-the-badge)](https://goreportcard.com/report/github.com/cj3636/dman)
[![Go Version](https://img.shields.io/github/go-mod/go-version/cj3636/dman?style=for-the-badge&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/cj3636/dman?style=for-the-badge&logo=docker)](https://hub.docker.com/r/cj3636/dman)
[![Release](https://img.shields.io/github/v/release/cj3636/dman?style=for-the-badge)](https://github.com/cj3636/dman/releases)

**A high-performance, secure dotfile synchronization tool for LAN environments**

dman provides enterprise-grade dotfile management with a modern client-server architecture, supporting multiple storage backends, compression, atomic operations, and comprehensive observability.

---

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
  - [End Users](#end-users)
  - [Developers](#developers)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Storage Backends](#storage-backends)
- [API Documentation](#api-documentation)
- [Docker Deployment](#docker-deployment)
- [Development](#development)
- [CI/CD Pipeline](#cicd-pipeline)
- [Contributing](#contributing)
- [Security](#security)
- [License](#license)

---

## Features

### Core Functionality
- **Client-Server Architecture** - Centralized dotfile storage with multi-client access
- **Atomic Operations** - Safe, transactional file operations with rollback capability
- **Bulk Operations** - Efficient tar-based bulk publish/install with streaming
- **Compression Support** - Optional gzip compression for bandwidth optimization
- **Multi-Platform** - Native support for Linux, macOS, and Windows

### Storage & Backends
- **Multiple Storage Backends** - Disk, Redis, MariaDB/MySQL support
- **Configurable Persistence** - Flexible storage configuration per environment
- **Path Safety** - Built-in traversal protection and sanitization
- **Metadata Tracking** - Comprehensive operation tracking and metrics

### Security & Authentication
- **Token-Based Authentication** - Secure Bearer token authentication
- **Path Validation** - Robust input validation and sanitization
- **Non-Root Containers** - Security-first Docker deployment
- **Audit Logging** - Comprehensive operation logging

### Operations & Monitoring
- **Health & Status Endpoints** - Built-in monitoring and observability
- **Metrics Collection** - In-memory metrics with persistence
- **Structured Logging** - Configurable log levels and formatting
- **Version Injection** - Build-time version and commit information

### Advanced Features
- **Delete Propagation** - Optional server-side file pruning
- **Change Detection** - Smart diff algorithm with optional same-file reporting
- **JSON Output** - Machine-readable output for automation
- **Race Condition Safety** - Comprehensive concurrency protection

---

## Quick Start

### End Users

#### 1. Installation

**Download Binary:**
```bash
# Linux/macOS
curl -L https://github.com/cj3636/dman/releases/latest/download/dman-linux-amd64.tar.gz | tar xz
sudo mv dman /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/cj3636/dman/releases/latest/download/dman-windows-amd64.zip" -OutFile "dman.zip"
```

**Docker:**
```bash
docker pull cj3636/dman:latest
```

#### 2. Server Setup

**Initialize Configuration:**
```bash
dman init
```

**Edit Configuration (`dman.yaml`):**
```yaml
auth_token: "your-secure-token-here"
server_url: "http://localhost:3626"
storage_driver: "disk"
users:
  alice:
    home: "/home/alice/"
    include:
      - ".bashrc"
      - ".vimrc"
      - ".gitconfig"
      - ".ssh/config"
  bob:
    home: "/home/bob/"
    include:
      - ".zshrc"
      - ".tmux.conf"
```

**Start Server:**
```bash
dman serve --addr :3626
```

#### 3. Client Operations

**Compare Files:**
```bash
dman compare --show-same
```

**Publish Changes:**
```bash
dman publish --bulk --gzip
```

**Install Updates:**
```bash
dman install --bulk --gzip
```

**Check Status:**
```bash
dman status --json
```

### Developers

#### 1. Environment Setup

**Clone Repository:**
```bash
git clone https://git.tyss.io/cj3636/dman.git
cd dman
```

**Install Dependencies:**
```bash
go mod download
```

**Run Tests:**
```bash
make test-race coverage
```

#### 2. Development Workflow

**Local Development:**
```bash
# Start development environment
docker compose up -d

# Run local CI checks
./scripts/ci-local.sh

# Build and test
make build
./bin/dman version
```

**Code Quality:**
```bash
make fmt vet
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

---

## Installation

### Binary Releases

Download pre-built binaries from [GitHub Releases](https://github.com/cj3636/dman/releases):

| Platform | Architecture | Download |
|----------|-------------|----------|
| Linux | AMD64 | `dman-linux-amd64.tar.gz` |
| Linux | ARM64 | `dman-linux-arm64.tar.gz` |
| macOS | AMD64 | `dman-darwin-amd64.tar.gz` |
| macOS | ARM64 | `dman-darwin-arm64.tar.gz` |
| Windows | AMD64 | `dman-windows-amd64.zip` |
| Windows | ARM64 | `dman-windows-arm64.zip` |

### Docker Images

```bash
docker pull cj3636/dman:latest          # Latest stable
docker pull cj3636/dman:v0.4.0          # Specific version
docker pull cj3636/dman:main            # Development
```

### Build from Source

**Prerequisites:**
- Go 1.24 or later
- Make

**Build:**
```bash
git clone https://git.tyss.io/cj3636/dman.git
cd dman
make build
```

---

## Usage

### Commands

| Command | Description | Example |
|---------|-------------|---------|
| `init` | Initialize configuration | `dman init` |
| `serve` | Start server | `dman serve --addr :3626` |
| `compare` | Compare local vs server | `dman compare --show-same --json` |
| `publish` | Upload changes | `dman publish --bulk --gzip --prune` |
| `install` | Download updates | `dman install --bulk --gzip` |
| `upload` | Upload single file | `dman upload --user alice --path .vimrc` |
| `download` | Download single file | `dman download --user alice --path .vimrc` |
| `status` | Server status | `dman status --json` |
| `login` | Authenticate client | `dman login --token TOKEN` |
| `logout` | Clear authentication | `dman logout` |
| `version` | Show version info | `dman version` |

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Configuration file path | `./dman.yaml` or `~/dman.yaml` |
| `--log-level` | Log level (info/debug) | `info` |
| `--verbose` | Enable verbose logging | `false` |

---

## Configuration

### Complete Configuration Example

```yaml
# Authentication
auth_token: "your-secure-token"
server_url: "http://localhost:3626"

# Storage configuration
storage_driver: "disk"  # Options: disk, redis, maria/mariadb/mysql

# Global file patterns (fallback for users without specific includes)
include:
  - ".bashrc"
  - ".profile"
  - ".gitconfig"

# User-specific configurations
users:
  alice:
    home: "/home/alice/"
    include:
      - ".bashrc"
      - ".vimrc"
      - ".gitconfig"
      - ".ssh/config"
      - ".tmux.conf"
  
  bob:
    home: "/home/bob/"
    include:
      - ".zshrc"
      - ".oh-my-zsh/"
      - ".config/nvim/"

# Redis configuration (when storage_driver: redis)
redis_addr: "127.0.0.1:6379"
redis_password: ""
redis_db: 0
redis_tls: false

# MariaDB configuration (when storage_driver: maria/mariadb/mysql)
maria_addr: "127.0.0.1:3306"
maria_db: "dman"
maria_user: "dman"
maria_password: "password"
maria_tls: false
```

### Environment Variables

- `DMAN_VERIFY_WRITE=1` - Enable post-write hash verification
- `DMAN_AUTH_TOKEN` - Override configuration auth token
- `DMAN_SERVER_URL` - Override server URL

---

## Storage Backends

### Disk Storage (Default)
- **Use Case:** Single-server deployments, development
- **Configuration:** `storage_driver: "disk"`
- **Data Location:** `./data/` directory
- **Status:** Stable

### Redis Storage
- **Use Case:** High-performance, distributed deployments
- **Configuration:** `storage_driver: "redis"`
- **Features:** In-memory performance, persistence, clustering
- **Status:** Experimental - Real Redis backend with chunked storage

### MariaDB/MySQL Storage
- **Use Case:** Enterprise deployments, complex queries
- **Configuration:** `storage_driver: "maria"`
- **Features:** ACID compliance, backups, replication
- **Status:** Scaffold - Currently delegates to disk storage

### Redis In-Memory (Testing)
- **Use Case:** Testing and development
- **Configuration:** `storage_driver: "redis-mem"`
- **Features:** No persistence, fast testing
- **Status:** Scaffold for development

---

## API Documentation

### Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/health` | No | Server health check |
| GET | `/status` | Yes | Detailed server status |
| POST | `/compare` | Yes | Compare file inventories |
| POST | `/publish` | Yes | Bulk file upload (tar) |
| POST | `/install` | Yes | Bulk file download (tar) |
| POST | `/prune` | Yes | Delete server files |
| PUT | `/upload` | Yes | Upload single file |
| GET | `/download` | Yes | Download single file |

### Response Examples

**Health Check:**
```json
{
  "ok": true,
  "version": "v0.4.0",
  "build_time": "2025-10-11T20:30:00Z",
  "commit": "89f140f",
  "server_time": "2025-10-11T20:35:00Z"
}
```

**Status:**
```json
{
  "files_total": 42,
  "bytes_total": 123456,
  "users": [
    {
      "user": "alice",
      "files": 30,
      "bytes": 90000
    }
  ],
  "last_publish": "2025-10-11T20:30:00Z",
  "last_install": "2025-10-11T20:32:00Z",
  "metrics": {
    "uptime": 1225877,
    "publish_requests": 10,
    "install_requests": 8
  }
}
```

---

## Docker Deployment

### Single Container

```bash
docker run -d \
  --name dman-server \
  -p 3626:3626 \
  -v dman_data:/app/data \
  -v ./dman.yaml:/app/dman.yaml:ro \
  cj3636/dman:latest serve --addr :3626
```

### Docker Compose

```yaml
services:
  dman:
    image: cj3636/dman:latest
    ports:
      - "3626:3626"
    volumes:
      - dman_data:/app/data
      - ./dman.yaml:/app/dman.yaml:ro
    environment:
      - DMAN_VERIFY_WRITE=1
    restart: unless-stopped

volumes:
  dman_data:
```

---

## Development

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Make

### Getting Started

```bash
# Clone repository
git clone https://git.tyss.io/cj3636/dman.git
cd dman

# Install dependencies
go mod download

# Run tests
make test test-race coverage

# Build
make build

# Local development
docker compose up -d
./scripts/ci-local.sh
```

### Testing

```bash
# Unit tests
make test

# Race condition tests
make test-race

# Coverage analysis
make coverage

# Coverage threshold check
make coverage-threshold

# Local CI simulation
./scripts/ci-local.sh

# Benchmark tests
go test -bench=. ./internal/storage/
```

---

## CI/CD Pipeline

The project uses GitLab CI/CD with comprehensive testing and deployment:

### Stages

1. **Validate** - Format, lint, and module verification
2. **Test** - Unit tests, race conditions, coverage analysis
3. **Build** - Multi-architecture binary compilation
4. **Security** - Security scanning and vulnerability detection
5. **Package** - Docker images and release artifacts
6. **Deploy** - Staging and production deployment

### Coverage Requirements

- Minimum 70% test coverage
- Race condition testing
- Property-based testing for core algorithms

See [CI/CD Documentation](docs/CI-CD.md) for detailed pipeline information.

---

## Contributing

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Run the test suite (`make test test-race coverage`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Standards

- Follow Go best practices and idioms
- Maintain test coverage above 70%
- Include comprehensive documentation
- Use structured logging
- Handle errors gracefully

---

## Security

### Reporting Vulnerabilities

Please report security vulnerabilities to [security@tyss.io](mailto:security@tyss.io).

### Security Features

- Bearer token authentication
- Path traversal protection
- Input validation and sanitization
- Non-root container execution
- Security scanning in CI/CD

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**For questions, issues, or contributions, visit [git.tyss.io/cj3636/dman](https://git.tyss.io/cj3636/dman)**
