package models

// TransformType represents the color palette change of the item for the character model graphics and inventory graphics.
type TransformType int

const (
	TransformTypeNoColorChange TransformType = iota
	TransformTypeGrey
	TransformTypeGrey2
	TransformTypeGold
	TransformTypeBrown
	TransformTypeGreyBrown
	TransformTypeInventoryGrey
	TransformTypeInventoryGrey2
	TransformTypeInventoryGreyBrown
)
