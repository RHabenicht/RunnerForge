//go:build windows

package main

import "unsafe"

type rectangle struct{ Left, Top, Right, Bottom int32 }

func workArea() rectangle {
	r := rectangle{0, 0, 1025, 800}
	user32.NewProc("SystemParametersInfoW").Call(0x0030, 0, uintptr(unsafe.Pointer(&r)), 0)
	return r
}
func resizeMain(width, height int) {
	if width < 1 || height < 1 {
		return
	}
	// Top controls remain at their readable size; lower panels grow with the window.
	left := max(250, width*62/100-20)
	rightX := left + 40
	panelHeight := max(50, height-378)
	move := user32.NewProc("MoveWindow")
	move.Call(app.controls[ID_RESULTS], 20, 358, uintptr(left), uintptr(panelHeight), 1)
	move.Call(app.controls[ID_SUMMARY], uintptr(rightX), 358, uintptr(max(100, width-rightX-20)), uintptr(panelHeight), 1)
}
