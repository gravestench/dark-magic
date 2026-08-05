package loot

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"
)

type AffixKind string

const (
	AffixPrefix AffixKind = "prefix"
	AffixSuffix AffixKind = "suffix"
)

// AffixModifier retains one property roll range from an affix record.
type AffixModifier struct {
	Code      string `json:"code"`
	Parameter int    `json:"parameter,omitempty"`
	Minimum   int    `json:"minimum"`
	Maximum   int    `json:"maximum"`
}

// Affix is the selection-relevant subset of MagicPrefix/MagicSuffix records.
type Affix struct {
	Name      string          `json:"name"`
	Kind      AffixKind       `json:"kind"`
	Version   int             `json:"version"`
	Spawnable bool            `json:"spawnable"`
	Rare      bool            `json:"rare"`
	Level     int             `json:"level"`
	MaxLevel  int             `json:"maxLevel,omitempty"`
	Frequency int             `json:"frequency"`
	Group     int             `json:"group,omitempty"`
	Includes  []string        `json:"includes"`
	Excludes  []string        `json:"excludes,omitempty"`
	Modifiers []AffixModifier `json:"modifiers,omitempty"`
}

// AffixOptions controls the bounded prefix/suffix selection for an item.
type AffixOptions struct {
	Version     int     `json:"version"`
	AffixLevel  int     `json:"affixLevel"`
	MaxPrefixes int     `json:"maxPrefixes"`
	MaxSuffixes int     `json:"maxSuffixes"`
	MaxTotal    int     `json:"maxTotal"`
	Quality     Quality `json:"quality"`
}

// GeneratedItem is the portable output of base, quality, special, and affix selection.
type GeneratedItem struct {
	Base     BaseItem     `json:"base"`
	Quality  Quality      `json:"quality"`
	Special  *SpecialItem `json:"special,omitempty"`
	Prefixes []Affix      `json:"prefixes,omitempty"`
	Suffixes []Affix      `json:"suffixes,omitempty"`
}

// ParseAffixesTSV parses either MagicPrefix.txt or MagicSuffix.txt.
func ParseAffixesTSV(input io.Reader, kind AffixKind) ([]Affix, error) {
	if kind != AffixPrefix && kind != AffixSuffix {
		return nil, fmt.Errorf("loot: unsupported affix kind %q", kind)
	}
	reader := csv.NewReader(input)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("loot: read %s header: %w", kind, err)
	}
	columns := make(map[string]int, len(header))
	for index, name := range header {
		columns[strings.TrimPrefix(name, "\ufeff")] = index
	}
	for _, required := range []string{"Name", "version", "spawnable", "rare", "level", "maxlevel", "frequency", "group", "itype1"} {
		if _, ok := columns[required]; !ok {
			return nil, fmt.Errorf("loot: %s table missing required column %q", kind, required)
		}
	}
	var affixes []Affix
	for rowNumber := 2; ; rowNumber++ {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("loot: read %s row %d: %w", kind, rowNumber, readErr)
		}
		name := field(row, columns, "Name")
		if name == "" {
			continue
		}
		affix := Affix{Name: name, Kind: kind, Spawnable: booleanField(row, columns, "spawnable"), Rare: booleanField(row, columns, "rare")}
		integerTargets := []struct {
			name string
			to   *int
		}{{"version", &affix.Version}, {"level", &affix.Level}, {"maxlevel", &affix.MaxLevel}, {"frequency", &affix.Frequency}, {"group", &affix.Group}}
		for _, target := range integerTargets {
			*target.to, err = integerField(row, columns, target.name, rowNumber)
			if err != nil {
				return nil, err
			}
		}
		for index := 1; index <= 7; index++ {
			if value := field(row, columns, fmt.Sprintf("itype%d", index)); value != "" {
				affix.Includes = append(affix.Includes, value)
			}
			if value := field(row, columns, fmt.Sprintf("etype%d", index)); value != "" {
				affix.Excludes = append(affix.Excludes, value)
			}
		}
		for index := 1; index <= 3; index++ {
			code := field(row, columns, fmt.Sprintf("mod%dcode", index))
			if code == "" {
				continue
			}
			modifier := AffixModifier{Code: code}
			for _, target := range []struct {
				name string
				to   *int
			}{{fmt.Sprintf("mod%dparam", index), &modifier.Parameter}, {fmt.Sprintf("mod%dmin", index), &modifier.Minimum}, {fmt.Sprintf("mod%dmax", index), &modifier.Maximum}} {
				*target.to, err = integerField(row, columns, target.name, rowNumber)
				if err != nil {
					return nil, err
				}
			}
			affix.Modifiers = append(affix.Modifiers, modifier)
		}
		affixes = append(affixes, affix)
	}
	return affixes, nil
}

// SelectAffixes chooses frequency-weighted affixes with a 50% check per slot.
func SelectAffixes(item BaseItem, types ItemTypes, prefixes, suffixes []Affix, options AffixOptions, seed uint64) ([]Affix, []Affix, error) {
	if options.MaxPrefixes < 0 || options.MaxSuffixes < 0 || options.MaxTotal < 0 || options.AffixLevel < 0 {
		return nil, nil, fmt.Errorf("loot: affix limits and level must not be negative")
	}
	rng := splitMix64(seed)
	usedGroups := make(map[int]bool)
	selectedPrefixes, err := selectAffixKind(item, types, prefixes, options, options.MaxPrefixes, &rng, usedGroups, nil)
	if err != nil {
		return nil, nil, err
	}
	selectedSuffixes, err := selectAffixKind(item, types, suffixes, options, options.MaxSuffixes, &rng, usedGroups, selectedPrefixes)
	if err != nil {
		return nil, nil, err
	}
	return selectedPrefixes, selectedSuffixes, nil
}

func selectAffixKind(item BaseItem, types ItemTypes, pool []Affix, options AffixOptions, maxPicks int, rng *splitMix64, usedGroups map[int]bool, already []Affix) ([]Affix, error) {
	selected := make([]Affix, 0, maxPicks)
	for attempt := 0; attempt < maxPicks && len(already)+len(selected) < options.MaxTotal; attempt++ {
		if rng.next()%2 == 0 {
			continue
		}
		candidates := make([]Affix, 0)
		total := 0
		for _, affix := range pool {
			if containsAffix(selected, affix.Name) {
				continue
			}
			eligible, err := affixEligible(affix, item, types, options, usedGroups)
			if err != nil {
				return nil, err
			}
			if eligible {
				candidates = append(candidates, affix)
				total += affix.Frequency
			}
		}
		if total == 0 {
			continue
		}
		sort.Slice(candidates, func(i, j int) bool { return candidates[i].Name < candidates[j].Name })
		roll := int(rng.next() % uint64(total))
		for _, candidate := range candidates {
			if roll < candidate.Frequency {
				selected = append(selected, candidate)
				if candidate.Group != 0 {
					usedGroups[candidate.Group] = true
				}
				break
			}
			roll -= candidate.Frequency
		}
	}
	return selected, nil
}

func containsAffix(affixes []Affix, name string) bool {
	for _, affix := range affixes {
		if affix.Name == name {
			return true
		}
	}
	return false
}

func affixEligible(affix Affix, item BaseItem, types ItemTypes, options AffixOptions, usedGroups map[int]bool) (bool, error) {
	if !affix.Spawnable || affix.Frequency <= 0 || affix.Level > options.AffixLevel || (affix.MaxLevel > 0 && options.AffixLevel > affix.MaxLevel) || (options.Version < 100 && affix.Version >= 100) || (options.Quality == QualityRare && !affix.Rare) || (affix.Group != 0 && usedGroups[affix.Group]) {
		return false, nil
	}
	for _, excluded := range affix.Excludes {
		match, err := itemMatchesType(item, types, excluded)
		if err != nil || match {
			return false, err
		}
	}
	for _, included := range affix.Includes {
		match, err := itemMatchesType(item, types, included)
		if err != nil || match {
			return match, err
		}
	}
	return false, nil
}
