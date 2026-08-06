package models

type ObjectTargetPriority int

const (
	NoChange ObjectTargetPriority = iota
	EqualCorpseWhenOpened
	AlwaysEqualCorpse
)
