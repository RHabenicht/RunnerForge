package srcapp

import (
	"encoding/json"
	"fmt"
)

// Individually checked book entries are separate from the legacy import.
// Regenerating catalog.json cannot discard these corrections or authorize
// unrelated rows in the original quarantine archive.
func (c *Catalogue) loadEquipmentSupplement() error {
	b, err := embeddedCatalog.ReadFile("equipment_supplement.json")
	if err != nil {
		return err
	}
	var supplement Catalogue
	if err := json.Unmarshal(b, &supplement); err != nil {
		return err
	}
	ids := make(map[int]bool, len(c.Entries))
	for _, item := range c.Entries {
		ids[item.ID] = true
	}
	for _, item := range supplement.Entries {
		if item.ID <= 0 || ids[item.ID] {
			return fmt.Errorf("equipment supplement has invalid or duplicate ID %d", item.ID)
		}
		ids[item.ID] = true
	}
	for _, book := range supplement.Books {
		found := false
		for i := range c.Books {
			if c.Books[i].Code == book.Code {
				c.Books[i] = book
				found = true
				break
			}
		}
		if !found {
			c.Books = append(c.Books, book)
		}
	}
	c.Entries = append(c.Entries, supplement.Entries...)
	return nil
}
