package srcapp

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

//go:embed catalog.json
var embeddedCatalog embed.FS

type Book struct {
	Name    string `json:"name"`
	Code    string `json:"code"`
	Allowed bool   `json:"allowed"`
}
type Item struct {
	ID       int                    `json:"id"`
	Name     string                 `json:"name"`
	Kind     string                 `json:"kind"`
	Group    string                 `json:"group"`
	Category string                 `json:"category"`
	Fields   map[string]interface{} `json:"fields"`
	Sources  []string               `json:"sources"`
	Origin   string                 `json:"origin"`
}
type Catalogue struct {
	Books   []Book `json:"books"`
	Entries []Item `json:"entries"`
	Schema  string `json:"schema"`
	Edition string `json:"edition"`
}

func LoadCatalogue() (*Catalogue, error) {
	b, e := embeddedCatalog.ReadFile("catalog.json")
	if e != nil {
		return nil, e
	}
	var x Catalogue
	if e = json.Unmarshal(b, &x); e != nil {
		return nil, e
	}
	return &x, nil
}
func (c *Catalogue) Search(q, kind string, limit int) []Item {
	q = strings.ToLower(strings.TrimSpace(q))
	xs := make([]Item, 0)
	if limit <= 0 {
		limit = 100
	}
	for _, v := range c.Entries {
		if kind != "" && kind != "all" && v.Kind != kind {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(v.Name+" "+v.Category+" "+strings.Join(v.Sources, " ")), q) {
			continue
		}
		xs = append(xs, v)
		if len(xs) >= limit {
			break
		}
	}
	return xs
}

type Rated struct {
	Name     string `json:"name"`
	Rating   int    `json:"rating"`
	Category string `json:"category,omitempty"`
}
type Possession struct {
	Name        string  `json:"name"`
	CatalogID   int     `json:"catalog_id,omitempty"`
	Kind        string  `json:"kind"`
	Notes       string  `json:"notes,omitempty"`
	EssenceCost float64 `json:"essence_cost,omitempty"`
	Modifiers   string  `json:"modifiers,omitempty"`
	Source      string  `json:"source,omitempty"`
	Raw         string  `json:"raw,omitempty"`
}
type Character struct {
	Format         string            `json:"format"`
	Edition        int               `json:"edition"`
	Name           string            `json:"name"`
	RealName       string            `json:"real_name,omitempty"`
	Race           string            `json:"race"`
	MagicType      string            `json:"magic_type,omitempty"`
	Priority       map[string]string `json:"priority"`
	Attributes     map[string]int    `json:"attributes"`
	Skills         []Rated           `json:"skills"`
	Gear           []Possession      `json:"gear"`
	Cyberware      []Possession      `json:"cyberware"`
	Bioware        []Possession      `json:"bioware,omitempty"`
	Spells         []Possession      `json:"spells,omitempty"`
	AdeptPowers    []Possession      `json:"adept_powers,omitempty"`
	Vehicles       []Possession      `json:"vehicles,omitempty"`
	Decks          []Possession      `json:"cyberdecks,omitempty"`
	EdgesFlaws     []Possession      `json:"edges_flaws,omitempty"`
	Contacts       []Possession      `json:"contacts,omitempty"`
	Totems         []Possession      `json:"totems,omitempty"`
	MagicGear      []Possession      `json:"magic_gear,omitempty"`
	Karma          int               `json:"karma"`
	KarmaPool      int               `json:"karma_pool"`
	Notes          string            `json:"notes"`
	LegacyText     string            `json:"legacy_text,omitempty"`
	LegacyBytes    []byte            `json:"legacy_bytes,omitempty"`
	LegacyWarnings []string          `json:"legacy_warnings,omitempty"`
	LegacyEssence  *float64          `json:"legacy_essence,omitempty"`
	LegacyMagic    *int              `json:"legacy_magic,omitempty"`
}

var attributeKeys = []string{"Body", "Quickness", "Strength", "Charisma", "Intelligence", "Willpower"}
var priorityNames = []string{"Resources", "Attributes", "Race", "Skills", "Magic"}
var raceMods = map[string]map[string]int{"Human": {}, "Dwarf": {"Body": 1, "Strength": 2, "Willpower": 1}, "Elf": {"Quickness": 1, "Charisma": 2}, "Ork": {"Body": 3, "Strength": 2, "Charisma": -1, "Intelligence": -1}, "Troll": {"Body": 5, "Quickness": -1, "Strength": 4, "Intelligence": -2, "Charisma": -2}}

func NewCharacter() *Character {
	v := &Character{Format: "nsrcgii.v2", Edition: 3, Name: "New Character", Race: "Human", Priority: map[string]string{"Resources": "A", "Attributes": "B", "Race": "E", "Skills": "C", "Magic": "D"}, Attributes: map[string]int{}, KarmaPool: 1}
	for _, k := range attributeKeys {
		v.Attributes[k] = 1
	}
	return v
}
func (ch *Character) Repair() {
	if ch.Priority == nil {
		ch.Priority = NewCharacter().Priority
	}
	if ch.Attributes == nil {
		ch.Attributes = NewCharacter().Attributes
	}
	if ch.Race == "" {
		ch.Race = "Human"
	}
	if ch.Edition == 0 {
		ch.Edition = 3
	}
	if ch.Format == "" {
		ch.Format = "nsrcgii.v2"
	}
}

var resourceLevels = map[string]int{"A": 1000000, "B": 400000, "C": 90000, "D": 20000, "E": 5000}
var attrLevels = map[string]int{"A": 30, "B": 27, "C": 24, "D": 21, "E": 18}
var skillLevels = map[string]int{"A": 50, "B": 40, "C": 34, "D": 30, "E": 27}

func (ch *Character) ResourceBudget() int  { return resourceLevels[ch.Priority["Resources"]] }
func (ch *Character) AttributeBudget() int { return attrLevels[ch.Priority["Attributes"]] }
func (ch *Character) SkillBudget() int     { return skillLevels[ch.Priority["Skills"]] }
func (ch *Character) AttributeSpent() int {
	s := 0
	for _, k := range attributeKeys {
		s += ch.Attributes[k]
	}
	return s
}
func (ch *Character) SkillSpent() int {
	s := 0
	for _, r := range ch.Skills {
		s += r.Rating
	}
	return s
} // allocation cost is approximate; specialization/knowledge rules separate
func (ch *Character) EffectiveAttribute(name string) int {
	return ch.Attributes[name] + raceMods[ch.Race][name]
}
func (ch *Character) Awakened() bool {
	return ch.Priority["Magic"] == "A" || ch.Priority["Magic"] == "B"
}
func (ch *Character) Essence() float64 {
	if ch.LegacyEssence != nil {
		return *ch.LegacyEssence
	}
	s := 6.0
	for _, cy := range ch.Cyberware {
		s -= cy.EssenceCost
	}
	if s < 0 {
		return 0
	}
	return s
}
func (ch *Character) Magic() int {
	if !ch.Awakened() {
		return 0
	}
	if ch.LegacyMagic != nil {
		return *ch.LegacyMagic
	}
	return int(ch.Essence())
}
func (ch *Character) Reaction() int {
	return (ch.EffectiveAttribute("Quickness")+ch.EffectiveAttribute("Intelligence"))/2 + ch.CyberReactionMod()
}

var reMod = regexp.MustCompile(`(?i)([+\-]\d+)\s*RCT`)
var reInit = regexp.MustCompile(`(?i)([+\-]\d+)\s*INI`)

func totalMods(cys []Possession, re *regexp.Regexp) int {
	s := 0
	for _, it := range cys {
		for _, v := range re.FindAllStringSubmatch(it.Modifiers, -1) {
			n, _ := strconv.Atoi(v[1])
			s += n
		}
	}
	return s
}
func (ch *Character) CyberReactionMod() int { return totalMods(ch.Cyberware, reMod) }
func (ch *Character) InitiativeDice() int   { return 1 + totalMods(ch.Cyberware, reInit) }
func (ch *Character) CombatPool() int {
	return (ch.EffectiveAttribute("Quickness") + ch.EffectiveAttribute("Intelligence") + ch.EffectiveAttribute("Willpower")) / 2
}
func (ch *Character) SpellPool() int {
	if !ch.Awakened() {
		return 0
	}
	return (ch.EffectiveAttribute("Intelligence") + ch.EffectiveAttribute("Willpower") + ch.Magic()) / 3
}
func (ch *Character) Validate() []string {
	issues := []string{}
	if ch.Edition != 3 {
		issues = append(issues, "Only Shadowrun Third Edition is supported")
	}
	found := map[string]string{}
	for _, k := range priorityNames {
		v := strings.ToUpper(strings.TrimSpace(ch.Priority[k]))
		if _, ok := resourceLevels[v]; !ok {
			issues = append(issues, k+" priority must be A–E")
		} else if old, ok := found[v]; ok {
			issues = append(issues, fmt.Sprintf("Priority %s assigned to %s and %s", v, old, k))
		} else {
			found[v] = k
		}
	}
	if _, ok := raceMods[ch.Race]; !ok {
		issues = append(issues, "Unknown race")
	}
	pr := ch.Priority["Race"]
	if pr == "E" && ch.Race != "Human" {
		issues = append(issues, "Race priority E only permits Human")
	}
	if pr == "D" && ch.Race != "Human" && ch.Race != "Ork" && ch.Race != "Dwarf" {
		issues = append(issues, "Race priority D permits Human, Ork and Dwarf")
	}
	if ch.AttributeSpent() > ch.AttributeBudget() {
		issues = append(issues, fmt.Sprintf("Attribute points exceed priority budget: %d / %d", ch.AttributeSpent(), ch.AttributeBudget()))
	}
	for _, k := range attributeKeys {
		if ch.Attributes[k] < 1 || ch.Attributes[k] > 6 {
			issues = append(issues, k+" base rating outside 1–6 (before racial adjustment)")
		}
		if ch.EffectiveAttribute(k) < 1 {
			issues = append(issues, k+" racial-adjusted rating must be at least 1")
		}
	}
	if ch.SkillSpent() > ch.SkillBudget() {
		issues = append(issues, fmt.Sprintf("Skill points exceed priority budget (simple ratings): %d / %d", ch.SkillSpent(), ch.SkillBudget()))
	}
	for _, v := range ch.Skills {
		if v.Rating < 1 || v.Rating > 6 {
			issues = append(issues, "Skill rating outside 1–6: "+v.Name)
		}
	}
	if ch.Essence() < 0.01 && len(ch.Cyberware) > 0 {
		issues = append(issues, "Essence depleted or below 0.01")
	}
	return issues
}
func SaveJSON(path string, c *Character) error {
	c.Repair()
	b, e := json.MarshalIndent(c, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, b, 0600)
}
func LoadJSON(path string) (*Character, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	var c Character
	dec := json.NewDecoder(bytes.NewReader(b))
	if e = dec.Decode(&c); e != nil {
		return nil, e
	}
	if c.Edition != 3 {
		return nil, errors.New("file targets unsupported edition (SR3 only)")
	}
	c.Repair()
	return &c, nil
}
func (ch *Character) AddItem(it Item) {
	p := Possession{Name: it.Name, CatalogID: it.ID, Kind: it.Kind, Source: strings.Join(it.Sources, ", ")}
	switch it.Kind {
	case "cyberware":
		p.EssenceCost, _ = strconv.ParseFloat(fmt.Sprint(it.Fields["EssCost"]), 64)
		p.Modifiers = fmt.Sprint(it.Fields["Mods"])
		ch.Cyberware = append(ch.Cyberware, p)
		if ch.LegacyEssence != nil {
			v := *ch.LegacyEssence - p.EssenceCost
			ch.LegacyEssence = &v
		}
	case "skills":
		ch.Skills = append(ch.Skills, Rated{Name: it.Name, Category: it.Group, Rating: 1})
	case "bioware":
		ch.Bioware = append(ch.Bioware, p)
	case "spells":
		ch.Spells = append(ch.Spells, p)
	case "adept_powers":
		ch.AdeptPowers = append(ch.AdeptPowers, p)
	case "vehicles":
		ch.Vehicles = append(ch.Vehicles, p)
	case "cyberdecks":
		ch.Decks = append(ch.Decks, p)
	case "edges_flaws":
		ch.EdgesFlaws = append(ch.EdgesFlaws, p)
	case "contacts":
		ch.Contacts = append(ch.Contacts, p)
	case "totems":
		ch.Totems = append(ch.Totems, p)
	case "magic_gear":
		ch.MagicGear = append(ch.MagicGear, p)
	default:
		ch.Gear = append(ch.Gear, p)
	}
}

var htmlTemplate = template.Must(template.New("sheet").Parse(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>{{.Name}} — NSRCG II</title><style>body{font:15px Arial,sans-serif;max-width:900px;margin:auto;padding:1.5rem;color:#181b22}h1{border-bottom:2px solid #222}h2{margin-top:1.5rem}table{border-collapse:collapse;width:100%}td,th{border-bottom:1px solid #ddd;text-align:left;padding:.35rem}small{color:#555}.col{display:grid;grid-template-columns:1fr 1fr;gap:1rem}@media print{body{margin:.3in;padding:0}h2{break-after:avoid}}</style></head><body><h1>{{.Name}}</h1><small>NSRCG II — Shadowrun Third Edition · {{.Race}} · Local HTML Export</small><div class="col"><section><h2>Attributes</h2><table>{{range .AttributesList}}<tr><td>{{.Key}}</td><td>{{.Value}}</td></tr>{{end}}</table></section><section><h2>Derived</h2><table><tr><td>Reaction</td><td>{{.Reaction}}</td></tr><tr><td>Initiative</td><td>{{.Reaction}} + {{.InitiativeDice}}D6</td></tr><tr><td>Essence</td><td>{{.Essence}}</td></tr><tr><td>Magic</td><td>{{.Magic}}</td></tr><tr><td>Combat Pool</td><td>{{.CombatPool}}</td></tr><tr><td>Spell Pool</td><td>{{.SpellPool}}</td></tr></table></section></div><h2>Priorities</h2><table>{{range .Priorities}}<tr><td>{{.Key}}</td><td>{{.Value}}</td></tr>{{end}}</table><h2>Skills</h2><table>{{range .Skills}}<tr><td>{{.Name}}</td><td>{{.Rating}}</td></tr>{{end}}</table><h2>Cyberware</h2><table>{{range .Cyberware}}<tr><td>{{.Name}}</td><td>{{.EssenceCost}}</td><td>{{.Source}}</td></tr>{{end}}</table><h2>Gear</h2><table>{{range .Gear}}<tr><td>{{.Name}}</td><td>{{.Source}}</td></tr>{{end}}</table><h2>Bioware</h2><table>{{range .Bioware}}<tr><td>{{.Name}}</td><td>{{.Source}}</td></tr>{{end}}</table><h2>Spells</h2><table>{{range .Spells}}<tr><td>{{.Name}}</td><td>{{.Source}}</td></tr>{{end}}</table><h2>Adept Powers</h2><table>{{range .AdeptPowers}}<tr><td>{{.Name}}</td><td>{{.Source}}</td></tr>{{end}}</table><h2>Vehicles</h2><table>{{range .Vehicles}}<tr><td>{{.Name}}</td></tr>{{end}}</table><h2>Cyberdecks</h2><table>{{range .Decks}}<tr><td>{{.Name}}</td></tr>{{end}}</table><h2>Edges / Flaws</h2><table>{{range .EdgesFlaws}}<tr><td>{{.Name}}</td></tr>{{end}}</table><h2>Totems / Traditions</h2><table>{{range .Totems}}<tr><td>{{.Name}}</td></tr>{{end}}</table><h2>Contacts</h2><table>{{range .Contacts}}<tr><td>{{.Name}}</td></tr>{{end}}</table><h2>Notes</h2><p style="white-space:pre-wrap">{{.Notes}}</p><small>Generated character record. Verify all calculations against the official SR3 rules, especially imported legacy cyberware and advanced options.</small></body></html>`))

type kv struct {
	Key   string
	Value interface{}
}

func (ch *Character) ExportHTML(w io.Writer) error {
	attrs := []kv{}
	for _, k := range attributeKeys {
		attrs = append(attrs, kv{k, ch.EffectiveAttribute(k)})
	}
	prs := []kv{}
	for _, k := range priorityNames {
		prs = append(prs, kv{k, ch.Priority[k]})
	}
	v := struct {
		*Character
		AttributesList []kv
		Priorities     []kv
		Reaction       int
		InitiativeDice int
		Essence        string
		Magic          int
		CombatPool     int
		SpellPool      int
	}{ch, attrs, prs, ch.Reaction(), ch.InitiativeDice(), fmt.Sprintf("%.2f", ch.Essence()), ch.Magic(), ch.CombatPool(), ch.SpellPool()}
	return htmlTemplate.Execute(w, v)
}
func (ch *Character) SaveHTML(path string) error {
	var b bytes.Buffer
	if e := ch.ExportHTML(&b); e != nil {
		return e
	}
	return os.WriteFile(path, b.Bytes(), 0600)
}
func (ch *Character) Summary() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Name: %s\r\nSR3 • %s\r\n", ch.Name, ch.Race)
	for _, k := range attributeKeys {
		fmt.Fprintf(&sb, "%s: %d (base %d)\r\n", k, ch.EffectiveAttribute(k), ch.Attributes[k])
	}
	fmt.Fprintf(&sb, "\r\nReaction %d • Initiative %d + %dD6\r\nEssence %.2f • Magic %d\r\nCombat Pool %d • Spell Pool %d\r\n", ch.Reaction(), ch.Reaction(), ch.InitiativeDice(), ch.Essence(), ch.Magic(), ch.CombatPool(), ch.SpellPool())
	fmt.Fprintf(&sb, "\r\nAttributes %d / %d points\r\nSkills %d / %d points (simple)\r\nResources: ¥%d\r\n", ch.AttributeSpent(), ch.AttributeBudget(), ch.SkillSpent(), ch.SkillBudget(), ch.ResourceBudget())
	fmt.Fprintf(&sb, "Skills: %d • Gear: %d • Cyberware: %d\r\n", len(ch.Skills), len(ch.Gear), len(ch.Cyberware))
	fmt.Fprintf(&sb, "Bioware %d • Spells %d • Adept powers %d • Vehicles %d • Decks %d\r\n", len(ch.Bioware), len(ch.Spells), len(ch.AdeptPowers), len(ch.Vehicles), len(ch.Decks))
	issues := ch.Validate()
	if len(issues) == 0 {
		sb.WriteString("\r\nBasic priority/attribute checks passed. Advanced SR3 validation incomplete.")
	} else {
		sb.WriteString("\r\nReview:\r\n - " + strings.Join(issues, "\r\n - "))
	}
	return sb.String()
}
func (ch *Character) SortedSkills() []Rated {
	out := append([]Rated(nil), ch.Skills...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
