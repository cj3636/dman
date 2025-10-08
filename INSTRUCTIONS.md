## **📘 Optimized Prompt for AI Agent**

**Goal:**
Create a **simple Go application** (CLI + optional lightweight API server) to sync and manage dotfiles between multiple
Ubuntu servers over a LAN.

---

### **🧩 Core Features**

#### **Client Functions**

1. **health / status check** – verify authentication and confirm server health.
2. **login / logout** – obtain and clear local auth tokens.
3. **publish / install** –

    * *publish*: upload new or changed dotfiles to the server.
    * *install*: download and apply newer server-side files to the local system.
4. **upload / download** – manually send or retrieve a file/folder.
5. **compare** – compare local and server versions to detect changes.

#### **Server Functions**

* Support all client operations.
* Implement **basic but secure token-based authentication** (no user accounts or passwords).
* Maintain per-user dotfile storage and metadata for sync comparisons.
* Optionally expose a `/health` endpoint and a `/status` summary.

---

### **⚙️ Configuration**

* Use a **single unified config file** (preferably YAML; JSON or .env also acceptable).
* Config defines:

    * `auth_token`
    * `server_url`
    * Users and their dotfile paths.
    * Optional per-user overrides.

#### **Default Configuration**

```yaml
auth_token: "<token>"
server_url: "http://localhost:7099"

users:
  root:
    home: /root/
    include:
      - .bash_aliases
      - .bashrc
      - .dircolors
      - .gitconfig
      - .nano/
      - .nanorc
      - .oh-my-zsh/plugins/
      - .profile
      - .selected_editor
      - .zshrc
      - .zprofile
      - .zlogin
      - .zlogout

  cjserver:
    home: /home/cjserver/
    include: *same as above*
```

* **Trailing `/`** = folder (include all contents).
* **`.oh-my-zsh/plugins/`** includes only the plugins subfolder (create missing parent folders if needed).

---

### **🚀 Behavior & Performance**

* Only scan the explicitly configured files/folders.
* Transfer **only changed files**, not full directories, to optimize speed and bandwidth.
* Optionally compress payloads (gzip) but keep it simple and local-first.
* Minimal dependencies — prefer standard library + small third-party packages when justified.
* The app should be **LAN-friendly** but maintain reasonable security (TLS optional, tokens mandatory).

---

### **🧭 Setup / Initialization**

* Include an optional `init` or `setup` command to create the initial config interactively (set URL, token, and user
  paths).
* Ensure first-time setup is simple and user-friendly.

---

### **💡 Design Philosophy**

* Focus only on the described scope — **no backups, web UI, databases, or extras**.
* Code should be **idiomatic Go**, modular, and maintainable.
* Follow modern Go best practices:

    * Use Go modules.
    * Separate packages cleanly (e.g. `cmd/`, `internal/`, `pkg/`).
    * Graceful shutdown for the server.
    * Consistent error handling and logging.

---

### **🧰 Implementation Notes**

* A single binary may act as both client and server, with subcommands (`--mode client|server`).
* The configuration file must work for both roles.
* Command structure example:

  ```
  dotman status
  dotman login
  dotman publish
  dotman install
  dotman compare
  dotman upload <file>
  dotman download <file>
  dotman serve
  ```
* If two binaries (client/server) are simpler, that’s acceptable as long as they share a unified config.

---

### **✅ Summary**

Create a minimal yet capable **Go dotfile sync tool** that:

* Syncs only configured files for specific users.
* Verifies health/auth.
* Publishes, installs, compares, and transfers files efficiently.
* Uses a single configuration for everything.
* Balances simplicity, security, and maintainability for LAN use.

---

Got it—here’s a clean, production-leaning blueprint the AI can follow immediately.

# Ideal Project Layout (Go 1.22+ *any version OK*)

```
dotman/
├─ cmd/
│  ├─ dotman/                # single binary (client + server via subcommands)
│  │  └─ main.go
│  └─ (optional) dotmand/    # if you decide to split server later
│     └─ main.go
├─ internal/
│  ├─ app/                   # orchestration wiring (DI, startup, shutdown)
│  │  ├─ app.go
│  │  └─ mode_client.go / mode_server.go
│  ├─ cli/                   # cobra/urfave or stdlib flag subcommands
│  │  ├─ root.go
│  │  ├─ cmd_status.go
│  │  ├─ cmd_login.go
│  │  ├─ cmd_logout.go
│  │  ├─ cmd_publish.go
│  │  ├─ cmd_install.go
│  │  ├─ cmd_compare.go
│  │  ├─ cmd_upload.go
│  │  ├─ cmd_download.go
│  │  └─ cmd_serve.go
│  ├─ config/                # config load/validate/expand homedir
│  │  ├─ config.go
│  │  └─ defaults.go
│  ├─ auth/                  # token handling (client) & middleware (server)
│  │  ├─ token.go
│  │  └─ middleware.go
│  ├─ scan/                  # fast, targeted file discovery
│  │  └─ scanner.go
│  ├─ fsio/                  # filesystem I/O (safe write, perms, atomic ops)
│  │  └─ fs.go
│  ├─ diff/                  # hashing/mtime compare and change sets
│  │  └─ diff.go
│  ├─ transfer/              # HTTP client, gzip, resumable support (future)
│  │  └─ transfer.go
│  ├─ server/                # HTTP server, routes, handlers
│  │  ├─ server.go
│  │  ├─ routes.go
│  │  └─ handlers.go
│  └─ logx/                  # structured logging glue (zap/slog)
│     └─ log.go
├─ pkg/                      # stable, reusable packages (kept minimal)
│  └─ model/                 # DTOs shared by client/server
│     └─ types.go
├─ api/                      # OpenAPI (optional), HTTP docs
│  └─ openapi.yaml
├─ testdata/                 # fixtures
├─ .gitignore
├─ go.mod
├─ go.sum
├─ Makefile
└─ README.md
```

## Naming & Style Conventions

* **Module name:** `github.com/you/dotman` (rename as needed).
* **Package rules:** business logic in `internal/*`, DTOs in `pkg/model`.
* **Errors:** wrap with context (`fmt.Errorf("publish %s: %w", path, err)`).
* **Logging:** `log/slog` or `uber-go/zap` (choose one; avoid mixing).
* **Config keys:** lower-snake in YAML → Go structs use `yaml:"field_name"` tags.
* **HTTP:** JSON request/response, `application/json`, deterministic enums.

## Config (single file; YAML preferred)

```yaml
# dotman.yaml
auth_token: "CHANGEME"
server_url: "http://cjserver:8080"

users:
  root:
    home: /root/
    include:
      - .bash_aliases
      - .bashrc
      - .dircolors
      - .gitconfig
      - .nano/
      - .nanorc
      - .oh-my-zsh/plugins/
      - .profile
      - .selected_editor
      - .zshrc
      - .zprofile
      - .zlogin
      - .zlogout

  cjserver:
    home: /home/cjserver/
    include:
      - .bash_aliases
      - .bashrc
      - .dircolors
      - .gitconfig
      - .nano/
      - .nanorc
      - .oh-my-zsh/plugins/
      - .profile
      - .selected_editor
      - .zshrc
      - .zprofile
      - .zlogin
      - .zlogout
```

> Rule: scan **only** `include` entries; treat a trailing `/` as a folder (recursive). For `.oh-my-zsh/plugins/`, create
parents if missing but sync only that subfolder.

## CLI UX (single binary example)

```
dotman status
dotman login
dotman logout
dotman publish           # upload changed files only
dotman install           # download newer server files only
dotman compare           # show change set (added/modified/removed)
dotman upload <path>     # explicit one-off
dotman download <path>   # explicit one-off
dotman serve             # start API server
dotman init              # guided first-time setup (optional)
```

## Minimal HTTP API (server)

* `GET  /health` → `{ ok: true, version, time }`
* `GET  /status` (auth) → summary: users, counts, last sync
* `POST /compare` (auth) → client sends inventory; server replies with delta
* `POST /publish` (auth) → stream changed files (gzip supported)
* `POST /install` (auth) → request changed files; server streams back
* `PUT  /upload?path=...` (auth)
* `GET  /download?path=...` (auth)

**Auth:** static bearer token via `Authorization: Bearer <token>`; reject if missing/mismatch. Rate limit not required (
LAN), but easy to add.

## Core Interfaces (keep them tiny)

```go
package model

type UserSpec struct {
	Name    string `json:"name"`
	Home    string `json:"home"`
	Include []string `json:"include"`
}

type InventoryItem struct {
	User   string `json:"user"`
	Path   string `json:"path"`   // relative to user home
	Size   int64  `json:"size"`
	MTime  int64  `json:"mtime_unix"`
	Hash   string `json:"sha256"` // content hash
	IsDir  bool   `json:"is_dir"`
}

type CompareRequest struct {
	Users     []string        `json:"users"`
	Inventory []InventoryItem `json:"inventory"`
}

type ChangeType string
const (
	ChangeAdd    ChangeType = "add"
	ChangeModify ChangeType = "modify"
	ChangeDelete ChangeType = "delete"
	ChangeSame   ChangeType = "same"
)

type Change struct {
	User string     `json:"user"`
	Path string     `json:"path"`
	Type ChangeType `json:"type"`
}
```

```go
package diff

type Comparator interface {
	// Compare client inventory vs server state; return actionable delta.
	Compare(req model.CompareRequest) ([]model.Change, error)
}
```

```go
package scan

type Scanner interface {
	// Walk only configured include entries; return inventory with hashes.
	InventoryFor(users []UserSpec) ([]model.InventoryItem, error)
}
```

```go
package transfer

type Client interface {
	Publish(changes []model.Change) error
	Install(changes []model.Change) error
	Upload(user, relPath string, r io.Reader) error
	Download(user, relPath string, w io.Writer) error
}
```

## Change Detection Strategy (simple & fast)

* Inventory = `{relative path, size, mtime, sha256}`.
* **Compare rule:** if path missing → `add`; if present and `sha256` differs → `modify`; if present on server but not in
  client set (for publish flow) → ignore or mark as server-only (shown in compare).
* Transfers send **only files that changed** (entire file content, not binary deltas). Keep code structured so
  rsync-style chunking can be added later.

## Server Skeleton

```go
package server

func New(addr string, authToken string, store Store, log *slog.Logger) *http.Server {
	r := chi.NewRouter()
	r.Use(middleware.RealIP, middleware.RequestID, middleware.Recoverer)
	r.Get("/health", healthHandler)
	r.Group(func(prot chi.Router) {
		prot.Use(Bearer(authToken))
		prot.Get("/status", statusHandler(store))
		prot.Post("/compare", compareHandler(store))
		prot.Post("/publish", publishHandler(store))
		prot.Post("/install", installHandler(store))
		prot.Put("/upload", uploadHandler(store))
		prot.Get("/download", downloadHandler(store))
	})
	return &http.Server{ Addr: addr, Handler: r }
}
```

> Storage: keep it **flat on disk** mirroring `/<hostname>/<user>/<relpath>` or just `<user>/<relpath>` for simplicity;
no database needed.

## Client Flow (publish / install)

1. **status/health** → ensure token good, server alive.
2. **scan** → build inventory from config includes.
3. **compare** → get change set.
4. **publish** → send only `add/modify` from client to server.
5. **install** → request server copies for `add/modify` (server-newer).

## Init (first-run)

* `dotman init`:

    * ask `server_url`, generate/store `auth_token`, confirm default users/paths, write `dotman.yaml`.
    * print next steps: `dotman serve` on host; `dotman status` on client.

## Dependencies (keep minimal)

* Router: `github.com/go-chi/chi/v5` (tiny, solid) **or** stdlib `http`.
* YAML: `gopkg.in/yaml.v3`.
* Logging: std `log/slog`.
* Hashing: std `crypto/sha256`.
* CLI: stdlib `flag` **or** `spf13/cobra` if you want nicer UX.

## Makefile (quality gates)

```make
GO ?= go

.PHONY: build run test fmt vet lint
build: ## Build dotman
	$(GO) build -o bin/dotman ./cmd/dotman
run: build
	./bin/dotman
test:
	$(GO) test ./...
fmt:
	$(GO) fmt ./...
vet:
	$(GO) vet ./...
lint:
	golangci-lint run
```

## Security Posture (LAN-friendly)

* Require bearer token on all mutating endpoints.
* Recommend running **behind NGINX/Caddy with TLS** if exposed beyond LAN.
* Validate user/path strictly; **no path traversal**. Enforce a safe root.
