package srcapp

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestSupplementPurchasesAndSave(t *testing.T) {
	cat, err := LoadCatalogue()
	if err != nil {
		t.Fatal(err)
	}
	ch := NewCharacter()
	purchases := []struct {
		name, kind string
		cost       int
	}{
		{"Mazda Dust Devil", "vehicles", 1500},
		{"Mitsuhama Butlerbot", "vehicles", 25000},
		{"Electronic Paper (standard sheet, 20 Mp)", "gear", 2},
		{"Electronic Paper (tabloid, 50 Mp)", "gear", 5},
		{"ColorChange Paint (1 liter)", "gear", 750},
		{"Paint Controller Unit", "gear", 2000},
		{"Counterfeiting Utility Rating [6]", "cyberdecks", 72000},
	}
	total := 0
	for _, buy := range purchases {
		matches := cat.Search(buy.name, buy.kind, 100)
		if len(matches) != 1 || matches[0].Name != buy.name {
			t.Fatalf("search %s: %v", buy.name, matches)
		}
		ch.AddItem(matches[0])
		total += buy.cost
	}
	if got := ch.ResourceAllocation(); got.Spent != total || len(got.Issues) != 0 {
		t.Fatalf("purchase total: %+v want %d", got, total)
	}
	if ch.Decks[0].Rating != 6 {
		t.Fatal("utility rating was lost")
	}
	if !strings.Contains(ch.Vehicles[1].Raw, "2 (4)/2") || !strings.Contains(ch.Vehicles[1].Raw, "75 PF") {
		t.Fatal("robot pilot or fuel data missing")
	}
	ch.Gear[2].Quantity = 2
	if ch.ResourceAllocation().Spent != total+750 {
		t.Fatal("paint quantity was not charged per liter")
	}
	path := filepath.Join(t.TempDir(), "equipment.json")
	if err := SaveJSON(path, ch); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(ch)
	after, _ := json.Marshal(loaded)
	if string(before) != string(after) {
		t.Fatal("equipment changed across save/load")
	}
}

func TestSupplementUnknownPricesRemainExplicit(t *testing.T) {
	cat, err := LoadCatalogue()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Altyerre", "Kuman-Nhepa", "Wudu'aku", "Churinga (unenchanted)", "Telecom Drone-Control Package", "Credstick Beacon Signaler"} {
		items := cat.Search(name, "all", 100)
		if len(items) != 1 {
			t.Fatalf("missing or duplicate %s", name)
		}
		ch := NewCharacter()
		ch.AddItem(items[0])
		allocation := ch.ResourceAllocation()
		if len(allocation.Issues) != 1 || !strings.Contains(allocation.Issues[0], "Price not set") {
			t.Fatalf("%s silently priced: %+v", name, allocation)
		}
	}
}

func TestSupplementCoverageAndLegacyIdentifiers(t *testing.T) {
	cat, err := LoadCatalogue()
	if err != nil {
		t.Fatal(err)
	}
	ids := map[int]Item{}
	for _, item := range cat.Entries {
		ids[item.ID] = item
	}
	for id := 5526; id <= 5536; id++ {
		if ids[id].Kind != "vehicles" {
			t.Fatalf("missing drone %d", id)
		}
	}
	if ids[5530].Name != "VOR Robotics House Helper" {
		t.Fatal("corrected House Helper name missing")
	}
	for rating := 1; rating <= 10; rating++ {
		item := ids[6539+rating]
		ch := NewCharacter()
		ch.AddItem(item)
		multiplier := 100
		if rating >= 4 {
			multiplier = 200
		}
		if rating >= 7 {
			multiplier = 500
		}
		if rating >= 10 {
			multiplier = 1000
		}
		if len(ch.Decks) != 1 || ch.Decks[0].UnitCost == nil || *ch.Decks[0].UnitCost != rating*rating*10*multiplier {
			t.Fatalf("incorrect utility price at rating %d", rating)
		}
	}
	if len(cat.Entries) != 6484 {
		t.Fatalf("unexpected active count: %d", len(cat.Entries))
	}
	// A future importer must not silently double-add a curated legacy identifier.
	if err := cat.loadEquipmentSupplement(); err == nil {
		t.Fatal("duplicate supplement accepted")
	}
}
