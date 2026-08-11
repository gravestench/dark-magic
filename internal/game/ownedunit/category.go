// Package ownedunit owns the authoritative relationship shared by summons,
// pets, traps, minions, and hirelings. It deliberately does not own their AI,
// combat statistics, inventory, rendering, or durable character records.
package ownedunit

import (
	"fmt"
	"strings"

	models "github.com/gravestench/dark-magic/internal/game/data/model"
)

// Replacement says what happens when one owner exceeds a category limit.
// Exact retail ordering varies by pet family, so every category must name it.
type Replacement string

const (
	Reject        Replacement = "reject"
	ReplaceOldest Replacement = "replace_oldest"
	ReplaceNewest Replacement = "replace_newest"
)

// Category is the immutable ownership/lifetime policy copied from typed data
// and an implementing behavior family. Effective skill/stat limits can be
// supplied later without changing the relationship schema.
type Category struct {
	ID                 string      `json:"id"`
	Group              int64       `json:"group"`
	BaseMax            int64       `json:"base_max"`
	Replacement        Replacement `json:"replacement"`
	Durable            bool        `json:"durable"`
	Unsummon           bool        `json:"unsummon"`
	WarpWithOwner      bool        `json:"warp_with_owner"`
	RangeLimited       bool        `json:"range_limited"`
	SurvivesOwnerDeath bool        `json:"survives_owner_death"`
}

func (category Category) validate() error {
	category.ID = strings.TrimSpace(category.ID)
	if category.ID == "" || category.Group < 0 || category.BaseMax < 1 {
		return fmt.Errorf("owned unit: category needs an ID, non-negative group, and positive maximum")
	}
	switch category.Replacement {
	case Reject, ReplaceOldest, ReplaceNewest:
		return nil
	default:
		return fmt.Errorf("owned unit: category %q has invalid replacement %q", category.ID, category.Replacement)
	}
}

// CategoryFromPetType preserves the ownership/lifetime facts verified in the
// legacy table. Replacement and durability are behavior-family policy because
// PetType alone does not establish one universal answer for them.
func CategoryFromPetType(record models.PetType, replacement Replacement, durable, survivesOwnerDeath bool) (Category, error) {
	category := Category{
		ID: strings.TrimSpace(record.PetType), Group: int64(record.Group), BaseMax: int64(record.BaseMax),
		Replacement: replacement, Durable: durable, Unsummon: record.Unsummon,
		WarpWithOwner: record.Warp, RangeLimited: record.Range, SurvivesOwnerDeath: survivesOwnerDeath,
	}
	if err := category.validate(); err != nil {
		return Category{}, err
	}
	return category, nil
}
