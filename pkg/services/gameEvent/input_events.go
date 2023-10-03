package service_template

// GameEvent represents an envent in the game engine
type GameEvent int

// Game events
const (
	ToggleGameMenu       GameEvent = iota
	ToggleCharacterPanel           // panel toggles
	ToggleInventoryPanel
	TogglePartyPanel
	ToggleSkillTreePanel
	ToggleHirelingPanel
	ToggleQuestLog
	ToggleHelpScreen
	ToggleChatOverlay
	ToggleMessageLog
	ToggleRightSkillSelector // these two are for left/right speed-skill panel toggles
	ToggleLeftSkillSelector
	ToggleAutomap
	CenterAutomap        // recenters the automap when opened
	FadeAutomap          // reduces the brightness of the map (not the players/npcs)
	TogglePartyOnAutomap // toggles the display of the party members on the automap
	ToggleNamesOnAutomap // toggles the display of party members names and npcs on the automap
	ToggleMiniMap
	UseSkill1 // there can be 16 hotkeys, each hotkey can have a skill assigned
	UseSkill2
	UseSkill3
	UseSkill4
	UseSkill5
	UseSkill6
	UseSkill7
	UseSkill8
	UseSkill9
	UseSkill10
	UseSkill11
	UseSkill12
	UseSkill13
	UseSkill14
	UseSkill15
	UseSkill16
	SelectPreviousSkill
	SelectNextSkill
	ToggleBelt
	UseBeltSlot1
	UseBeltSlot2
	UseBeltSlot3
	UseBeltSlot4
	SwapWeapons
	ToggleChatBox
	ToggleRunWalk
	SayHelp
	SayFollowMe
	SayThisIsForYou
	SayThanks
	SaySorry
	SayBye
	SayNowYouDie
	SayRetreat
	HoldRun // these events are fired while a player holds the corresponding key
	HoldStandStill
	HoldShowGroundItems
	HoldShowPortraits
	TakeScreenShot
	ClearScreen // closes all active menus/panels
	ClearMessages
)

func (e GameEvent) String() string {
	table := map[GameEvent]string{
		ToggleGameMenu:           "ToggleGameMenu",
		ToggleCharacterPanel:     "ToggleCharacterPanel",
		ToggleInventoryPanel:     "ToggleInventoryPanel",
		TogglePartyPanel:         "TogglePartyPanel",
		ToggleSkillTreePanel:     "ToggleSkillTreePanel",
		ToggleHirelingPanel:      "ToggleHirelingPanel",
		ToggleQuestLog:           "ToggleQuestLog",
		ToggleHelpScreen:         "ToggleHelpScreen",
		ToggleChatOverlay:        "ToggleChatOverlay",
		ToggleMessageLog:         "ToggleMessageLog",
		ToggleRightSkillSelector: "ToggleRightSkillSelector",
		ToggleLeftSkillSelector:  "ToggleLeftSkillSelector",
		ToggleAutomap:            "ToggleAutomap",
		CenterAutomap:            "CenterAutomap",
		FadeAutomap:              "FadeAutomap",
		TogglePartyOnAutomap:     "TogglePartyOnAutomap",
		ToggleNamesOnAutomap:     "ToggleNamesOnAutomap",
		ToggleMiniMap:            "ToggleMiniMap",
		UseSkill1:                "UseSkill1",
		UseSkill2:                "UseSkill2",
		UseSkill3:                "UseSkill3",
		UseSkill4:                "UseSkill4",
		UseSkill5:                "UseSkill5",
		UseSkill6:                "UseSkill6",
		UseSkill7:                "UseSkill7",
		UseSkill8:                "UseSkill8",
		UseSkill9:                "UseSkill9",
		UseSkill10:               "UseSkill10",
		UseSkill11:               "UseSkill11",
		UseSkill12:               "UseSkill12",
		UseSkill13:               "UseSkill13",
		UseSkill14:               "UseSkill14",
		UseSkill15:               "UseSkill15",
		UseSkill16:               "UseSkill16",
		SelectPreviousSkill:      "SelectPreviousSkill",
		SelectNextSkill:          "SelectNextSkill",
		ToggleBelt:               "ToggleBelt",
		UseBeltSlot1:             "UseBeltSlot1",
		UseBeltSlot2:             "UseBeltSlot2",
		UseBeltSlot3:             "UseBeltSlot3",
		UseBeltSlot4:             "UseBeltSlot4",
		SwapWeapons:              "SwapWeapons",
		ToggleChatBox:            "ToggleChatBox",
		ToggleRunWalk:            "ToggleRunWalk",
		SayHelp:                  "SayHelp",
		SayFollowMe:              "SayFollowMe",
		SayThisIsForYou:          "SayThisIsForYou",
		SayThanks:                "SayThanks",
		SaySorry:                 "SaySorry",
		SayBye:                   "SayBye",
		SayNowYouDie:             "SayNowYouDie",
		SayRetreat:               "SayRetreat",
		HoldRun:                  "HoldRun",
		HoldStandStill:           "HoldStandStill",
		HoldShowGroundItems:      "HoldShowGroundItems",
		HoldShowPortraits:        "HoldShowPortraits",
		TakeScreenShot:           "TakeScreenShot",
		ClearScreen:              "ClearScreen",
		ClearMessages:            "ClearMessages",
	}

	if str, found := table[e]; found {
		return str
	}

	return "UNKNOWN GAME EVENT"
}
