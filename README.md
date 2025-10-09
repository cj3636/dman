# dman

Minimal dotfile sync tool (work-in-progress).

## Build

```bash
go build ./cmd/dman
```

## Usage

Provide a `dman.yaml` config then:

```
./dman serve
./dman status
./dman compare
./dman publish
./dman install
```

## Config Example

```yaml
auth_token: "CHANGEME"
server_url: "http://localhost:7099"
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
users:
  root:
    home: /root/
  cjserver:
    home: /home/cjserver/
```

## Bulk Sync

Use tar-based high throughput endpoints instead of per-file upload/download:

```
./dman publish --bulk
./dman install --bulk
```

These use:
- POST /publish  (Content-Type: application/x-tar) tar entries named user/relpath
- POST /install  (JSON CompareRequest -> tar stream response)

## Endpoints Summary

| Method | Path      | Auth | Description |
|--------|-----------|------|-------------|
| GET    | /health   | no   | Liveness check |
| GET    | /status   | yes  | Stored file count |
| POST   | /compare  | yes  | Compute change set |
| POST   | /publish  | yes  | Bulk upload tar |
| POST   | /install  | yes  | Bulk download tar (changed server files) |
| PUT    | /upload   | yes  | Single file upload |
| GET    | /download | yes  | Single file download |
