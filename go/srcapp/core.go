package srcapp

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"math"
	"os"
	"path/filepath"
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
	AdditionalSpecializations map[string]int `json:"additional_specializations,omitempty"`
	SpecializationValue       *int           `json:"specialization_rating,omitempty"`
	Name                      string         `json:"name"`
	Rating                    int            `json:"rating"`
	Category                  string         `json:"category,omitempty"`
	Specialization            string         `json:"specialization,omitempty"`
	LinkedAttribute           string         `json:"linked_attribute,omitempty"`
}
type Possession struct {
	Grade             string   `json:"grade,omitempty"`
	BioIndex          *float64 `json:"bio_index,omitempty"`
	EssenceUnresolved bool     `json:"essence_unresolved,omitempty"`
	PointCost         *float64 `json:"point_cost,omitempty"`
	Name              string   `json:"name"`
	UnitCost          *int     `json:"unit_cost,omitempty"`
	Quantity          int      `json:"quantity,omitempty"`
	Rating            int      `json:"rating,omitempty"`
	CatalogID         int      `json:"catalog_id,omitempty"`
	Kind              string   `json:"kind"`
	Notes             string   `json:"notes,omitempty"`
	EssenceCost       float64  `json:"essence_cost,omitempty"`
	Modifiers         string   `json:"modifiers,omitempty"`
	Source            string   `json:"source,omitempty"`
	Raw               string   `json:"raw,omitempty"`
}
type Character struct {
	Career               bool                  `json:"career,omitempty"`
	CreationSnapshot     json.RawMessage       `json:"creation_snapshot,omitempty"`
	Advancements         []Advancement         `json:"advancements,omitempty"`
	SpellPointsPurchased int                   `json:"spell_points_purchased,omitempty"`
	SpellPointsSold      int                   `json:"spell_points_sold,omitempty"`
	Tradition            string                `json:"tradition,omitempty"`
	Format               string                `json:"format"`
	Edition              int                   `json:"edition"`
	Name                 string                `json:"name"`
	RealName             string                `json:"real_name,omitempty"`
	Race                 string                `json:"race"`
	CustomRace           *RaceDefinition       `json:"custom_race,omitempty"`
	MagicType            string                `json:"magic_type,omitempty"`
	BuildPoints          *BuildPointAllocation `json:"build_points,omitempty"`
	Priority             map[string]string     `json:"priority"`
	Attributes           map[string]int        `json:"attributes"`
	Skills               []Rated               `json:"skills"`
	Gear                 []Possession          `json:"gear"`
	Cyberware            []Possession          `json:"cyberware"`
	Bioware              []Possession          `json:"bioware,omitempty"`
	Spells               []Possession          `json:"spells,omitempty"`
	AdeptPowers          []Possession          `json:"adept_powers,omitempty"`
	Vehicles             []Possession          `json:"vehicles,omitempty"`
	Decks                []Possession          `json:"cyberdecks,omitempty"`
	EdgesFlaws           []Possession          `json:"edges_flaws,omitempty"`
	Contacts             []Possession          `json:"contacts,omitempty"`
	Totems               []Possession          `json:"totems,omitempty"`
	MagicGear            []Possession          `json:"magic_gear,omitempty"`
	Karma                int                   `json:"karma"`
	KarmaPool            int                   `json:"karma_pool"`
	Notes                string                `json:"notes"`
	LegacyText           string                `json:"legacy_text,omitempty"`
	LegacyBytes          []byte                `json:"legacy_bytes,omitempty"`
	LegacyWarnings       []string              `json:"legacy_warnings,omitempty"`
	LegacyEssence        *float64              `json:"legacy_essence,omitempty"`
	LegacyMagic          *int                  `json:"legacy_magic,omitempty"`
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

func (ch *Character) ResourceBudget() int {
	if ch.BuildPoints != nil {
		return ch.BuildPoints.Resources
	}
	return resourceLevels[ch.Priority["Resources"]]
}
func (ch *Character) AttributeBudget() int { return attrLevels[ch.Priority["Attributes"]] }
func (ch *Character) SkillBudget() int     { return skillLevels[ch.Priority["Skills"]] }
func (ch *Character) AttributeSpent() int {
	s := 0
	for _, k := range attributeKeys {
		s += ch.Attributes[k]
	}
	return s
}
func (ch *Character) SkillSpent() int { return ch.SkillAllocation().Active }
func (ch *Character) EffectiveAttribute(name string) int {
	race, _ := ch.raceDefinition()
	return ch.Attributes[name] + race.Modifiers[name]
}
func (ch *Character) Awakened() bool {
	cost, ok := buildMagicCosts[ch.MagicChoice()]
	return ok && cost > 0
}
func (ch *Character) Essence() float64 {
	if ch.LegacyEssence != nil {
		return *ch.LegacyEssence
	}
	s := 6.0
	for _, cy := range ch.Cyberware {
		s -= cy.ChargedEssence()
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
	return max(0, int(math.Floor(ch.Essence()-ch.BioIndex()/2+1e-9)))
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
	if !ch.Awakened() || ch.MagicChoice() == "Adept" {
		return 0
	}
	return (ch.EffectiveAttribute("Intelligence") + ch.EffectiveAttribute("Willpower") + ch.Magic()) / 3
}
func (ch *Character) Validate() []string {
	issues := []string{}
	if ch.Edition != 3 {
		issues = append(issues, "Only Shadowrun Third Edition is supported")
	}
	if !ch.Career {
		if ch.BuildPoints != nil {
			_, bpIssues := ch.buildPointTotal()
			issues = append(issues, bpIssues...)
		} else {
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
			expected := map[string]string{"Human": "E", "Dwarf": "D", "Ork": "D", "Elf": "C", "Troll": "C"}[ch.Race]
			if expected != "" && ch.Priority["Race"] != expected {
				issues = append(issues, ch.Race+" requires Race priority "+expected)
			}
			if ch.AttributeSpent() > ch.AttributeBudget() {
				issues = append(issues, fmt.Sprintf("Attribute points exceed priority budget: %d / %d", ch.AttributeSpent(), ch.AttributeBudget()))
			}
			if ch.SkillSpent() > ch.SkillBudget() {
				issues = append(issues, fmt.Sprintf("Skill points exceed priority budget: %d / %d", ch.SkillSpent(), ch.SkillBudget()))
			}
		}
	}
	if _, ok := ch.raceDefinition(); !ok {
		issues = append(issues, "Unknown race")
	}
	if _, standard := raceMods[ch.Race]; !standard && ch.BuildPoints == nil {
		issues = append(issues, "Choose Build Points for Companion metavariants or custom races.")
	}
	if ch.CustomRace != nil {
		if err := ch.CustomRace.validate(); err != nil {
			issues = append(issues, err.Error())
		}
	}
	for _, k := range attributeKeys {
		if ch.Attributes[k] < 1 || (!ch.Career && ch.Attributes[k] > 6) {
			issues = append(issues, k+" base rating outside 1–6 (before racial adjustment)")
		}
		if ch.EffectiveAttribute(k) < 1 {
			issues = append(issues, k+" racial-adjusted rating must be at least 1")
		}
	}
	issues = append(issues, ch.SkillAllocation().Issues...)
	issues = append(issues, ch.ResourceAllocation().Issues...)
	if !ch.Career {
		issues = append(issues, ch.MagicAllocation().Issues...)
	}
	issues = append(issues, ch.AugmentationIssues()...)
	if !ch.Career {
		_, edgeIssues := ch.EdgeFlawPoints()
		issues = append(issues, edgeIssues...)
	}
	if ch.Essence() < 0.01 && len(ch.Cyberware) > 0 {
		issues = append(issues, "Essence depleted or below 0.01")
	}
	return issues
}
func SaveJSON(path string, c *Character) error {
	if c == nil {
		return errors.New("cannot save a nil character")
	}
	if c.Edition != 3 {
		return errors.New("file targets unsupported edition (SR3 only)")
	}
	if c.Format != "" && c.Format != "nsrcgii.v2" {
		return fmt.Errorf("unsupported character format %q", c.Format)
	}
	if c.CustomRace != nil {
		if err := c.CustomRace.validate(); err != nil {
			return err
		}
	}
	c.Repair()
	b, e := json.MarshalIndent(c, "", "  ")
	if e != nil {
		return e
	}
	// Stage beside the destination so a failed write leaves the existing save intact.
	temp, e := os.CreateTemp(filepath.Dir(path), ".nsrcg-save-*")
	if e != nil {
		return e
	}
	defer os.Remove(temp.Name())
	if _, e = temp.Write(b); e != nil {
		temp.Close()
		return e
	}
	if e = temp.Sync(); e != nil {
		temp.Close()
		return e
	}
	if e = temp.Close(); e != nil {
		return e
	}
	return os.Rename(temp.Name(), path)
}
func LoadJSON(path string) (*Character, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	var c Character
	dec := json.NewDecoder(bytes.NewReader(b))
	// Refuse fields this version cannot round-trip, including nested item fields.
	dec.DisallowUnknownFields()
	if e = dec.Decode(&c); e != nil {
		return nil, e
	}
	var trailing interface{}
	if e = dec.Decode(&trailing); e != io.EOF {
		return nil, errors.New("character file must contain exactly one JSON object")
	}
	if c.Edition != 3 {
		return nil, errors.New("file targets unsupported edition (SR3 only)")
	}
	if c.Format != "" && c.Format != "nsrcgii.v2" {
		return nil, fmt.Errorf("unsupported character format %q", c.Format)
	}
	if c.CustomRace != nil {
		if err := c.CustomRace.validate(); err != nil {
			return nil, err
		}
	}
	c.Repair()
	return &c, nil
}
func (ch *Character) AddItem(it Item) {
	p := Possession{Name: it.Name, CatalogID: it.ID, Kind: it.Kind, Source: strings.Join(it.Sources, ", "), Quantity: 1}
	if price, err := strconv.Atoi(strings.ReplaceAll(fmt.Sprint(it.Fields["$Cost"]), ",", "")); err == nil && price >= 0 {
		p.UnitCost = &price
	}
	if notes, ok := it.Fields["Notes"].(string); ok {
		p.Notes = notes
	}
	if raw, err := json.Marshal(it.Fields); err == nil {
		p.Raw = string(raw)
	}
	switch it.Kind {
	case "cyberware":
		var parseError error
		p.EssenceCost, parseError = strconv.ParseFloat(fmt.Sprint(it.Fields["EssCost"]), 64)
		p.EssenceUnresolved = parseError != nil
		p.Modifiers = fmt.Sprint(it.Fields["Mods"])
		ch.Cyberware = append(ch.Cyberware, p)
		if ch.LegacyEssence != nil {
			v := *ch.LegacyEssence - p.ChargedEssence()
			ch.LegacyEssence = &v
		}
	case "skills":
		ch.Skills = append(ch.Skills, Rated{Name: it.Name, Category: it.Category, Rating: 1})
	case "bioware":
		if index, err := strconv.ParseFloat(fmt.Sprint(it.Fields["BioIndex"]), 64); err == nil {
			p.BioIndex = &index
		}
		ch.Bioware = append(ch.Bioware, p)
	case "spells":
		p.Rating = 1
		ch.Spells = append(ch.Spells, p)
	case "adept_powers":
		p.Rating = 1
		if cost, err := strconv.ParseFloat(fmt.Sprint(it.Fields["AdeptCost"]), 64); err == nil {
			p.PointCost = &cost
		}
		ch.AdeptPowers = append(ch.AdeptPowers, p)
	case "vehicles":
		ch.Vehicles = append(ch.Vehicles, p)
	case "cyberdecks":
		ch.Decks = append(ch.Decks, p)
	case "edges_flaws":
		if p.UnitCost != nil {
			cost := float64(*p.UnitCost)
			if fmt.Sprint(it.Fields["EorF"]) == "F" {
				cost = -cost
			}
			p.PointCost = &cost
			p.UnitCost = nil
		}
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

//go:embed sheet.html
var sheetTemplate string
var htmlTemplate = template.Must(template.New("sheet").Parse(sheetTemplate))

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
	if ch.BuildPoints != nil {
		total, _ := ch.buildPointTotal()
		prs = []kv{{"Build Points used (known costs)", fmt.Sprintf("%d / %d", total, ch.BuildPoints.Budget)}, {"Magic type", ch.BuildPoints.Magic}, {"Resources", ch.BuildPoints.Resources}}
	} else {
		for _, k := range priorityNames {
			prs = append(prs, kv{k, ch.Priority[k]})
		}
	}
	type section struct {
		Name  string
		Items []Possession
	}
	sections := []section{}
	for _, group := range PossessionGroups {
		sections = append(sections, section{group, *ch.Possessions(group)})
	}
	v := struct {
		*Character
		AttributesList []kv
		Priorities     []kv
		Sections       []section
		Resources      ResourceAllocation
		Allocation     SkillAllocation
		Issues         []string
	}{ch, attrs, prs, sections, ch.ResourceAllocation(), ch.SkillAllocation(), ch.Validate()}
	return htmlTemplate.Execute(w, v)
}
func (ch *Character) SaveHTML(path string) error {
	var b bytes.Buffer
	if e := ch.ExportHTML(&b); e != nil {
		return e
	}
	return atomicWrite(path, b.Bytes())
}
func (ch *Character) Summary() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Name: %s\r\nSR3 • %s\r\n", ch.Name, ch.Race)
	if traits := ch.RacialTraits(); traits != "" {
		fmt.Fprintf(&sb, "Racial traits: %s\r\n", traits)
	}
	for _, k := range attributeKeys {
		fmt.Fprintf(&sb, "%s: %d (base %d)\r\n", k, ch.EffectiveAttribute(k), ch.Attributes[k])
	}
	fmt.Fprintf(&sb, "\r\nReaction %d • Initiative %d + %dD6\r\nEssence %.2f • Magic %d\r\nCombat Pool %d • Spell Pool %d\r\n", ch.Reaction(), ch.Reaction(), ch.InitiativeDice(), ch.Essence(), ch.Magic(), ch.CombatPool(), ch.SpellPool())
	if ch.BuildPoints != nil {
		total, _ := ch.buildPointTotal()
		fmt.Fprintf(&sb, "\r\nBuild Points used: %d / %d (known costs)\r\nResources: ¥%d\r\n", total, ch.BuildPoints.Budget, ch.ResourceBudget())
	} else {
		fmt.Fprintf(&sb, "\r\nAttributes %d / %d points\r\nActive skills %d / %d points\r\nResources: ¥%d\r\n", ch.AttributeSpent(), ch.AttributeBudget(), ch.SkillSpent(), ch.SkillBudget(), ch.ResourceBudget())
	}
	allocation := ch.SkillAllocation()
	fmt.Fprintf(&sb, "Knowledge %d / %d • Languages %d / %d (Knowledge may cover excess)\r\n", allocation.Knowledge, allocation.KnowledgeBudget, allocation.Language, allocation.LanguageBudget)
	fmt.Fprintf(&sb, "Skills: %d • Gear: %d • Cyberware: %d\r\n", len(ch.Skills), len(ch.Gear), len(ch.Cyberware))
	fmt.Fprintf(&sb, "Bioware %d • Spells %d • Adept powers %d • Vehicles %d • Decks %d\r\n", len(ch.Bioware), len(ch.Spells), len(ch.AdeptPowers), len(ch.Vehicles), len(ch.Decks))
	magic := ch.MagicAllocation()
	fmt.Fprintf(&sb, "Spell Points %d / %d; Power Points %.2f / %d\r\n", magic.SpellSpent, magic.SpellBudget, magic.PowerSpent, ch.Magic())
	money := ch.ResourceAllocation()
	fmt.Fprintf(&sb, "Resources spent (priced entries): ¥%d; remaining ¥%d\r\n", money.Spent, ch.ResourceBudget()-money.Spent)
	issues := ch.Validate()
	if len(issues) == 0 {
		sb.WriteString("\r\nBasic creation checks passed. Advanced SR3 validation incomplete.")
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
