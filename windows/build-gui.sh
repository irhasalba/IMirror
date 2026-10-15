#!/bin/bash
# Build iMirror.exe (Go GUI) from WSL straight into the Windows bundle.
set -e
cd "$(dirname "$0")/.."
go test ./...
# Windows-only tests (main_windows_test.go) must run on Windows itself
GOOS=windows GOARCH=amd64 go test -c -o /mnt/d/imirror-build/imirror-test.exe .
powershell.exe -NoProfile -Command '& D:\imirror-build\imirror-test.exe; exit $LASTEXITCODE' | tr -d '\r'
GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui -s -w" -o /mnt/d/imirror-build/dist/iMirror/iMirror.exe .
echo "OK -> D:\\imirror-build\\dist\\iMirror\\iMirror.exe"
