package models

// MonsterUniqueModifier represents the monster modifier for special monsters, including Unique and Champion monsters.
type MonsterUniqueModifier struct {
	UniqueMod string `csv:"uniquemod"` // Reference field to define the monster modifier.
	ID        int    `csv:"id"`        // Unique numeric ID for the monster modifier. Used as a reference in other data files.
	Enabled   int    `csv:"enabled"`   // If 1, this monster modifier will be available for monsters to spawn with. If 0, it will never be used.
	Version   int    `csv:"version"`   // Defines which game version to use this monster modifier (<100 = Classic mode | 100 = Expansion mode).

	Xfer     int `csv:"xfer"`     // If 1, this monster modifier can be transferred from the Boss monster to Minion monsters, including auras. If 0, it will never be transferred.
	Champion int `csv:"champion"` // If 1, this monster modifier will only be used by Champion monsters. If 0, it can be used by any type of special monster.
	FPick    int `csv:"fPick"`    // Controls if this monster modifier is allowed on the monster based on the function code and the parameters it checks.

	Exclude1 string `csv:"exclude1"` // Controls which Monster Types should not have this monster modifier (Uses the "type" field from MonType.txt).
	Exclude2 string `csv:"exclude2"` // Additional exclusion for Monster Types.

	CPick  int `csv:"cpick"`     // Modifies the chances that this monster modifier will be chosen for a Champion monster compared to other monster modifiers.
	CPickN int `csv:"cpick (N)"` // Modifies the chances in Nightmare difficulty.
	CPickH int `csv:"cpick (H)"` // Modifies the chances in Hell difficulty.

	UPick  int `csv:"upick"`     // Modifies the chances that this monster modifier will be chosen for a Unique monster compared to other monster modifiers.
	UPickN int `csv:"upick (N)"` // Modifies the chances in Nightmare difficulty.
	UPickH int `csv:"upick (H)"` // Modifies the chances in Hell difficulty.

	Constants string `csv:"constants"` // Special list of numeric parameters for special monsters.
}
