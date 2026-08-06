package models

// MonsterEquipment represents the equipment setup for a monster.
type MonsterEquipment struct {
	Monster string `csv:"monster"` // Defines the monster that should be equipped. Points to the matching "Id" value in the monstats.txt file.
	OnInit  int    `csv:"oninit"`  // Defines if the monster equipment is added on initialization during the monster's creation.
	Level   int    `csv:"level"`   // Defines the level requirement for the monster to gain this equipment.

	Item1 string `csv:"item1"` // Item to be equipped on the monster. Uses ID pointer from Weapons.txt, Armor.txt, or Misc.txt.
	Loc1  string `csv:"loc1"`  // Specifies the inventory slot where the item will be equipped.
	Mod1  int    `csv:"mod1"`  // Controls the quality level of the related item.

	Item2 string `csv:"item2"` // Second item to be equipped on the monster.
	Loc2  string `csv:"loc2"`  // Specifies the inventory slot for the second item.
	Mod2  int    `csv:"mod2"`  // Quality level for the second item.

	Item3 string `csv:"item3"` // Third item to be equipped on the monster.
	Loc3  string `csv:"loc3"`  // Specifies the inventory slot for the third item.
	Mod3  int    `csv:"mod3"`  // Quality level for the third item.
}
