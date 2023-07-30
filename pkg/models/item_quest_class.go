package models

// QuestClass represents the quest class tied to the item.
type QuestClass int

const (
	QuestClassNotAQuestItem QuestClass = iota
	QuestClassAct1Prologue
	QuestClassDenOfEvil
	QuestClassSistersBurialGrounds
	QuestClassToolsOfTheTrade
	QuestClassTheSearchForCain
	// Add other quest classes here.
)
