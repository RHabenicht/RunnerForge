//go:build windows

package main

import (
	"context"
	"fmt"
	"nsrcgii/srcapp"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_BORDER           = 0x00800000
	WS_VSCROLL          = 0x00200000
	WS_TABSTOP          = 0x00010000
	ES_MULTILINE        = 4
	ES_AUTOVSCROLL      = 0x40
	ES_READONLY         = 0x800
	ES_AUTOHSCROLL      = 0x80
	BS_PUSHBUTTON       = 0
	CBS_DROPDOWNLIST    = 3
	LBS_NOTIFY          = 1
	WS_EX_CLIENTEDGE    = 0x200
	WM_CREATE           = 0x0001
	WM_DESTROY          = 0x0002
	WM_COMMAND          = 0x0111
	WM_CLOSE            = 0x0010
	WM_SETFONT          = 0x0030
	MB_OK               = 0
	MB_ICONWARNING      = 0x30
	MB_YESNOCANCEL      = 3
	IDYES               = 6
	IDCANCEL            = 2
	CB_ADDSTRING        = 0x0143
	CB_RESETCONTENT     = 0x014B
	CB_SETCURSEL        = 0x014E
	CB_GETCURSEL        = 0x0147
	CB_GETLBTEXT        = 0x0148
	LB_ADDSTRING        = 0x0180
	LB_RESETCONTENT     = 0x0184
	LB_GETCURSEL        = 0x0188
	LB_ERR              = 0xFFFFFFFF
	SW_SHOW             = 5
	CS_HREDRAW          = 2
	CS_VREDRAW          = 1
	CW_USEDEFAULT       = 0x80000000
	EN_CHANGE           = 0x0300
	CBN_SELCHANGE       = 1
	ID_NEW              = 100
	ID_SAVE             = 101
	ID_OPEN             = 102
	ID_IMPORT           = 103
	ID_HTML             = 104
	ID_PDF              = 105
	ID_CALC             = 106
	ID_SEARCH           = 107
	ID_ADD              = 108
	ID_SKILL            = 109
	ID_REMOVE           = 110
	ID_CUSTOM_RACE      = 111
	ID_EDIT_RACE        = 112
	ID_NAME             = 201
	ID_RACE             = 202
	ID_ATTR             = 220
	ID_PRIORITY         = 240
	ID_SEARCHTEXT       = 260
	ID_KIND             = 261
	ID_RESULTS          = 262
	ID_SUMMARY          = 263
	ID_SKILLNAME        = 264
	ID_SKILLRATING      = 265
	ID_METHOD           = 266
	ID_BP_BUDGET        = 267
	ID_BP_MAGIC         = 268
	ID_BP_RESOURCES     = 269
)

var (
	user32        = syscall.NewLazyDLL("user32.dll")
	kernel32      = syscall.NewLazyDLL("kernel32.dll")
	gdi32         = syscall.NewLazyDLL("gdi32.dll")
	comdlg32      = syscall.NewLazyDLL("comdlg32.dll")
	procReg       = user32.NewProc("RegisterClassExW")
	procCreate    = user32.NewProc("CreateWindowExW")
	procDef       = user32.NewProc("DefWindowProcW")
	procMessage   = user32.NewProc("MessageBoxW")
	procSend      = user32.NewProc("SendMessageW")
	procSet       = user32.NewProc("SetWindowTextW")
	procGet       = user32.NewProc("GetWindowTextW")
	procGetLen    = user32.NewProc("GetWindowTextLengthW")
	procGetMsg    = user32.NewProc("GetMessageW")
	procTranslate = user32.NewProc("TranslateMessage")
	procDispatch  = user32.NewProc("DispatchMessageW")
	procQuit      = user32.NewProc("PostQuitMessage")
	procDestroy   = user32.NewProc("DestroyWindow")
	procShow      = user32.NewProc("ShowWindow")
	procUpdate    = user32.NewProc("UpdateWindow")
	procModule    = kernel32.NewProc("GetModuleHandleW")
	procStock     = gdi32.NewProc("GetStockObject")
	title         = "RunnerForge — Shadowrun Third Edition"
	window        uintptr
	app           *application
)

type point struct{ X, Y int32 }
type msg struct {
	Hwnd    uintptr
	Message uint32
	Pad     uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
	Private uint32
}
type wndclassex struct {
	CbSize     uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uintptr
}
type ofn struct {
	Size            uint32
	Owner           uintptr
	Instance        uintptr
	Filter          *uint16
	CustomFilter    *uint16
	MaxCustomFilter uint32
	FilterIndex     uint32
	File            *uint16
	MaxFile         uint32
	FileTitle       *uint16
	MaxFileTitle    uint32
	InitialDir      *uint16
	Title           *uint16
	Flags           uint32
	FileOffset      uint16
	FileExtension   uint16
	DefExt          *uint16
	CustData        uintptr
	Hook            uintptr
	TemplateName    *uint16
	Reserved        uintptr
	Reserved2       uint32
	FlagsEx         uint32
}
type application struct {
	char      *srcapp.Character
	catalogue *srcapp.Catalogue
	controls  map[int]uintptr
	matches   []srcapp.Item
	dirty     bool
	savePath  string
	loading   bool
}

func ptr(s string) *uint16 { return syscall.StringToUTF16Ptr(s) }
func send(hwnd uintptr, m uint32, w, l uintptr) uintptr {
	r, _, _ := procSend.Call(hwnd, uintptr(m), w, l)
	return r
}
func set(hwnd uintptr, s string) { procSet.Call(hwnd, uintptr(unsafe.Pointer(ptr(s)))) }
func get(hwnd uintptr) string {
	if hwnd == 0 {
		return ""
	}
	n, _, _ := procGetLen.Call(hwnd)
	b := make([]uint16, n+2)
	procGet.Call(hwnd, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)))
	return syscall.UTF16ToString(b)
}
func msgbox(s string, flags uintptr) int {
	v, _, _ := procMessage.Call(window, uintptr(unsafe.Pointer(ptr(s))), uintptr(unsafe.Pointer(ptr(title))), flags)
	return int(v)
}
func ctl(class, label string, style uint32, x, y, w, h, id int) uintptr {
	v, _, _ := procCreate.Call(0, uintptr(unsafe.Pointer(ptr(class))), uintptr(unsafe.Pointer(ptr(label))), uintptr(style|WS_CHILD|WS_VISIBLE), uintptr(x), uintptr(y), uintptr(w), uintptr(h), window, uintptr(id), 0, 0)
	font, _, _ := procStock.Call(17)
	send(v, WM_SETFONT, font, 1)
	if id > 0 {
		app.controls[id] = v
	}
	return v
}
func text(label string, x, y, w, h int) { ctl("STATIC", label, 0, x, y, w, h, 0) }
func edit(value string, x, y, w, h, id int, multiline bool) uintptr {
	style := uint32(WS_BORDER | ES_AUTOHSCROLL | WS_TABSTOP)
	if multiline {
		style = WS_BORDER | WS_VSCROLL | ES_MULTILINE | ES_AUTOVSCROLL
	}
	return ctl("EDIT", value, style, x, y, w, h, id)
}
func button(label string, x, y, w, h, id int) { ctl("BUTTON", label, BS_PUSHBUTTON, x, y, w, h, id) }
func combo(values []string, x, y, w, h, id int, selected string) {
	c := ctl("COMBOBOX", "", CBS_DROPDOWNLIST|WS_VSCROLL|WS_TABSTOP, x, y, w, h, id)
	for i, v := range values {
		send(c, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr(v))))
		if v == selected {
			send(c, CB_SETCURSEL, uintptr(i), 0)
		}
	}
}
func comboVal(id int) string {
	h := app.controls[id]
	i := send(h, CB_GETCURSEL, 0, 0)
	if i == LB_ERR {
		return ""
	}
	s := make([]uint16, 256)
	send(h, CB_GETLBTEXT, i, uintptr(unsafe.Pointer(&s[0])))
	return syscall.UTF16ToString(s)
}

var creationMethods = []string{"Priority", "Build Points"}
var buildMagicOptions = []string{"Mundane", "Full Magician", "Aspected Magician", "Adept"}
var buildResourceOptions = []string{"500", "5000", "20000", "90000", "200000", "400000", "650000", "1000000"}

func selectCombo(id int, selected string) {
	h := app.controls[id]
	if h == 0 {
		return
	}
	values := map[int][]string{ID_METHOD: creationMethods, ID_BP_MAGIC: buildMagicOptions, ID_BP_RESOURCES: buildResourceOptions}
	if id >= ID_PRIORITY && id < ID_PRIORITY+5 {
		values[id] = []string{"A", "B", "C", "D", "E"}
	}
	for i, v := range values[id] {
		if v == selected {
			send(h, CB_SETCURSEL, uintptr(i), 0)
			return
		}
	}
}

func pickFile(save bool, filter string, def string) string {
	buf := make([]uint16, 1024)
	f := append(utf16.Encode([]rune(filter)), 0)
	var o ofn
	o.Size = uint32(unsafe.Sizeof(o))
	o.Owner = window
	if raceWindow != 0 {
		o.Owner = raceWindow
	}
	o.File = &buf[0]
	o.MaxFile = uint32(len(buf))
	o.Filter = &f[0]
	o.Flags = 0x00080000 | 0x00000004 | 0x00000800
	if save {
		o.Flags |= 0x00000002
	} else {
		o.Flags |= 0x00001000
	}
	if def != "" {
		o.DefExt = ptr(def)
	}
	target := "GetOpenFileNameW"
	if save {
		target = "GetSaveFileNameW"
	}
	ret, _, _ := comdlg32.NewProc(target).Call(uintptr(unsafe.Pointer(&o)))
	runtime.KeepAlive(f)
	if ret == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

var attrs = []string{"Body", "Quickness", "Strength", "Charisma", "Intelligence", "Willpower"}
var pri = []string{"Resources", "Attributes", "Race", "Skills", "Magic"}

func setupUI() {
	text("Character name", 20, 17, 150, 18)
	edit("", 20, 36, 240, 25, ID_NAME, false)
	text("Metatype", 280, 17, 130, 18)
	combo(nil, 280, 34, 145, 260, ID_RACE, "")
	button("New", 455, 30, 73, 32, ID_NEW)
	button("Open", 535, 30, 73, 32, ID_OPEN)
	button("Save", 615, 30, 73, 32, ID_SAVE)
	button("Import .sr3", 694, 30, 100, 32, ID_IMPORT)
	button("Export HTML", 800, 30, 108, 32, ID_HTML)
	button("PDF", 914, 30, 70, 32, ID_PDF)
	text("Priorities (A–E)", 20, 75, 135, 18)
	combo(creationMethods, 155, 70, 125, 100, ID_METHOD, "Priority")
	text("BP budget", 300, 75, 75, 18)
	edit("120", 378, 70, 55, 24, ID_BP_BUDGET, false)
	text("Magic", 454, 75, 55, 18)
	combo(buildMagicOptions, 510, 70, 170, 140, ID_BP_MAGIC, "Mundane")
	text("Resources (¥)", 700, 75, 108, 18)
	combo(buildResourceOptions, 812, 70, 150, 190, ID_BP_RESOURCES, "5000")
	for i, k := range pri {
		xx := 20 + i*193
		text(k, xx, 98, 155, 18)
		sel := app.char.Priority[k]
		combo([]string{"A", "B", "C", "D", "E"}, xx, 119, 178, 145, ID_PRIORITY+i, sel)
	}
	text("Attributes — unaugmented base ratings", 20, 162, 430, 20)
	for i, k := range attrs {
		xx := 20 + i*161
		text(k, xx, 190, 150, 20)
		edit(strconv.Itoa(app.char.Attributes[k]), xx, 211, 145, 25, ID_ATTR+i, false)
	}
	button("Recalculate / Validate", 20, 247, 220, 32, ID_CALC)
	text("Add skill", 265, 250, 85, 19)
	edit("", 345, 246, 220, 27, ID_SKILLNAME, false)
	text("Rating", 578, 250, 65, 19)
	edit("1", 636, 246, 45, 27, ID_SKILLRATING, false)
	button("Add skill", 692, 246, 100, 31, ID_SKILL)
	button("Edit race…", 800, 246, 85, 31, ID_EDIT_RACE)
	button("Load race…", 895, 246, 85, 31, ID_CUSTOM_RACE)
	text("Reference catalogue — approved sources only", 20, 300, 410, 20)
	edit("", 20, 325, 390, 25, ID_SEARCHTEXT, false)
	combo([]string{"all", "gear", "cyberware", "bioware", "adept_powers", "skills", "spells", "vehicles", "cyberdecks", "edges_flaws", "totems", "magic_gear"}, 420, 325, 185, 180, ID_KIND, "all")
	button("Search", 616, 323, 100, 30, ID_SEARCH)
	button("Add selected", 725, 323, 125, 30, ID_ADD)
	ctl("LISTBOX", "", WS_BORDER|LBS_NOTIFY|WS_VSCROLL, 20, 358, 610, 365, ID_RESULTS)
	text("Character summary and validation", 651, 300, 340, 22)
	edit("", 651, 358, 333, 365, ID_SUMMARY, true)
	button("Remove last gear/cyber", 862, 323, 122, 30, ID_REMOVE)
	app.search()
	app.populate()
}
func (a *application) populate() {
	a.loading = true
	defer func() { a.loading = false; a.dirty = false }()
	set(a.controls[ID_NAME], a.char.Name)
	a.populateRaces()
	method := "Priority"
	allocation := srcapp.BuildPointAllocation{Budget: 120, Resources: 5000, Magic: "Mundane"}
	if a.char.BuildPoints != nil {
		method = "Build Points"
		allocation = *a.char.BuildPoints
	}
	selectCombo(ID_METHOD, method)
	set(a.controls[ID_BP_BUDGET], strconv.Itoa(allocation.Budget))
	selectCombo(ID_BP_MAGIC, allocation.Magic)
	selectCombo(ID_BP_RESOURCES, strconv.Itoa(allocation.Resources))
	for i, k := range pri {
		selectCombo(ID_PRIORITY+i, a.char.Priority[k])
	}
	for i, k := range attrs {
		set(a.controls[ID_ATTR+i], strconv.Itoa(a.char.Attributes[k]))
	}
	set(a.controls[ID_SUMMARY], a.char.Summary())
}

func (a *application) populateRaces() {
	control := a.controls[ID_RACE]
	send(control, CB_RESETCONTENT, 0, 0)
	names := srcapp.RaceNames()
	if a.char.CustomRace != nil {
		names = append(names, a.char.CustomRace.Name)
	}
	for _, name := range names {
		index := send(control, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr(name))))
		if name == a.char.Race {
			send(control, CB_SETCURSEL, index, 0)
		}
	}
}

func (a *application) sync() {
	a.char.Name = get(a.controls[ID_NAME])
	if race := comboVal(ID_RACE); race != "" {
		a.char.Race = race
	}
	if comboVal(ID_METHOD) == "Build Points" {
		budget, _ := strconv.Atoi(strings.TrimSpace(get(a.controls[ID_BP_BUDGET])))
		resources, _ := strconv.Atoi(comboVal(ID_BP_RESOURCES))
		a.char.BuildPoints = &srcapp.BuildPointAllocation{
			Budget: budget, Resources: resources, Magic: comboVal(ID_BP_MAGIC),
		}
	} else {
		a.char.BuildPoints = nil
	}
	for i, k := range pri {
		a.char.Priority[k] = comboVal(ID_PRIORITY + i)
	}
	for i, k := range attrs {
		n, e := strconv.Atoi(strings.TrimSpace(get(a.controls[ID_ATTR+i])))
		if e == nil {
			a.char.Attributes[k] = n
		}
	}
}
func (a *application) update() { a.sync(); set(a.controls[ID_SUMMARY], a.char.Summary()) }
func (a *application) search() {
	q := get(a.controls[ID_SEARCHTEXT])
	kind := comboVal(ID_KIND)
	a.matches = a.catalogue.Search(q, kind, 400)
	h := a.controls[ID_RESULTS]
	send(h, LB_RESETCONTENT, 0, 0)
	for _, v := range a.matches {
		s := fmt.Sprintf("%s  [%s]  (%s)", v.Name, v.Kind, strings.Join(v.Sources, ","))
		send(h, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr(s))))
	}
}
func (a *application) beforeReplace() bool {
	if !a.dirty {
		return true
	}
	switch msgbox("Unsaved changes will be lost. Continue?", MB_YESNOCANCEL|MB_ICONWARNING) {
	case IDYES:
		return true
	default:
		return false
	}
}
func (a *application) perform(id int) {
	switch id {
	case ID_NEW:
		if !a.beforeReplace() {
			return
		}
		a.char = srcapp.NewCharacter()
		a.savePath = ""
		a.populate()
	case ID_EDIT_RACE:
		openRaceEditor()
	case ID_CUSTOM_RACE:
		path := pickFile(false, "Custom race (*.json)\x00*.json\x00\x00", "")
		if path == "" {
			return
		}
		race, err := srcapp.LoadRaceDefinition(path)
		if err != nil {
			msgbox("Could not load custom race: "+err.Error(), MB_OK|MB_ICONWARNING)
			return
		}
		a.sync()
		a.char.CustomRace = race
		a.char.Race = race.Name
		a.populateRaces()
		selectCombo(ID_METHOD, "Build Points")
		a.dirty = true
		a.update()
	case ID_CALC:
		a.update()
	case ID_SEARCH:
		a.search()
	case ID_SKILL:
		a.sync()
		n := strings.TrimSpace(get(a.controls[ID_SKILLNAME]))
		rating, e := strconv.Atoi(get(a.controls[ID_SKILLRATING]))
		if n == "" || e != nil || rating < 1 || rating > 6 {
			msgbox("Enter a skill name and rating from 1 to 6.", MB_OK|MB_ICONWARNING)
			return
		}
		a.char.Skills = append(a.char.Skills, srcapp.Rated{Name: n, Rating: rating, Category: "Active"})
		a.dirty = true
		a.update()
	case ID_ADD:
		a.sync()
		index := send(a.controls[ID_RESULTS], LB_GETCURSEL, 0, 0)
		if index == LB_ERR || int(index) >= len(a.matches) {
			return
		}
		a.char.AddItem(a.matches[index])
		a.dirty = true
		a.update()
	case ID_REMOVE:
		a.sync()
		if len(a.char.Cyberware) > 0 {
			a.char.Cyberware = a.char.Cyberware[:len(a.char.Cyberware)-1]
		} else if len(a.char.Gear) > 0 {
			a.char.Gear = a.char.Gear[:len(a.char.Gear)-1]
		}
		a.dirty = true
		a.update()
	case ID_SAVE:
		a.sync()
		path := a.savePath
		if path == "" {
			path = pickFile(true, "RunnerForge Character (*.nsrcg2.json)\x00*.nsrcg2.json\x00JSON Files (*.json)\x00*.json\x00\x00", "json")
		}
		if path == "" {
			return
		}
		if e := srcapp.SaveJSON(path, a.char); e != nil {
			msgbox("Save failed: "+e.Error(), MB_OK|MB_ICONWARNING)
		} else {
			a.savePath = path
			a.dirty = false
			msgbox("Character saved.", MB_OK)
		}
	case ID_OPEN:
		if !a.beforeReplace() {
			return
		}
		path := pickFile(false, "RunnerForge Character (*.json)\x00*.json\x00All files (*.*)\x00*.*\x00\x00", "")
		if path == "" {
			return
		}
		c, e := srcapp.LoadJSON(path)
		if e != nil {
			msgbox("Open failed: "+e.Error(), MB_OK|MB_ICONWARNING)
		} else {
			a.char = c
			a.savePath = path
			a.populate()
		}
	case ID_IMPORT:
		if !a.beforeReplace() {
			return
		}
		path := pickFile(false, "Legacy NSRCG Character (*.sr3)\x00*.sr3\x00All files (*.*)\x00*.*\x00\x00", "")
		if path == "" {
			return
		}
		c, e := srcapp.ReadLegacy(path)
		if e != nil {
			msgbox("Import failed: "+e.Error(), MB_OK|MB_ICONWARNING)
		} else {
			a.char = c
			a.savePath = ""
			a.populate()
			msgbox(fmt.Sprintf("Imported with %d review notes. Original .sr3 content preserved in the save file.", len(c.LegacyWarnings)), MB_OK)
		}
	case ID_HTML:
		a.sync()
		path := pickFile(true, "HTML Character Sheet (*.html)\x00*.html\x00\x00", "html")
		if path == "" {
			return
		}
		if e := a.char.SaveHTML(path); e != nil {
			msgbox("HTML export failed: "+e.Error(), MB_OK|MB_ICONWARNING)
		} else {
			msgbox("HTML character sheet exported.", MB_OK)
		}
	case ID_PDF:
		a.sync()
		path := pickFile(true, "PDF Character Sheet (*.pdf)\x00*.pdf\x00\x00", "pdf")
		if path == "" {
			return
		}
		if e := pdfFromHTML(a.char, path); e != nil {
			msgbox("PDF export failed: "+e.Error()+"\r\nTry HTML export instead.", MB_OK|MB_ICONWARNING)
		} else {
			msgbox("PDF exported using local Microsoft Edge rendering.", MB_OK)
		}
	}
}
func pdfFromHTML(c *srcapp.Character, target string) error {
	f, e := os.CreateTemp("", "nsrcgii-*.html")
	if e != nil {
		return e
	}
	path := f.Name()
	f.Close()
	defer os.Remove(path)
	if e = c.SaveHTML(path); e != nil {
		return e
	}
	candidates := []string{`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`, `C:\Program Files\Microsoft\Edge\Application\msedge.exe`}
	if p, e := exec.LookPath("msedge.exe"); e == nil {
		candidates = append([]string{p}, candidates...)
	}
	var exe string
	for _, p := range candidates {
		if _, e := os.Stat(p); e == nil {
			exe = p
			break
		}
	}
	if exe == "" {
		return fmt.Errorf("Microsoft Edge not found; HTML export is available")
	}
	abs, e := filepath.Abs(path)
	if e != nil {
		return e
	}
	abs = strings.ReplaceAll(abs, "\\", "/")
	url := "file:///" + strings.ReplaceAll(abs, " ", "%20")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	args := []string{"--headless", "--disable-gpu", "--no-first-run", "--no-pdf-header-footer", "--print-to-pdf=" + target, url}
	cmd := exec.CommandContext(ctx, exe, args...)
	if b, e := cmd.CombinedOutput(); e != nil {
		return fmt.Errorf("Edge printing: %v (%s)", e, string(b))
	}
	if _, e := os.Stat(target); e != nil {
		return fmt.Errorf("Edge did not create PDF")
	}
	return nil
}
func wndProc(hwnd uintptr, ev uint32, wp, lp uintptr) uintptr {
	switch ev {
	case WM_CREATE:
		window = hwnd
		cat, e := srcapp.LoadCatalogue()
		if e != nil {
			msgbox("Could not load bundled catalog: "+e.Error(), MB_OK|MB_ICONWARNING)
			return 0
		}
		app = &application{char: srcapp.NewCharacter(), catalogue: cat, controls: map[int]uintptr{}}
		setupUI()
		return 0
	case WM_COMMAND:
		if app == nil {
			return 0
		}
		id := int(wp & 0xFFFF)
		notif := int((wp >> 16) & 0xFFFF)
		if id >= 100 && id <= ID_EDIT_RACE {
			app.perform(id)
		} else if !app.loading && (notif == EN_CHANGE || notif == CBN_SELCHANGE) && id != ID_SUMMARY && id != ID_SEARCHTEXT && id != ID_RESULTS {
			app.dirty = true
		}
		return 0
	case WM_CLOSE:
		if app != nil && !app.beforeReplace() {
			return 0
		}
		procDestroy.Call(hwnd)
		return 0
	case WM_DESTROY:
		procQuit.Call(0)
		return 0
	}
	r, _, _ := procDef.Call(hwnd, uintptr(ev), wp, lp)
	return r
}
func main() {
	runtime.LockOSThread()
	inst, _, _ := procModule.Call(0)
	class := ptr("NSRCGIIWindow")
	wc := wndclassex{Style: CS_HREDRAW | CS_VREDRAW, WndProc: syscall.NewCallback(wndProc), Instance: inst, Background: 6 + 1, ClassName: class}
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	v, _, e := procReg.Call(uintptr(unsafe.Pointer(&wc)))
	if v == 0 {
		fmt.Println("Register window class failed", e)
		return
	}
	hwnd, _, e := procCreate.Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(ptr(title))), WS_OVERLAPPEDWINDOW|WS_VISIBLE, 65, 65, 1025, 800, 0, 0, inst, 0)
	if hwnd == 0 {
		fmt.Println("Window creation failed", e)
		return
	}
	window = hwnd
	procShow.Call(hwnd, SW_SHOW)
	procUpdate.Call(hwnd)
	var m msg
	for {
		r, _, _ := procGetMsg.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		dialog := window
		if raceWindow != 0 {
			dialog = raceWindow
		}
		handled, _, _ := user32.NewProc("IsDialogMessageW").Call(dialog, uintptr(unsafe.Pointer(&m)))
		if handled != 0 {
			continue
		}
		procTranslate.Call(uintptr(unsafe.Pointer(&m)))
		procDispatch.Call(uintptr(unsafe.Pointer(&m)))
	}
}
