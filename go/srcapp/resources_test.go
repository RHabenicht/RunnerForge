package srcapp

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestContactAndLifestylePurchases(t *testing.T) {
	ch := NewCharacter()
	contact, monthly, free := 5000, 1000, 0
	ch.Contacts = []Possession{{Name: "Fixer", UnitCost: &contact}, {Name: "Free contact", UnitCost: &free}}
	ch.Lifestyles = []Possession{{Name: "Apartment", UnitCost: &monthly, Quantity: 3, Source: "House rules", Notes: "Three prepaid months"}}
	if a := ch.ResourceAllocation(); a.Spent != 8000 || len(a.Issues) != 0 {
		t.Fatalf("allocation: %+v", a)
	}
	path := filepath.Join(t.TempDir(), "character.json")
	if err := SaveJSON(path, ch); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if a := loaded.ResourceAllocation(); a.Spent != 8000 || len(a.Issues) != 0 {
		t.Fatalf("round trip: %+v", a)
	}
	var sheet bytes.Buffer
	if err := loaded.ExportHTML(&sheet); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"Lifestyles", "Apartment", "Three prepaid months", "House rules", "8000"} {
		if !strings.Contains(sheet.String(), s) {
			t.Errorf("sheet missing %q", s)
		}
	}
	if err := loaded.ReplacePossession("Lifestyles", 0, nil); err != nil {
		t.Fatal(err)
	}
	if loaded.ResourceAllocation().Spent != 5000 {
		t.Fatal("removed lifestyle still charged")
	}
}

func TestContactLifestylePriceValidation(t *testing.T) {
	for _, group := range []string{"Contacts", "Lifestyles"} {
		t.Run(group, func(t *testing.T) {
			ch := NewCharacter()
			list := ch.Possessions(group)
			*list = []Possession{{Name: "Unknown"}}
			if a := ch.ResourceAllocation(); len(a.Issues) != 1 || !strings.Contains(a.Issues[0], "Price not set") {
				t.Fatalf("missing price: %+v", a)
			}
			negative := -1
			(*list)[0].UnitCost = &negative
			if a := ch.ResourceAllocation(); len(a.Issues) != 1 || a.Spent != 0 {
				t.Fatalf("negative price: %+v", a)
			}
			price := 600000
			(*list)[0].UnitCost = &price
			(*list)[0].Quantity = 2
			if a := ch.ResourceAllocation(); a.Spent != 1200000 || len(a.Issues) != 1 || !strings.Contains(a.Issues[0], "budget exceeded") {
				t.Fatalf("overspend: %+v", a)
			}
			ch.Career = true
			if a := ch.ResourceAllocation(); len(a.Issues) != 0 {
				t.Fatalf("career creation budget: %+v", a)
			}
			(*list)[0].Quantity = -2
			if a := ch.ResourceAllocation(); len(a.Issues) != 1 || a.Spent != 0 {
				t.Fatalf("negative quantity: %+v", a)
			}
		})
	}
}

func TestLegacyContactLifestyleTotals(t *testing.T) {
	ch, err := ReadLegacy("testdata/archStreetSamurai.sr3")
	if err != nil {
		t.Fatal(err)
	}
	if len(ch.Contacts) != 3 || len(ch.Lifestyles) != 2 {
		t.Fatalf("contacts %d lifestyles %d", len(ch.Contacts), len(ch.Lifestyles))
	}
	if ch.Contacts[2].Name != "Fixer" || ch.Contacts[2].Rating != 1 || *ch.Contacts[2].UnitCost != 5000 {
		t.Fatalf("contact: %+v", ch.Contacts[2])
	}
	for _, p := range ch.Lifestyles {
		if p.Count() != 1 || p.UnitCost == nil || *p.UnitCost != 15000 || p.Raw == "" {
			t.Fatalf("lifestyle: %+v", p)
		}
	}
	if got := ch.ResourceAllocation().Spent; got != 35000 {
		t.Fatalf("saved purchase totals charged incorrectly: %d", got)
	}
	path := filepath.Join(t.TempDir(), "legacy.json")
	if err := SaveJSON(path, ch); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Lifestyles[0].Raw != ch.Lifestyles[0].Raw || loaded.Contacts[2].Raw != ch.Contacts[2].Raw || !bytes.Equal(loaded.LegacyBytes, ch.LegacyBytes) {
		t.Fatal("lost original legacy data")
	}
}

func TestLegacyMissingPurchasePrices(t *testing.T) {
	ch, err := ParseLegacy("[General]\nCharName=Example\nLifeStyle1=Unknown~0\n[Contacts]\nContact0=Unknown~~~~Unknown~1~~~\nContact1=Free~~~~Free~1~~~0\nContact2=Invalid~~~~Invalid~1~~~-5\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(ch.Lifestyles) != 1 || len(ch.Contacts) != 3 {
		t.Fatal("missing entries")
	}
	if ch.Lifestyles[0].UnitCost != nil || ch.Contacts[0].UnitCost != nil || ch.Contacts[2].UnitCost != nil {
		t.Fatal("unresolved prices treated as known")
	}
	if ch.Contacts[1].UnitCost == nil || *ch.Contacts[1].UnitCost != 0 {
		t.Fatal("explicit free contact lost")
	}
	if a := ch.ResourceAllocation(); len(a.Issues) != 3 {
		t.Fatalf("missing review issues: %+v", a)
	}
}
