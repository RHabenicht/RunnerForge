package srcapp

import "fmt"

var PossessionGroups = []string{"Gear", "Cyberware", "Bioware", "Spells", "Adept powers", "Vehicles", "Cyberdecks", "Edges / Flaws", "Contacts", "Totems", "Magic gear"}

func (ch *Character) Possessions(group string) *[]Possession {
	switch group {
	case "Gear":
		return &ch.Gear
	case "Cyberware":
		return &ch.Cyberware
	case "Bioware":
		return &ch.Bioware
	case "Spells":
		return &ch.Spells
	case "Adept powers":
		return &ch.AdeptPowers
	case "Vehicles":
		return &ch.Vehicles
	case "Cyberdecks":
		return &ch.Decks
	case "Edges / Flaws":
		return &ch.EdgesFlaws
	case "Contacts":
		return &ch.Contacts
	case "Totems":
		return &ch.Totems
	case "Magic gear":
		return &ch.MagicGear
	}
	return nil
}
func (p Possession) Count() int {
	if p.Quantity == 0 {
		return 1
	}
	return p.Quantity
}
func (ch *Character) ReplacePossession(group string, index int, p *Possession) error {
	list := ch.Possessions(group)
	if list == nil || index < 0 || index >= len(*list) {
		return fmt.Errorf("no selected item")
	}
	old := (*list)[index]
	if p == nil {
		*list = append((*list)[:index], (*list)[index+1:]...)
	} else {
		(*list)[index] = *p
	}
	if group == "Cyberware" && ch.LegacyEssence != nil {
		change := old.ChargedEssence()
		if p != nil {
			change -= p.ChargedEssence()
		}
		value := *ch.LegacyEssence + change
		ch.LegacyEssence = &value
		if change != 0 && ch.LegacyMagic != nil {
			ch.LegacyWarnings = append(ch.LegacyWarnings, "Cyberware changed: imported Magic requires review.")
		}
	}
	return nil
}

type ResourceAllocation struct {
	Spent  int
	Issues []string
}

func (ch *Character) ResourceAllocation() ResourceAllocation {
	var result ResourceAllocation
	if ch.SpellPointsPurchased > 0 && ch.SpellPointsPurchased <= 50 {
		result.Spent += 25000 * ch.SpellPointsPurchased
	}
	for _, group := range []string{"Gear", "Cyberware", "Bioware", "Vehicles", "Cyberdecks", "Magic gear"} {
		for _, p := range *ch.Possessions(group) {
			if p.UnitCost == nil {
				result.Issues = append(result.Issues, "Price not set: "+p.Name)
				continue
			}
			if *p.UnitCost < 0 || *p.UnitCost > 1000000000 || p.Count() < 1 || p.Count() > 1000000 {
				result.Issues = append(result.Issues, "Invalid price or quantity: "+p.Name)
				continue
			}
			result.Spent += p.ChargedUnitCost() * p.Count()
		}
	}
	if result.Spent > ch.ResourceBudget() {
		result.Issues = append(result.Issues, fmt.Sprintf("Resource budget exceeded (¥%d / ¥%d).", result.Spent, ch.ResourceBudget()))
	}
	return result
}
