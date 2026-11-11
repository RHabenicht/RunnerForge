package srcapp

import (
	"fmt"
	"strings"
)

func (ch *Character) MagicChoice() string {
	if ch.BuildPoints != nil {
		return ch.BuildPoints.Magic
	}
	if _, ok := buildMagicCosts[ch.MagicType]; ok {
		return ch.MagicType
	}
	if strings.Contains(strings.ToLower(ch.MagicType), "adept") {
		return "Adept"
	}
	if ch.Priority["Magic"] == "A" {
		return "Full Magician"
	}
	if ch.Priority["Magic"] == "B" {
		return "Aspected Magician"
	}
	return "Mundane"
}

type MagicAllocation struct {
	SpellBudget int
	SpellSpent  int
	PowerSpent  float64
	Issues      []string
}

func (ch *Character) MagicAllocation() MagicAllocation {
	a := MagicAllocation{}
	choice := ch.MagicChoice()
	switch choice {
	case "Full Magician":
		a.SpellBudget = 25
	case "Aspected Magician":
		a.SpellBudget = 35
	}
	base := a.SpellBudget
	if ch.SpellPointsPurchased < 0 || ch.SpellPointsPurchased+base > 50 {
		a.Issues = append(a.Issues, "Purchased Spell Points must leave at most 50 total before sales")
	}
	if ch.SpellPointsSold < 0 || ch.SpellPointsSold%5 != 0 || ch.SpellPointsSold > base || (ch.SpellPointsSold > 0 && ch.BuildPoints == nil) {
		a.Issues = append(a.Issues, "Spell Point sales require Build Point creation and multiples of five, within the starting allowance")
	}
	if base == 0 && (ch.SpellPointsPurchased != 0 || ch.SpellPointsSold != 0) {
		a.Issues = append(a.Issues, "Only full or aspected magicians can buy or sell Spell Points")
	}
	a.SpellBudget += ch.SpellPointsPurchased - ch.SpellPointsSold
	if ch.BuildPoints == nil {
		expected := map[string]string{"Full Magician": "A", "Aspected Magician": "B", "Adept": "B"}[choice]
		if expected != "" && ch.Priority["Magic"] != expected {
			a.Issues = append(a.Issues, choice+" requires Magic priority "+expected)
		}
	}
	for _, spell := range ch.Spells {
		if spell.Rating < 1 || spell.Rating > 6 {
			a.Issues = append(a.Issues, "Starting spell Force must be 1–6: "+spell.Name)
		}
		a.SpellSpent += spell.Rating
	}
	if a.SpellSpent > a.SpellBudget || (len(ch.Spells) > 0 && a.SpellBudget == 0) {
		a.Issues = append(a.Issues, fmt.Sprintf("Spell Point allowance exceeded (%d / %d).", a.SpellSpent, a.SpellBudget))
	}
	for _, power := range ch.AdeptPowers {
		if power.PointCost == nil {
			a.Issues = append(a.Issues, "Power Point cost not set: "+power.Name)
		} else if *power.PointCost < 0 {
			a.Issues = append(a.Issues, "Invalid Power Point cost: "+power.Name)
		} else {
			a.PowerSpent += *power.PointCost
		}
	}
	if len(ch.AdeptPowers) > 0 && choice != "Adept" {
		a.Issues = append(a.Issues, "Adept powers require adept magical ability")
	}
	if a.PowerSpent > float64(ch.Magic()) {
		a.Issues = append(a.Issues, "Power Points exceed Magic rating")
	}
	return a
}
func (ch *Character) EdgeFlawPoints() (float64, []string) {
	var edges, flaws float64
	edgeCount, flawCount := 0, 0
	var issues []string
	for _, entry := range ch.EdgesFlaws {
		if entry.PointCost == nil {
			issues = append(issues, "Build Point cost not set: "+entry.Name)
			continue
		}
		cost := *entry.PointCost
		if cost != float64(int(cost)) {
			issues = append(issues, "Edge/Flaw cost must be a whole number: "+entry.Name)
		}
		if cost >= 0 {
			edges += cost
			edgeCount++
		} else {
			flaws -= cost
			flawCount++
		}
	}
	if edgeCount > 5 || flawCount > 5 {
		issues = append(issues, "Starting characters may have at most five Edges and five Flaws")
	}
	if edges > 6 {
		issues = append(issues, "Edges exceed the recommended 6-point limit")
	}
	if flaws > 6 {
		issues = append(issues, "Flaws exceed the recommended 6-point limit")
	}
	if ch.BuildPoints == nil && edges != flaws {
		issues = append(issues, "Priority creation requires equal Edge and Flaw points")
	}
	return edges - flaws, issues
}
