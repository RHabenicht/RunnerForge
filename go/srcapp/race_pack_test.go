package srcapp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNightLifePack(t *testing.T) {
	p, err := LoadRacePack("../../race-packs/NightLife-Rebuilt-Races.dat")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Races) != 16 {
		t.Fatal(len(p.Races))
	}
	costs := map[string]int{"Halfling": 5, "Cyclops": 10, "Koborkuru": 5, "Formori": 10, "Menehene": 5, "Hobgoblin": 5, "Giant": 10, "Gnome": 5, "Oni": 5, "Wakyambi": 10, "Ogre": 5, "Minotaur": 10, "Satyr": 5, "Night Ones": 10, "Dryad": 10, "Sasquatch": 15}
	for _, r := range p.Races {
		name := strings.TrimSuffix(r.Name, " (NightLife)")
		if cost, ok := costs[name]; !ok || cost != r.BuildPointCost {
			t.Fatalf("surcharge or bad cost: %+v", r)
		}
		want := "C"
		if r.BuildPointCost == 5 {
			want = "D"
		}
		if name == "Sasquatch" {
			want = "B"
		}
		if r.Priority != want {
			t.Fatal(r)
		}
		ch := NewCharacter()
		for k := range ch.Attributes {
			ch.Attributes[k] = 3
		}
		if err := ch.SelectPackRace(r); err != nil {
			t.Fatal(err)
		}
		for k := range ch.Attributes {
			if ch.EffectiveAttribute(k) != 3+r.Modifiers[k] {
				t.Fatalf("modifier %s %s", name, k)
			}
		}
		ch.Priority["Race"] = want
		for _, issue := range ch.Validate() {
			if strings.Contains(issue, "requires Race") || strings.Contains(issue, "Choose Build Points") {
				t.Fatal(issue)
			}
		}
		ch.Priority["Race"] = "E"
		if !strings.Contains(strings.Join(ch.Validate(), ";"), "requires Race priority "+want) {
			t.Fatal("priority not enforced")
		}
		f := filepath.Join(t.TempDir(), "character.json")
		if err := SaveJSON(f, ch); err != nil {
			t.Fatal(err)
		}
		loaded, err := LoadJSON(f)
		if err != nil {
			t.Fatal(err)
		}
		if loaded.CustomRace.Traits != r.Traits || loaded.CustomRace.Priority != want {
			t.Fatal("saved custom definition lost")
		}
		ch.CustomRace.Modifiers["Body"] = 99
		if r.Modifiers["Body"] == 99 {
			t.Fatal("pack modified through character")
		}
	}
	// Built-ins stay unchanged after loading a pack.
	ch := NewCharacter()
	ch.Race = "Cyclops"
	r, _ := ch.raceDefinition()
	if r.BuildPointCost != 15 || r.Modifiers["Willpower"] != 0 {
		t.Fatal("default Cyclops changed")
	}
}
func TestRejectInvalidRacePack(t *testing.T) {
	for _, data := range []string{`{"format":"wrong","name":"x","races":[]}`, `{"format":"runnerforge.race-pack.v1","name":"x","races":[{"name":"Test","priority":"Z"}]}`, `{"format":"runnerforge.race-pack.v1","name":"x","races":[{"name":"Test"},{"name":"test"}]}`} {
		f := filepath.Join(t.TempDir(), "bad.dat")
		os.WriteFile(f, []byte(data), 0600)
		if _, err := LoadRacePack(f); err == nil {
			t.Fatal("invalid pack accepted")
		}
	}
}
