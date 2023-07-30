package models

// ArmorData represents the additional data fields exclusive to armor type items.
type ArmorData struct {
	MinAc int // The minimum amount of Defense that an armor item type can have.
	MaxAc int // The maximum amount of Defense that an armor item type can have.
	Block int // Controls the block percent chance that the item provides (out of 100, but caps at 75).
	RArm  int // Controls the character's graphics and animations for the Right Arm component when wearing the armor, where the value 0 = Light or "lit", 1 = Medium or "med", and 2 = Heavy or "hvy".
	LArm  int // Controls the character's graphics and animations for the Left Arm component when wearing the armor, where the value 0 = Light or "lit", 1 = Medium or "med", and 2 = Heavy or "hvy".
	Torso int // Controls the character's graphics and animations for the Torso component when wearing the armor, where the value 0 = Light or "lit", 1 = Medium or "med", and 2 = Heavy or "hvy".
	Legs  int // Controls the character's graphics and animations for the Legs component when wearing the armor, where the value 0 = Light or "lit", 1 = Medium or "med", and 2 = Heavy or "hvy".
	RSPad int // Controls the character's graphics and animations for the Right Shoulder Pad component when wearing the armor, where the value 0 = Light or "lit", 1 = Medium or "med", and 2 = Heavy or "hvy".
	LSPad int // Controls the character's graphics and animations for the Left Shoulder Pad component when wearing the armor, where the value 0 = Light or "lit", 1 = Medium or "med", and 2 = Heavy or "hvy".
}
