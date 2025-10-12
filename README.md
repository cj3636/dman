# dman

Hardened dotfile sync tool (publish/install/compare/upload/download) with bulk tar, gzip, prune, status metrics.

## Build

```bash
make build
```

## Usage

```
./bin/dman serve --log-level debug
./bin/dman status --json
./bin/dman compare --show-same --json
./bin/dman publish --bulk --gzip --prune --json
./bin/dman install --bulk --gzip --json
./bin/dman version
```

## Config Example

```yaml
auth_token: "CHANGEME"
server_url: "http://localhost:7099"
storage_driver: "disk" # options: disk | redis | mariadb (redis/mariadb are scaffolds)
include:
  - .bashrc
users:
  me:
    home: ~/ # tilde expanded
# Redis configuration example (when storage_driver: redis)
redis_addr: "10.0.0.36:6379"   # OR omit and use redis_socket
# redis_socket: "/var/run/redis/redis.sock"  # mutually exclusive with redis_addr
redis_username: ""             # optional ACL user
redis_password: "PASSWORD"      # optional password
redis_db: 0
redis_tls: false                # set true to enable TLS
redis_tls_ca: "ssl/ca.crt"      # optional CA bundle path
redis_tls_insecure_skip_verify: false
redis_tls_server_name: "redis.internal" # optional SNI override
# MariaDB configuration example (when storage_driver: mariadb)
maria_addr: "10.0.0.36:3306"   # OR omit and use maria_socket
# maria_socket: "/run/mysqld/mysqld.sock" # mutually exclusive with maria_addr
maria_db: "DATABASE"
maria_user: "DBUSER"
maria_password: "PASSWORD"
maria_tls: false                 # enable TLS
maria_tls_ca: "ssl/ca.crt"      # optional CA
maria_tls_cert: "ssl/client.crt" # optional client cert
maria_tls_key: "ssl/client.key"  # optional client key
maria_tls_insecure_skip_verify: false
maria_tls_server_name: "db.internal" # optional SNI override
```

## Storage Backends

| Driver    | Status     | Notes |
|-----------|------------|-------|
| disk      | Stable     | Files stored under data/<user>/... |
| redis     | Experimental | Real Redis backend (supports TCP/unix socket + optional TLS); stores each file as a binary value with key user/rel. |
| redis-mem | Scaffold   | In-memory map (no persistence) mainly for tests / dev without Redis server. |
| mariadb   | Scaffold   | Currently delegates to an on-disk store at data/maria; placeholder for future SQL implementation. |

Future work will replace the redis scaffold with a real Redis client (hash/object layout) and implement MariaDB schema for metadata + blob storage.

## Endpoints

| Method | Path      | Notes |
|--------|-----------|-------|
| GET    | /health   | build info |
| GET    | /status   | metrics, per-user counts |
| POST   | /compare  | inventory diff |
| POST   | /publish  | bulk upload tar (gzip optional) |
| POST   | /install  | bulk download tar (Accept-Encoding gzip) |
| POST   | /prune    | delete server-only files (opt-in) |
| PUT    | /upload   | single file upload |
| GET    | /download | single file download |

All except /health require Authorization: Bearer token.

## Testing

```
make test
make test-race
make coverage
```

To quick-test alternate backends:
```
# redis in-memory scaffold (no external server required)
./bin/dman serve --config dman.yaml --log-level debug  # with storage_driver: redis-mem
# real redis (ensure redis server reachable / socket path configured)
./bin/dman serve --config dman.yaml --log-level debug  # with storage_driver: redis
# mariadb scaffold (disk delegate currently)
./bin/dman serve --config dman.yaml --log-level debug  # with storage_driver: mariadb
```

## Roadmap
- Real Redis backend (key layout & TTL policies)
- Real MariaDB backend (schema, migrations, streaming blobs)
- Optional VCS-style version history (see internal/vcs)
- Negative tests for malformed gzip / truncated streams
- Coverage threshold gate

## Backend Implementation Notes

### Redis (chunked storage)
Each stored file is written in 256KB chunks by default:
- Base key: `user/rel` contains a JSON manifest `{"chunks":N,"v":1}`.
- Chunk keys: `user/rel:chunk:0`, `user/rel:chunk:1`, ...
- Legacy (pre-chunk) single-value objects are still readable; on first overwrite they are migrated to chunked form.
- Deletes remove manifest and all chunk keys.
- List filtering excludes any key containing `:chunk:` so high-level inventory remains clean.
- Simple exponential backoff (up to 3 attempts) is applied for Redis `SET` / `GET` operations.

Future enhancements (not yet implemented): configurable chunk size, pipelined writes, side-car metadata hashes, key TTL policies.

### MariaDB
Current implementation stores the entire file in a single LONGBLOB row (`dman_files`). TLS (CA, optional client cert/key, SNI, insecure skip) supported via a custom registered TLS profile. Retries with exponential backoff wrap CRUD queries. Future work: chunked / streaming blobs, metadata columns (hash/mtime), migrations for schema evolution.

## Benchmarks
Run basic storage benchmarks (disk & in-memory redis):
```bash
go test -bench=BenchmarkSaveDisk -run ^$ ./internal/storage
go test -bench=BenchmarkSaveRedisMem -run ^$ ./internal/storage
```
Optional real Redis benchmark (set address first):
```bash
set DMAN_BENCH_REDIS_ADDR=127.0.0.1:6379
# or export DMAN_BENCH_REDIS_ADDR=127.0.0.1:6379 (Unix shells)
go test -bench=BenchmarkSaveRedisRealOptional -run ^$ ./internal/storage
```

## Coverage Threshold
A convenience Make target enforces a minimum total coverage (default 70%):
```bash
make coverage-threshold
```
Adjust the threshold by editing `thresh` in the Makefile target.
