package srcapp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// Race packs are opt-in external data, never merged with built-in rules.
type RacePack struct {
	Format string           `json:"format"`
	Name   string           `json:"name"`
	Races  []RaceDefinition `json:"races"`
}

func LoadRacePack(path string) (*RacePack, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	d := json.NewDecoder(io.LimitReader(f, 2<<20))
	d.DisallowUnknownFields()
	var p RacePack
	if err = d.Decode(&p); err != nil {
		return nil, err
	}
	var extra any
	if err = d.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("race pack must contain one JSON object")
	}
	if p.Format != "runnerforge.race-pack.v1" || strings.TrimSpace(p.Name) == "" || len(p.Races) == 0 || len(p.Races) > 1000 {
		return nil, fmt.Errorf("invalid race pack format, name or race count")
	}
	seen := map[string]bool{}
	for _, r := range p.Races {
		if err = r.Validate(); err != nil {
			return nil, err
		}
		key := strings.ToLower(r.Name)
		if seen[key] {
			return nil, fmt.Errorf("duplicate race %s", r.Name)
		}
		seen[key] = true
	}
	return &p, nil
}

// SelectPackRace copies the definition into the save; the original file is
// unnecessary when the character is reopened on another computer.
func (ch *Character) SelectPackRace(r RaceDefinition) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if ch.Shapeshifter != nil || ch.Ghoul != nil {
		return fmt.Errorf("remove the special-character configuration before selecting a custom race")
	}
	copy := r
	copy.Modifiers = map[string]int{}
	for k, v := range r.Modifiers {
		copy.Modifiers[k] = v
	}
	ch.CustomRace = &copy
	ch.Race = copy.Name
	return nil
}
