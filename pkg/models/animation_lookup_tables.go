package models

type PlayerClassCode struct {
	Name string `csv:"Player Class"`
	Code string `csv:"Code"`
}

type PlayerMode struct {
	Name  string `csv:"Name"`
	Token string `csv:"Token"`
}
type PlayerType struct {
	Name  string `csv:"Name"`
	Token string `csv:"Token"`
}
type MonsterMode struct {
	Name  string `csv:"Name"`
	Token string `csv:"Token"`
}
type MonsterPlace struct {
	Code string `csv:"code"`
}
