// iMirror: Phone-Link-style iPhone screen mirroring for Windows.
//
// iPhone -> AirPlay (Screen Mirroring) -> UxPlay receiver -> GStreamer window.
// Phone Link can't mirror iPhone (see evidence/REPORT.md), so we act as an
// AirPlay receiver instead, which is the only path iOS opens to third parties
// without a companion iOS app.
package main

import (
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
