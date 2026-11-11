//go:build windows

package main

import (
	"nsrcgii/srcapp"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

func TestNativeCustomRaceEditor(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	inst, _, _ := procModule.Call(0)
	class := ptr("RunnerForgeEditorTest")
	wc := wndclassex{WndProc: syscall.NewCallback(wndProc), Instance: inst, Background: 7, ClassName: class}
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	if ok, _, err := procReg.Call(uintptr(unsafe.Pointer(&wc))); ok == 0 {
		t.Fatal(err)
	}
	hwnd, _, err := procCreate.Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(ptr(title))), WS_OVERLAPPEDWINDOW, 0, 0, 1025, 800, 0, 0, inst, 0)
	if hwnd == 0 {
		t.Fatal(err)
	}
	defer procDestroy.Call(hwnd)
	app.char.Race = "Night One"
	app.populateRaces()
	openRaceEditor()
	if raceWindow == 0 {
		t.Fatal("editor did not open")
	}
	defer func() {
		if raceWindow != 0 {
			procDestroy.Call(raceWindow)
		}
	}()
	if get(raceFields[1]) != "Night One (custom)" || get(raceFields[2]) != "15" || get(raceFields[11]) != "2" {
		t.Fatal("built-in template not loaded")
	}
	set(raceFields[1], "Night One (rebuilt)")
	set(raceFields[2], "10")
	set(raceFields[3], "House rules")
	send(raceWindow, WM_COMMAND, 100, 0)
	if raceWindow != 0 || !app.dirty || app.char.Race != "Night One (rebuilt)" || app.char.BuildPoints == nil {
		t.Fatal("editor did not apply race")
	}
	if app.char.CustomRace.BuildPointCost != 10 || app.char.EffectiveAttribute("Quickness") != 3 {
		t.Fatal("custom mechanics not applied")
	}
	set(app.controls[ID_BP_BUDGET], "20")
	app.update()
	if !strings.Contains(get(app.controls[ID_SUMMARY]), "budget exceeded") {
		t.Fatal("lowered budget did not flag overspend")
	}
	set(app.controls[ID_BP_BUDGET], "120")
	app.update()
	if strings.Contains(get(app.controls[ID_SUMMARY]), "budget exceeded") {
		t.Fatal("raised budget retained overspend")
	}
	path := filepath.Join(t.TempDir(), "character.json")
	if err := srcapp.SaveJSON(path, app.char); err != nil {
		t.Fatal(err)
	}
	loaded, err := srcapp.LoadJSON(path)
	if err != nil || loaded.CustomRace == nil || loaded.CustomRace.BuildPointCost != 10 {
		t.Fatalf("save/load failed: %v", err)
	}
	openRaceEditor()
	set(raceFields[2], "1")
	send(raceWindow, WM_COMMAND, 102, 0)
	if app.char.CustomRace.BuildPointCost != 10 {
		t.Fatal("cancel modified character")
	}
	openContents()
	if contentsWindow == 0 {
		t.Fatal("contents editor did not open")
	}
	defer func() {
		if contentsWindow != 0 {
			contentsDirty = false
			procDestroy.Call(contentsWindow)
		}
	}()
	set(contentsFields[10], "Physics")
	set(contentsFields[11], "3")
	set(contentsFields[12], "Knowledge")
	if err := contentsApply(); err != nil {
		t.Fatal(err)
	}
	if len(app.char.Skills) != 1 || app.char.Skills[0].Class() != "Knowledge" {
		t.Fatal("knowledge skill not added")
	}
	contentsIndex = 0
	contentsLoad()
	set(contentsFields[11], "4")
	if err := contentsApply(); err != nil {
		t.Fatal(err)
	}
	if len(app.char.Skills) != 1 || app.char.Skills[0].Rating != 4 {
		t.Fatal("skill edit duplicated entry")
	}
	contentsIndex = 0
	send(contentsWindow, WM_COMMAND, 102, 0)
	if len(app.char.Skills) != 0 {
		t.Fatal("selected skill not removed")
	}
	contentsGroup = "Character"
	contentsRefresh()
	set(contentsFields[10], "Test real name")
	set(contentsFields[11], "30")
	set(contentsFields[15], "Test character notes")
	set(contentsFields[16], "2")
	if err := contentsApply(); err != nil {
		t.Fatal(err)
	}
	if app.char.RealName != "Test real name" || app.char.Notes != "Test character notes" || app.char.BuildPoints.ExtraKnowledge != 2 {
		t.Fatal("character details were not saved")
	}
	app.char.Career = true
	app.char.Attributes["Quickness"] = 4
	app.char.Skills = []srcapp.Rated{{Name: "Pistols", Rating: 3, Specialization: "Beretta"}}
	app.populate()
	app.dirty = true
	contentsGroup = "Skills"
	contentsRefresh()
	contentsIndex = 0
	contentsLoad()
	send(contentsWindow, WM_COMMAND, 104, 0)
	if app.char.Karma != 24 || app.char.Skills[0].Rating != 4 || app.char.Skills[0].SpecializationRating() != 5 {
		t.Fatal("native advancement action failed")
	}
	contentsGroup = "Attributes"
	contentsRefresh()
	contentsIndex = 0
	contentsLoad()
	send(contentsWindow, WM_COMMAND, 104, 0)
	if app.char.Attributes["Body"] != 2 || app.char.Karma != 20 {
		t.Fatal("attribute advancement overwritten by form sync")
	}
	contentsDirty = false
	procDestroy.Call(contentsWindow)
	resizeMain(1000, 550)
	var rect rectangle
	user32.NewProc("GetWindowRect").Call(app.controls[ID_RESULTS], uintptr(unsafe.Pointer(&rect)))
	if rect.Bottom-rect.Top > 180 {
		t.Fatal("lower panel exceeds resized window")
	}

}
