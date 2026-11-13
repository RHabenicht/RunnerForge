package srcapp

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// Jenny's worked example, Companion printed pp.14 and 34.
func jennyGhoul(t *testing.T) *Character {
	t.Helper()
	ch := NewCharacter()
	ch.Race = "Ork"
	ch.BuildPoints = &BuildPointAllocation{Budget: 120, Resources: 5000, Magic: "Adept"}
	ch.Attributes = map[string]int{"Body": 1, "Quickness": 5, "Strength": 1, "Charisma": 3, "Intelligence": 5, "Willpower": 5}
	if err := ch.ConfigureGhoul(&GhoulCreation{Successes: 1, GMAlternative: true, Sacrifices: []string{"Willpower"}}); err != nil {
		t.Fatal(err)
	}
	return ch
}
func TestGhoulWorkedExample(t *testing.T) {
	ch := jennyGhoul(t)
	for name, want := range map[string]int{"Body": 6, "Quickness": 5, "Strength": 4, "Charisma": 1, "Intelligence": 2, "Willpower": 3} {
		if got := ch.EffectiveAttribute(name); got != want {
			t.Errorf("%s=%d want %d", name, got, want)
		}
	}
	if ch.Essence() != 5 || ch.Magic() != 5 || ch.Reaction() != 3 || ch.GhoulTestDice() != 7 {
		t.Fatal("worked example derived values differ")
	}
	for name, want := range map[string]int{"Body": 11, "Quickness": 6, "Strength": 9, "Charisma": 4, "Intelligence": 4, "Willpower": 6} {
		if got := ch.RacialModifiedLimit(name); got != want {
			t.Errorf("%s limit=%d want %d", name, got, want)
		}
	}
	ch.Skills = []Rated{{Name: "Athletics", Rating: 6, LinkedAttribute: "Body"}}
	a := ch.SkillAllocation()
	if a.Active != 8 || a.KnowledgeBudget != 20 || a.LanguageBudget != 6 {
		t.Fatalf("pre-transformation creation costs/pools: %+v", a)
	}
	cost, _ := ch.buildPointTotal()
	ghoul := ch.Ghoul
	ch.Ghoul = nil
	ordinary, _ := ch.buildPointTotal()
	ch.Ghoul = ghoul
	if cost-ordinary != 10 {
		t.Fatalf("ghoul BP surcharge: %d", cost-ordinary)
	}
	if issues := ch.Validate(); len(issues) != 0 {
		t.Fatalf("worked example review: %v", issues)
	}
}
func TestGhoulInfectionResultsAndLimits(t *testing.T) {
	for successes := 0; successes <= 4; successes++ {
		ch := NewCharacter()
		ch.BuildPoints = &BuildPointAllocation{Budget: 120, Resources: 5000, Magic: "Mundane"}
		for _, name := range attributeKeys {
			ch.Attributes[name] = 5
		}
		if err := ch.ConfigureGhoul(&GhoulCreation{Successes: successes}); err != nil {
			t.Fatal(err)
		}
		wantInt := []int{2, 3, 4, 4, 4}[successes]
		wantCha := []int{2, 3, 3, 4, 4}[successes]
		wantWill := []int{5, 5, 5, 5, 6}[successes]
		if ch.EffectiveAttribute("Intelligence") != wantInt || ch.EffectiveAttribute("Charisma") != wantCha || ch.EffectiveAttribute("Willpower") != wantWill {
			t.Fatalf("wrong result for %d successes", successes)
		}
		if ch.RacialModifiedLimit("Willpower") != 6 || ch.RacialModifiedLimit("Intelligence") != 5 {
			t.Fatal("test result changed racial maximum")
		}
		if ch.Magic() != 0 {
			t.Fatal("infection awakened mundane")
		}
	}
}
func TestGhoulSubstitutionsAndCareer(t *testing.T) {
	ch := jennyGhoul(t)
	ch.Ghoul.GMAlternative = false
	ch.Ghoul.Sacrifices = nil
	if len(ch.GhoulIssues()) == 0 || ch.EffectiveAttribute("Charisma") != 0 {
		t.Fatal("silently rescued NPC result")
	}
	if err := ch.BeginCareer(); err == nil {
		t.Fatal("NPC result entered career")
	}
	ch.Ghoul.GMAlternative = true
	if len(ch.GhoulIssues()) == 0 {
		t.Fatal("missing sacrifice accepted")
	}
	ch.Ghoul.Sacrifices = []string{"Willpower"}
	if err := ch.BeginCareer(); err != nil {
		t.Fatal(err)
	}
	before := append([]byte(nil), ch.CreationSnapshot...)
	ch.Karma = 100
	if err := ch.AdvanceAttribute("Charisma"); err != nil {
		t.Fatal(err)
	}
	if ch.EffectiveAttribute("Charisma") != 2 || ch.Karma != 96 {
		t.Fatal("prevented loss swallowed advancement")
	}
	if !bytes.Equal(before, ch.CreationSnapshot) {
		t.Fatal("advancement altered creation snapshot")
	}
	var creation Character
	if err := json.Unmarshal(ch.CreationSnapshot, &creation); err != nil {
		t.Fatal(err)
	}
	if creation.EffectiveAttribute("Charisma") != 1 || creation.Career {
		t.Fatal("wrong creation snapshot")
	}
	if err := ch.ConfigureGhoul(nil); err == nil {
		t.Fatal("career ghoul choices unlocked")
	}
	if issues := ch.GhoulIssues(); len(issues) != 0 {
		t.Fatalf("career substitution: %v", issues)
	}
	path := filepath.Join(t.TempDir(), "ghoul.json")
	if err := SaveJSON(path, ch); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.EffectiveAttribute("Charisma") != 2 || loaded.Ghoul.PreventedLoss["Charisma"] != 1 {
		t.Fatal("lost preserved substitution")
	}
	var sheet bytes.Buffer
	if err := loaded.ExportHTML(&sheet); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sheet.String(), "Starting ghoul") || !strings.Contains(sheet.String(), "Companion pp.33–34") {
		t.Fatal("missing ghoul export")
	}
}
func TestGhoulEssenceAndEligibility(t *testing.T) {
	ch := jennyGhoul(t)
	ch.Cyberware = []Possession{{Name: "Alpha implant", EssenceCost: 1, Grade: "Alpha"}}
	if ch.Essence() != 3.4 || ch.Magic() != 3 {
		t.Fatalf("post-transformation alpha implant: Essence %v Magic %d", ch.Essence(), ch.Magic())
	}
	var sheet bytes.Buffer
	if err := ch.ExportHTML(&sheet); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sheet.String(), "Essence cost: 1.6") {
		t.Fatal("sheet hides doubled Essence")
	}
	ch.Cyberware[0].EssenceCost = 4
	if ch.Essence() != 0 || len(ch.GhoulIssues()) == 0 {
		t.Fatal("lethal Essence result not flagged")
	}
	ordinary := NewCharacter()
	if err := ordinary.ConfigureGhoul(&GhoulCreation{Successes: 1}); err == nil {
		t.Fatal("Priority ghoul accepted")
	}
	ordinary.BuildPoints = &BuildPointAllocation{Budget: 120, Resources: 5000, Magic: "Mundane"}
	if err := ordinary.ConfigureGhoul(&GhoulCreation{Successes: 4}); err == nil {
		t.Fatal("more successes than test dice")
	}
	legacy := 6.0
	ordinary.LegacyEssence = &legacy
	if err := ordinary.ConfigureGhoul(&GhoulCreation{Successes: 1}); err == nil {
		t.Fatal("import override silently reused")
	}
}
