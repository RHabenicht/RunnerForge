package srcapp

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogFilteredAndSearch(t *testing.T) {
	c, e := LoadCatalogue()
	if e != nil {
		t.Fatal(e)
	}

	approved := map[string]bool{"sr3": true, "src": true, "cc": true, "mits": true, "mm": true, "mat": true, "sota": true, "sta2": true, "twl": true, "yotc": true, "ssg": true, "tal": true}
	for _, book := range c.Books {
		if book.Allowed != approved[book.Code] {
			t.Fatalf("incorrect approval for %s", book.Code)
		}
	}
	ids := map[int]bool{}
	for _, it := range c.Entries {
		if it.ID <= 0 || ids[it.ID] {
			t.Fatalf("invalid or duplicated ID %d", it.ID)
		}
		ids[it.ID] = true
		if len(it.Sources) == 0 {
			t.Fatalf("missing provenance: %s", it.Name)
		}
		for _, source := range it.Sources {
			if !approved[source] {
				t.Fatalf("unapproved source %s on %s", source, it.Name)
			}
		}
	}

	if len(c.Search("Datajack", "cyberware", 100)) < 1 {
		t.Fatal("datajack not found")
	}
	if len(c.Search("", "contacts", 999)) != 0 {
		t.Fatal("unattributed contacts escaped quarantine")
	}
}
func TestPriorityRacialAndModifiers(t *testing.T) {
	c := NewCharacter()
	if len(c.Validate()) != 0 {
		t.Fatal(c.Validate())
	}
	if c.AttributeBudget() != 27 || c.SkillBudget() != 34 || c.ResourceBudget() != 1000000 {
		t.Fatal("priority matrix")
	}
	c.Priority["Race"] = "D"
	c.Priority["Magic"] = "E"
	c.Race = "Ork"
	if c.EffectiveAttribute("Body") != 4 || c.EffectiveAttribute("Charisma") != 0 {
		t.Fatal("racial modifiers not applied")
	}
	if c.Reaction() != 0 {
		t.Fatal("expected floor(Quickness + Intelligence)/2 for Ork starting values")
	}
	c.Attributes["Quickness"] = 4
	c.Attributes["Intelligence"] = 4
	c.Attributes["Willpower"] = 3
	if c.CombatPool() != (4+3+3)/2 {
		t.Fatal("combat pool")
	}
	c.Cyberware = append(c.Cyberware, Possession{Name: "Wired Reflexes II", Kind: "cyberware", EssenceCost: 2.4, Modifiers: "+4RCT,+2INI,"})
	if c.CyberReactionMod() != 4 || c.InitiativeDice() != 3 {
		t.Fatal("cyberware reaction/initiative")
	}
	if math.Abs(c.Essence()-3.6) > 0.001 {
		t.Fatal("essence")
	}
	c.Priority["Magic"] = "A"
	c.Priority["Resources"] = "E"
	c.Priority["Attributes"] = "B"
	c.Priority["Skills"] = "C"
	c.Priority["Race"] = "D"
	if c.Magic() != 3 {
		t.Fatalf("magic %d", c.Magic())
	}
	c.Priority["Race"] = "E"
	if len(c.Validate()) == 0 {
		t.Fatal("expected validation duplicate or invalid metatype")
	}
}
func TestLegacyAllArchetypesSaveAndHTML(t *testing.T) {
	files, e := filepath.Glob("testdata/arch*.sr3")
	if e != nil || len(files) != 17 {
		t.Fatalf("expected 17 archetypes, %d %v", len(files), e)
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			b, e := os.ReadFile(f)
			if e != nil {
				t.Fatal(e)
			}
			ch, e := ReadLegacy(f)
			if e != nil {
				t.Fatal(e)
			}
			if len(ch.LegacyBytes) != len(b) || !bytes.Equal(ch.LegacyBytes, b) {
				t.Fatal("legacy byte loss")
			}
			if ch.Edition != 3 {
				t.Fatal("SR3 only")
			}
			if ch.Name == "" {
				t.Fatal("name empty")
			}
			tmp := filepath.Join(t.TempDir(), "character.nsrcg2.json")
			if e = SaveJSON(tmp, ch); e != nil {
				t.Fatal(e)
			}
			x, e := LoadJSON(tmp)
			if e != nil {
				t.Fatal(e)
			}
			if !bytes.Equal(x.LegacyBytes, b) {
				t.Fatal("JSON archival bytes changed")
			}
			var h bytes.Buffer
			if e = x.ExportHTML(&h); e != nil {
				t.Fatal(e)
			}
			if !strings.Contains(h.String(), "Shadowrun Third Edition") || !strings.Contains(h.String(), "<meta charset=\"UTF-8\"") {
				t.Fatal("HTML output missing required fields")
			}
		})
	}
}
func TestSamuraiDerived(t *testing.T) {
	c, e := ReadLegacy("testdata/archStreetSamurai.sr3")
	if e != nil {
		t.Fatal(e)
	}
	if c.Name != "Street Samurai" {
		t.Fatal(c.Name)
	}
	if c.Reaction() != 11 || c.InitiativeDice() != 3 || c.CombatPool() != 7 {
		t.Fatalf("derived wrong: reaction=%d init=%d combat=%d", c.Reaction(), c.InitiativeDice(), c.CombatPool())
	}
	if math.Abs(c.Essence()-0.09) > 0.001 {
		t.Fatal("essence")
	}
}
func TestUnicodeAndHTMLSafety(t *testing.T) {
	c := NewCharacter()
	c.Name = `José Ægir <img src=x onerror=alert(1)>`
	c.Notes = `<script>alert("x")</script> שלום`
	var h bytes.Buffer
	if e := c.ExportHTML(&h); e != nil {
		t.Fatal(e)
	}
	s := h.String()
	if !strings.Contains(s, "José Ægir") || !strings.Contains(s, "&lt;script&gt;") {
		t.Fatal("unicode or escaping failed")
	}
	if strings.Contains(s, "<script>") || strings.Contains(s, "<img src=x") {
		t.Fatal("HTML injection")
	}
	b, e := json.Marshal(c)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Contains(b, []byte("José")) {
		t.Fatal("unicode missing in JSON")
	}
}
func TestAddItemSourceAndLegacyEssence(t *testing.T) {
	c := NewCharacter()
	e := float64(4)
	c.LegacyEssence = &e
	c.AddItem(Item{ID: 50, Name: "Datajack", Kind: "cyberware", Fields: map[string]interface{}{"EssCost": "0.2", "Mods": ""}, Sources: []string{"sr3"}})
	if len(c.Cyberware) != 1 || c.Cyberware[0].Source != "sr3" || math.Abs(c.Essence()-3.8) > 0.001 {
		t.Fatal("cyber add")
	}
	c.AddItem(Item{ID: 22, Name: "Magic Missile", Kind: "spells"})
	if len(c.Spells) != 1 || len(c.Gear) != 0 {
		t.Fatal("spell added to wrong collection")
	}
}
