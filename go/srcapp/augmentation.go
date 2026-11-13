package srcapp

import (
	"math"
	"strings"
)

// Man & Machine pp. 44–45; an empty grade retains stored/imported costs.
func (p Possession) CyberGrade() (essenceFactor, priceFactor float64, valid bool) {
	grade := strings.ToLower(strings.TrimSpace(p.Grade))
	used := strings.HasPrefix(grade, "used ")
	grade = strings.TrimPrefix(grade, "used ")
	switch grade {
	case "", "basic", "standard":
		essenceFactor = 1
		priceFactor = 1
	case "alpha":
		essenceFactor = .8
		priceFactor = 2
	case "beta":
		essenceFactor = .6
		priceFactor = 4
	case "delta":
		essenceFactor = .5
		priceFactor = 8
	default:
		return 1, 1, false
	}
	if used {
		if grade == "delta" {
			return 1, 1, false
		}
		priceFactor *= .5
	}
	return essenceFactor, priceFactor, true
}
func (p Possession) ChargedEssence() float64 {
	factor, _, _ := p.CyberGrade()
	cost := p.EssenceCost
	if factor != 1 && cost > 0 {
		cost = math.Max(.01, math.Ceil(cost*factor*100-1e-9)/100)
	}
	return cost * float64(p.Count())
}

func (ch *Character) CyberwareEssence(p Possession) float64 {
	cost := p.ChargedEssence()
	if ch.Ghoul != nil {
		cost *= 2
	}
	return cost
}
func (p Possession) ChargedUnitCost() int {
	if p.UnitCost == nil {
		return 0
	}
	factor := 1.0
	if p.Grade != "" {
		_, factor, _ = p.CyberGrade()
	}
	return int(math.Ceil(float64(*p.UnitCost) * factor))
}
func (ch *Character) BioIndex() float64 {
	total := 0.0
	for _, p := range ch.Bioware {
		if p.BioIndex != nil {
			total += *p.BioIndex * float64(p.Count())
		}
	}
	return total
}
func (ch *Character) AugmentationIssues() []string {
	var issues []string
	for _, p := range ch.Cyberware {
		_, _, valid := p.CyberGrade()
		if !valid {
			issues = append(issues, "Unknown or unavailable cyberware grade: "+p.Name)
		}
		if !ch.Career && (strings.Contains(strings.ToLower(p.Grade), "beta") || strings.Contains(strings.ToLower(p.Grade), "delta")) {
			issues = append(issues, "Beta/delta cyberware is unavailable at standard character creation: "+p.Name)
		}
		if p.EssenceUnresolved {
			issues = append(issues, "Essence cost requires review: "+p.Name)
		}
		if p.EssenceCost < 0 {
			issues = append(issues, "Negative Essence cost: "+p.Name)
		}
	}
	for _, p := range ch.Bioware {
		if p.BioIndex == nil {
			issues = append(issues, "Bio Index not set: "+p.Name)
		} else if *p.BioIndex < 0 {
			issues = append(issues, "Negative Bio Index: "+p.Name)
		}
	}
	if ch.BioIndex() > ch.Essence() {
		issues = append(issues, "Bio Index exceeds Essence; apply the Man & Machine excessive-bioware rules")
	}
	return issues
}
