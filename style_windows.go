package main

import (
	"syscall"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

// Palette and building blocks for the layout in ui_windows.go.
var (
	colBg      = walk.RGB(243, 244, 246)
	colCard    = walk.RGB(255, 255, 255)
	colSubtle  = walk.RGB(243, 244, 246)
	colText    = walk.RGB(17, 24, 39)
	colMuted   = walk.RGB(107, 114, 128)
	colBlue    = walk.RGB(29, 78, 216)
	colRed     = walk.RGB(220, 38, 38)
	colGreen   = walk.RGB(21, 128, 61)
	colOrange  = walk.RGB(194, 65, 12)
	colAmber   = walk.RGB(230, 140, 0)
	colGray    = walk.RGB(156, 163, 175)
	colLogBg   = walk.RGB(17, 19, 24)
	colLogText = walk.RGB(209, 213, 219)
	colLogHead = walk.RGB(229, 231, 235)
	colLogLink = walk.RGB(147, 197, 253)
)

// Segoe MDL2 Assets ships with Windows 10 and 11.
const (
	iconFamily   = "Segoe MDL2 Assets"
	glyphMonitor = ""
	glyphOK      = ""
	glyphWarn    = ""
	glyphClip    = ""
)

func font(size int, bold bool) Font { return Font{Family: "Segoe UI", PointSize: size, Bold: bold} }
func mono(size int) Font            { return Font{Family: "Consolas", PointSize: size} }
func icon(size int) Font            { return Font{Family: iconFamily, PointSize: size} }

func pad(l, t, r, b int) Margins { return Margins{Left: l, Top: t, Right: r, Bottom: b} }
func fill(c walk.Color) Brush    { return SolidColorBrush{Color: c} }

func text(s string, f Font, c walk.Color) Label { return Label{Text: s, Font: f, TextColor: c} }
func heading(s string) Label                    { return text(s, font(11, true), colText) }
func body(s string) Label                       { return text(s, font(9, false), colText) }
func muted(s string) Label                      { return text(s, font(9, false), colMuted) }

// link is a clickable label; give it walk.CursorHand() after Create.
func link(assign **walk.Label, s string, c walk.Color, onClick func()) Label {
	return Label{AssignTo: assign, Text: s, TextColor: c, OnMouseUp: func(int, int, walk.MouseButton) { onClick() }}
}

func row(spacing int, children ...Widget) Composite {
	return Composite{Layout: HBox{MarginsZero: true, Spacing: spacing}, Children: children}
}

func column(spacing int, children ...Widget) Composite {
	return Composite{Layout: VBox{MarginsZero: true, Spacing: spacing}, Children: children}
}

// card is a white panel; flat corners, a Win32 child can't be rounded cheaply.
func card(layout Layout, children ...Widget) Composite {
	return Composite{Background: fill(colCard), Layout: layout, Children: children}
}

func cardBox() VBox { return VBox{Margins: pad(20, 16, 20, 18), Spacing: 8} }

// paintRounded fills w with a rounded rect and centers text on it.
func paintRounded(c *walk.Canvas, w *walk.CustomWidget, bg walk.Color, s string, f *walk.Font, fg walk.Color) error {
	brush, err := walk.NewSolidColorBrush(bg)
	if err != nil {
		return err
	}
	defer brush.Dispose()
	b := w.ClientBoundsPixels()
	r := walk.IntFrom96DPI(10, c.DPI())
	if err := c.FillRoundedRectanglePixels(brush, b, walk.Size{Width: r, Height: r}); err != nil {
		return err
	}
	return c.DrawTextPixels(s, f, fg, b, walk.TextCenter|walk.TextVCenter|walk.TextSingleLine)
}

// flatten drops the 3D client edge and uses the dark scrollbar theme (Win10 1809+) for the dark log.
func flatten(w walk.Widget) {
	h := w.Handle()
	win.SetWindowTheme(h, syscall.StringToUTF16Ptr("DarkMode_Explorer"), nil)
	win.SetWindowLong(h, win.GWL_EXSTYLE, win.GetWindowLong(h, win.GWL_EXSTYLE)&^win.WS_EX_CLIENTEDGE)
	win.SetWindowPos(h, 0, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_NOZORDER|win.SWP_FRAMECHANGED)
}
