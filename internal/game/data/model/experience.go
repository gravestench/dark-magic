package models

// ExperienceData represents the data fields for experience required for each level by each class.
type ExperienceData struct {
	//Level       string `csv:"Level"`       // This is a reference field to define the level
	Amazon      int `csv:"Amazon"`      // Controls the experience required for each level with the Amazon class
	Sorceress   int `csv:"Sorceress"`   // Controls the experience required for each level with the Sorceress class
	Necromancer int `csv:"Necromancer"` // Controls the experience required for each level with the Necromancer class
	Paladin     int `csv:"Paladin"`     // Controls the experience required for each level with the Paladin class
	Barbarian   int `csv:"Barbarian"`   // Controls the experience required for each level with the Barbarian class
	Druid       int `csv:"Druid"`       // Controls the experience required for each level with the Druid class
	Assassin    int `csv:"Assassin"`    // Controls the experience required for each level with the Assassin class
	ExpRatio    int `csv:"ExpRatio"`    // This multiplier affects the percentage of experienced earned, based on the level
}
