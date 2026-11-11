//go:build windows

package main

import (
	"fmt"
	"math"
	"nsrcgii/srcapp"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

var contentsWindow uintptr
var contentsFields map[int]uintptr
var contentsLabels map[int]uintptr
var contentsRegistered bool
var contentsGroup string
var contentsIndex = -1
var contentsLoading bool
var contentsDirty bool

func openContents() {
	if contentsWindow != 0 {
		return
	}
	app.sync()
	class := ptr("RunnerForgeContents")
	inst, _, _ := procModule.Call(0)
	if !contentsRegistered {
		wc := wndclassex{WndProc: syscall.NewCallback(contentsProc), Instance: inst, Background: 7, ClassName: class}
		wc.CbSize = uint32(unsafe.Sizeof(wc))
		if ok, _, _ := procReg.Call(uintptr(unsafe.Pointer(&wc))); ok == 0 {
			msgbox("Could not open character editor", MB_ICONWARNING)
			return
		}
		contentsRegistered = true
	}
	contentsWindow, _, _ = procCreate.Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(ptr("Character entries — RunnerForge"))), WS_OVERLAPPEDWINDOW|WS_VISIBLE, 80, 80, 760, 620, window, 0, inst, 0)
	if contentsWindow != 0 {
		user32.NewProc("EnableWindow").Call(window, 0)
	}
}
func contentCtl(parent uintptr, class, label string, style uint32, x, y, w, h, id int) uintptr {
	v, _, _ := procCreate.Call(0, uintptr(unsafe.Pointer(ptr(class))), uintptr(unsafe.Pointer(ptr(label))), uintptr(style|WS_CHILD|WS_VISIBLE), uintptr(x), uintptr(y), uintptr(w), uintptr(h), parent, uintptr(id), 0, 0)
	font, _, _ := procStock.Call(17)
	send(v, WM_SETFONT, font, 1)
	if id > 0 {
		contentsFields[id] = v
	}
	return v
}
func contentComboValue(id int) string {
	h := contentsFields[id]
	i := send(h, CB_GETCURSEL, 0, 0)
	if i == LB_ERR {
		return ""
	}
	b := make([]uint16, 256)
	send(h, CB_GETLBTEXT, i, uintptr(unsafe.Pointer(&b[0])))
	return syscall.UTF16ToString(b)
}
func contentsRefresh() {
	contentsLoading = true
	defer func() { contentsLoading = false }()
	list := contentsFields[2]
	send(list, LB_RESETCONTENT, 0, 0)
	if contentsGroup == "Skills" {
		for _, s := range app.char.Skills {
			label := fmt.Sprintf("%s  %d", s.Name, s.Rating)
			if s.Specialization != "" {
				label += fmt.Sprintf(" (%s %d)", s.Specialization, s.SpecializationRating())
			}
			send(list, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr(label))))
		}
	} else {
		for _, p := range *app.char.Possessions(contentsGroup) {
			send(list, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr(p.Name))))
		}
	}
	contentsIndex = -1
	contentsLoad()
}
func contentsLoad() {
	contentsLoading = true
	defer func() { contentsLoading = false; contentsDirty = false }()
	for id := 10; id <= 17; id++ {
		set(contentsFields[id], "")
	}
	labels := map[int]string{10: "Name", 11: "Rating / Force (0 if unused)", 12: "Source", 13: "Modifiers", 14: "Essence cost (cyberware)", 15: "Notes", 16: "Unit price ¥ (blank = unresolved)", 17: "Quantity"}
	skill := contentsGroup == "Skills"
	if skill {
		labels[11] = "Base rating (after specialization)"
		labels[12] = "Category: Active, Knowledge, Language"
		labels[13] = "Specialization (optional)"
		labels[14] = "Linked attribute (blank = catalogue)"
	}
	for id, label := range labels {
		set(contentsLabels[id], label)
		show := uintptr(5)
		if skill && id >= 15 {
			show = 0
		}
		procShow.Call(contentsLabels[id], show)
		procShow.Call(contentsFields[id], show)
	}
	if skill {
		s := srcapp.Rated{Rating: 1, Category: "Active"}
		if contentsIndex >= 0 && contentsIndex < len(app.char.Skills) {
			s = app.char.Skills[contentsIndex]
		}
		set(contentsFields[10], s.Name)
		set(contentsFields[11], strconv.Itoa(s.Rating))
		set(contentsFields[12], s.Class())
		set(contentsFields[13], s.Specialization)
		set(contentsFields[14], s.LinkedAttribute)
	} else {
		p := srcapp.Possession{Quantity: 1}
		list := *app.char.Possessions(contentsGroup)
		if contentsIndex >= 0 && contentsIndex < len(list) {
			p = list[contentsIndex]
		}
		set(contentsFields[10], p.Name)
		set(contentsFields[11], strconv.Itoa(p.Rating))
		set(contentsFields[12], p.Source)
		set(contentsFields[13], p.Modifiers)
		set(contentsFields[14], strconv.FormatFloat(p.EssenceCost, 'f', -1, 64))
		set(contentsFields[15], p.Notes)
		set(contentsFields[17], strconv.Itoa(p.Count()))
		if p.UnitCost != nil {
			set(contentsFields[16], strconv.Itoa(*p.UnitCost))
		}
	}
}
func contentsApply() error {
	name := strings.TrimSpace(get(contentsFields[10]))
	if name == "" {
		return fmt.Errorf("enter a name")
	}
	rating, err := strconv.Atoi(strings.TrimSpace(get(contentsFields[11])))
	if err != nil || rating < 0 || rating > 100 {
		return fmt.Errorf("enter a rating from 0 to 100")
	}
	if contentsGroup == "Skills" {
		category := strings.TrimSpace(get(contentsFields[12]))
		switch category {
		case "Active", "Knowledge", "Language", "Literacy":
		default:
			return fmt.Errorf("choose Active, Knowledge or Language")
		}
		s := srcapp.Rated{Name: name, Rating: rating, Category: category, Specialization: strings.TrimSpace(get(contentsFields[13])), LinkedAttribute: strings.TrimSpace(get(contentsFields[14]))}
		if s.Rating < 1 || s.PurchasedRating() > 6 {
			return fmt.Errorf("creation rating must be 1–6; a specialized base must be 1–5")
		}
		if s.Class() == "Active" && s.Link() == "" {
			return fmt.Errorf("enter the linked attribute for this custom skill")
		}
		if s.LinkedAttribute != "" {
			valid := s.LinkedAttribute == "Reaction"
			for _, a := range attrs {
				valid = valid || a == s.LinkedAttribute
			}
			if !valid {
				return fmt.Errorf("linked attribute must be Body, Quickness, Strength, Charisma, Intelligence, Willpower or Reaction")
			}
		}
		if contentsIndex < 0 {
			app.char.Skills = append(app.char.Skills, s)
		} else {
			app.char.Skills[contentsIndex] = s
		}
	} else {
		list := app.char.Possessions(contentsGroup)
		p := srcapp.Possession{}
		if contentsIndex >= 0 {
			p = (*list)[contentsIndex]
		}
		essence, err := strconv.ParseFloat(strings.TrimSpace(get(contentsFields[14])), 64)
		if err != nil || math.IsNaN(essence) || math.IsInf(essence, 0) || essence < 0 || essence > 100 {
			return fmt.Errorf("enter a nonnegative Essence cost")
		}
		quantity, err := strconv.Atoi(strings.TrimSpace(get(contentsFields[17])))
		if err != nil || quantity < 1 || quantity > 1000000 {
			return fmt.Errorf("quantity must be 1–1,000,000")
		}
		p.UnitCost = nil
		if price := strings.TrimSpace(get(contentsFields[16])); price != "" {
			v, err := strconv.Atoi(price)
			if err != nil || v < 0 || v > 1000000000 {
				return fmt.Errorf("enter a nonnegative whole-number unit price")
			}
			p.UnitCost = &v
		}
		p.Name = name
		p.Rating = rating
		p.Source = get(contentsFields[12])
		p.Modifiers = get(contentsFields[13])
		p.EssenceCost = essence
		p.Notes = get(contentsFields[15])
		p.Quantity = quantity
		if contentsIndex < 0 {
			*list = append(*list, p)
			if contentsGroup == "Cyberware" && app.char.LegacyEssence != nil {
				value := *app.char.LegacyEssence - essence
				app.char.LegacyEssence = &value
			}
		} else {
			if err := app.char.ReplacePossession(contentsGroup, contentsIndex, &p); err != nil {
				return err
			}
		}
	}
	app.dirty = true
	app.update()
	contentsRefresh()
	return nil
}
func contentsMayDiscard(hwnd uintptr) bool {
	if !contentsDirty {
		return true
	}
	result, _, _ := procMessage.Call(hwnd, uintptr(unsafe.Pointer(ptr("Discard unapplied entry edits?"))), uintptr(unsafe.Pointer(ptr("Character entries"))), MB_YESNOCANCEL|MB_ICONWARNING)
	return result == IDYES
}
func contentsProc(hwnd uintptr, ev uint32, wp, lp uintptr) uintptr {
	switch ev {
	case WM_CREATE:
		contentsFields = map[int]uintptr{}
		contentsLabels = map[int]uintptr{}
		contentsGroup = "Skills"
		contentsIndex = -1
		group := contentCtl(hwnd, "COMBOBOX", "", CBS_DROPDOWNLIST|WS_TABSTOP|WS_VSCROLL, 15, 15, 230, 300, 1)
		for _, s := range append([]string{"Skills"}, srcapp.PossessionGroups...) {
			send(group, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr(s))))
		}
		send(group, CB_SETCURSEL, 0, 0)
		contentCtl(hwnd, "LISTBOX", "", WS_BORDER|WS_VSCROLL|WS_TABSTOP|LBS_NOTIFY, 15, 55, 230, 440, 2)
		for id := 10; id <= 17; id++ {
			x := 265
			y := 15 + (id-10)*57
			w := 455
			if id == 17 {
				x = 540
				y = 414
				w = 180
			}
			if id == 16 {
				y = 414
				w = 255
			}
			if id == 15 {
				y = 300
			}
			contentsLabels[id] = contentCtl(hwnd, "STATIC", "", 0, x, y, w, 20, 0)
			style := uint32(WS_BORDER | WS_TABSTOP | ES_AUTOHSCROLL)
			height := 25
			if id == 15 {
				style = WS_BORDER | WS_TABSTOP | WS_VSCROLL | ES_MULTILINE | ES_AUTOVSCROLL
				height = 80
			}
			contentCtl(hwnd, "EDIT", "", style, x, y+21, w, height, id)
		}
		for i, label := range []string{"New entry", "Apply", "Remove", "Close"} {
			contentCtl(hwnd, "BUTTON", label, WS_TABSTOP, 15+i*180, 520, 165, 32, 100+i)
		}
		contentsRefresh()
		return 0
	case WM_COMMAND:
		id := int(wp & 0xffff)
		notification := int(wp >> 16 & 0xffff)
		if !contentsLoading && notification == EN_CHANGE && id >= 10 && id <= 17 {
			contentsDirty = true
		}
		if id == 1 && notification == CBN_SELCHANGE {
			if contentsMayDiscard(hwnd) {
				contentsGroup = contentComboValue(1)
				contentsRefresh()
			} else {
				for i, g := range append([]string{"Skills"}, srcapp.PossessionGroups...) {
					if g == contentsGroup {
						send(contentsFields[1], CB_SETCURSEL, uintptr(i), 0)
					}
				}
			}
		}
		if id == 2 && notification == 1 {
			if contentsMayDiscard(hwnd) {
				i := send(contentsFields[2], LB_GETCURSEL, 0, 0)
				if i != LB_ERR {
					contentsIndex = int(i)
					contentsLoad()
				}
			} else {
				send(contentsFields[2], 0x0186, uintptr(contentsIndex), 0)
			}
		}
		switch id {
		case 100:
			if contentsMayDiscard(hwnd) {
				contentsIndex = -1
				contentsLoad()
			}
		case 101:
			if err := contentsApply(); err != nil {
				procMessage.Call(hwnd, uintptr(unsafe.Pointer(ptr(err.Error()))), uintptr(unsafe.Pointer(ptr("Character entry"))), MB_ICONWARNING)
			}
		case 102:
			if contentsIndex >= 0 && contentsMayDiscard(hwnd) {
				if contentsGroup == "Skills" {
					app.char.Skills = append(app.char.Skills[:contentsIndex], app.char.Skills[contentsIndex+1:]...)
				} else {
					app.char.ReplacePossession(contentsGroup, contentsIndex, nil)
				}
				app.dirty = true
				app.update()
				contentsRefresh()
			}
		case 103:
			if contentsMayDiscard(hwnd) {
				procDestroy.Call(hwnd)
			}
		}
		return 0
	case WM_CLOSE:
		if contentsMayDiscard(hwnd) {
			procDestroy.Call(hwnd)
		}
		return 0
	case WM_DESTROY:
		contentsWindow = 0
		user32.NewProc("EnableWindow").Call(window, 1)
		user32.NewProc("SetForegroundWindow").Call(window)
		return 0
	}
	r, _, _ := procDef.Call(hwnd, uintptr(ev), wp, lp)
	return r
}
