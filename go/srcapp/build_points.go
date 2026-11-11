package srcapp

import (
	"fmt"
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
	skills := ch.SkillAllocation()
	skillCost := skills.Active
	if skillCost > 60 {
		issues = append(issues, "Active skills exceed the Companion 60 BP cap.")
	}
	race, _ := ch.raceDefinition()
	total := race.BuildPointCost + magicCost + resourceCost + attributeCost + skillCost
	if total > bp.Budget {
		issues = append(issues, fmt.Sprintf("Build Point budget exceeded (%d / %d).", total, bp.Budget))
	}
	return total, issues
}
