package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

const tokenMask = '●'

var modeDot = map[string]walk.Color{"stopped": colGray, "waiting": colAmber, "connected": colGreen}

func (a *app) t(key string) string { return tr(a.lang, key) }

// label renders key through style and re-translates it on language change.
func (a *app) label(style func(string) Label, key string) Label {
	l := style(a.t(key))
	p := new(*walk.Label)
	l.AssignTo = p
	a.retext = append(a.retext, func() { (*p).SetText(a.t(key)) })
	return l
}

func (a *app) buildUI() error {
	a.lang = loadLang()
	a.btnFont, _ = walk.NewFont("Segoe UI", 10, walk.FontBold)
	a.iconFont, _ = walk.NewFont(iconFamily, 20, 0)
	err := MainWindow{
		AssignTo:   &a.mw,
		Title:      "iMirror",
		Icon:       2, // icon group ID from rsrc (manifest=1); walk's default looks for 7
		Size:       Size{Width: 820, Height: 780},
		MinSize:    Size{Width: 700, Height: 660},
		Background: fill(colBg),
		Font:       font(9, false),
		Layout:     VBox{Margins: pad(20, 12, 20, 20), Spacing: 14},
		ToolBar:    a.langToolBar(),
		Children: []Widget{
			a.headerCard(),
			row(14, a.deviceCard(), a.checksCard()),
			a.clipCard(),
			a.logPanel(),
		},
	}.Create()
	if err != nil {
		return err
	}
	for _, w := range []walk.Widget{a.btn, a.reveal, a.clear} {
		w.SetCursor(walk.CursorHand())
	}
	a.token.SendMessage(win.EM_SETPASSWORDCHAR, tokenMask, 0)
	flatten(a.log)
	a.show("stopped")
	a.renderReady()
	return nil
}

// langToolBar is a "Language: English" dropdown in the window's toolbar.
func (a *app) langToolBar() ToolBar {
	items := []MenuItem{}
	for _, l := range langs {
		code := l.code
		act := new(*walk.Action)
		a.langActs = append(a.langActs, act)
		items = append(items, Action{
			AssignTo: act, Text: l.name, Checkable: true, Checked: code == a.lang,
			OnTriggered: func() { a.setLang(code) },
		})
	}
	return ToolBar{
		ButtonStyle: ToolBarButtonTextOnly,
		Items:       []MenuItem{Menu{AssignActionTo: &a.langMenu, Text: a.langMenuText(), Items: items}},
	}
}

func (a *app) langMenuText() string {
	for _, l := range langs {
		if l.code == a.lang {
			return a.t("language") + ": " + l.name
		}
	}
	return a.t("language")
}

func (a *app) headerCard() Widget {
	return card(HBox{Margins: pad(20, 18, 20, 18), Spacing: 16},
		CustomWidget{
			AssignTo: &a.tile, MinSize: Size{Width: 56, Height: 56}, MaxSize: Size{Width: 56, Height: 56},
			ClearsBackground: true,
			PaintPixels: func(c *walk.Canvas, _ walk.Rectangle) error {
				return paintRounded(c, a.tile, colSubtle, glyphMonitor, a.iconFont, colText)
			},
		},
		column(2,
			row(6,
				Label{AssignTo: &a.dot, Text: "●", Font: font(8, false)},
				Label{AssignTo: &a.caption, Font: font(8, true), TextColor: colMuted},
				HSpacer{},
			),
			Label{AssignTo: &a.title, Font: font(15, true), TextColor: colText},
			Label{AssignTo: &a.sub, Font: font(10, false), TextColor: colMuted},
		),
		HSpacer{},
		CustomWidget{
			AssignTo:            &a.btn,
			MinSize:             Size{Width: 132, Height: 46},
			MaxSize:             Size{Width: 132, Height: 46},
			ClearsBackground:    true,
			InvalidatesOnResize: true,
			PaintPixels:         a.paintBtn,
			OnMouseUp: func(_, _ int, b walk.MouseButton) {
				if b == walk.LeftButton {
					a.toggle()
				}
			},
		},
	)
}

func (a *app) deviceCard() Widget {
	return card(cardBox(),
		a.label(heading, "device"),
		a.label(muted, "device.name"),
		LineEdit{AssignTo: &a.name, Text: "iMirror", Font: font(11, false), OnTextChanged: a.nameChanged},
		Composite{
			Background: fill(colSubtle),
			Layout:     VBox{Margins: pad(14, 12, 14, 12), Spacing: 4},
			Children: []Widget{
				a.label(func(s string) Label { return text(s, font(9, true), colText) }, "howto"),
				a.label(body, "step1"),
				a.label(body, "step2"),
				Label{AssignTo: &a.step3, Text: fmt.Sprintf(a.t("step3"), "iMirror"), TextColor: colText},
				HSpacer{}, // stretch the box to the card width
			},
		},
		VSpacer{},
	)
}

func (a *app) checksCard() Widget {
	ws := []Widget{
		row(0,
			a.label(heading, "checks"),
			HSpacer{},
			Label{AssignTo: &a.ready, Font: font(9, true)},
		),
	}
	for _, c := range a.checks {
		glyph, color := glyphOK, colGreen
		if !c.ok {
			glyph, color = glyphWarn, colOrange
		}
		subFont := font(9, false)
		if c.mono {
			subFont = mono(9)
		}
		ws = append(ws, Composite{
			Layout: HBox{MarginsZero: true, Spacing: 10, Alignment: AlignHNearVNear},
			Children: []Widget{
				text(glyph, icon(12), color),
				column(0,
					a.label(func(s string) Label { return text(s, font(10, false), colText) }, c.title),
					a.label(func(s string) Label { return text(s, subFont, colMuted) }, c.sub),
				),
				HSpacer{},
			},
		})
	}
	return card(cardBox(), append(ws, VSpacer{})...)
}

func (a *app) renderReady() {
	ready := 0
	for _, c := range a.checks {
		if c.ok {
			ready++
		}
	}
	color := colGreen
	if ready < len(a.checks) {
		color = colOrange
	}
	a.ready.SetText(fmt.Sprintf(a.t("checks.ready"), ready, len(a.checks)))
	a.ready.SetTextColor(color)
}

func (a *app) clipCard() Widget {
	field := func(assign **walk.LineEdit, span int) LineEdit {
		return LineEdit{AssignTo: assign, ReadOnly: true, ColumnSpan: span, Font: mono(11), Background: fill(colSubtle)}
	}
	copyBtn := func(src **walk.LineEdit) PushButton {
		p := new(*walk.PushButton)
		a.retext = append(a.retext, func() { (*p).SetText(a.t("copy")) })
		return PushButton{AssignTo: p, Text: a.t("copy"), OnClicked: func() { walk.Clipboard().SetText((*src).Text()) }}
	}
	return card(cardBox(),
		row(8,
			text(glyphClip, icon(12), colText),
			heading("Clipboard sync"),
			HSpacer{},
			a.label(muted, "clip.hint"),
		),
		Composite{
			Layout: Grid{Columns: 4, MarginsZero: true, Spacing: 8},
			Children: []Widget{
				Label{Text: "Endpoint", TextColor: colMuted, MinSize: Size{Width: 70}},
				field(&a.endpoint, 2),
				copyBtn(&a.endpoint),
				muted("X-Token"),
				field(&a.token, 1),
				link(&a.reveal, a.t("show"), colBlue, a.toggleReveal),
				copyBtn(&a.token),
			},
		},
	)
}

func (a *app) logPanel() Widget {
	return Composite{
		Background: fill(colLogBg),
		Layout:     VBox{Margins: pad(16, 12, 16, 12), Spacing: 8},
		Children: []Widget{
			row(0,
				text("LOG", font(8, true), colLogHead),
				HSpacer{},
				link(&a.clear, a.t("clear"), colLogLink, func() { a.log.SetText("") }),
			),
			TextEdit{
				AssignTo: &a.log, ReadOnly: true, VScroll: true, MinSize: Size{Height: 110},
				Font: mono(9), TextColor: colLogText, Background: fill(colLogBg),
			},
		},
	}
}

// show renders the header for mode: "stopped", "waiting" or "connected".
func (a *app) show(mode string) {
	a.mode = mode
	sub := a.t(mode + ".sub")
	if mode == "waiting" {
		sub = fmt.Sprintf(sub, a.deviceName())
	}
	a.dot.SetTextColor(modeDot[mode])
	a.caption.SetText(a.t(mode + ".caption"))
	a.title.SetText(a.t(mode + ".title"))
	a.sub.SetText(sub)
	// window text is the UI Automation name (screen readers, windows/guitest.ps1)
	a.btn.SendMessage(win.WM_SETTEXT, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(a.btnText()))))
	a.btn.Invalidate()
}

func (a *app) btnText() string {
	if a.proc != nil {
		return a.t("stop")
	}
	return a.t("start")
}

func (a *app) paintBtn(c *walk.Canvas, _ walk.Rectangle) error {
	color, glyph := colBlue, "▶  "
	if a.proc != nil {
		color, glyph = colRed, "■  "
	}
	return paintRounded(c, a.btn, color, glyph+a.btnText(), a.btnFont, colCard)
}

func (a *app) deviceName() string {
	if name := strings.TrimSpace(a.name.Text()); name != "" {
		return name
	}
	return "iMirror"
}

func (a *app) nameChanged() {
	if a.step3 != nil {
		a.step3.SetText(fmt.Sprintf(a.t("step3"), a.deviceName()))
	}
}

func (a *app) toggleReveal() {
	a.revealed = !a.revealed
	mask, key := tokenMask, "show"
	if a.revealed {
		mask, key = 0, "hide"
	}
	a.token.SendMessage(win.EM_SETPASSWORDCHAR, uintptr(mask), 0)
	a.reveal.SetText(a.t(key))
	a.token.Invalidate()
}

func (a *app) setLang(code string) {
	if code == a.lang {
		return
	}
	a.lang = code
	os.WriteFile(langFile(), []byte(code), 0o600)
	a.langMenu.SetText(a.langMenuText())
	for i, act := range a.langActs {
		(*act).SetChecked(langs[i].code == code)
	}
	for _, f := range a.retext {
		f()
	}
	a.show(a.mode)
	a.nameChanged()
	a.renderReady()
	a.clear.SetText(a.t("clear"))
	a.revealed = !a.revealed
	a.toggleReveal() // re-applies the Show/Hide text without flipping state
}

func langFile() string { return filepath.Join(os.Getenv("LOCALAPPDATA"), "iMirror-lang.txt") }

var procUILang = syscall.NewLazyDLL("kernel32.dll").NewProc("GetUserDefaultUILanguage")

// loadLang returns the saved choice, else Indonesian on an Indonesian Windows, else English.
func loadLang() string {
	if b, err := os.ReadFile(langFile()); err == nil {
		if code := strings.TrimSpace(string(b)); strs[code] != nil {
			return code
		}
	}
	if id, _, _ := procUILang.Call(); id&0x3ff == 0x21 { // LANG_INDONESIAN
		return "id"
	}
	return "en"
}
