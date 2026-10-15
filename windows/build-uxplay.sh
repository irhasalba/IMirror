#!/bin/bash
# Rebuild uxplay.exe + GStreamer bundle. Run inside MSYS2 UCRT64 (D:\msys64\ucrt64.exe):
#   bash //wsl.localhost/Debian/home/irhasalba/projects/personal/poc_iphone_mirror/windows/build-uxplay.sh
set -e
(cd /d/imirror-build/UxPlay && mkdir -p build && cd build && cmake -G Ninja -DCMAKE_BUILD_TYPE=Release .. >/dev/null && ninja)
bash "$(dirname "$0")/bundle.sh"
