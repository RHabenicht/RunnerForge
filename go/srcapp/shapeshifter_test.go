package srcapp

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func makeShifter(t *testing.T, species string) *Character {
	t.Helper()
	ch := NewCharacter()
	ch.BuildPoints = &BuildPointAllocation{Budget: 120, Resources: 5000, Magic: "Mundane"}
	ch.Attributes = map[string]int{"Body": 2, "Quickness": 3, "Strength": 4, "Charisma": 3, "Intelligence": 3, "Willpower": 3}
	s := &ShapeshifterCreation{Species: species, Form: "Human", Movement: "Ground", AnimalAttributes: map[string]int{"Body": 2, "Quickness": 3, "Strength": 4}}
	if species == "Seal" {
		s.SealLimits = "Modifiers"
	}
	if err := ch.ConfigureShapeshifter(s); err != nil {
		t.Fatal(err)
	}
	return ch
}
func TestShapeshifterSpeciesForms(t *testing.T) {
	// Independently transcribed final values for purchases 2/3/4, from p.36.
	for _, tc := range []struct {
		name                     string
		human, animal            [3]int
		intelligence, animalDice int
	}{
		{"Bear", [3]int{4, 3, 5}, [3]int{6, 4, 8}, 3, 1},
		{"Eagle", [3]int{2, 4, 4}, [3]int{2, 5, 6}, 4, 1},
		{"Fox", [3]int{2, 3, 4}, [3]int{2, 4, 4}, 4, 2},
		{"Leopard", [3]int{3, 3, 5}, [3]int{4, 4, 6}, 3, 3},
		{"Seal", [3]int{3, 3, 5}, [3]int{4, 4, 5}, 3, 1},
		{"Tiger", [3]int{3, 4, 5}, [3]int{5, 5, 7}, 3, 3},
		{"Wolf", [3]int{3, 3, 5}, [3]int{4, 4, 5}, 3, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ch := makeShifter(t, tc.name)
			for i, name := range []string{"Body", "Quickness", "Strength"} {
				if ch.FormAttribute(name, "Human") != tc.human[i] || ch.FormAttribute(name, "Animal") != tc.animal[i] {
					t.Fatalf("%s form values wrong", name)
				}
			}
			if ch.FormAttribute("Intelligence", "Human") != tc.intelligence || ch.FormAttribute("Intelligence", "Animal") != tc.intelligence {
				t.Fatal("mental attributes not shared")
			}
			if ch.AttributeSpent() != 27 {
				t.Fatalf("independent physical purchases missing: %d", ch.AttributeSpent())
			}
			bp, issues := ch.buildPointTotal()
			if bp != 79 || len(issues) != 0 {
				t.Fatalf("BP %d, issues %v", bp, issues)
			}
			if issues := ch.Validate(); len(issues) != 0 {
				t.Fatalf("species validation: %v", issues)
			}
			humanSkill := ch.SkillAllocation()
			ch.Shapeshifter.Form = "Animal"
			if ch.Reaction() != (tc.animal[1]+tc.intelligence)/2+1 || ch.InitiativeDice() != tc.animalDice {
				t.Fatal("animal reaction/initiative")
			}
			animalSkill := ch.SkillAllocation()
			if humanSkill.KnowledgeBudget != tc.intelligence*3 || humanSkill.LanguageBudget != tc.intelligence || animalSkill.KnowledgeBudget != humanSkill.KnowledgeBudget {
				t.Fatal("special free skill pools")
			}
			if ch.Essence() != 8 || ch.Magic() != 0 {
				t.Fatal("mundane Essence/Magic")
			}
		})
	}
}
func TestShapeshifterSkillsUseHumanForm(t *testing.T) {
	ch := makeShifter(t, "Bear")
	ch.Skills = []Rated{{Name: "Athletics", Rating: 6, LinkedAttribute: "Body"}}
	ch.Shapeshifter.Form = "Animal"
	if ch.SkillAllocation().Active != 8 {
		t.Fatal("animal strength reduced creation skill cost")
	}
	if err := ch.BeginCareer(); err != nil {
		t.Fatal(err)
	}
	ch.Karma = 100
	if err := ch.AdvanceSkill(0, false); err != nil {
		t.Fatal(err)
	}
	// Rating 7 is above human Body 4 and below twice Body: 7 * 2 = 14.
	if ch.Karma != 86 {
		t.Fatalf("skill advancement did not use human attribute: %d", ch.Karma)
	}
	if err := ch.AdvanceAttribute("Body"); err != nil {
		t.Fatal(err)
	}
	if ch.Shapeshifter.AnimalAttributes["Body"] != 3 || ch.Attributes["Body"] != 2 || ch.Advancements[1].Name != "Animal Body" {
		t.Fatal("animal advancement changed human purchase")
	}
	saved := append([]byte(nil), ch.CreationSnapshot...)
	s := *ch.Shapeshifter
	s.Form = "Human"
	if err := ch.ConfigureShapeshifter(&s); err != nil {
		t.Fatal(err)
	}
	if err := ch.AdvanceAttribute("Body"); err != nil {
		t.Fatal(err)
	}
	if ch.Attributes["Body"] != 3 || ch.Shapeshifter.AnimalAttributes["Body"] != 3 || ch.Advancements[2].Name != "Human Body" {
		t.Fatal("human advancement changed animal purchase")
	}
	if !bytes.Equal(saved, ch.CreationSnapshot) {
		t.Fatal("form changes altered snapshot")
	}
}
func TestShapeshifterConditionalInitiativeAndMagic(t *testing.T) {
	for _, tc := range []struct{ species, movement string }{{"Eagle", "Flying"}, {"Seal", "Swimming"}} {
		ch := makeShifter(t, tc.species)
		ch.BuildPoints.Magic = "Adept"
		if ch.Magic() != 6 {
			t.Fatal("Essence 8 incorrectly granted Magic 8")
		}
		s := *ch.Shapeshifter
		s.Form = "Animal"
		s.Movement = tc.movement
		if err := ch.ConfigureShapeshifter(&s); err != nil {
			t.Fatal(err)
		}
		if ch.InitiativeDice() != 2 || ch.AdeptPowersActive() {
			t.Fatal("conditional initiative or animal adept powers")
		}
		s.Movement = "Ground"
		if err := ch.ConfigureShapeshifter(&s); err != nil {
			t.Fatal(err)
		}
		if ch.InitiativeDice() != 1 {
			t.Fatal("conditional bonus applied on ground")
		}
	}
}
func TestShapeshifterSealConflictAndEligibility(t *testing.T) {
	ch := makeShifter(t, "Seal")
	s := *ch.Shapeshifter
	s.SealLimits = ""
	if err := ch.ConfigureShapeshifter(&s); err == nil {
		t.Fatal("silently guessed seal limits")
	}
	s.SealLimits = "Limits"
	s.Form = "Animal"
	if err := ch.ConfigureShapeshifter(&s); err != nil {
		t.Fatal(err)
	}
	if ch.RacialModifiedLimit("Body") != 7 || ch.RacialModifiedLimit("Quickness") != 6 {
		t.Fatal("printed p.37 limits")
	}
	s.SealLimits = "Modifiers"
	if err := ch.ConfigureShapeshifter(&s); err != nil {
		t.Fatal(err)
	}
	if ch.RacialModifiedLimit("Body") != 8 || ch.RacialModifiedLimit("Quickness") != 7 {
		t.Fatal("p.36 derived limits")
	}
	price := 0
	ch.Contacts = []Possession{{Name: "Free contact", UnitCost: &price}}
	ch.Cyberware = []Possession{{Name: "Implant", EssenceCost: .1, UnitCost: &price}}
	ch.BuildPoints.Resources = 20000
	issues := strings.Join(ch.ShapeshifterIssues(), " ")
	for _, s := range []string{"¥5,000", "purchase all", "cannot receive cyberware"} {
		if !strings.Contains(issues, s) {
			t.Errorf("missing %q", s)
		}
	}
	if ch.ResourceBudget() != 5000 {
		t.Fatal("bought extra shapeshifter resources")
	}
	if err := ch.ConfigureGhoul(&GhoulCreation{Successes: 3}); err == nil {
		t.Fatal("combined special types")
	}
}
func TestShapeshifterPriorityPersistenceAndExport(t *testing.T) {
	ch := makeShifter(t, "Fox")
	ch.BuildPoints = nil
	ch.Priority = map[string]string{"Resources": "A", "Attributes": "B", "Race": "C", "Skills": "D", "Magic": "E"}
	if issues := ch.Validate(); len(issues) != 0 {
		t.Fatalf("Priority shapeshifter: %v", issues)
	}
	ch.Priority = map[string]string{"Resources": "B", "Attributes": "D", "Race": "C", "Skills": "E", "Magic": "A"}
	ch.MagicType = "Full Magician"
	ch.Attributes["Strength"] = 1
	ch.Attributes["Charisma"] = 1
	ch.Attributes["Willpower"] = 2
	if issues := ch.Validate(); len(issues) != 0 {
		t.Fatalf("Priority magician shapeshifter: %v", issues)
	}
	if ch.Magic() != 6 || ch.ResourceBudget() != 5000 {
		t.Fatal("Priority resources/magic")
	}
	path := filepath.Join(t.TempDir(), "shifter.json")
	if err := SaveJSON(path, ch); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.FormAttribute("Quickness", "Animal") != 4 || loaded.FormAttribute("Quickness", "Human") != 3 {
		t.Fatal("lost independent form purchases")
	}
	var out bytes.Buffer
	if err := loaded.ExportHTML(&out); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"Fox shapeshifter", "Human: Body 2", "Animal: Body 2", "Companion pp.13–14,34–37"} {
		if !strings.Contains(out.String(), s) {
			t.Errorf("sheet missing %q", s)
		}
	}
}
