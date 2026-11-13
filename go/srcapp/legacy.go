package srcapp

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var iniKvp = regexp.MustCompile(`^([^=]+)=(.*)$`)
var attrNum = regexp.MustCompile(`^[+-]?\d+`)
var skillNum = regexp.MustCompile(`\[(\d+)(?:/\d+)?\]`)

// ReadLegacy uses the NSRCG 3.6x .sr3 files from the user-owned 2005 installation.
// Preserve the original text losslessly, including unrecognized sections.
func ReadLegacy(filename string) (*Character, error) {
	b, e := os.ReadFile(filename)
	if e != nil {
		return nil, e
	}
	return ParseLegacy(string(b))
}
func ParseLegacy(text string) (*Character, error) {
	c := NewCharacter()
	c.LegacyText = text
	c.LegacyBytes = []byte(text)
	sec := ""
	m := map[string]map[string]string{}
	for _, l := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, ";") {
			continue
		}
		if strings.HasPrefix(l, "[") && strings.HasSuffix(l, "]") {
			sec = strings.Trim(l, "[]")
			if m[sec] == nil {
				m[sec] = map[string]string{}
			}
			continue
		}
		mm := iniKvp.FindStringSubmatch(l)
		if len(mm) == 3 && sec != "" {
			m[sec][strings.TrimSpace(mm[1])] = strings.TrimSpace(mm[2])
		}
	}
	if _, ok := m["General"]; !ok {
		return nil, fmt.Errorf("not a recognized NSRCG .sr3 character file")
	}
	g := m["General"]
	c.Name = decUnderscore(g["CharName"])
	friendly := map[string]string{"archCovertOpsSpec": "Covert Operations Specialist", "archTheFace": "The Face", "archTheInvestigator": "The Investigator", "archWeaponsSpec": "Weapons Specialist"}
	if name, ok := friendly[c.Name]; ok {
		c.Name = name
	}
	c.RealName = decUnderscore(g["RealName"])
	c.Race = m["Race"]["RaceName"]
	if _, ok := raceMods[c.Race]; !ok {
		c.Race = "Human"
		c.LegacyWarnings = append(c.LegacyWarnings, "unrecognized race in original")
	}
	for _, k := range attributeKeys {
		val := strings.Split(m["Attributes"][k], "|")[0]
		if n, e := strconv.Atoi(strings.TrimSpace(val)); e == nil {
			c.Attributes[k] = n
		}
	}
	// NSRCG's priority values encode category allocations by numeric slot; the recovered
	// raw priority codes are preserved, not guessed when their mapping is uncertain.
	c.Priority = map[string]string{"Resources": "A", "Attributes": "B", "Race": "E", "Skills": "C", "Magic": "D"}
	c.LegacyWarnings = append(c.LegacyWarnings, "Priority letters require confirmation against original file; provisional defaults used")
	if v := strings.Split(m["Attributes"]["Essence"], "|"); len(v) > 0 {
		tail := v[len(v)-1]
		if n, e := strconv.ParseFloat(tail, 64); e == nil {
			c.LegacyEssence = &n
		}
	}
	if v := strings.Split(m["Attributes"]["Magic"], "|"); len(v) > 0 {
		tail := v[len(v)-1]
		if n, e := strconv.Atoi(tail); e == nil {
			c.LegacyMagic = &n
		}
	}
	if c.LegacyMagic != nil && *c.LegacyMagic > 0 {
		c.Priority["Magic"] = "A"
		c.Priority["Resources"] = "D"
		c.Priority["Skills"] = "B"
		c.Priority["Attributes"] = "C"
		c.Priority["Race"] = "E"
		c.LegacyWarnings = append(c.LegacyWarnings, "Awakened import: check priority values")
	}
	for _, k := range sortedKeys(m["Skills"], "Skill") {
		val := m["Skills"][k]
		if !strings.HasPrefix(k, "Skill") {
			continue
		}
		fields := strings.SplitN(val, "~", 2)
		name := decUnderscore(fields[0])
		rating := 0
		if len(fields) > 1 {
			sm := skillNum.FindStringSubmatch(fields[1])
			if len(sm) == 2 {
				rating, _ = strconv.Atoi(sm[1])
			}
		}
		if rating == 0 {
			continue
		}
		group := "Active"
		if strings.Contains(name, "(KNO)") {
			group = "Knowledge"
		}
		if strings.Contains(name, "(LAN)") {
			group = "Language"
		}
		if strings.Contains(name, "(RW)") {
			group = "Literacy"
		}
		c.Skills = append(c.Skills, Rated{Name: name, Rating: rating, Category: group})
	}
	for _, k := range sortedKeys(m["Gear"], "Gear") {
		val := m["Gear"][k]
		if !strings.HasPrefix(k, "Gear") || k == "Gear" {
			continue
		}
		piece := strings.SplitN(val, "~", 2)
		name := decUnderscore(piece[0])
		name = strings.TrimPrefix(name, "X:")
		c.Gear = append(c.Gear, Possession{Name: strings.TrimSpace(name), Kind: "gear", Raw: val})
	}
	for _, k := range sortedKeys(m["Cyberware"], "Cyber") {
		val := m["Cyberware"][k]
		if !strings.HasPrefix(k, "Cyber") || k == "SensitiveSystem" {
			continue
		}
		piece := strings.SplitN(val, "~", 2)
		name := decUnderscore(piece[0])
		p := Possession{Name: name, Kind: "cyberware", Raw: val}
		if len(piece) > 1 {
			f := strings.Split(piece[1], "~")
			if len(f) > 0 {
				p.EssenceCost, _ = strconv.ParseFloat(f[0], 64)
			}
			if len(f) > 3 {
				p.Modifiers = f[3]
			}
		}
		c.Cyberware = append(c.Cyberware, p)
	}
	c.Notes = decUnderscore(g["Notes"])
	// These saved prices are purchase totals, not catalog unit prices. Import
	// each lifestyle as one lump sum so prepaid months are not charged twice.
	for _, k := range sortedKeys(g, "LifeStyle") {
		if _, err := strconv.Atoi(strings.TrimPrefix(k, "LifeStyle")); err != nil {
			continue
		}
		fields := strings.Split(g[k], "~")
		if strings.TrimSpace(fields[0]) == "" {
			continue
		}
		p := Possession{Name: decUnderscore(fields[0]), Kind: "lifestyle", Quantity: 1, Raw: g[k], Notes: "Imported as one purchase total; original duration and details retained below."}
		p.UnitCost = legacyPrice(fields, 5)
		c.Lifestyles = append(c.Lifestyles, p)
	}
	for _, k := range sortedKeys(m["Contacts"], "Contact") {
		if _, err := strconv.Atoi(strings.TrimPrefix(k, "Contact")); err != nil {
			continue
		}
		fields := strings.Split(m["Contacts"][k], "~")
		if strings.TrimSpace(fields[0]) == "" {
			continue
		}
		p := Possession{Name: decUnderscore(fields[0]), Kind: "contact", Quantity: 1, Raw: m["Contacts"][k]}
		p.UnitCost = legacyPrice(fields, 8)
		if len(fields) > 5 {
			p.Rating, _ = strconv.Atoi(strings.TrimSpace(fields[5]))
		}
		c.Contacts = append(c.Contacts, p)
	}
	if v, e := strconv.Atoi(m["Karma"]["RemainKarma"]); e == nil {
		c.Karma = v
	}
	if v, e := strconv.Atoi(m["Karma"]["KarmaPool"]); e == nil {
		c.KarmaPool = v
	}
	if c.Name == "" {
		c.Name = decUnderscore(g["Archtype"])
	}
	if c.Name == "" {
		c.Name = "Imported Character"
	}
	return c, nil
}

// Missing or malformed prices must stay unresolved, including negative values.
// Explicit zero prices in the save represent free entries and remain zero.
func legacyPrice(fields []string, index int) *int {
	if index >= len(fields) {
		return nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(fields[index]))
	if err != nil || value < 0 || value > 1000000000 {
		return nil
	}
	return &value
}
func decUnderscore(s string) string { return strings.ReplaceAll(s, "_", " ") }
func ParseLeadingInt(s string) int {
	v := attrNum.FindString(strings.TrimSpace(s))
	n, _ := strconv.Atoi(v)
	return n
}

func sortedKeys(items map[string]string, prefix string) []string {
	var keys []string
	for k := range items {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		ni, _ := strconv.Atoi(strings.TrimPrefix(keys[i], prefix))
		nj, _ := strconv.Atoi(strings.TrimPrefix(keys[j], prefix))
		return ni < nj
	})
	return keys
}
