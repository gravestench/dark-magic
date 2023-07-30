package models

// SetBonusData represents the data fields for Set bonus statistics when the player has equipped enough Set Items.
type SetBonusData struct {
	Index   string `csv:"index"`   // Defines the specific Set ID
	Name    string `csv:"name"`    // Uses a string for displaying the Set name in the inventory tooltip
	Version int    `csv:"version"` // Defines which game version to use this Set bonus (0 = Classic mode | 100 = Expansion mode)

	// Controls the each of the different pairs of Partial Set item properties. These are applied when the player has equipped the related # of Set items.
	PCode2a string `csv:"PCode2a"`
	PCode3a string `csv:"PCode3a"`
	PCode4a string `csv:"PCode4a"`
	PCode5a string `csv:"PCode5a"`

	// The stat's "parameter" value associated with the relative property (PCode#a). Usage depends on the property function (See the "func" field on Properties.txt)
	PParam2a string `csv:"PParam2a"`
	PParam3a string `csv:"PParam3a"`
	PParam4a string `csv:"PParam4a"`
	PParam5a string `csv:"PParam5a"`

	// The stat's "min" value associated with the listed relative (PCode#a). Usage depends on the property function (See the "func" field on Properties.txt)
	PMin2a int `csv:"PMin2a"`
	PMin3a int `csv:"PMin3a"`
	PMin4a int `csv:"PMin4a"`
	PMin5a int `csv:"PMin5a"`

	// The stat's "max" value to assign to the listed relative (PCode#a). Usage depends on the property function (See the "func" field on Properties.txt)
	PMax2a int `csv:"PMax2a"`
	PMax3a int `csv:"PMax3a"`
	PMax4a int `csv:"PMax4a"`
	PMax5a int `csv:"PMax5a"`

	// Controls the each of the different pairs of Partial Set item properties. These are applied when the player has equipped the related # of Set items.
	PCode2b string `csv:"PCode2b"`
	PCode3b string `csv:"PCode3b"`
	PCode4b string `csv:"PCode4b"`
	PCode5b string `csv:"PCode5b"`

	// The stat's "parameter" value associated with the relative property (PCode#b). Usage depends on the property function (See the "func" field on Properties.txt)
	PParam2b string `csv:"PParam2b"`
	PParam3b string `csv:"PParam3b"`
	PParam4b string `csv:"PParam4b"`
	PParam5b string `csv:"PParam5b"`

	// The stat's "min" value associated with the listed relative (PCode#b). Usage depends on the property function (See the "func" field on Properties.txt)
	PMin2b int `csv:"PMin2b"`
	PMin3b int `csv:"PMin3b"`
	PMin4b int `csv:"PMin4b"`
	PMin5b int `csv:"PMin5b"`

	// The stat's "max" value to assign to the listed relative (PCode#b). Usage depends on the property function (See the "func" field on Properties.txt)
	PMax2b int `csv:"PMax2b"`
	PMax3b int `csv:"PMax3b"`
	PMax4b int `csv:"PMax4b"`
	PMax5b int `csv:"PMax5b"`

	// Controls the each of the different Full Set item properties. These are applied when the player has all Set item pieces equipped.
	FCode1 string `csv:"FCode1"`
	FCode2 string `csv:"FCode2"`
	FCode3 string `csv:"FCode3"`
	FCode4 string `csv:"FCode4"`
	FCode5 string `csv:"FCode5"`
	FCode6 string `csv:"FCode6"`
	FCode7 string `csv:"FCode7"`
	FCode8 string `csv:"FCode8"`

	// The stat's "parameter" value associated with the relative property (FCode#b). Usage depends on the property function (See the "func" field on Properties.txt)
	FParam1 string `csv:"FParam1"`
	FParam2 string `csv:"FParam2"`
	FParam3 string `csv:"FParam3"`
	FParam4 string `csv:"FParam4"`
	FParam5 string `csv:"FParam5"`
	FParam6 string `csv:"FParam6"`
	FParam7 string `csv:"FParam7"`
	FParam8 string `csv:"FParam8"`

	// The stat's "min" value associated with the listed relative (FCode#b). Usage depends on the property function (See the "func" field on Properties.txt)
	FMin1 int `csv:"FMin1"`
	FMin2 int `csv:"FMin2"`
	FMin3 int `csv:"FMin3"`
	FMin4 int `csv:"FMin4"`
	FMin5 int `csv:"FMin5"`
	FMin6 int `csv:"FMin6"`
	FMin7 int `csv:"FMin7"`
	FMin8 int `csv:"FMin8"`

	// The stat's "max" value to assign to the listed relative (FCode#b). Usage depends on the property function (See the "func" field on Properties.txt)
	FMax1 int `csv:"FMax1"`
	FMax2 int `csv:"FMax2"`
	FMax3 int `csv:"FMax3"`
	FMax4 int `csv:"FMax4"`
	FMax5 int `csv:"FMax5"`
	FMax6 int `csv:"FMax6"`
	FMax7 int `csv:"FMax7"`
	FMax8 int `csv:"FMax8"`
}
