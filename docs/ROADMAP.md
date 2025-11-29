# Project Roadmap

This roadmap outlines the planned work needed to fully realize the capabilities described in the docs and config examples, along with pragmatic enhancements to strengthen the dotfile manager.

## Phase 1: Configuration & Tracking Reliability
- Harden configuration validation for `track`/exclusion patterns (e.g., `docs/` with `!docs/config.yaml`), including startup linting and user-level overrides.
- Expand glob compatibility and benchmarking for common shell patterns (e.g., `.oh-my-zsh/plugins/**/*.zsh` and theme files) to guarantee predictable matching across platforms.
- Provide migration tooling and guidance for legacy `include` configs, ensuring safe upgrades to the current `track` syntax.

## Phase 2: Storage Backends & Data Integrity
- Implement and productionize Redis connector options (connection pooling, TLS toggles, socket support) for `storage_driver: redis` flows.
- Complete MariaDB/MySQL integration using the `db` block (TLS options, socket connections, pooled prepared statements) while keeping disk storage stable.
- Add integration tests and fault-injection cases covering driver selection, fallback, and data durability guarantees.

## Phase 3: Sync Engine Robustness
- Strengthen diff/compare logic for large trees (streamed hashing, concurrency tuning) and ensure atomic publish/install semantics for bulk operations.
- Support selective publish/install by glob, honoring exclusion rules, with clear user feedback when patterns skip files.
- Improve conflict detection and recovery (e.g., server-precedence toggles, same-file reporting) to avoid accidental overwrites.

## Phase 4: Observability & Operations
- Enrich structured logging with operation IDs and per-user context; add log sampling for noisy scenarios.
- Extend metrics/health endpoints with latency breakdowns, storage-driver status, and queue depths; ship simple dashboards for local monitoring.
- Add audit trail persistence for sensitive events (auth failures, destructive actions) consistent with the existing token model.

## Phase 5: UX, CLI, and Docs
- Polish CLI UX (progress bars for bulk operations, clearer error strings, and machine-readable JSON schemas for outputs).
- Provide wizard-style `dman init` prompts that surface the new config blocks (redis/db) and track defaults for quick starts.
- Keep docs in sync with defaults (including `docs/config.yaml` as a live example), and add troubleshooting guides for pattern matching and storage setup.

## Phase 6: Deployment, Release, and QA
- Expand automated CI to cover race detectors, static analysis, and cross-platform build matrix.
- Deliver signed release artifacts and container images with non-root defaults and configurable ports.
- Codify smoke tests for Docker Compose flows and rolling upgrades to prevent regressions during releases.
