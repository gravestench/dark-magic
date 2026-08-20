package models

// MonsterUniqueModifier represents the monster modifier for special monsters, including Unique and Champion monsters.
type MonsterUniqueModifier struct {
	UniqueMod string `csv:"uniquemod"` // Reference field to define the monster modifier.
	// Unique numeric ID for the monster modifier. Used as a reference in other data files.
	ID int `csv:"id"`
	// If 1, this monster modifier will be available for monsters to spawn with. If 0, it will never be used.
	Enabled int `csv:"enabled"`
	// Defines which game version to use this monster modifier (<100 = Classic mode | 100 = Expansion mode).
	Version int `csv:"version"`

	// If 1, this monster modifier can be transferred from the Boss monster to Minion monsters, including auras. If 0, it
	// will never be transferred.
	Xfer int `csv:"xfer"`
	// If 1, this monster modifier will only be used by Champion monsters. If 0, it can be used by any type of special
	// monster.
	Champion int `csv:"champion"`
	// Controls if this monster modifier is allowed on the monster based on the function code and the parameters it checks.
	FPick int `csv:"fPick"`

	// Controls which Monster Types should not have this monster modifier (Uses the "type" field from MonType.txt).
	Exclude1 string `csv:"exclude1"`
	Exclude2 string `csv:"exclude2"` // Additional exclusion for Monster Types.

	// Modifies the chances that this monster modifier will be chosen for a Champion monster compared to other monster
	// modifiers.
	CPick  int `csv:"cpick"`
	CPickN int `csv:"cpick (N)"` // Modifies the chances in Nightmare difficulty.
	CPickH int `csv:"cpick (H)"` // Modifies the chances in Hell difficulty.

	// Modifies the chances that this monster modifier will be chosen for a Unique monster compared to other monster
	// modifiers.
	UPick  int `csv:"upick"`
	UPickN int `csv:"upick (N)"` // Modifies the chances in Nightmare difficulty.
	UPickH int `csv:"upick (H)"` // Modifies the chances in Hell difficulty.

	Constants string `csv:"constants"` // Special list of numeric parameters for special monsters.
}
