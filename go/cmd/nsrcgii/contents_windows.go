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

func editorGroups() []string {
	return append([]string{"Skills", "Character", "Attributes", "Ghoul creation", "Shapeshifter creation"}, srcapp.PossessionGroups...)
}

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
	bounds := workArea()
	height := min(620, int(bounds.Bottom-bounds.Top))
	contentsWindow, _, _ = procCreate.Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(ptr("Character entries — RunnerForge"))), (WS_OVERLAPPEDWINDOW&^0x00050000)|WS_VISIBLE, uintptr(bounds.Left), uintptr(bounds.Top), 760, uintptr(height), window, 0, inst, 0)
	if contentsWindow != 0 {
		user32.NewProc("EnableWindow").Call(window, 0)
	}
}
func contentCtl(parent uintptr, class, label string, style uint32, x, y, w, h, id int) uintptr {
	bounds := workArea()
	height := min(620, int(bounds.Bottom-bounds.Top)) - 40
	y = y * height / 580
	h = max(18, h*height/580)
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
			if specs := s.SpecializationText(); specs != "" {
				label += " (" + specs + ")"
			}
			send(list, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr(label))))
		}
	} else if contentsGroup == "Attributes" {
		for _, name := range attrs {
			send(list, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr(name))))
		}
	} else if contentsGroup == "Shapeshifter creation" {
		send(list, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr("Shapeshifter forms and purchases"))))
	} else if contentsGroup == "Ghoul creation" {
		send(list, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr("Starting ghoul settings"))))
	} else if contentsGroup == "Character" {
		send(list, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr("Character details"))))
	} else {
		for _, p := range *app.char.Possessions(contentsGroup) {
			send(list, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr(p.Name))))
		}
	}
	contentsIndex = -1
	if contentsGroup == "Character" {
		set(contentsFields[100], "Begin career")
	} else {
		set(contentsFields[100], "New entry")
	}
	contentsLoad()
}
func contentsLoad() {
	enable := user32.NewProc("EnableWindow")
	for _, id := range []int{100, 101, 102} {
		allowed := uintptr(1)
		if contentsGroup == "Attributes" || ((contentsGroup == "Ghoul creation" || contentsGroup == "Shapeshifter creation") && (id != 101 || (contentsGroup == "Ghoul creation" && app.char.Career))) {
			allowed = 0
		}
		enable.Call(contentsFields[id], allowed)
	}
	set(contentsFields[104], "Raise skill (karma)")
	if contentsGroup == "Attributes" {
		set(contentsFields[104], "Raise attribute (karma)")
	}

	contentsLoading = true
	defer func() { contentsLoading = false; contentsDirty = false }()
	for id := 10; id <= 17; id++ {
		set(contentsFields[id], "")
	}
	send(contentsFields[15], 0x00CF, 0, 0) // EM_SETREADONLY; reset when changing groups.
	labels := map[int]string{10: "Name", 11: "Rating / Force (0 if unused)", 12: "Source", 13: "Modifiers (RCT / INI supported)", 14: "Essence cost (cyberware)", 15: "Notes", 16: "Unit price ¥ (blank = unresolved)", 17: "Quantity"}
	if contentsGroup == "Contacts" {
		labels[16] = "Purchase price ¥ (0 = free, blank = unresolved)"
	}
	if contentsGroup == "Lifestyles" {
		labels[16] = "Price per paid period ¥ (blank = unresolved)"
		labels[17] = "Paid periods (1 for a lump-sum purchase)"
	}
	if contentsGroup == "Shapeshifter creation" {
		loadShapeshifterContents()
		return
	}
	if contentsGroup == "Ghoul creation" {
		send(contentsFields[15], 0x00CF, 1, 0)
		ghoulLabels := map[int]string{10: "Character type: Ghoul or None", 11: "Third-stage test successes (enter actual result)", 12: "GM attribute alternative approved: yes / no", 13: "Two-point sacrifices (comma-separated attributes)", 15: "Companion pp.33–34"}
		for id := 10; id <= 17; id++ {
			show := uintptr(0)
			if label, ok := ghoulLabels[id]; ok {
				set(contentsLabels[id], label)
				show = 5
			}
			procShow.Call(contentsLabels[id], show)
			procShow.Call(contentsFields[id], show)
		}
		set(contentsFields[10], "None")
		set(contentsFields[11], "0")
		set(contentsFields[12], "no")
		if g := app.char.Ghoul; g != nil {
			set(contentsFields[10], "Ghoul")
			set(contentsFields[11], strconv.Itoa(g.Successes))
			if g.GMAlternative {
				set(contentsFields[12], "yes")
			}
			set(contentsFields[13], strings.Join(g.Sacrifices, ", "))
		}
		set(contentsFields[15], fmt.Sprintf("Build the ordinary character first. Roll %d dice against TN 6 (Willpower + 2); no Karma. Ghoul costs +10 BP. If Intelligence/Charisma fall below 1, the GM may permit a two-point sacrifice per lost point. Choices lock at career start. Starting implants are post-transformation.", app.char.GhoulTestDice()))
		return
	}
	if contentsGroup == "Attributes" {
		for id := 10; id <= 17; id++ {
			show := uintptr(0)
			if id <= 11 {
				show = 5
			}
			procShow.Call(contentsLabels[id], show)
			procShow.Call(contentsFields[id], show)
		}
		set(contentsLabels[10], "Attribute")
		if app.char.Shapeshifter != nil {
			set(contentsLabels[10], "Attribute (active "+app.char.Shapeshifter.Form+" form)")
		}
		set(contentsLabels[11], "Current natural rating")
		if contentsIndex >= 0 && contentsIndex < len(attrs) {
			name := attrs[contentsIndex]
			set(contentsFields[10], name)
			set(contentsFields[11], strconv.Itoa(app.char.EffectiveAttribute(name)))
		}
		contentsDirty = false
		return
	}
	if contentsGroup == "Character" {
		labels = map[int]string{10: "Real name (optional)", 11: "Unspent karma", 12: "Karma pool", 13: "Spell Points purchased (¥25,000 each)", 14: "Spell Points sold (5 for 1 BP)", 15: "Character notes", 16: "Extra Knowledge BP (optional)", 17: "Tradition / aspect"}
		for id, label := range labels {
			set(contentsLabels[id], label)
			procShow.Call(contentsLabels[id], 5)
			procShow.Call(contentsFields[id], 5)
		}
		set(contentsFields[10], app.char.RealName)
		set(contentsFields[11], strconv.Itoa(app.char.Karma))
		set(contentsFields[12], strconv.Itoa(app.char.KarmaPool))
		set(contentsFields[13], strconv.Itoa(app.char.SpellPointsPurchased))
		set(contentsFields[14], strconv.Itoa(app.char.SpellPointsSold))
		set(contentsFields[15], app.char.Notes)
		set(contentsFields[16], strconv.Itoa(extraKnowledge(app.char)))
		set(contentsFields[17], app.char.Tradition)
		return
	}
	skill := contentsGroup == "Skills"
	pointEntry := contentsGroup == "Adept powers" || contentsGroup == "Edges / Flaws"
	if contentsGroup == "Cyberware" {
		labels[12] = "Grade: Basic, Alpha, Beta, Delta (used prefix)"
		labels[14] = "Base Essence cost (before grade)"
		labels[16] = "Base unit price (before grade)"
	}
	if contentsGroup == "Bioware" {
		labels[14] = "Bio Index (blank = unresolved)"
	}
	if pointEntry {
		labels[16] = "Total point cost (Flaws: negative)"
	}
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
		if contentsGroup == "Cyberware" {
			set(contentsFields[12], p.Grade)
		}
		set(contentsFields[13], p.Modifiers)
		set(contentsFields[14], strconv.FormatFloat(p.EssenceCost, 'f', -1, 64))
		if contentsGroup == "Bioware" {
			set(contentsFields[14], "")
			if p.BioIndex != nil {
				set(contentsFields[14], strconv.FormatFloat(*p.BioIndex, 'f', -1, 64))
			}
		}
		set(contentsFields[15], p.Notes)
		set(contentsFields[17], strconv.Itoa(p.Count()))
		if pointEntry && p.PointCost != nil {
			set(contentsFields[16], strconv.FormatFloat(*p.PointCost, 'f', -1, 64))
		} else if p.UnitCost != nil {
			set(contentsFields[16], strconv.Itoa(*p.UnitCost))
		}
	}
}
func extraKnowledge(ch *srcapp.Character) int {
	if ch.BuildPoints != nil {
		return ch.BuildPoints.ExtraKnowledge
	}
	return 0
}
func contentsApply() error {
	if contentsGroup == "Shapeshifter creation" {
		return applyShapeshifterContents()
	}
	if contentsGroup == "Ghoul creation" {
		var g *srcapp.GhoulCreation
		switch strings.ToLower(strings.TrimSpace(get(contentsFields[10]))) {
		case "none":
		case "ghoul":
			n, err := strconv.Atoi(strings.TrimSpace(get(contentsFields[11])))
			if err != nil {
				return fmt.Errorf("enter the actual third-stage test successes")
			}
			approved := strings.ToLower(strings.TrimSpace(get(contentsFields[12])))
			if approved != "yes" && approved != "no" {
				return fmt.Errorf("enter yes or no for GM permission")
			}
			g = &srcapp.GhoulCreation{Successes: n, GMAlternative: approved == "yes"}
			for _, name := range strings.Split(get(contentsFields[13]), ",") {
				name = strings.TrimSpace(name)
				if name != "" {
					g.Sacrifices = append(g.Sacrifices, name)
				}
			}
		default:
			return fmt.Errorf("choose Ghoul or None")
		}
		if err := app.char.ConfigureGhoul(g); err != nil {
			return err
		}
		app.dirty = true
		app.update()
		contentsRefresh()
		return nil
	}
	if contentsGroup == "Character" {
		numbers := map[int]int{}
		for _, id := range []int{11, 12, 13, 14, 16} {
			v, err := strconv.Atoi(strings.TrimSpace(get(contentsFields[id])))
			if err != nil || v < 0 || v > 1000000 {
				return fmt.Errorf("karma and point fields require nonnegative whole numbers")
			}
			numbers[id] = v
		}
		if app.char.BuildPoints == nil && numbers[16] > 0 {
			return fmt.Errorf("extra Knowledge points require Build Point creation")
		}
		app.char.RealName = get(contentsFields[10])
		app.char.Karma = numbers[11]
		app.char.KarmaPool = numbers[12]
		app.char.SpellPointsPurchased = numbers[13]
		app.char.SpellPointsSold = numbers[14]
		app.char.Notes = get(contentsFields[15])
		app.char.Tradition = get(contentsFields[17])
		if app.char.BuildPoints != nil {
			app.char.BuildPoints.ExtraKnowledge = numbers[16]
		}
		app.dirty = true
		app.update()
		contentsRefresh()
		return nil
	}

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
		if s.Rating < 1 || (!app.char.Career && s.PurchasedRating() > 6) {
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
		if app.char.Career && contentsIndex >= 0 && app.char.Skills[contentsIndex].Specialization != s.Specialization {
			return fmt.Errorf("use Learn / raise specialization to add a specialization during play")
		}
		if contentsIndex >= 0 && app.char.Skills[contentsIndex].Specialization == s.Specialization {
			s.SpecializationValue = app.char.Skills[contentsIndex].SpecializationValue
			s.AdditionalSpecializations = app.char.Skills[contentsIndex].AdditionalSpecializations
		}
		if contentsIndex < 0 && app.char.Career {
			if s.Rating != 1 || s.Specialization != "" {
				return fmt.Errorf("learn a new skill at rating 1, then advance it with karma")
			}
			if app.char.Karma < 1 {
				return fmt.Errorf("learning a new skill costs 1 karma")
			}
			app.char.Karma--
			app.char.Advancements = append(app.char.Advancements, srcapp.Advancement{Kind: "Skill", Name: s.Name, From: 0, To: 1, Karma: 1})
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
		essenceText := strings.TrimSpace(get(contentsFields[14]))
		if contentsGroup == "Bioware" && essenceText == "" {
			essenceText = "0"
		}
		essence, err := strconv.ParseFloat(essenceText, 64)
		if err != nil || math.IsNaN(essence) || math.IsInf(essence, 0) || essence < 0 || essence > 100 {
			return fmt.Errorf("enter a nonnegative Essence cost")
		}
		quantity, err := strconv.Atoi(strings.TrimSpace(get(contentsFields[17])))
		if err != nil || quantity < 1 || quantity > 1000000 {
			return fmt.Errorf("quantity must be 1–1,000,000")
		}
		p.UnitCost = nil
		p.PointCost = nil
		if price := strings.TrimSpace(get(contentsFields[16])); price != "" {
			if contentsGroup == "Adept powers" || contentsGroup == "Edges / Flaws" {
				value, err := strconv.ParseFloat(price, 64)
				if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < -1000000 || value > 1000000 {
					return fmt.Errorf("enter a valid point cost")
				}
				p.PointCost = &value
			} else {
				value, err := strconv.Atoi(price)
				if err != nil || value < 0 || value > 1000000000 {
					return fmt.Errorf("enter a nonnegative whole-number unit price")
				}
				p.UnitCost = &value
			}
		}

		p.Name = name
		p.Rating = rating
		if contentsGroup == "Cyberware" {
			p.Grade = strings.TrimSpace(get(contentsFields[12]))
			if _, _, valid := p.CyberGrade(); !valid {
				return fmt.Errorf("use Basic, Alpha, Beta, Delta, or Used Basic/Alpha/Beta")
			}
		} else {
			p.Source = get(contentsFields[12])
		}
		p.Modifiers = get(contentsFields[13])
		if contentsGroup == "Bioware" {
			p.BioIndex = nil
			if strings.TrimSpace(get(contentsFields[14])) != "" {
				p.BioIndex = &essence
			}
		} else {
			p.EssenceCost = essence
			p.EssenceUnresolved = false
		}
		p.Notes = get(contentsFields[15])
		p.Quantity = quantity
		if contentsIndex < 0 {
			*list = append(*list, p)
			if contentsGroup == "Cyberware" && app.char.LegacyEssence != nil {
				value := *app.char.LegacyEssence - p.ChargedEssence()
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
		for _, s := range editorGroups() {
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
		contentCtl(hwnd, "BUTTON", "Raise skill (karma)", WS_TABSTOP, 265, 475, 210, 30, 104)
		contentCtl(hwnd, "BUTTON", "Learn / raise specialization", WS_TABSTOP, 490, 475, 230, 30, 105)
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
				for i, g := range editorGroups() {
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
		case 104, 105:
			if contentsGroup == "Attributes" && id == 104 && contentsIndex >= 0 && contentsIndex < len(attrs) {
				if err := app.char.AdvanceAttribute(attrs[contentsIndex]); err != nil {
					msgbox(err.Error(), MB_ICONWARNING)
				} else {
					app.populate()
					app.dirty = true
					contentsRefresh()
				}
				return 0
			}
			if contentsGroup == "Skills" && (!contentsDirty || id == 105) {
				var err error
				if id == 105 {
					if contentsIndex >= 0 && contentsIndex < len(app.char.Skills) {
						selected := app.char.Skills[contentsIndex]
						if get(contentsFields[10]) != selected.Name || get(contentsFields[11]) != strconv.Itoa(selected.Rating) || get(contentsFields[12]) != selected.Class() || get(contentsFields[14]) != selected.LinkedAttribute {
							msgbox("Apply other skill edits before buying a specialization.", MB_ICONWARNING)
							return 0
						}
					}
					err = app.char.LearnOrAdvanceSpecialization(contentsIndex, get(contentsFields[13]))
				} else {
					err = app.char.AdvanceSkill(contentsIndex, false)
				}
				if err != nil {
					msgbox(err.Error(), MB_ICONWARNING)
				} else {
					app.dirty = true
					app.update()
					contentsRefresh()
				}
			} else {
				msgbox("Select a saved skill entry before advancing it.", MB_ICONWARNING)
			}
		case 100:
			if contentsGroup == "Character" {
				if contentsDirty {
					msgbox("Apply character edits before beginning career mode.", MB_ICONWARNING)
					return 0
				}
				if err := app.char.BeginCareer(); err != nil {
					msgbox(err.Error(), MB_ICONWARNING)
				} else {
					app.dirty = true
					app.update()
				}
				return 0
			}
			if contentsMayDiscard(hwnd) {
				contentsIndex = -1
				contentsLoad()
			}
		case 101:
			if err := contentsApply(); err != nil {
				procMessage.Call(hwnd, uintptr(unsafe.Pointer(ptr(err.Error()))), uintptr(unsafe.Pointer(ptr("Character entry"))), MB_ICONWARNING)
			}
		case 102:
			if contentsGroup != "Character" && contentsIndex >= 0 && contentsMayDiscard(hwnd) {
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
