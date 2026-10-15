// iMirror: Phone-Link-style iPhone screen mirroring for Windows.
//
// iPhone -> AirPlay (Screen Mirroring) -> UxPlay receiver -> GStreamer window.
// Phone Link can't mirror iPhone (see evidence/REPORT.md), so we act as an
// AirPlay receiver instead, which is the only path iOS opens to third parties
// without a companion iOS app.
package main

import (
	"crypto/subtle"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// Fixed legacy AirPlay ports (uxplay -p): TCP 7100,7000,7001  UDP 6000,6001,7011.
// Fixed ports are what make a Windows firewall rule possible.
var uxplayArgs = []string{"-p", "-nh", "-vs", "d3d11videosink", "-as", "wasapisink"}

// uxplayCmd returns exe, args and extra env for the bundle layout
// <app>/uxplay/{bin/uxplay.exe, lib/gstreamer-1.0/*.dll}.
func uxplayCmd(bundle, name, path, localAppData string) (string, []string, []string) {
	exe := filepath.Join(bundle, "bin", "uxplay.exe")
	args := append([]string{"-n", name}, uxplayArgs...)
	env := []string{
		"PATH=" + filepath.Join(bundle, "bin") + ";" + path,
		"GST_PLUGIN_SYSTEM_PATH=" + filepath.Join(bundle, "lib", "gstreamer-1.0"),
		"GST_PLUGIN_SCANNER=" + filepath.Join(bundle, "libexec", "gstreamer-1.0", "gst-plugin-scanner.exe"),
		"GST_REGISTRY=" + filepath.Join(localAppData, "iMirror-gst-registry.bin"),
	}
	return exe, args, env
}

type state int

const (
	stateNone state = iota
	stateConnected
	stateWaiting
)

// statusFor maps a uxplay log line to a UI state.
// ponytail: status from uxplay log text; breaks if upstream rewords logs, upgrade = parse -d debug output
func statusFor(line string) state {
	low := strings.ToLower(line)
	switch {
	case strings.Contains(low, "connection request"), strings.Contains(low, "accepted"):
		return stateConnected
	case strings.Contains(low, "connection closed"), strings.Contains(low, "teardown"):
		return stateWaiting
	}
	return stateNone
}

// Clipboard sync: iPhone has no API to read our clipboard (no Continuity on Windows),
// so an iOS Shortcut calls GET/POST /clip on the LAN with the X-Token header.
const (
	clipPort = "7123"
	maxClip  = 1 << 20
)

func clipHandler(token string, get func() (string, error), set func(string) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token == "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Token")), []byte(token)) != 1 {
			http.Error(w, "bad token", http.StatusUnauthorized)
			return
		}
		switch r.Method {
		case http.MethodGet:
			text, err := get()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			io.WriteString(w, text)
		case http.MethodPost:
			body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxClip))
			if err != nil {
				http.Error(w, "too large", http.StatusRequestEntityTooLarge)
				return
			}
			if err := set(string(body)); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "GET or POST", http.StatusMethodNotAllowed)
		}
	})
}
