package models

import (
	"fmt"
	"log"
	"strconv"
)

// Hero is used for different types of hero's
type Hero int

// Heroes
const (
	HeroNone        Hero = iota //
	HeroBarbarian               // Barbarian
	HeroNecromancer             // Necromancer
	HeroPaladin                 // Paladin
	HeroAssassin                // Assassin
	HeroSorceress               // Sorceress
	HeroAmazon                  // Amazon
	HeroDruid                   // Druid
)

func HeroFromString(input string) Hero {
	switch input {
	case HeroBarbarian.String():
		return HeroBarbarian
	case HeroNecromancer.String():
		return HeroNecromancer
	case HeroPaladin.String():
		return HeroPaladin
	case HeroAssassin.String():
		return HeroAssassin
	case HeroSorceress.String():
		return HeroSorceress
	case HeroAmazon.String():
		return HeroAmazon
	case HeroDruid.String():
		return HeroDruid
	}

	return HeroNone
}

func (h Hero) String() string {
	switch h {
	case HeroBarbarian:
		return "Barbarian"
	case HeroNecromancer:
		return "Necromancer"
	case HeroPaladin:
		return "Paladin"
	case HeroAssassin:
		return "Assassin"
	case HeroSorceress:
		return "Sorceress"
	case HeroAmazon:
		return "Amazon"
	case HeroDruid:
		return "Druid"
	default:
		log.Fatalf("Unknown hero token: %d", h)
	}

	return ""
}

// GetToken returns a 2 letter token
func (h Hero) GetToken() string {
	switch h {
	case HeroBarbarian:
		return "BA"
	case HeroNecromancer:
		return "NE"
	case HeroPaladin:
		return "PA"
	case HeroAssassin:
		return "AI"
	case HeroSorceress:
		return "SO"
	case HeroAmazon:
		return "AM"
	case HeroDruid:
		return "DZ"
	default:
		log.Fatalf("Unknown hero token: %d", h)
	}

	return ""
}

// GetToken3 returns a 3 letter token
func (h Hero) GetToken3() string {
	switch h {
	case HeroBarbarian:
		return "BAR"
	case HeroNecromancer:
		return "NEC"
	case HeroPaladin:
		return "PAL"
	case HeroAssassin:
		return "ASS"
	case HeroSorceress:
		return "SOR"
	case HeroAmazon:
		return "AMA"
	case HeroDruid:
		return "DRU"
	default:
		log.Fatalf("Unknown hero token: %d", h)
	}

	return ""
}

// MarshalJSON converts CustomString to a JSON string when marshaling.
func (h Hero) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, h.String())), nil
}

// UnmarshalJSON parses a JSON string and sets the value of CustomString.
func (h *Hero) UnmarshalJSON(data []byte) error {
	// Remove surrounding double quotes if present
	strData := string(data)
	if len(strData) >= 2 && strData[0] == '"' && strData[len(strData)-1] == '"' {
		strData = strData[1 : len(strData)-1]
	}

	// Parse the string and set the value
	val, err := strconv.Atoi(strData)
	if err != nil {
		return err
	}
	*h = Hero(val)
	return nil
}
