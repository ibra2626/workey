#!/usr/bin/env bash
# Build the workey binary: generate templ code, compile Tailwind CSS, then build.
# Run from anywhere; paths are resolved relative to the repo root.
set -euo pipefail

cd "$(dirname "$0")/.."

# Tailwind standalone CLI lives in ./bin (gitignored). On Windows it is .exe.
tailwind="./bin/tailwindcss"
if [ -x "./bin/tailwindcss.exe" ]; then
	tailwind="./bin/tailwindcss.exe"
fi

echo "==> templ generate"
go tool templ generate

echo "==> tailwind (input.css -> static/app.css, minified)"
"$tailwind" -i web/styles/input.css -o web/static/app.css --minify

echo "==> go build"
go build ./...

echo "==> done"
