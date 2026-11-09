package srcapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

// RaceDefinition is a self-contained house rule stored with the character.
type RaceDefinition struct {
	Name           string         `json:"name"`
	BuildPointCost int            `json:"build_point_cost"`
	Modifiers      map[string]int `json:"modifiers"`
	Traits         string         `json:"traits,omitempty"`
}

type metavariant struct {
	base      string
	modifiers map[string]int
	traits    string
}

// Revised Shadowrun Companion, pp. 38–43. Modifiers replace the parent modifiers.
var companionVariants = map[string]metavariant{
	"Cyclops":    {"Troll", map[string]int{"Body": 5, "Quickness": -1, "Strength": 6, "Intelligence": -2, "Charisma": -2}, "+1 Reach, +2 ranged attack target numbers, no dermal armor or thermographic vision"},
	"Dryad":      {"Elf", map[string]int{"Body": -1, "Quickness": 1, "Strength": -1, "Charisma": 3}, "Low-light vision, Mild urban allergy, Animal Empathy for birds and small tree-dwelling animals"},
	"Fomori":     {"Troll", map[string]int{"Body": 4, "Quickness": -1, "Strength": 3, "Intelligence": -2}, "+1 Reach, thermographic vision, no dermal armor"},
	"Giant":      {"Troll", map[string]int{"Body": 5, "Quickness": -1, "Strength": 5, "Intelligence": -2, "Charisma": -2}, "+1 Reach, thermographic vision, no dermal armor"},
	"Gnome":      {"Dwarf", map[string]int{"Body": 1, "Strength": 1, "Willpower": 2}, "Thermographic vision, no dwarf disease or toxin resistance"},
	"Hobgoblin":  {"Ork", map[string]int{"Body": 2, "Strength": 2, "Charisma": -1}, "Low-light vision"},
	"Koborokuru": {"Dwarf", nil, "Standard dwarf traits, running multiplier ×3, both Natural Immunity Edges with toxins chosen at creation"},
	"Menehune":   {"Dwarf", map[string]int{"Body": 2, "Strength": 1, "Willpower": 1}, "Thermographic vision, dwarf disease and toxin resistance"},
	"Minotaur":   {"Troll", map[string]int{"Body": 4, "Quickness": -1, "Strength": 3, "Intelligence": -1, "Charisma": -1}, "+1 Reach, thermographic vision, dermal armor"},
	"Night One":  {"Elf", map[string]int{"Quickness": 2, "Charisma": 2}, "Low-light vision, Mild sunlight allergy"},
	"Ogre":       {"Ork", map[string]int{"Body": 3, "Strength": 2, "Intelligence": -1}, "No low-light vision"},
	"Oni":        {"Ork", map[string]int{"Body": 2, "Strength": 2, "Charisma": -1, "Intelligence": -1, "Willpower": 1}, "Low-light vision"},
	"Satyr":      {"Ork", map[string]int{"Body": 3, "Strength": 2, "Quickness": -1, "Intelligence": -1, "Charisma": -1, "Willpower": 1}, "Low-light vision, running multiplier ×4"},
	"Wakyambi":   {"Elf", map[string]int{"Charisma": 2, "Willpower": 1}, "Low-light vision"},
}

func RaceNames() []string {
	return []string{"Human", "Dwarf", "Elf", "Ork", "Troll", "Cyclops", "Dryad", "Fomori", "Giant", "Gnome", "Hobgoblin", "Koborokuru", "Menehune", "Minotaur", "Night One", "Ogre", "Oni", "Satyr", "Wakyambi"}
}

func (ch *Character) raceDefinition() (RaceDefinition, bool) {
	if ch.CustomRace != nil && ch.CustomRace.Name == ch.Race {
		return *ch.CustomRace, true
	}
	if variant, ok := companionVariants[ch.Race]; ok {
		mods := variant.modifiers
		if mods == nil {
			mods = raceMods[variant.base]
		}
		return RaceDefinition{ch.Race, buildRaceCosts[variant.base] + 5, mods, variant.traits}, true
	}
	mods, ok := raceMods[ch.Race]
	return RaceDefinition{ch.Race, buildRaceCosts[ch.Race], mods, ""}, ok
}

func (ch *Character) RacialTraits() string {
	race, _ := ch.raceDefinition()
	return race.Traits
}

func (race RaceDefinition) validate() error {
	if race.Name == "" || strings.TrimSpace(race.Name) != race.Name ||
		len([]rune(race.Name)) > 80 || strings.ContainsFunc(race.Name, unicode.IsControl) {
		return errors.New("race name must be 1–80 characters with no control characters or surrounding spaces")
	}
	for _, name := range RaceNames() {
		if strings.EqualFold(name, race.Name) {
			return fmt.Errorf("use a distinct name for a custom race based on %s", name)
		}
	}
	if race.BuildPointCost < 0 || race.BuildPointCost > 1000000 {
		return errors.New("race cost must be between 0 and 1,000,000 Build Points")
	}
	for attribute, modifier := range race.Modifiers {
		if !containsAttribute(attribute) || modifier < -5 || modifier > 100 {
			return fmt.Errorf("invalid racial modifier for %q (allowed range -5 to +100)", attribute)
		}
	}
	return nil
}

func containsAttribute(name string) bool {
	for _, attribute := range attributeKeys {
		if name == attribute {
			return true
		}
	}
	return false
}

func LoadRaceDefinition(path string) (*RaceDefinition, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var race RaceDefinition
	if err := decoder.Decode(&race); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, errors.New("race file must contain exactly one JSON object")
	}
	if err := race.validate(); err != nil {
		return nil, err
	}
	return &race, nil
}
