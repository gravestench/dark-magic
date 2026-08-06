package models

// RareSuffix represents the data for rare item name suffixes
type RareSuffix struct {
	Name         string `csv:"name"`    // Uses a string key to define the Rare Suffix name
	Version      int    `csv:"version"` // Defines which game version to use this Set bonus (0 = Classic mode | 100 = Expansion mode)
	IncludeType1 string `csv:"itype1"`  // Controls what item types are allowed for this Rare Suffix to spawn on (Uses the ID field from ItemTypes.txt)
	IncludeType2 string `csv:"itype2"`
	IncludeType3 string `csv:"itype3"`
	IncludeType4 string `csv:"itype4"`
	IncludeType5 string `csv:"itype5"`
	IncludeType6 string `csv:"itype6"`
	IncludeType7 string `csv:"itype7"`
	ExcludeType1 string `csv:"etype1"` // Controls what item types are excluded for this Rare Suffix to spawn on (Uses the ID field from ItemTypes.txt)
	ExcludeType2 string `csv:"etype2"`
	ExcludeType3 string `csv:"etype3"`
	ExcludeType4 string `csv:"etype4"`
}
