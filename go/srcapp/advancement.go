package srcapp

import (
	"encoding/json"
	"fmt"
)

type Advancement struct {
	Kind  string `json:"kind"`
	Name  string `json:"name"`
	From  int    `json:"from"`
	To    int    `json:"to"`
	Karma int    `json:"karma"`
}

func (ch *Character) BeginCareer() error {
	if ch.Career {
		return nil
	}
	if issues := ch.Validate(); len(issues) > 0 {
		return fmt.Errorf("resolve creation review items before starting play: %s", issues[0])
	}
	ch.freezeGhoulCreation()
	snapshot, err := json.Marshal(ch)
	if err != nil {
		return err
	}
	ch.CreationSnapshot = snapshot
	ch.Career = true
	return nil
}
func (ch *Character) AdvanceAttribute(name string) error {
	if !ch.Career {
		return fmt.Errorf("start career mode before spending karma")
	}
	if !containsAttribute(name) {
		return fmt.Errorf("only physical and mental attributes can be improved with karma")
	}
	limit := ch.RacialModifiedLimit(name)
	current := ch.EffectiveAttribute(name)
	next := current + 1
	if next > (limit*3+1)/2 {
		return fmt.Errorf("attribute maximum reached")
	}
	cost := next * 2
	if next > limit {
		cost = next * 3
	}
	if ch.Karma < cost {
		return fmt.Errorf("requires %d unspent karma", cost)
	}
	ch.Karma -= cost
	ch.Attributes[name]++
	ch.Advancements = append(ch.Advancements, Advancement{"Attribute", name, current, next, cost})
	return nil
}
func (ch *Character) AdvanceSkill(index int, specialization bool) error {
	if !ch.Career {
		return fmt.Errorf("start career mode before spending karma")
	}
	if index < 0 || index >= len(ch.Skills) {
		return fmt.Errorf("select a skill")
	}
	if specialization {
		return ch.LearnOrAdvanceSpecialization(index, ch.Skills[index].Specialization)
	}
	skill := &ch.Skills[index]
	link := skill.Link()
	if link == "" {
		return fmt.Errorf("set the linked attribute first")
	}
	attribute := ch.EffectiveAttribute(link)
	if link == "Reaction" {
		attribute = (ch.EffectiveAttribute("Quickness") + ch.EffectiveAttribute("Intelligence")) / 2
	}
	current := skill.Rating
	name := skill.Name
	kind := "Skill"
	next := current + 1
	// Multipliers are represented in halves; integer division rounds down.
	factor := 3
	if skill.Class() != "Active" {
		factor = 2
	}

	if next > attribute {
		factor++
	}
	if next > 2*attribute {
		factor++
	}
	cost := next * factor / 2
	if ch.Karma < cost {
		return fmt.Errorf("requires %d unspent karma", cost)
	}
	// Preserve the specialization's independent rating when its base advances.
	if skill.Specialization != "" && skill.SpecializationValue == nil {
		value := skill.SpecializationRating()
		skill.SpecializationValue = &value
	}
	skill.Rating = next
	ch.Karma -= cost
	ch.Advancements = append(ch.Advancements, Advancement{kind, name, current, next, cost})
	return nil
}
