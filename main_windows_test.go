package main

import "testing"

// HideWindow sets SW_HIDE in STARTUPINFO, which hides uxplay's first window:
// the D3D11 video window. iPhone connects but nothing shows.
func TestUxplayProcAttrKeepsVideoWindowVisible(t *testing.T) {
	a := uxplayProcAttr()
	if a.HideWindow {
		t.Fatal("HideWindow must be false or the mirroring window is hidden")
	}
	if a.CreationFlags&createNoWindow == 0 {
		t.Fatal("CREATE_NO_WINDOW needed to suppress uxplay's console")
	}
}
