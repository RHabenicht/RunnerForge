package srcapp

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompanionVariantsUsePublishedCostsAndModifiers(t *testing.T) {
	ch := NewCharacter()
	ch.BuildPoints = &BuildPointAllocation{Budget: 120, Magic: "Mundane", Resources: 5000}
	for _, tc := range []struct {
		name, attribute string
		cost, modifier  int
	}{
		{"Cyclops", "Strength", 15, 6},
		{"Dryad", "Charisma", 15, 3},
		{"Koborokuru", "Willpower", 10, 1},
		{"Oni", "Willpower", 10, 1},
		{"Hobgoblin", "Intelligence", 10, 0},
	} {
		ch.Race = tc.name
		points, _ := ch.buildPointTotal()
		if points != tc.cost+12 || ch.EffectiveAttribute(tc.attribute) != 1+tc.modifier {
			t.Errorf("%s has incorrect cost or attribute modifier", tc.name)
		}
	}
}

func TestCustomRaceRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "variant.json")
	definition := `{"name":"Night One (house rules)","build_point_cost":10,"modifiers":{"Quickness":2,"Charisma":2},"traits":"Mild sunlight allergy"}`
	if err := os.WriteFile(path, []byte(definition), 0600); err != nil {
		t.Fatal(err)
	}
	race, err := LoadRaceDefinition(path)
	if err != nil {
		t.Fatal(err)
	}
	ch := NewCharacter()
	ch.Race, ch.CustomRace = race.Name, race
	ch.BuildPoints = &BuildPointAllocation{Budget: 120, Magic: "Mundane", Resources: 5000}
	save := filepath.Join(dir, "character.json")
	if err := SaveJSON(save, ch); err != nil {
		t.Fatal(err)
	}
	ch, err = LoadJSON(save)
	if err != nil {
		t.Fatal(err)
	}
	points, issues := ch.buildPointTotal()
	if points != 22 || len(issues) != 0 || ch.EffectiveAttribute("Quickness") != 3 {
		t.Fatalf("custom race was not applied: %d BP, %v", points, issues)
	}
	var output bytes.Buffer
	if err := ch.ExportHTML(&output); err != nil || !strings.Contains(output.String(), "Mild sunlight allergy") {
		t.Fatalf("racial traits missing from export: %v", err)
	}
}

func TestRaceTemplateAndExport(t *testing.T) {
	ch := NewCharacter()
	ch.Race = "Night One"
	race := ch.RaceTemplate()
	race.Name = "Night One (rebuilt)"
	race.BuildPointCost = 10
	race.Modifiers["Quickness"] = 3
	if ch.EffectiveAttribute("Quickness") != 3 {
		t.Fatal("editing a template changed the built-in race")
	}
	path := filepath.Join(t.TempDir(), "race.json")
	if err := SaveRaceDefinition(path, race); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadRaceDefinition(path)
	if err != nil || loaded.BuildPointCost != 10 || loaded.Modifiers["Quickness"] != 3 {
		t.Fatalf("export round trip: %v, %v", loaded, err)
	}
	race.Name = "Elf"
	if err := SaveRaceDefinition(path, race); err == nil {
		t.Fatal("accepted built-in collision")
	}
	loaded, err = LoadRaceDefinition(path)
	if err != nil || loaded.Name != "Night One (rebuilt)" {
		t.Fatal("invalid export damaged existing race")
	}
}
