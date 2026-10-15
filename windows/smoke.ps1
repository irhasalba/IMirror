$b = "D:\imirror-build\dist\iMirror\uxplay"
$env:PATH = "$b\bin;C:\Windows\System32;C:\Windows"
$env:GST_PLUGIN_SYSTEM_PATH = "$b\lib\gstreamer-1.0"
$env:GST_PLUGIN_SCANNER = "$b\libexec\gstreamer-1.0\gst-plugin-scanner.exe"
$env:GST_REGISTRY = "$env:TEMP\imirror-smoke-registry.bin"
Remove-Item $env:GST_REGISTRY -ErrorAction SilentlyContinue
$p = Start-Process "$b\bin\uxplay.exe" -ArgumentList "-n","iMirrorTest","-p","-nh","-vs","d3d11videosink","-as","wasapisink" -RedirectStandardOutput "$env:TEMP\ux.out" -RedirectStandardError "$env:TEMP\ux.err" -PassThru -NoNewWindow
Start-Sleep 8
if (Get-Command dns-sd.exe -ErrorAction SilentlyContinue) { "--- dns-sd browse"; $d = Start-Process dns-sd.exe -ArgumentList "-B","_airplay._tcp" -RedirectStandardOutput "$env:TEMP\dnssd.out" -PassThru -NoNewWindow; Start-Sleep 4; Stop-Process $d -Force; Get-Content "$env:TEMP\dnssd.out" } else { "dns-sd.exe not available" }
"--- alive: " + (-not $p.HasExited)
Stop-Process $p -Force -ErrorAction SilentlyContinue
"--- stdout"; Get-Content "$env:TEMP\ux.out"; "--- stderr"; Get-Content "$env:TEMP\ux.err"
"--- registry bytes: " + (Get-Item $env:GST_REGISTRY -ErrorAction SilentlyContinue).Length
