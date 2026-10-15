; iMirror installer. Build: windows/build-installer.sh (or D:\InnoSetup\ISCC.exe installer.iss)
#define AppName "iMirror"
#define AppVersion "1.1.0"
#define Dist "D:\imirror-build\dist\iMirror"
#define FwRule "iMirror (uxplay)"
#define FwClipRule "iMirror (clipboard)"

[Setup]
AppId={{6B0E6A3C-3F57-4C8E-9B5E-1D2C7A9F0E41}
AppName={#AppName}
AppVersion={#AppVersion}
AppPublisher=iMirror
DefaultDirName={autopf}\{#AppName}
DefaultGroupName={#AppName}
DisableProgramGroupPage=yes
; admin by default (firewall rule); /CURRENTUSER or the dialog allows a no-admin install
PrivilegesRequired=admin
PrivilegesRequiredOverridesAllowed=dialog commandline
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
OutputDir=D:\imirror-build
OutputBaseFilename={#AppName}-Setup-{#AppVersion}
Compression=lzma2/ultra64
SolidCompression=yes
UninstallDisplayIcon={app}\iMirror.exe
LicenseFile={#Dist}\UXPLAY-LICENSE-GPL3.txt
WizardStyle=modern

[Tasks]
Name: desktopicon; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"

[Files]
Source: "{#Dist}\*"; DestDir: "{app}"; Flags: recursesubdirs ignoreversion

[Icons]
Name: "{group}\{#AppName}"; Filename: "{app}\iMirror.exe"
Name: "{group}\Uninstall {#AppName}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#AppName}"; Filename: "{app}\iMirror.exe"; Tasks: desktopicon

[Run]
; pre-allow inbound AirPlay on private networks so users never see the firewall popup
Filename: "netsh.exe"; Parameters: "advfirewall firewall add rule name=""{#FwRule}"" dir=in action=allow program=""{app}\uxplay\bin\uxplay.exe"" profile=private enable=yes"; Flags: runhidden; Check: IsAdminInstallMode
Filename: "netsh.exe"; Parameters: "advfirewall firewall add rule name=""{#FwClipRule}"" dir=in action=allow program=""{app}\iMirror.exe"" protocol=TCP localport=7123 profile=private enable=yes"; Flags: runhidden; Check: IsAdminInstallMode
Filename: "{app}\iMirror.exe"; Description: "{cm:LaunchProgram,{#AppName}}"; Flags: nowait postinstall skipifsilent

[UninstallRun]
Filename: "netsh.exe"; Parameters: "advfirewall firewall delete rule name=""{#FwRule}"""; Flags: runhidden; RunOnceId: "DelFwRule"; Check: IsAdminInstallMode
Filename: "netsh.exe"; Parameters: "advfirewall firewall delete rule name=""{#FwClipRule}"""; Flags: runhidden; RunOnceId: "DelFwClipRule"; Check: IsAdminInstallMode

[UninstallDelete]
Type: files; Name: "{localappdata}\iMirror-gst-registry.bin"
Type: files; Name: "{localappdata}\iMirror-clip-token.txt"

[Code]
function InitializeSetup(): Boolean;
begin
  Result := True;
  if not RegKeyExists(HKLM, 'SYSTEM\CurrentControlSet\Services\Bonjour Service') then
    MsgBox('Bonjour Service belum terpasang. iMirror butuh Bonjour agar iPhone bisa menemukan PC ini.' + #13#10 +
           'Install "Bonjour Print Services for Windows" dari situs Apple, lalu jalankan iMirror.',
           mbInformation, MB_OK);
end;
