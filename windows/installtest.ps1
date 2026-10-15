$d = "D:\imirror-installtest"
Start-Process D:\imirror-build\iMirror-Setup-1.0.0.exe -ArgumentList "/VERYSILENT","/SUPPRESSMSGBOXES","/CURRENTUSER","/DIR=$d","/TASKS=desktopicon" -Wait
"installed exe: " + (Test-Path "$d\iMirror.exe") + "  uxplay: " + (Test-Path "$d\uxplay\bin\uxplay.exe") + "  uninstaller: " + (Test-Path "$d\unins000.exe")
"start menu: " + (Test-Path "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\iMirror\iMirror.lnk") + "  desktop: " + (Test-Path ([Environment]::GetFolderPath("Desktop") + "\iMirror.lnk"))
"uninstall entry: " + ((Get-ChildItem HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall | % { Get-ItemProperty $_.PSPath } | ? DisplayName -like "iMirror*").DisplayName)
& powershell -NoProfile -ExecutionPolicy Bypass -File D:\imirror-build\guitest.ps1 "$d\iMirror.exe"
Start-Process "$d\unins000.exe" -ArgumentList "/VERYSILENT","/SUPPRESSMSGBOXES" -Wait; Start-Sleep 3
"after uninstall: dir=" + (Test-Path "$d\iMirror.exe") + " startmenu=" + (Test-Path "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\iMirror") + " desktop=" + (Test-Path ([Environment]::GetFolderPath("Desktop") + "\iMirror.lnk"))
