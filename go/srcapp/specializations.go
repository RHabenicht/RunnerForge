package srcapp

import (
	"fmt"
	"sort"
	"strings"
)

func (s Rated) SpecializationText() string {
	entries := []string{}
	if s.Specialization != "" {
		entries = append(entries, fmt.Sprintf("%s %d", s.Specialization, s.SpecializationRating()))
	}
	for name, rating := range s.AdditionalSpecializations {
		entries = append(entries, fmt.Sprintf("%s %d", name, rating))
	}
	sort.Strings(entries)
	return strings.Join(entries, "; ")
}
func (ch *Character) LearnOrAdvanceSpecialization(index int, name string) error {
	if !ch.Career {
		return fmt.Errorf("start career mode before spending karma")
	}
	if index < 0 || index >= len(ch.Skills) {
		return fmt.Errorf("select a base skill")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("enter a specialization name")
	}
	s := &ch.Skills[index]
	link := s.Link()
	if link == "" {
		return fmt.Errorf("set the linked attribute first")
	}
	attr := ch.skillAttribute(link)
	if link == "Reaction" {
		attr = (ch.skillAttribute("Quickness") + ch.skillAttribute("Intelligence")) / 2
	}
	current := s.Rating
	exists := false
	primary := false
	if strings.EqualFold(name, s.Specialization) {
		name = s.Specialization
		current = s.SpecializationRating()
		exists = true
		primary = true
	}
	for n, v := range s.AdditionalSpecializations {
		if strings.EqualFold(name, n) {
			name = n
			current = v
			exists = true
		}
	}
	count := len(s.AdditionalSpecializations)
	if s.Specialization != "" {
		count++
	}
	if !exists && count >= attr {
		return fmt.Errorf("specialization count cannot exceed the linked attribute")
	}
	next := current + 1
	if next > max(3, s.Rating*2) {
		return fmt.Errorf("raise the base skill before increasing this specialization")
	}
	factor := 1
	if next > attr {
		factor++
	}
	if next > attr*2 {
		factor++
	}
	cost := next * factor / 2
	if ch.Karma < cost {
		return fmt.Errorf("requires %d unspent karma", cost)
	}
	if primary {
		s.SpecializationValue = &next
	} else {
		if s.AdditionalSpecializations == nil {
			s.AdditionalSpecializations = map[string]int{}
		}
		s.AdditionalSpecializations[name] = next
	}
	ch.Karma -= cost
	ch.Advancements = append(ch.Advancements, Advancement{"Specialization", s.Name + " (" + name + ")", current, next, cost})
	return nil
}
