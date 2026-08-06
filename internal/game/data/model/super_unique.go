package models

// SuperUnique describes a static special-boss encounter from SuperUniques.txt.
// Field meanings and cross-table references follow the bundled Diablo II Data
// File Guide; unknown and patch-specific columns remain available generically.
type SuperUnique struct {
	ID                     string `csv:"Superunique"`
	Name                   string `csv:"Name"`
	Class                  string `csv:"Class"`
	HardcodedID            int    `csv:"hcIdx"`
	MonsterSound           string `csv:"MonSound"`
	Modifier1              int    `csv:"Mod1"`
	Modifier2              int    `csv:"Mod2"`
	Modifier3              int    `csv:"Mod3"`
	MinGroup               int    `csv:"MinGrp"`
	MaxGroup               int    `csv:"MaxGrp"`
	AutoPosition           bool   `csv:"AutoPos"`
	Stacks                 bool   `csv:"Stacks"`
	Replaceable            bool   `csv:"Replaceable"`
	TransformNormal        int    `csv:"Utrans"`
	TransformNightmare     int    `csv:"Utrans(N)"`
	TransformHell          int    `csv:"Utrans(H)"`
	TreasureClassNormal    string `csv:"TC"`
	TreasureClassNightmare string `csv:"TC(N)"`
	TreasureClassHell      string `csv:"TC(H)"`
}
