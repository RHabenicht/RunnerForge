package srcapp

import (
	"bytes"
	"path/filepath"
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

func TestMagicAndAugmentationBudgets(t *testing.T) {
	ch := NewCharacter()
	ch.BuildPoints = &BuildPointAllocation{Budget: 120, Resources: 90000, Magic: "Full Magician"}
	ch.SpellPointsPurchased = 2
	ch.SpellPointsSold = 5
	if ch.MagicAllocation().SpellBudget != 22 || ch.ResourceAllocation().Spent != 50000 {
		t.Fatal("spell purchase/sale accounting")
	}
	index := 1.0
	cost := 10000
	ch.Cyberware = []Possession{{Name: "Test", EssenceCost: .25, UnitCost: &cost, Grade: "Alpha"}}
	ch.Bioware = []Possession{{Name: "Test bio", BioIndex: &index, UnitCost: &cost}}
	if ch.Essence() != 5.8 || ch.Magic() != 5 || ch.ResourceAllocation().Spent != 80000 {
		t.Fatalf("augmentation calculation: Essence %v Magic %d resources %d", ch.Essence(), ch.Magic(), ch.ResourceAllocation().Spent)
	}
	edge, flaw := 2.0, -1.0
	ch.EdgesFlaws = []Possession{{Name: "E", PointCost: &edge}, {Name: "F", PointCost: &flaw}}
	total, _ := ch.buildPointTotal()
	// 12 attributes +30 magic +10 resources +1 net Edge/Flaw -1 sold spell points.
	if total != 52 {
		t.Fatalf("BP total %d", total)
	}
	ch.Bioware[0].BioIndex = nil
	if len(ch.AugmentationIssues()) == 0 {
		t.Fatal("missing bio index was silently accepted")
	}
}

func TestAdvancementKeepsSpecializationIndependent(t *testing.T) {
	ch := NewCharacter()
	ch.Career = true
	ch.Karma = 30
	ch.Attributes["Quickness"] = 4
	ch.Skills = []Rated{{Name: "Pistols", Rating: 3, Specialization: "Beretta"}}
	if err := ch.AdvanceSkill(0, false); err != nil {
		t.Fatal(err)
	}
	if ch.Karma != 24 || ch.Skills[0].Rating != 4 || ch.Skills[0].SpecializationRating() != 5 {
		t.Fatal("base advancement changed specialization or wrong cost")
	}
	if err := ch.AdvanceSkill(0, true); err != nil {
		t.Fatal(err)
	}
	if ch.Karma != 18 || ch.Skills[0].SpecializationRating() != 6 {
		t.Fatal("specialization cost")
	}
	ch.Attributes["Strength"] = 6
	if err := ch.AdvanceAttribute("Strength"); err == nil {
		t.Fatal("spent more karma than available")
	}
	ch.Karma = 21
	if err := ch.AdvanceAttribute("Strength"); err != nil {
		t.Fatal(err)
	}
	if ch.Karma != 0 || ch.Attributes["Strength"] != 7 {
		t.Fatal("attribute advancement above racial modified limit")
	}
}

func TestCareerSnapshotIsStable(t *testing.T) {
	ch := NewCharacter()
	ch.Karma = 20
	if err := ch.BeginCareer(); err != nil {
		t.Fatal(err)
	}
	snapshot := string(ch.CreationSnapshot)
	if err := ch.AdvanceAttribute("Body"); err != nil {
		t.Fatal(err)
	}
	if string(ch.CreationSnapshot) != snapshot {
		t.Fatal("advancement rewrote creation record")
	}
	if err := ch.BeginCareer(); err != nil {
		t.Fatal(err)
	}
	if string(ch.CreationSnapshot) != snapshot {
		t.Fatal("second begin career rewrote creation record")
	}
	path := filepath.Join(t.TempDir(), "career.json")
	if err := SaveJSON(path, ch); err != nil {
		t.Fatal(err)
	}
	saved, err := LoadJSON(path)
	if err != nil || !saved.Career || len(saved.Advancements) != 1 || string(saved.CreationSnapshot) == "" {
		t.Fatalf("career round trip: %v", err)
	}
}

func TestLearningMultipleSpecializations(t *testing.T) {
	ch := NewCharacter()
	ch.Career = true
	ch.Karma = 10
	ch.Attributes["Quickness"] = 6
	ch.Skills = []Rated{{Name: "Stealth", Rating: 5}}
	if err := ch.LearnOrAdvanceSpecialization(0, "Sneaking"); err != nil {
		t.Fatal(err)
	}
	if err := ch.LearnOrAdvanceSpecialization(0, "Awareness"); err != nil {
		t.Fatal(err)
	}
	if ch.Karma != 4 || ch.Skills[0].Rating != 5 || len(ch.Skills[0].AdditionalSpecializations) != 2 {
		t.Fatal("new specializations changed base or had incorrect costs")
	}
	if err := ch.LearnOrAdvanceSpecialization(0, "Sneaking"); err == nil {
		t.Fatal("accepted advancement without enough karma")
	}
}
