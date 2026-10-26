# WKY-12 — Deploy: Linux build, systemd unit, Nginx example

## Goal
Everything needed to run the single binary on Ubuntu behind Nginx with systemd; no Docker.

## Scope
- `scripts/build.sh` target for `GOOS=linux GOARCH=amd64 CGO_ENABLED=0` with `-trimpath -ldflags "-s -w -X main.version=..."`.
- `deploy/workey.service`: dedicated `workey` user, `EnvironmentFile=/etc/workey/workey.env` (mode 0600), `StateDirectory=workey` for the DB, `Restart=on-failure`, hardening (`NoNewPrivileges`, `ProtectSystem=strict`, `ProtectHome`, `PrivateTmp`, `ReadWritePaths` for state only).
- `deploy/nginx/workey.conf.example`: HTTPS server (certbot paths as placeholders), HTTP→HTTPS redirect, `proxy_pass` to 127.0.0.1:8080, forwarded headers, `/mcp` with `proxy_buffering off`, long `proxy_read_timeout`, `proxy_http_version 1.1` for streaming; HSTS; client body size limit.
- Server trusts `X-Forwarded-Proto` only from loopback (for Secure cookie logic) — implement or verify.
- `docs/deploy.md`: install steps, generating the token (`openssl rand`), `workey hash-password`, env file, enabling the unit, SQLite backup (`sqlite3 .backup` or copy with WAL checkpoint), upgrade steps.
- Non-goals: actually deploying to a server.

## Acceptance criteria
- [ ] Linux build succeeds from Windows.
- [ ] `systemd-analyze verify` and `nginx -t` run if available (WSL or a Linux box); otherwise state they were not run.
- [ ] `.env.example` lists every variable the unit needs; no secret in any file.

## Checks
- `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/workey-linux-amd64 ./cmd/workey`
- `go vet ./...`

## Progress
Blocked (2026-10-26, user): v1 runs only locally on Windows; a Linux server may come later. Do not start until the user lifts the hold. When resumed, re-check Go version, Nginx/systemd assumptions and the Secure-cookie/proxy notes against the code as it is then.
