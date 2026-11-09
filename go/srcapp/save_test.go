package srcapp

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveReplacesExistingCharacter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "character.json")
	character := NewCharacter()
	if err := SaveJSON(path, character); err != nil {
		t.Fatal(err)
	}
	character.Name = "Updated"
	if err := SaveJSON(path, character); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Name != character.Name {
		t.Fatal("replacement was not saved")
	}
}

func TestSaveFailurePreservesExistingBytes(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "character.json")
	character := NewCharacter()
	if err := SaveJSON(path, character); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	character.Cyberware = []Possession{{EssenceCost: math.NaN()}}
	if err := SaveJSON(path, character); err == nil {
		t.Fatal("invalid value accepted")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("existing save changed on failure")
	}
	// A directory destination forces the final replacement to fail.
	if err := SaveJSON(directory, NewCharacter()); err == nil {
		t.Fatal("directory replaced")
	}
	matches, _ := filepath.Glob(filepath.Join(filepath.Dir(directory), ".nsrcg-save-*"))
	if len(matches) != 0 {
		t.Fatal("temporary save leaked")
	}
}

func TestLoadRejectsDataItCannotPreserve(t *testing.T) {
	cases := []string{
		`{"edition":3,"future_field":true}`,
		`{"edition":3,"gear":[{"name":"item","future_field":true}]}`,
		`{"edition":3,"format":"nsrcgii.v99"}`,
		`{"edition":3} {"edition":3}`,
		`{"edition":3} garbage`,
		`null`,
	}
	for _, content := range cases {
		t.Run(content, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "character.json")
			if err := os.WriteFile(path, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadJSON(path); err == nil {
				t.Fatal("unsafe input accepted")
			}
			after, _ := os.ReadFile(path)
			if string(after) != content {
				t.Fatal("input modified")
			}
		})
	}
}

func TestLoadLegacyMissingFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "character.json")
	if err := os.WriteFile(path, []byte(`{"edition":3,"name":"Older save"}`), 0600); err != nil {
		t.Fatal(err)
	}
	character, err := LoadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if character.Format != "nsrcgii.v2" || character.Name != "Older save" {
		t.Fatal("repair failed")
	}
}
