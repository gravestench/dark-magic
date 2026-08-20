package models

// NPC represents the NPCs in each Act that have specific item data.
type NPC string

// NPCTrade represents the data for how each town NPC manipulates their store prices.
type NPCTrade struct {
	// Points to the matching "Id" value in the monstats.txt file. Should not be changed.
	NPC string `csv:"npc"`
	// Used to calculate the item's price when bought by the NPC from the player. [cost] * [buy mult] / 1024
	BuyMult int `csv:"buy mult"`
	// Used to calculate the item's price when sold by the NPC to the player. [cost] * [sell mult] / 1024
	SellMult int `csv:"sell mult"`
	// Used to calculate the cost to repair an item. [cost] * [rep mult] / 1024. Influences repair cost based on item
	// durability and charges.
	RepMult int `csv:"rep mult"`
	// Relative additional price calculations based on the player's quest flag progress (Quest Progress).
	QuestFlagA int `csv:"questflag A"`
	QuestFlagB int `csv:"questflag B"` // Additional quest flag progress (Quest Progress).
	QuestFlagC int `csv:"questflag C"` // Additional quest flag progress (Quest Progress).
	// Same functionality as "buy mult", but relies on "questflag" and applies after "buy mult" calculation.
	QuestBuyMultA int `csv:"questbuymult A"`
	// Same functionality as "buy mult", but relies on "questflag" and applies after "buy mult" calculation.
	QuestBuyMultB int `csv:"questbuymult B"`
	// Same functionality as "buy mult", but relies on "questflag" and applies after "buy mult" calculation.
	QuestBuyMultC int `csv:"questbuymult C"`
	// Same functionality as "sell mult", but relies on "questflag" and applies after "sell mult" calculation.
	QuestSellMultA int `csv:"questsellmult A"`
	// Same functionality as "sell mult", but relies on "questflag" and applies after "sell mult" calculation.
	QuestSellMultB int `csv:"questsellmult B"`
	// Same functionality as "sell mult", but relies on "questflag" and applies after "sell mult" calculation.
	QuestSellMultC int `csv:"questsellmult C"`
	// Same functionality as "rep mult", but relies on "questflag" and applies after "rep mult" calculation.
	QuestRepMultA int `csv:"questrepmult A"`
	// Same functionality as "rep mult", but relies on "questflag" and applies after "rep mult" calculation.
	QuestRepMultB int `csv:"questrepmult B"`
	// Same functionality as "rep mult", but relies on "questflag" and applies after "rep mult" calculation.
	QuestRepMultC int `csv:"questrepmult C"`
	// Sets the maximum price that the NPC will pay when the player sells an item in Normal Difficulty.
	MaxBuy int `csv:"max buy"`
	// Sets the maximum price that the NPC will pay when the player sells an item in Nightmare Difficulty.
	MaxBuyN int `csv:"max buy (N)"`
	// Sets the maximum price that the NPC will pay when the player sells an item in Hell Difficulty.
	MaxBuyH int `csv:"max buy (H)"`
}
