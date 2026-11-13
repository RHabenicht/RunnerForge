package srcapp

import (
	"fmt"
	"strings"
)

// Companion printed pp.13–14 and 34–37. Human physical purchases remain in
// Character.Attributes; animal physical purchases are independent.
type ShapeshifterCreation struct {
	Species          string         `json:"species"`
	Form             string         `json:"form"`
	Movement         string         `json:"movement,omitempty"`
	AnimalAttributes map[string]int `json:"animal_attributes"`
	// Seal tables conflict. Require a recorded GM choice, never assume errata.
	SealLimits string `json:"seal_limits,omitempty"`
}

type shifterSpecies struct {
	human, animal, mental map[string]int
	initiative            int
	movement              string
	traits                string
}

var shifterSpeciesTable = map[string]shifterSpecies{
	"Bear":    {map[string]int{"Body": 2, "Strength": 1}, map[string]int{"Body": 4, "Quickness": 1, "Strength": 4}, nil, 0, "", "Animal: dermal armor 1, Reach +1, unarmed Power +1 and Damage Level +1; running ×4"},
	"Eagle":   {map[string]int{"Quickness": 1}, map[string]int{"Quickness": 2, "Strength": 2}, map[string]int{"Charisma": 1, "Intelligence": 1}, 1, "Flying", "Animal: flying ×5; unarmed Power +1 and Damage Level +1; extra initiative die only while flying"},
	"Fox":     {nil, map[string]int{"Quickness": 1}, map[string]int{"Charisma": 2, "Intelligence": 1, "Willpower": 1}, 1, "", "Animal: running ×5"},
	"Leopard": {map[string]int{"Body": 1, "Strength": 1}, map[string]int{"Body": 2, "Quickness": 1, "Strength": 2}, map[string]int{"Charisma": 1}, 2, "", "Animal: running ×5; unarmed Power +1 and Damage Level +1"},
	"Seal":    {map[string]int{"Body": 1, "Strength": 1}, map[string]int{"Body": 2, "Quickness": 1, "Strength": 1}, map[string]int{"Charisma": 1}, 1, "Swimming", "Animal: swimming ×4; extra initiative die only while swimming; pp.36/37 limits conflict requires GM choice"},
	"Tiger":   {map[string]int{"Body": 1, "Quickness": 1, "Strength": 1}, map[string]int{"Body": 3, "Quickness": 2, "Strength": 3}, nil, 2, "", "Animal: running ×5, Reach +1, unarmed Power +1 and Damage Level +1"},
	"Wolf":    {map[string]int{"Body": 1, "Strength": 1}, map[string]int{"Body": 2, "Quickness": 1, "Strength": 1}, map[string]int{"Charisma": 1}, 1, "", "Animal: running ×5; unarmed Power +1"},
}

func physicalAttribute(name string) bool {
	return name == "Body" || name == "Quickness" || name == "Strength"
}
func (ch *Character) FormAttribute(name, form string) int {
	if ch.Shapeshifter == nil {
		return ch.ordinaryAttribute(name)
	}
	species := shifterSpeciesTable[ch.Shapeshifter.Species]
	if !physicalAttribute(name) {
		return ch.Attributes[name] + species.mental[name]
	}
	if form == "Animal" {
		return ch.Shapeshifter.AnimalAttributes[name] + species.animal[name]
	}
	return ch.Attributes[name] + species.human[name]
}
func (ch *Character) formLimit(name, form string) int {
	species := shifterSpeciesTable[ch.Shapeshifter.Species]
	if !physicalAttribute(name) {
		return 6 + species.mental[name]
	}
	if form == "Animal" {
		if ch.Shapeshifter.Species == "Seal" && ch.Shapeshifter.SealLimits == "Limits" {
			if name == "Body" {
				return 7
			}
			if name == "Quickness" {
				return 6
			}
		}
		return 6 + species.animal[name]
	}
	return 6 + species.human[name]
}
func (ch *Character) FormReaction(form string) int {
	if ch.Shapeshifter == nil {
		return ch.Reaction()
	}
	n := (ch.FormAttribute("Quickness", form) + ch.FormAttribute("Intelligence", form)) / 2
	if form == "Animal" {
		n++
	}
	return n
}
func (ch *Character) FormInitiativeDice(form string) int {
	if ch.Shapeshifter == nil {
		return ch.InitiativeDice()
	}
	n := 1
	if form == "Animal" {
		s := shifterSpeciesTable[ch.Shapeshifter.Species]
		if s.movement == "" || s.movement == ch.Shapeshifter.Movement {
			n += s.initiative
		}
	}
	return n
}
func (ch *Character) AdeptPowersActive() bool {
	return ch.Shapeshifter == nil || ch.Shapeshifter.Form == "Human"
}
func (ch *Character) ConfigureShapeshifter(s *ShapeshifterCreation) error {
	if s == nil {
		if ch.Career {
			return fmt.Errorf("shapeshifter species and purchases lock at career start")
		}
		if ch.Shapeshifter != nil {
			ch.Shapeshifter = nil
			ch.Race = "Human"
		}
		return nil
	}
	if _, ok := shifterSpeciesTable[s.Species]; !ok {
		return fmt.Errorf("choose Bear, Eagle, Fox, Leopard, Seal, Tiger or Wolf")
	}
	if s.Form != "Human" && s.Form != "Animal" {
		return fmt.Errorf("choose Human or Animal form")
	}
	if s.Movement != "Ground" && s.Movement != "Flying" && s.Movement != "Swimming" {
		return fmt.Errorf("choose Ground, Flying or Swimming movement")
	}
	if s.Movement == "Flying" && (s.Species != "Eagle" || s.Form != "Animal") {
		return fmt.Errorf("flying initiative applies to eagle animal form only")
	}
	if s.Species == "Seal" && s.SealLimits != "Modifiers" && s.SealLimits != "Limits" {
		return fmt.Errorf("GM must choose Modifiers (p.36-derived limits) or Limits (printed p.37 limits) for seal")
	}
	if s.Species != "Seal" && s.SealLimits != "" {
		return fmt.Errorf("seal table choice is only for seals")
	}
	if ch.Ghoul != nil || ch.LegacyEssence != nil || ch.LegacyMagic != nil {
		return fmt.Errorf("shapeshifters cannot combine this creation mode with ghouls or imported overrides")
	}
	for name, n := range s.AnimalAttributes {
		if !physicalAttribute(name) || n < 1 || n > 100 {
			return fmt.Errorf("invalid animal physical attribute: %s", name)
		}
	}
	for _, name := range []string{"Body", "Quickness", "Strength"} {
		if s.AnimalAttributes[name] < 1 {
			return fmt.Errorf("purchase animal %s separately", name)
		}
	}
	if ch.Career {
		old := ch.Shapeshifter
		if old == nil || old.Species != s.Species || old.SealLimits != s.SealLimits {
			return fmt.Errorf("shapeshifter species and limits lock at career start")
		}
		for _, name := range []string{"Body", "Quickness", "Strength"} {
			if s.AnimalAttributes[name] != old.AnimalAttributes[name] {
				return fmt.Errorf("use karma advancement to improve animal attributes")
			}
		}
	}
	copy := *s
	copy.AnimalAttributes = map[string]int{}
	for name, n := range s.AnimalAttributes {
		copy.AnimalAttributes[name] = n
	}
	ch.Shapeshifter = &copy
	ch.Race = s.Species + " shapeshifter"
	return nil
}
func (ch *Character) ShapeshifterIssues() []string {
	s := ch.Shapeshifter
	if s == nil {
		return nil
	}
	var issues []string
	trial := *ch
	if err := trial.ConfigureShapeshifter(s); err != nil {
		issues = append(issues, err.Error())
	}
	if ch.Race != s.Species+" shapeshifter" {
		issues = append(issues, "Metatype must match the configured shapeshifter species")
	}
	if !ch.Career {
		if ch.BuildPoints != nil {
			if ch.BuildPoints.Resources != 5000 {
				issues = append(issues, "Shapeshifters cannot spend BP on Resources: starting resources are ¥5,000")
			}
		} else {
			priority := "A"
			if ch.MagicChoice() == "Full Magician" {
				priority = "B"
			}
			if ch.Priority["Race"] != "C" || ch.Priority["Resources"] != priority {
				issues = append(issues, "Shapeshifter creation requires Race C and Resources "+priority+", with ¥5,000 actual resources")
			}
		}
		for _, name := range []string{"Body", "Quickness", "Strength"} {
			if s.AnimalAttributes[name] < 1 || s.AnimalAttributes[name] > 6 || ch.FormAttribute(name, "Animal") > ch.formLimit(name, "Animal") {
				issues = append(issues, "Animal "+name+" exceeds its creation limits")
			}
		}
		for _, p := range ch.Contacts {
			if p.UnitCost != nil && *p.UnitCost == 0 {
				issues = append(issues, "Shapeshifters must purchase all starting contacts: "+p.Name)
			}
		}
	}
	if len(ch.Cyberware) > 0 {
		issues = append(issues, "Shapeshifters cannot receive cyberware implants (Companion pp.14,34)")
	}
	if len(ch.Bioware) > 0 {
		issues = append(issues, "Shapeshifter bioware/regeneration interaction requires source and GM review")
	}
	return issues
}
func (ch *Character) ShapeshifterDescription() string {
	if ch.Shapeshifter == nil {
		return ""
	}
	s := ch.Shapeshifter
	notes := shifterSpeciesTable[s.Species].traits
	if s.Species == "Seal" {
		notes += "; recorded GM limit choice: " + s.SealLimits
	}
	return fmt.Sprintf("%s shapeshifter — Companion pp.13–14,34–37. Active form: %s (%s). Human running ×4. Dual nature; regeneration; Severe silver allergy and vulnerability. Animal form cannot speak or use Social Skills except Intimidation; adept powers are inactive. Full-magician astral projection requires human form. %s. Regeneration and silver combat effects require manual adjudication.", s.Species, s.Form, s.Movement, notes)
}
func (ch *Character) ShapeFormSummary() string {
	if ch.Shapeshifter == nil {
		return ""
	}
	var b strings.Builder
	for _, form := range []string{"Human", "Animal"} {
		fmt.Fprintf(&b, "%s: Body %d, Quickness %d, Strength %d, Reaction %d, Initiative %dD6. ", form, ch.FormAttribute("Body", form), ch.FormAttribute("Quickness", form), ch.FormAttribute("Strength", form), ch.FormReaction(form), ch.FormInitiativeDice(form))
	}
	return b.String()
}
