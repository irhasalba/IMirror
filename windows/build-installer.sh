#!/bin/bash
# Rebuild GUI, then package D:\imirror-build\dist\iMirror into an installer (from WSL).
set -e
cd "$(dirname "$0")"
./build-gui.sh
cp installer.iss /mnt/d/imirror-build/installer.iss
powershell.exe -NoProfile -Command '& D:\InnoSetup\ISCC.exe /Q D:\imirror-build\installer.iss; exit $LASTEXITCODE' | tr -d '\r'
ls -la /mnt/d/imirror-build/iMirror-Setup-*.exe
