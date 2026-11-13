package srcapp

import (
	"fmt"
	"strings"
)

// GhoulCreation records a starting ghoul's third-stage test, not a later
// infection. Attributes remain the ordinary, pre-transformation purchases.
// Revised Shadowrun Companion, printed pp. 33–34.
type GhoulCreation struct {
	Successes     int  `json:"successes"`
	GMAlternative bool `json:"gm_alternative,omitempty"`
	// Each entry replaces one point below 1 in Charisma/Intelligence with a
	// two-point loss from the named attribute, with explicit GM permission.
	Sacrifices    []string       `json:"sacrifices,omitempty"`
	PreventedLoss map[string]int `json:"prevented_loss,omitempty"`
}

func (ch *Character) ConfigureGhoul(g *GhoulCreation) error {
	if ch.Career {
		return fmt.Errorf("starting-ghoul choices are locked after beginning a career")
	}
	if g != nil {
		if ch.Shapeshifter != nil {
			return fmt.Errorf("ghoul creation requires a metahuman, not a shapeshifter")
		}
		if ch.BuildPoints == nil {
			return fmt.Errorf("choose Build Points before starting ghoul creation")
		}
		if ch.LegacyEssence != nil || ch.LegacyMagic != nil {
			return fmt.Errorf("recreate the ordinary character without imported Essence/Magic overrides first")
		}
		if g.Successes < 0 || g.Successes > ch.GhoulTestDice() {
			return fmt.Errorf("enter actual successes from 0 to %d; roll Willpower (6) with two extra dice, no Karma", ch.GhoulTestDice())
		}
		for _, name := range g.Sacrifices {
			if !containsAttribute(name) {
				return fmt.Errorf("unknown sacrifice attribute: %s", name)
			}
		}
		if len(g.Sacrifices) > 0 && !g.GMAlternative {
			return fmt.Errorf("attribute substitutions require GM permission")
		}
		copy := *g
		copy.Sacrifices = append([]string(nil), g.Sacrifices...)
		copy.PreventedLoss = nil
		ch.Ghoul = &copy
	} else {
		ch.Ghoul = nil
	}
	return nil
}

func (ch *Character) ghoulPreventedLoss(name string) int {
	if name != "Charisma" && name != "Intelligence" {
		return 0
	}
	if ch.Career {
		return ch.Ghoul.PreventedLoss[name]
	}
	return max(0, 1-ch.ghoulAttribute(name))
}

func (ch *Character) freezeGhoulCreation() {
	if ch.Ghoul == nil || !ch.Ghoul.GMAlternative {
		return
	}
	ch.Ghoul.PreventedLoss = map[string]int{}
	for _, name := range []string{"Charisma", "Intelligence"} {
		ch.Ghoul.PreventedLoss[name] = ch.ghoulPreventedLoss(name)
	}
}

var ghoulCommon = map[string]int{"Body": 2, "Strength": 1, "Charisma": -1, "Intelligence": -1}

func (ch *Character) ordinaryAttribute(name string) int {
	race, _ := ch.raceDefinition()
	return ch.Attributes[name] + race.Modifiers[name]
}

func (ch *Character) ghoulAttribute(name string) int {
	n := ch.ordinaryAttribute(name) + ghoulCommon[name]
	if ch.Ghoul == nil {
		return ch.ordinaryAttribute(name)
	}
	switch {
	case ch.Ghoul.Successes == 0:
		if name == "Intelligence" || name == "Charisma" {
			n -= 2
		}
	case ch.Ghoul.Successes == 1:
		if name == "Intelligence" || name == "Charisma" {
			n--
		}
	case ch.Ghoul.Successes == 2:
		if name == "Charisma" {
			n--
		}
	case ch.Ghoul.Successes >= 4:
		if name == "Willpower" {
			n++
		}
	}
	return n
}

// Creation costs/pools are assigned before ghoul transformation (Companion p.14).
func (ch *Character) skillAttribute(name string) int {
	if ch.Shapeshifter != nil {
		return ch.FormAttribute(name, "Human")
	}
	if ch.Ghoul != nil && !ch.Career {
		return ch.ordinaryAttribute(name)
	}
	return ch.EffectiveAttribute(name)
}

// Infection-test changes never change racial limits (Companion p.34).
func (ch *Character) RacialModifiedLimit(name string) int {
	if ch.Shapeshifter != nil {
		return ch.formLimit(name, ch.Shapeshifter.Form)
	}
	race, _ := ch.raceDefinition()
	n := 6 + race.Modifiers[name]
	if ch.Ghoul != nil {
		n += ghoulCommon[name]
	}
	return n
}

func (ch *Character) GhoulTestDice() int { return ch.ordinaryAttribute("Willpower") + 2 }

func (ch *Character) GhoulIssues() []string {
	if ch.Ghoul == nil {
		return nil
	}
	var issues []string
	if ch.Shapeshifter != nil {
		issues = append(issues, "Ghoul creation requires a metahuman, not a shapeshifter")
	}
	if !ch.Career {
		for _, name := range attributeKeys {
			if ch.ordinaryAttribute(name) < 1 {
				issues = append(issues, "Build a valid ordinary character before ghoul transformation: "+name+" is below 1")
			}
		}
	}
	if ch.BuildPoints == nil {
		issues = append(issues, "Starting ghouls require Build Point creation (Companion p.33)")
	}
	if ch.Ghoul.Successes < 0 || ch.Ghoul.Successes > 100 || (!ch.Career && ch.Ghoul.Successes > ch.GhoulTestDice()) {
		issues = append(issues, "Record the third-stage Willpower (6) test successes; starting ghouls roll two extra dice and cannot spend Karma")
	}
	if ch.LegacyEssence != nil || ch.LegacyMagic != nil {
		issues = append(issues, "Starting-ghoul creation cannot use imported Essence/Magic overrides; recreate the ordinary character first")
	}
	required := 0
	for _, name := range []string{"Charisma", "Intelligence"} {
		required += ch.ghoulPreventedLoss(name)
	}
	for name, n := range ch.Ghoul.PreventedLoss {
		if (name != "Charisma" && name != "Intelligence") || n < 0 || n > 100 {
			issues = append(issues, "Invalid preserved ghoul attribute substitution")
		}
	}
	if ch.Ghoul.GMAlternative {
		if len(ch.Ghoul.Sacrifices) != required {
			issues = append(issues, fmt.Sprintf("GM alternative requires %d explicit two-point attribute sacrifices", required))
		}
		for _, name := range ch.Ghoul.Sacrifices {
			if !containsAttribute(name) {
				issues = append(issues, "Unknown ghoul sacrifice attribute: "+name)
			}
		}
	} else {
		if required > 0 {
			issues = append(issues, "Ghoul transformation reduces Charisma/Intelligence below 1: character becomes an NPC unless the GM permits the explicit two-point alternative")
		}
		if len(ch.Ghoul.Sacrifices) > 0 {
			issues = append(issues, "Ghoul attribute sacrifices require explicit GM permission")
		}
	}
	if ch.Essence() <= 0 {
		issues = append(issues, "Ghoul transformation leaves no Essence: character dies (Companion p.34)")
	}
	return issues
}

func (ch *Character) GhoulDescription() string {
	if ch.Ghoul == nil {
		return ""
	}
	alternative := "No GM attribute substitution."
	if ch.Ghoul.GMAlternative {
		alternative = "GM-approved two-point sacrifices: " + strings.Join(ch.Ghoul.Sacrifices, ", ") + "."
	}
	return fmt.Sprintf("Starting ghoul — Companion pp.33–34. Third-stage Willpower (6) test: %d successes, no Karma. %s Mild sunlight allergy; blindness; Sensitive System; raw-flesh diet; dual nature; VITAS immunity; hearing/smell Perception TN −2; running multiplier +1. Starting cyberware is installed after transformation and costs double Essence.", ch.Ghoul.Successes, alternative)
}
