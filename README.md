# iMirror

Mirror your iPhone screen to Windows, Phone Link style (Phone Link itself can't mirror an iPhone).

```
iPhone -> AirPlay (Screen Mirroring) -> UxPlay -> GStreamer window (D3D11)
```

iMirror is a Go GUI ([lxn/walk](https://github.com/lxn/walk)) that runs [UxPlay](https://github.com/FDH2/UxPlay) as an AirPlay receiver, plus text clipboard sync over the LAN.

## Requirements (Windows)

- [Bonjour Print Services for Windows](https://support.apple.com/kb/DL999) (the "Bonjour Service" must be running)
- iPhone and PC on the same network
- Firewall: allow iMirror/uxplay on Private networks (the installer adds the rules)

On launch, the **System checks** panel shows the status of the above.

The UI is available in English and Bahasa Indonesia (**Language** dropdown in the toolbar; defaults to Indonesian on an Indonesian Windows, otherwise English, and the choice is saved in `%LOCALAPPDATA%\iMirror-lang.txt`). The log is always English.

## Usage

1. Open iMirror, set **Name shown on iPhone**, click **Start**.
2. On the iPhone: Control Center > Screen Mirroring > pick that name.

Fixed ports (`uxplay -p`): TCP 7000, 7001, 7100 · UDP 6000, 6001, 7011.

## Clipboard sync (iOS Shortcuts)

iMirror serves `http://<pc-ip>:7123/clip`, guarded by an `X-Token` header. The URL and token are shown in the **Clipboard sync** panel (with **Copy** buttons); the token is saved in `%LOCALAPPDATA%\iMirror-clip-token.txt`.

| Method | Action |
|--------|--------|
| `GET /clip`  | read the PC clipboard (text) |
| `POST /clip` | set the PC clipboard from the body (max 1 MB) |

In Shortcuts, use **Get Contents of URL** with the `X-Token` header.

## Build (from WSL)

Bundle layout: `iMirror.exe` + `uxplay/{bin,lib/gstreamer-1.0,libexec/gstreamer-1.0}`, output in `D:\imirror-build\dist\iMirror`.

```sh
go test ./...                       # logic tests (run on Linux)
windows/build-gui.sh                # Windows tests + build iMirror.exe
windows/build-installer.sh          # build GUI + Inno Setup installer
```

UxPlay + GStreamer are built in MSYS2 UCRT64: `windows/build-uxplay.sh` (calls `windows/bundle.sh`).

On non-Windows platforms the binary only prints a "Windows-only" message.

## License

UxPlay is GPL-3.0 licensed (shipped in the bundle as `UXPLAY-LICENSE-GPL3.txt`).
