package models

type Book struct {
	Name            string `csv:"Name"`            // Reference field to define the book
	ScrollSpellCode string `csv:"ScrollSpellCode"` // Item's code for the related scroll item
	BookSpellCode   string `csv:"BookSpellCode"`   // Item's code for the book item
	// Item spell function to use when using the book (See Code-Description mapping)
	PSpell int `csv:"pSpell"`
	// Numeric index to pick the DC6 file for the mouse cursor when using the scroll or book
	SpellIcon int `csv:"SpellIcon"`
	// Skill to use for the scroll item (uses the "skill" field from skills.txt)
	ScrollSkill string `csv:"ScrollSkill"`
	// Skill to use for the book item (uses the "skill" field from skills.txt)
	BookSkill string `csv:"BookSkill"`
	BaseCost  int    `csv:"BaseCost"` // Starting gold cost to buy the book from an NPC
	// Additional gold cost added with the book's "BaseCost" value, based on how many charges the book has
	CostPerCharge int `csv:"CostPerCharge"`
}
