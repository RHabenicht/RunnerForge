package srcapp

import (
	"fmt"
	"strings"
)

// Shadowrun Companion, pp. 13–14. A nil allocation selects Priority creation.
type BuildPointAllocation struct {
	Budget    int    `json:"budget"`
	Resources int    `json:"resources"`
	Magic     string `json:"magic"`
}

var buildRaceCosts = map[string]int{"Human": 0, "Dwarf": 5, "Ork": 5, "Elf": 10, "Troll": 10}
var buildMagicCosts = map[string]int{"Mundane": 0, "Full Magician": 30, "Aspected Magician": 25, "Adept": 25}
var buildResourceCosts = map[int]int{500: -5, 5000: 0, 20000: 5, 90000: 10, 200000: 15, 400000: 20, 650000: 25, 1000000: 30}

func (ch *Character) buildPointTotal() (int, []string) {
	bp := ch.BuildPoints
	issues := []string{}
	resourceCost, resourceOK := buildResourceCosts[bp.Resources]
	magicCost, magicOK := buildMagicCosts[bp.Magic]
	if !resourceOK || !magicOK || bp.Budget < 1 {
		issues = append(issues, "Choose a valid Build Point budget, resource tier and magic type.")
	}
	attributeCost := 2 * ch.AttributeSpent()
	if attributeCost > 60 {
		issues = append(issues, "Attributes exceed the default 60 BP cap.")
	}
	skillCost := 0
	catalog, err := LoadCatalogue()
	if err != nil {
		return 0, append(issues, "Cannot verify skill costs without the catalogue.")
	}
	links := make(map[string]string)
	for _, entry := range catalog.Entries {
		if entry.Kind == "skills" {
			links[strings.ToLower(entry.Name)] = entry.Category
		}
	}
	attributes := map[string]string{"BOD": "Body", "QCK": "Quickness", "STR": "Strength", "CHA": "Charisma", "INT": "Intelligence", "WIL": "Willpower"}
	for _, skill := range ch.Skills {
		code := links[strings.ToLower(strings.TrimSpace(skill.Name))]
		group := strings.ToLower(skill.Category)
		if code == "KNO" || strings.Contains(group, "knowledge") || strings.Contains(group, "language") {
			continue
		}
		attribute, known := attributes[code]
		if !known {
			issues = append(issues, "Linked attribute unavailable for "+skill.Name+".")
			continue
		}
		skillCost += skill.Rating + max(0, skill.Rating-ch.EffectiveAttribute(attribute))
	}
	race, _ := ch.raceDefinition()
	total := race.BuildPointCost + magicCost + resourceCost + attributeCost + skillCost
	if total > bp.Budget {
		issues = append(issues, fmt.Sprintf("Build Point budget exceeded (%d / %d).", total, bp.Budget))
	}
	return total, issues
}
