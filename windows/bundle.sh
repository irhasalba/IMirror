#!/bin/bash
# Assemble portable iMirror folder from MSYS2 UCRT64 build.
set -e
OUT=/d/imirror-build/dist/iMirror/uxplay
rm -rf "$OUT"; mkdir -p "$OUT/bin" "$OUT/lib/gstreamer-1.0" "$OUT/libexec/gstreamer-1.0"
cp /d/imirror-build/UxPlay/build/uxplay.exe "$OUT/bin/"
cp /ucrt64/libexec/gstreamer-1.0/gst-plugin-scanner.exe "$OUT/libexec/gstreamer-1.0/"
for p in coreelements app audioconvert audioresample volume autodetect libav audioparsers videoparsersbad \
         d3d11 d3d12 wasapi directsound videoconvertscale jpeg imagefreeze pango level videofilter playback typefindfunctions; do
  cp /ucrt64/lib/gstreamer-1.0/libgst$p.dll "$OUT/lib/gstreamer-1.0/"
done
# every non-system DLL any of these needs, recursively
ntldd -R "$OUT"/bin/*.exe "$OUT"/libexec/gstreamer-1.0/*.exe "$OUT"/lib/gstreamer-1.0/*.dll \
  | grep -io '[a-z]:\\msys64\\ucrt64\\bin\\[^ ]*\.dll' | sort -u | while read -r d; do cp "$(cygpath -u "$d")" "$OUT/bin/"; done
du -sh "$OUT"; ls "$OUT/bin" | wc -l
