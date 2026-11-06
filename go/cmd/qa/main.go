package main

import (
	"fmt"
	"nsrcgii/srcapp"
	"os"
	"path/filepath"
)

func main() {
	out := "sample_exports"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	if e := os.MkdirAll(out, 0755); e != nil {
		panic(e)
	}
	c, e := srcapp.ReadLegacy("srcapp/testdata/archStreetSamurai.sr3")
	if e != nil {
		panic(e)
	}
	if e = c.SaveHTML(filepath.Join(out, "street_samurai_enhanced.html")); e != nil {
		panic(e)
	}
	if e = srcapp.SaveJSON(filepath.Join(out, "street_samurai_enhanced.nsrcg2.json"), c); e != nil {
		panic(e)
	}
	fmt.Println("Name", c.Name, "Reaction", c.Reaction(), "Initiative Dice", c.InitiativeDice(), "Combat Pool", c.CombatPool(), "Essence", c.Essence())
}
