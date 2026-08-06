package models

type TransformColor struct {
	Name string    `csv:"Transform Color"`
	Code ColorCode `csv:"Code"`
}
type ComponentCode struct {
	Name string `csv:"component"`
	Code string `csv:"code"`
}
type ElementType struct {
	Name string `csv:"Elemental Type"`
	Code string `csv:"Code"`
}
type UnitEvent struct {
	Event       string `csv:"event"`
	Description string `csv:"*desc"`
}
type MissileCalculation struct {
	Code        string `csv:"code"`
	Description string `csv:"*desc"`
}
type SkillCalculation struct {
	Code        string `csv:"code"`
	Description string `csv:"*desc"`
}
