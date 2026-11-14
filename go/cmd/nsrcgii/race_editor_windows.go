//go:build windows

package main

import (
	"fmt"
	"nsrcgii/srcapp"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

var raceWindow uintptr
var raceFields map[int]uintptr
var raceClassRegistered bool

func openRaceEditor() {
	if raceWindow != 0 {
		user32.NewProc("SetForegroundWindow").Call(raceWindow)
		return
	}
	app.sync()
	class := ptr("RunnerForgeRaceEditor")
	inst, _, _ := procModule.Call(0)
	if !raceClassRegistered {
		wc := wndclassex{WndProc: syscall.NewCallback(raceEditorProc), Instance: inst, Background: 7, ClassName: class}
		wc.CbSize = uint32(unsafe.Sizeof(wc))
		ok, _, _ := procReg.Call(uintptr(unsafe.Pointer(&wc)))
		if ok == 0 {
			msgbox("Could not open race editor", MB_ICONWARNING)
			return
		}
		raceClassRegistered = true
	}
	raceWindow, _, _ = procCreate.Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(ptr("Custom race — RunnerForge"))), (WS_OVERLAPPEDWINDOW&^0x00050000)|WS_VISIBLE, 0, 0, 560, 500, window, 0, inst, 0)
	if raceWindow != 0 {
		user32.NewProc("EnableWindow").Call(window, 0)
	}
}

func raceControl(parent uintptr, class, label string, style uint32, x, y, w, h, id int) uintptr {
	v, _, _ := procCreate.Call(0, uintptr(unsafe.Pointer(ptr(class))), uintptr(unsafe.Pointer(ptr(label))), uintptr(style|WS_CHILD|WS_VISIBLE), uintptr(x), uintptr(y), uintptr(w), uintptr(h), parent, uintptr(id), 0, 0)
	font, _, _ := procStock.Call(17)
	send(v, WM_SETFONT, font, 1)
	if id > 0 {
		raceFields[id] = v
	}
	return v
}

func editorRace() (*srcapp.RaceDefinition, error) {
	cost, err := strconv.Atoi(strings.TrimSpace(get(raceFields[2])))
	if err != nil {
		return nil, fmt.Errorf("Build Point cost must be a whole number")
	}
	r := &srcapp.RaceDefinition{Name: strings.TrimSpace(get(raceFields[1])), BuildPointCost: cost, Modifiers: map[string]int{}, Traits: get(raceFields[3]), Priority: app.char.RaceTemplate().Priority}
	for i, k := range attrs {
		n, err := strconv.Atoi(strings.TrimSpace(get(raceFields[10+i])))
		if err != nil {
			return nil, fmt.Errorf("%s modifier must be a whole number", k)
		}
		r.Modifiers[k] = n
	}
	return r, r.Validate()
}

func raceEditorProc(hwnd uintptr, ev uint32, wp, lp uintptr) uintptr {
	switch ev {
	case WM_CREATE:
		raceFields = map[int]uintptr{}
		r := app.char.RaceTemplate()
		if app.char.CustomRace == nil || app.char.CustomRace.Name != app.char.Race {
			r.Name += " (custom)"
		}
		label := func(s string, x, y, w int) { raceControl(hwnd, "STATIC", s, 0, x, y, w, 20, 0) }
		field := func(s string, x, y, w, id int) {
			raceControl(hwnd, "EDIT", s, WS_BORDER|WS_TABSTOP|ES_AUTOHSCROLL, x, y, w, 25, id)
		}
		label("Race name", 20, 15, 200)
		field(r.Name, 20, 38, 340, 1)
		label("Build Point cost", 380, 15, 140)
		field(strconv.Itoa(r.BuildPointCost), 380, 38, 130, 2)
		label("Racial modifiers (added to base attributes)", 20, 80, 490)
		for i, k := range attrs {
			x := 20 + (i%3)*165
			y := 110 + (i/3)*65
			label(k, x, y, 150)
			field(strconv.Itoa(r.Modifiers[k]), x, y+23, 145, 10+i)
		}
		label("Traits and notes (descriptive; no automatic rule effects)", 20, 245, 500)
		raceControl(hwnd, "EDIT", r.Traits, WS_BORDER|WS_TABSTOP|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL, 20, 270, 490, 95, 3)
		raceControl(hwnd, "BUTTON", "Use race", WS_TABSTOP, 20, 390, 140, 32, 100)
		raceControl(hwnd, "BUTTON", "Export race…", WS_TABSTOP, 180, 390, 150, 32, 101)
		raceControl(hwnd, "BUTTON", "Cancel", WS_TABSTOP, 350, 390, 140, 32, 102)
		return 0
	case WM_COMMAND:
		id := int(wp & 0xffff)
		if id == 102 {
			procDestroy.Call(hwnd)
			return 0
		}
		if id == 100 || id == 101 {
			r, err := editorRace()
			if err != nil {
				procMessage.Call(hwnd, uintptr(unsafe.Pointer(ptr(err.Error()))), uintptr(unsafe.Pointer(ptr("Custom race"))), MB_ICONWARNING)
				return 0
			}
			if id == 101 {
				path := pickFile(true, "Custom race (*.json)\x00*.json\x00\x00", "json")
				if path != "" {
					if err := srcapp.SaveRaceDefinition(path, *r); err != nil {
						msgbox(err.Error(), MB_ICONWARNING)
					}
				}
			} else {
				app.char.CustomRace = r
				app.char.Race = r.Name
				app.populateRaces()
				selectCombo(ID_METHOD, "Build Points")
				app.dirty = true
				app.update()
				procDestroy.Call(hwnd)
			}
		}
		return 0
	case WM_CLOSE:
		procDestroy.Call(hwnd)
		return 0
	case WM_DESTROY:
		raceWindow = 0
		user32.NewProc("EnableWindow").Call(window, 1)
		user32.NewProc("SetForegroundWindow").Call(window)
		return 0
	}
	r, _, _ := procDef.Call(hwnd, uintptr(ev), wp, lp)
	return r
}
