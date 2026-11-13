//go:build windows

package main

import (
	"fmt"
	"nsrcgii/srcapp"
	"strconv"
	"strings"
)

func loadShapeshifterContents() {
	labels := map[int]string{10: "Species: Bear, Eagle, Fox, Leopard, Seal, Tiger, Wolf / None", 11: "Active form: Human / Animal", 12: "Movement: Ground / Flying / Swimming", 13: "Animal Body (base purchase)", 14: "Animal Quickness (base purchase)", 15: "Companion pp.13–14,34–37", 16: "Animal Strength (base purchase)", 17: "Seal limits choice"}
	for id, label := range labels {
		set(contentsLabels[id], label)
		procShow.Call(contentsLabels[id], 5)
		procShow.Call(contentsFields[id], 5)
	}
	send(contentsFields[15], 0x00CF, 1, 0)
	s := app.char.Shapeshifter
	if s == nil {
		s = &srcapp.ShapeshifterCreation{Species: "None", Form: "Human", Movement: "Ground", AnimalAttributes: map[string]int{"Body": 1, "Quickness": 1, "Strength": 1}}
	}
	set(contentsFields[10], s.Species)
	set(contentsFields[11], s.Form)
	set(contentsFields[12], s.Movement)
	for id, name := range map[int]string{13: "Body", 14: "Quickness", 16: "Strength"} {
		set(contentsFields[id], strconv.Itoa(s.AnimalAttributes[name]))
	}
	set(contentsFields[17], s.SealLimits)
	set(contentsFields[15], "Main-window attributes are human-form purchases; buy animal physical attributes separately here. Skills use human form. Resources are ¥5,000; no cyberware or free contacts. After career start, use Raise attribute in the active form. Seal tables conflict: GM must enter Modifiers (p.36-derived limits) or Limits (printed p.37 limits), blank for other species. Animal adept powers are inactive.")
}
func applyShapeshifterContents() error {
	species := strings.TrimSpace(get(contentsFields[10]))
	var s *srcapp.ShapeshifterCreation
	if strings.EqualFold(species, "None") {
		if err := app.char.ConfigureShapeshifter(nil); err != nil {
			return err
		}
	} else {
		s = &srcapp.ShapeshifterCreation{Species: species, Form: strings.TrimSpace(get(contentsFields[11])), Movement: strings.TrimSpace(get(contentsFields[12])), SealLimits: strings.TrimSpace(get(contentsFields[17])), AnimalAttributes: map[string]int{}}
		for id, name := range map[int]string{13: "Body", 14: "Quickness", 16: "Strength"} {
			n, err := strconv.Atoi(strings.TrimSpace(get(contentsFields[id])))
			if err != nil {
				return fmt.Errorf("enter a base purchase for animal %s", name)
			}
			s.AnimalAttributes[name] = n
		}
		if err := app.char.ConfigureShapeshifter(s); err != nil {
			return err
		}
	}
	app.populateRaces()
	app.dirty = true
	app.update()
	contentsRefresh()
	return nil
}
