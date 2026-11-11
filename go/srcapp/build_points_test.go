package srcapp

import (
	"path/filepath"
	"testing"
)

func TestBuildPointAllocation(t *testing.T) {
	ch := NewCharacter()
	ch.Race = "Elf"
	ch.Attributes["Body"] = 3
	ch.Skills = []Rated{{Name: "Athletics", Rating: 6}}
	ch.BuildPoints = &BuildPointAllocation{Budget: 55, Resources: 500, Magic: "Adept"}
	// Elf 10, adept 25, eight attribute points 16, Athletics 9, resources -5.
	if total, issues := ch.buildPointTotal(); total != 55 || len(issues) != 0 {
		t.Fatalf("cost %d, issues %v", total, issues)
	}
	if !ch.Awakened() || ch.ResourceBudget() != 500 {
		t.Fatal("Priority selections overrode Build Point allocation")
	}
	ch.BuildPoints.Budget--
	if _, issues := ch.buildPointTotal(); len(issues) == 0 {
		t.Fatal("accepted an overspent allocation")
	}
	path := filepath.Join(t.TempDir(), "character.json")
	if err := SaveJSON(path, ch); err != nil {
		t.Fatal(err)
	}
	saved, err := LoadJSON(path)
	if err != nil || saved.BuildPoints == nil || *saved.BuildPoints != *ch.BuildPoints {
		t.Fatalf("allocation failed to round-trip: %v", err)
	}
}
