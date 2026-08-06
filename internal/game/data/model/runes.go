package models

// RuneWordData represents the data fields for Rune Words and their modifiers.
type RuneWordData struct {
	Name              string `csv:"Name"`              // Controls the string key used to display the name of the item when the Rune Word is complete
	Complete          int    `csv:"complete"`          // Boolean field. If equals 1, then the Rune Word can be crafted in-game. If equals 0, then the Rune Word cannot be crafted in-game.
	FirstLadderSeason int    `csv:"firstLadderSeason"` // Integer field. The first ladder season the Rune Word can be made on (inclusive). If blank or 0 then it is available in non-ladder.
	LastLadderSeason  int    `csv:"lastLadderSeason"`  // Integer field. The last ladder season the Rune Word is ladder-only (inclusive). Must be used in conjunction with firstLadderSeason.

	// Controls what item types are allowed for this Rune Word (Uses the ID field from ItemTypes.txt)
	ItemType1 string `csv:"itype1"`
	ItemType2 string `csv:"itype2"`
	ItemType3 string `csv:"itype3"`
	ItemType4 string `csv:"itype4"`
	ItemType5 string `csv:"itype5"`
	ItemType6 string `csv:"itype6"`

	// Controls what item types are excluded for this Rune Word (Uses the ID field from ItemTypes.txt)
	ExcludedType1 string `csv:"etype1"`
	ExcludedType2 string `csv:"etype2"`
	ExcludedType3 string `csv:"etype3"`

	// Controls what runes are required to make the Rune Word. The order of each of these fields matters. (Uses the ID field from misc.txt)
	Rune1 string `csv:"Rune1"`
	Rune2 string `csv:"Rune2"`
	Rune3 string `csv:"Rune3"`
	Rune4 string `csv:"Rune4"`
	Rune5 string `csv:"Rune5"`
	Rune6 string `csv:"Rune6"`

	// Controls the item properties that the Rune Word provides (Uses the "code" field from Properties.txt)
	T1Code1 string `csv:"T1Code1"`
	T1Code2 string `csv:"T1Code2"`
	T1Code3 string `csv:"T1Code3"`
	T1Code4 string `csv:"T1Code4"`
	T1Code5 string `csv:"T1Code5"`
	T1Code6 string `csv:"T1Code6"`
	T1Code7 string `csv:"T1Code7"`

	// The stat's "parameter" value associated with the related property (T1Code). Usage depends on the property function (See the "func" field on Properties.txt)
	T1Param1 string `csv:"T1Param1"`
	T1Param2 string `csv:"T1Param2"`
	T1Param3 string `csv:"T1Param3"`
	T1Param4 string `csv:"T1Param4"`
	T1Param5 string `csv:"T1Param5"`
	T1Param6 string `csv:"T1Param6"`
	T1Param7 string `csv:"T1Param7"`

	// The stat's "min" value to assign to the related property (T1Code). Usage depends on the property function (See the "func" field on Properties.txt)
	T1Min1 int `csv:"T1Min1"`
	T1Min2 int `csv:"T1Min2"`
	T1Min3 int `csv:"T1Min3"`
	T1Min4 int `csv:"T1Min4"`
	T1Min5 int `csv:"T1Min5"`
	T1Min6 int `csv:"T1Min6"`
	T1Min7 int `csv:"T1Min7"`

	// The stat's "max" value to assign to the related property (T1Code). Usage depends on the property function (See the "func" field on Properties.txt)
	T1Max1 int `csv:"T1Max1"`
	T1Max2 int `csv:"T1Max2"`
	T1Max3 int `csv:"T1Max3"`
	T1Max4 int `csv:"T1Max4"`
	T1Max5 int `csv:"T1Max5"`
	T1Max6 int `csv:"T1Max6"`
	T1Max7 int `csv:"T1Max7"`
}
