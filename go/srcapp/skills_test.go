package srcapp

import (
	"bytes"
	"strings"
	"testing"
)

func TestSkillCostsAndFreePools(t *testing.T) {
	ch := NewCharacter()
	ch.Attributes["Body"] = 3
	ch.Attributes["Intelligence"] = 4
	ch.Attributes["Quickness"] = 5
	ch.Skills = []Rated{{Name: "Athletics", Rating: 6}, {Name: "Electronics", Rating: 3, Specialization: "Maglocks"}, {Name: "Corporate politics", Rating: 4, Category: "Knowledge"}, {Name: "English", Rating: 4, Category: "Language"}, {Name: "Japanese", Rating: 4, Category: "Language"}}
	a := ch.SkillAllocation()
	if a.Active != 13 || a.Knowledge != 4 || a.Language != 8 || a.KnowledgeBudget != 20 || a.LanguageBudget != 6 || len(a.Issues) != 0 {
		t.Fatalf("allocation: %+v", a)
	}
	if ch.Skills[1].SpecializationRating() != 5 || ch.Skills[3].ReadWriteRating() != 2 {
		t.Fatal("specialization or literacy")
	}
	ch.Skills = append(ch.Skills, Rated{Name: "Physics", Category: "Knowledge", Rating: 6}, Rated{Name: "Chemistry", Category: "Knowledge", Rating: 6})
	if len(ch.SkillAllocation().Issues) == 0 {
		t.Fatal("language overflow did not consume knowledge allowance")
	}
	ch.Skills = []Rated{{Name: "Car", Rating: 6}}
	if ch.SkillSpent() != 8 {
		t.Fatalf("Reaction-linked skill cost %d", ch.SkillSpent())
	}
}
func TestSelectedInventoryAndCompleteExport(t *testing.T) {
	ch := NewCharacter()
	base := 3.5
	ch.LegacyEssence = &base
	cost := 100
	ch.Cyberware = []Possession{{Name: "First", EssenceCost: 1, UnitCost: &cost}, {Name: "Second", EssenceCost: .5, UnitCost: &cost}}
	if err := ch.ReplacePossession("Cyberware", 0, nil); err != nil {
		t.Fatal(err)
	}
	if ch.Cyberware[0].Name != "Second" || ch.Essence() != 4.5 {
		t.Fatal("wrong item removed or essence not restored")
	}
	ch.MagicGear = []Possession{{Name: "Focus test", Notes: "Focus notes <script>", Source: "Test source", UnitCost: &cost, Quantity: 3}}
	ch.Contacts = []Possession{{Name: "Contact test", Notes: "Contact details"}}
	ch.Karma = 17
	ch.Notes = "Character notes"
	ch.RealName = "Real identity"
	if ch.ResourceAllocation().Spent != 400 {
		t.Fatal("quantity not included in spending")
	}
	var out bytes.Buffer
	if err := ch.ExportHTML(&out); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"Focus test", "Focus notes &lt;script&gt;", "Contact details", "Character notes", "Real identity", "Karma", "17"} {
		if !strings.Contains(out.String(), s) {
			t.Errorf("missing %q", s)
		}
	}
}
