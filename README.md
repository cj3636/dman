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
users:
  me:
    home: /home/me/
    include:
      - .bashrc
      - .zshrc
```
