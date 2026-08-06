package models

// WeaponClassID represents the base weapon class
type WeaponClassID string

const (
	// HandToHand represents the Hand to Hand weapon class
	HandToHand WeaponClassID = "hth"
	// OneHandedSwing represents the One Handed Swing weapon class
	OneHandedSwing WeaponClassID = "1hs"
	// OneHandedThrust represents the One Handed Thrust weapon class
	OneHandedThrust WeaponClassID = "1ht"
	// Bow represents the Bow weapon class
	Bow WeaponClassID = "bow"
	// TwoHandedSwing represents the Two Handed Swing weapon class
	TwoHandedSwing WeaponClassID = "2hs"
	// TwoHandedThrust represents the Two Handed Thrust weapon class
	TwoHandedThrust WeaponClassID = "2ht"
	// DualWieldingLeftJabRightSwing represents the Dual Wielding - Left Jab Right Swing weapon class
	DualWieldingLeftJabRightSwing WeaponClassID = "1js"
	// DualWieldingLeftJabRightThrust represents the Dual Wielding - Left Jab Right Thrust weapon class
	DualWieldingLeftJabRightThrust WeaponClassID = "1jt"
	// DualWieldingLeftSwingRightSwing represents the Dual Wielding - Left Swing Right Swing weapon class
	DualWieldingLeftSwingRightSwing WeaponClassID = "1ss"
	// DualWieldingLeftSwingRightThrust represents the Dual Wielding - Left Swing Right Thrust weapon class
	DualWieldingLeftSwingRightThrust WeaponClassID = "1st"
	// Staff represents the Staff weapon class
	Staff WeaponClassID = "stf"
	// Crossbow represents the Crossbow weapon class
	Crossbow WeaponClassID = "xbw"
	// OneHandToHand represents the One Hand to Hand weapon class
	OneHandToHand WeaponClassID = "ht1"
	// TwoHandToHand represents the Two Hand to Hand weapon class
	TwoHandToHand WeaponClassID = "ht2"
)

// WeaponClass represents different weapon classes.
type WeaponClass struct {
	Name string        `csv:"Weapon Class"` // The name of the weapon class.
	Code WeaponClassID `csv:"Code"`         // The unique 3-letter/number code for the weapon class.
}
