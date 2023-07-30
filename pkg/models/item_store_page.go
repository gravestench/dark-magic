package models

// ItemStorePage represents the UI tab page on the NPC shop UI to display specific item types
type ItemStorePage string

const (
	StorePageArmor   ItemStorePage = "armo"
	StorePageWeapons ItemStorePage = "weap"
	StorePageMagic   ItemStorePage = "mag"
	StorePageMisc    ItemStorePage = "misc"
)
