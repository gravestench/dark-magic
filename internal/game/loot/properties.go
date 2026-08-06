package loot

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"
)

// PropertyStep is one of up to seven Properties.txt function/stat mappings.
type PropertyStep struct {
	Function int    `json:"function"`
	Stat     string `json:"stat,omitempty"`
	Set      bool   `json:"set,omitempty"`
	Value    int    `json:"value,omitempty"`
}

type PropertyDefinition struct {
	Code  string         `json:"code"`
	Steps []PropertyStep `json:"steps"`
}

type PropertyCatalog map[string]PropertyDefinition

// StatDefinition is the ItemStatCost metadata needed by portable item stats.
type StatDefinition struct {
	Code     string `json:"code"`
	Signed   bool   `json:"signed,omitempty"`
	ValShift int    `json:"valueShift,omitempty"`
	Minimum  int    `json:"minimum,omitempty"`
}

type StatCatalog map[string]StatDefinition

// ItemStat is an interpreted runtime stat operation. Set distinguishes an
// assignment from the normal additive behavior.
type ItemStat struct {
	Code      string `json:"code"`
	Parameter int    `json:"parameter,omitempty"`
	Value     int    `json:"value"`
	Set       bool   `json:"set,omitempty"`
	Function  int    `json:"function"`
}

type ItemEffect struct {
	Kind       string `json:"kind"`
	Property   string `json:"property"`
	Stat       string `json:"stat,omitempty"`
	Value      int    `json:"value,omitempty"`
	SkillID    int    `json:"skillId,omitempty"`
	SkillLevel int    `json:"skillLevel,omitempty"`
	Chance     int    `json:"chance,omitempty"`
	Charges    int    `json:"charges,omitempty"`
	Class      int    `json:"class,omitempty"`
	SkillTab   int    `json:"skillTab,omitempty"`
}

// PropertyApplication retains a known property whose function needs a more
// specialized runtime implementation.
type PropertyApplication struct {
	Property  string `json:"property"`
	Function  int    `json:"function"`
	Stat      string `json:"stat,omitempty"`
	Parameter int    `json:"parameter,omitempty"`
	Value     int    `json:"value"`
}

func ParsePropertiesTSV(input io.Reader) (PropertyCatalog, error) {
	reader := csv.NewReader(input)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("loot: read properties header: %w", err)
	}
	columns := columnsByName(header)
	if _, ok := columns["code"]; !ok {
		return nil, fmt.Errorf("loot: properties table missing required column %q", "code")
	}
	catalog := make(PropertyCatalog)
	for rowNumber := 2; ; rowNumber++ {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("loot: read properties row %d: %w", rowNumber, readErr)
		}
		code := field(row, columns, "code")
		if code == "" {
			continue
		}
		if _, duplicate := catalog[code]; duplicate {
			return nil, fmt.Errorf("loot: duplicate property %q at row %d", code, rowNumber)
		}
		definition := PropertyDefinition{Code: code}
		for index := 1; index <= 7; index++ {
			function, err := integerField(row, columns, fmt.Sprintf("func%d", index), rowNumber)
			if err != nil {
				return nil, err
			}
			stat := field(row, columns, fmt.Sprintf("stat%d", index))
			if function == 0 && stat == "" {
				continue
			}
			value, err := integerField(row, columns, fmt.Sprintf("val%d", index), rowNumber)
			if err != nil {
				return nil, err
			}
			definition.Steps = append(definition.Steps, PropertyStep{
				Function: function, Stat: stat, Set: booleanField(row, columns, fmt.Sprintf("set%d", index)), Value: value,
			})
		}
		catalog[code] = definition
	}
	return catalog, nil
}

func ParseItemStatCostTSV(input io.Reader) (StatCatalog, error) {
	reader := csv.NewReader(input)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("loot: read item stat cost header: %w", err)
	}
	columns := columnsByName(header)
	if _, ok := columns["Stat"]; !ok {
		return nil, fmt.Errorf("loot: item stat cost table missing required column %q", "Stat")
	}
	catalog := make(StatCatalog)
	for rowNumber := 2; ; rowNumber++ {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("loot: read item stat cost row %d: %w", rowNumber, readErr)
		}
		code := field(row, columns, "Stat")
		if code == "" {
			continue
		}
		if _, duplicate := catalog[code]; duplicate {
			return nil, fmt.Errorf("loot: duplicate item stat %q at row %d", code, rowNumber)
		}
		valShift, err := integerField(row, columns, "ValShift", rowNumber)
		if err != nil {
			return nil, err
		}
		minimum, err := integerField(row, columns, "MinAccr", rowNumber)
		if err != nil {
			return nil, err
		}
		catalog[code] = StatDefinition{Code: code, Signed: booleanField(row, columns, "Signed"), ValShift: valShift, Minimum: minimum}
	}
	return catalog, nil
}

// InterpretItemProperties resolves rolled affix property codes into runtime stats.
func InterpretItemProperties(item GeneratedItem, properties PropertyCatalog, stats StatCatalog) (GeneratedItem, error) {
	return InterpretItemPropertiesWithSkills(item, properties, stats, nil, 0)
}

// InterpretItemPropertiesWithSkills additionally resolves Properties function
// 36 deterministically against class-owned Skills.txt records.
func InterpretItemPropertiesWithSkills(item GeneratedItem, properties PropertyCatalog, stats StatCatalog, skills Skills, seed uint64) (GeneratedItem, error) {
	rng := splitMix64(seed)
	item.Stats = nil
	item.Effects = nil
	item.Ethereal = false
	item.Indestructible = false
	item.Unsupported = nil
	for _, affix := range append(append([]RolledAffix{}, item.Prefixes...), item.Suffixes...) {
		for _, modifier := range affix.Modifiers {
			definition, ok := properties[modifier.Code]
			if !ok {
				return GeneratedItem{}, fmt.Errorf("loot: affix %q references unknown property %q", affix.Name, modifier.Code)
			}
			previous := 0
			for _, step := range definition.Steps {
				function := step.Function
				if function == 3 || function == 9 {
					function = previous
				} else if function != 0 {
					previous = function
				}
				if directPropertyFunction(function) {
					if _, ok := stats[step.Stat]; !ok {
						return GeneratedItem{}, fmt.Errorf("loot: property %q references unknown item stat %q", modifier.Code, step.Stat)
					}
					item.Stats = append(item.Stats, ItemStat{Code: step.Stat, Parameter: modifier.Parameter, Value: modifier.Value, Set: step.Set, Function: function})
					continue
				}
				handled, err := interpretSpecialProperty(&item, modifier, step, function, stats, skills, &rng)
				if err != nil {
					return GeneratedItem{}, err
				}
				if handled {
					continue
				}
				item.Unsupported = append(item.Unsupported, PropertyApplication{Property: modifier.Code, Function: function, Stat: step.Stat, Parameter: modifier.Parameter, Value: modifier.Value})
			}
		}
	}
	sort.SliceStable(item.Stats, func(i, j int) bool {
		if item.Stats[i].Code == item.Stats[j].Code {
			return item.Stats[i].Parameter < item.Stats[j].Parameter
		}
		return item.Stats[i].Code < item.Stats[j].Code
	})
	return item, nil
}

func interpretSpecialProperty(item *GeneratedItem, modifier RolledModifier, step PropertyStep, function int, stats StatCatalog, skills Skills, rng *splitMix64) (bool, error) {
	requireStat := func() error {
		if _, ok := stats[step.Stat]; !ok {
			return fmt.Errorf("loot: property %q references unknown item stat %q", modifier.Code, step.Stat)
		}
		return nil
	}
	addEffect := func(effect ItemEffect) {
		effect.Property = modifier.Code
		effect.Stat = step.Stat
		item.Effects = append(item.Effects, effect)
	}
	switch function {
	case 5:
		addEffect(ItemEffect{Kind: "minimum_damage", Value: modifier.Value})
	case 6:
		addEffect(ItemEffect{Kind: "maximum_damage", Value: modifier.Value})
	case 7:
		addEffect(ItemEffect{Kind: "damage_percent", Value: modifier.Value})
	case 10:
		if err := requireStat(); err != nil {
			return false, err
		}
		addEffect(ItemEffect{Kind: "skill_tab", Value: modifier.Value, Class: modifier.Parameter / 3, SkillTab: modifier.Parameter % 3})
	case 11:
		if err := requireStat(); err != nil {
			return false, err
		}
		addEffect(ItemEffect{Kind: "proc", SkillID: modifier.Parameter, Chance: modifier.Minimum, SkillLevel: modifier.Maximum})
	case 12:
		if err := requireStat(); err != nil {
			return false, err
		}
		addEffect(ItemEffect{Kind: "random_skill", SkillID: modifier.Value, SkillLevel: modifier.Parameter})
	case 18:
		if err := requireStat(); err != nil {
			return false, err
		}
		item.Stats = append(item.Stats, ItemStat{Code: step.Stat, Value: modifier.Parameter, Set: step.Set, Function: function})
	case 19:
		if err := requireStat(); err != nil {
			return false, err
		}
		addEffect(ItemEffect{Kind: "charged_skill", SkillID: modifier.Parameter, Charges: modifier.Minimum, SkillLevel: modifier.Maximum})
	case 20:
		item.Indestructible = item.Indestructible || modifier.Value > 0
	case 21:
		if err := requireStat(); err != nil {
			return false, err
		}
		item.Stats = append(item.Stats, ItemStat{Code: step.Stat, Parameter: step.Value, Value: modifier.Value, Set: step.Set, Function: function})
	case 23:
		item.Ethereal = item.Ethereal || modifier.Value > 0
	case 24:
		addEffect(ItemEffect{Kind: "state", Value: modifier.Value})
	case 36:
		if len(skills) == 0 {
			return false, nil
		}
		var candidates []Skill
		for _, skill := range skills {
			if skill.Class == modifier.Parameter {
				candidates = append(candidates, skill)
			}
		}
		if len(candidates) == 0 {
			return false, fmt.Errorf("loot: property %q has no skills for class %d", modifier.Code, modifier.Parameter)
		}
		selected := candidates[int(rng.next()%uint64(len(candidates)))]
		addEffect(ItemEffect{Kind: "random_class_skill", SkillID: selected.ID, SkillLevel: modifier.Value, Class: selected.Class})
	default:
		return false, nil
	}
	return true, nil
}

func directPropertyFunction(function int) bool {
	switch function {
	case 1, 2, 8, 13, 14, 15, 16, 17, 22:
		return true
	default:
		return false
	}
}

func columnsByName(header []string) map[string]int {
	columns := make(map[string]int, len(header))
	for index, name := range header {
		columns[strings.TrimPrefix(name, "\ufeff")] = index
	}
	return columns
}
