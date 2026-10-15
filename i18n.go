package main

// UI strings. The log stays English-only so it reads the same in bug reports.
var strs = map[string]map[string]string{
	"en": {
		"stopped.caption":   "STOPPED",
		"stopped.title":     "Mirroring is off",
		"stopped.sub":       "Press Start, then connect from your iPhone",
		"waiting.caption":   "WAITING",
		"waiting.title":     "Waiting for iPhone",
		"waiting.sub":       "On iPhone: Screen Mirroring > choose “%s”",
		"connected.caption": "CONNECTED",
		"connected.title":   "iPhone is mirroring",
		"connected.sub":     "Press Stop to disconnect",
		"start":             "Start",
		"stop":              "Stop",
		"device":            "Device",
		"device.name":       "Name shown on iPhone",
		"howto":             "How to connect",
		"step1":             "1. Open Control Center on iPhone",
		"step2":             "2. Tap Screen Mirroring",
		"step3":             "3. Choose “%s”",
		"language":          "Language",
		"checks":            "System checks",
		"checks.ready":      "%d of %d ready",
		"uxplay.ok":         "uxplay.exe found",
		"uxplay.miss":       "uxplay.exe not found",
		"uxplay.sub":        `uxplay\bin`,
		"bonjour.ok":        "Bonjour Service running",
		"bonjour.ok.sub":    "Bonjour Print Services for Windows",
		"bonjour.miss":      "Bonjour Service not running",
		"bonjour.miss.sub":  "Install Bonjour Print Services for Windows from Apple",
		"firewall":          "Windows Firewall access",
		"firewall.ok.sub":   "Rule iMirror (uxplay) is in place",
		"firewall.miss.sub": "Allow it in the popup on first Start (Private network)",
		"clip.hint":         "Send text from iPhone with a Shortcut",
		"clip.off":          "off: %s",
		"copy":              "Copy",
		"show":              "Show",
		"hide":              "Hide",
		"clear":             "Clear",
	},
	"id": {
		"stopped.caption":   "BERHENTI",
		"stopped.title":     "Mirroring belum aktif",
		"stopped.sub":       "Tekan Mulai, lalu sambungkan dari iPhone",
		"waiting.caption":   "MENUNGGU",
		"waiting.title":     "Menunggu iPhone",
		"waiting.sub":       "Di iPhone: Screen Mirroring > pilih “%s”",
		"connected.caption": "TERHUBUNG",
		"connected.title":   "iPhone sedang di-mirror",
		"connected.sub":     "Tekan Berhenti untuk memutus",
		"start":             "Mulai",
		"stop":              "Berhenti",
		"device":            "Perangkat",
		"device.name":       "Nama yang tampil di iPhone",
		"howto":             "Cara menyambung",
		"step1":             "1. Buka Control Center di iPhone",
		"step2":             "2. Ketuk Screen Mirroring",
		"step3":             "3. Pilih “%s”",
		"language":          "Bahasa",
		"checks":            "Pemeriksaan sistem",
		"checks.ready":      "%d dari %d siap",
		"uxplay.ok":         "uxplay.exe ditemukan",
		"uxplay.miss":       "uxplay.exe tidak ditemukan",
		"uxplay.sub":        `uxplay\bin`,
		"bonjour.ok":        "Bonjour Service berjalan",
		"bonjour.ok.sub":    "Bonjour Print Services for Windows",
		"bonjour.miss":      "Bonjour Service tidak berjalan",
		"bonjour.miss.sub":  "Install Bonjour Print Services for Windows dari Apple",
		"firewall":          "Izin Windows Firewall",
		"firewall.ok.sub":   "Rule iMirror (uxplay) sudah ada",
		"firewall.miss.sub": "Izinkan di popup saat Start pertama (Private network)",
		"clip.hint":         "Kirim teks dari iPhone lewat Shortcut",
		"clip.off":          "mati: %s",
		"copy":              "Salin",
		"show":              "Tampilkan",
		"hide":              "Sembunyikan",
		"clear":             "Bersihkan",
	},
}

// langs is the language picker order.
var langs = []struct{ code, name string }{{"en", "English"}, {"id", "Bahasa Indonesia"}}

// tr looks up key in lang, falling back to English, then to the key itself.
func tr(lang, key string) string {
	if s, ok := strs[lang][key]; ok {
		return s
	}
	if s, ok := strs["en"][key]; ok {
		return s
	}
	return key
}
