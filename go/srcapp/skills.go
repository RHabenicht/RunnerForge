package srcapp

import (
	"fmt"
	"strings"
	"sync"
)

// Rating is the final base rating. At creation a specialization has Rating+2,
// and the purchased rating (used for cost) is Rating+1. SR3 pp. 57–59.
func (s Rated) SpecializationRating() int {
	if s.SpecializationValue != nil {
		return *s.SpecializationValue
	}
	if strings.TrimSpace(s.Specialization) == "" {
		return 0
	}
	return s.Rating + 2
}
func (s Rated) PurchasedRating() int {
	if strings.TrimSpace(s.Specialization) != "" {
		return s.Rating + 1
	}
	return s.Rating
}
func (s Rated) ReadWriteRating() int {
	if s.Class() == "Language" {
		return s.Rating / 2
	}
	return 0
}

var skillCodesOnce sync.Once
var skillCodes map[string]string

func skillCatalogCode(name string) string {
	skillCodesOnce.Do(func() {
		skillCodes = map[string]string{}
		cat, err := LoadCatalogue()
		if err != nil {
			return
		}
		for _, it := range cat.Entries {
			if it.Kind == "skills" {
				skillCodes[strings.ToLower(it.Name)] = it.Category
			}
		}
	})
	return skillCodes[strings.ToLower(strings.TrimSpace(name))]
}
func (s Rated) Class() string {
	category := strings.ToLower(s.Category)
	switch {
	case strings.Contains(category, "literacy"):
		return "Literacy"
	case category == "lan" || strings.Contains(category, "language"):
		return "Language"
	case category == "kno" || strings.Contains(category, "knowledge"):
		return "Knowledge"
	}
	switch skillCatalogCode(s.Name) {
	case "KNO":
		return "Knowledge"
	case "LAN":
		return "Language"
	}
	return "Active"
}
func (s Rated) Link() string {
	if s.Class() != "Active" {
		return "Intelligence"
	}
	if s.LinkedAttribute != "" {
		return s.LinkedAttribute
	}
	return map[string]string{"BOD": "Body", "QCK": "Quickness", "STR": "Strength", "CHA": "Charisma", "INT": "Intelligence", "WIL": "Willpower", "REA": "Reaction"}[skillCatalogCode(s.Name)]
}

type SkillAllocation struct {
	Active          int
	Knowledge       int
	Language        int
	KnowledgeBudget int
	LanguageBudget  int
	Issues          []string
}

func (ch *Character) SkillAllocation() SkillAllocation {
	intelligence := max(0, ch.EffectiveAttribute("Intelligence"))
	result := SkillAllocation{KnowledgeBudget: intelligence * 5, LanguageBudget: intelligence * 3 / 2}
	if ch.BuildPoints != nil && ch.BuildPoints.ExtraKnowledge > 0 {
		result.KnowledgeBudget += ch.BuildPoints.ExtraKnowledge
	}
	names := map[string]bool{}
	for _, skill := range ch.Skills {
		name := strings.TrimSpace(skill.Name)
		class := skill.Class()
		key := strings.ToLower(class + ":" + name)
		if names[key] {
			result.Issues = append(result.Issues, "Duplicate skill: "+name)
		}
		names[key] = true
		if name == "" {
			result.Issues = append(result.Issues, "Skill name is required")
		}
		rating := skill.PurchasedRating()
		if skill.Rating < 1 || (!ch.Career && rating > 6) {
			result.Issues = append(result.Issues, "Creation skill ratings must be 1–6 before specialization: "+name)
		}
		if class == "Literacy" {
			continue
		} // Preserve explicitly imported reading/writing ratings.
		link := skill.Link()
		if link == "" || (!containsAttribute(link) && link != "Reaction") {
			result.Issues = append(result.Issues, "Choose a linked attribute for "+name)
			continue
		}
		linked := ch.EffectiveAttribute(link)
		if link == "Reaction" {
			linked = (ch.EffectiveAttribute("Quickness") + ch.EffectiveAttribute("Intelligence")) / 2
		}
		cost := rating + max(0, rating-linked)
		switch class {
		case "Knowledge":
			result.Knowledge += cost
		case "Language":
			result.Language += cost
		default:
			result.Active += cost
		}
		if (strings.EqualFold(name, "Sorcery") || strings.EqualFold(name, "Conjuring")) && !ch.Awakened() {
			result.Issues = append(result.Issues, name+" requires magical ability")
		}
	}
	// Unspent Knowledge points may pay for Language skills; the reverse is not allowed.
	usedKnowledge := result.Knowledge + max(0, result.Language-result.LanguageBudget)
	if !ch.Career && usedKnowledge > result.KnowledgeBudget {
		result.Issues = append(result.Issues, fmt.Sprintf("Knowledge/Language points exceeded (%d / %d Knowledge points after Language allowance).", usedKnowledge, result.KnowledgeBudget))
	}
	return result
}
