$b = "D:\imirror-build\dist\iMirror\uxplay"
Copy-Item D:\msys64\ucrt64\bin\gst-inspect-1.0.exe "$b\bin\"
$env:PATH = "$b\bin;C:\Windows\System32;C:\Windows"
$env:GST_PLUGIN_SYSTEM_PATH = "$b\lib\gstreamer-1.0"
$env:GST_PLUGIN_SCANNER = "$b\libexec\gstreamer-1.0\gst-plugin-scanner.exe"
$env:GST_REGISTRY = "$env:TEMP\imirror-smoke-registry.bin"
foreach ($e in "appsrc","queue","h264parse","h265parse","aacparse","avdec_h264","avdec_h265","avdec_aac","avdec_alac","videoconvert","videoscale","videoflip","d3d11videosink","d3d12videosink","autovideosink","wasapisink","directsoundsink","autoaudiosink","audioconvert","audioresample","volume","level","jpegdec","imagefreeze","textoverlay","playbin","fakesink") {
  & "$b\bin\gst-inspect-1.0.exe" --exists $e; "{0,-16} {1}" -f $e, $(if ($LASTEXITCODE -eq 0) {"OK"} else {"MISSING"})
}
& "$b\bin\gst-inspect-1.0.exe" -b | Select -First 5
Remove-Item "$b\bin\gst-inspect-1.0.exe"
