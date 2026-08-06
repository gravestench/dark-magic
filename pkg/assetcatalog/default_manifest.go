package assetcatalog

const (
	paletteSky     = "data/global/Palette/sky/pal.dat"
	paletteUnits   = "data/global/Palette/units/pal.dat"
	paletteFechar  = "data/global/Palette/fechar/pal.dat"
	paletteLoading = "data/global/Palette/loading/pal.dat"
)

// DefaultManifest is the first curated set of high-value screen hypotheses.
// It intentionally includes disputed paths so a scan records which files are
// actually present instead of turning one project's assumption into doctrine.
func DefaultManifest() Manifest {
	ref := func(names ...string) []string { return names }
	manifest := Manifest{Version: 1, Assets: []Hypothesis{
		{ID: "main-game-select-exp", Screen: "main-menu", Path: "data/global/ui/FrontEnd/gameselectscreenEXP.dc6", Palette: paletteSky, Meaning: "expansion game-select background", References: ref("OpenDiablo2")},
		{ID: "main-title-screen", Screen: "main-menu", Path: "data/global/ui/FrontEnd/TitleScreen.dc6", Palette: paletteSky, Meaning: "title background candidate", References: ref("Community research")},
		{ID: "main-trademark-exp", Screen: "main-menu", Path: "data/global/ui/FrontEnd/trademarkscreenEXP.dc6", Palette: paletteSky, Meaning: "expansion trademark screen", References: ref("Community research")},
		{ID: "main-logo-black-left", Screen: "main-menu", Path: "data/global/ui/FrontEnd/D2logoBlackLeft.DC6", Palette: paletteUnits, Meaning: "left opaque logo layer", References: ref("OpenDiablo2", "Community research")},
		{ID: "main-logo-fire-left", Screen: "main-menu", Path: "data/global/ui/FrontEnd/D2logoFireLeft.DC6", Palette: paletteUnits, Meaning: "left luminous animated logo layer", References: ref("OpenDiablo2", "Community research")},
		{ID: "main-logo-black-right", Screen: "main-menu", Path: "data/global/ui/FrontEnd/D2logoBlackRight.DC6", Palette: paletteUnits, Meaning: "right opaque logo layer", References: ref("OpenDiablo2", "Community research")},
		{ID: "main-logo-fire-right", Screen: "main-menu", Path: "data/global/ui/FrontEnd/D2logoFireRight.DC6", Palette: paletteUnits, Meaning: "right luminous animated logo layer", References: ref("OpenDiablo2", "Community research")},
		{ID: "button-wide", Screen: "front-end-common", Path: "data/global/ui/FrontEnd/WideButtonBlank.dc6", Palette: paletteUnits, Meaning: "wide blank button states", References: ref("OpenDiablo2")},
		{ID: "button-3wide", Screen: "front-end-common", Path: "data/global/ui/FrontEnd/3WideButtonBlank.dc6", Palette: paletteUnits, Meaning: "three-wide blank button states", References: ref("Community research")},
		{ID: "button-medium", Screen: "front-end-common", Path: "data/global/ui/FrontEnd/MediumButtonBlank.dc6", Palette: paletteUnits, Meaning: "medium blank button states", References: ref("OpenDiablo2", "Community research")},
		{ID: "popup-ok-cancel", Screen: "front-end-common", Path: "data/global/ui/FrontEnd/PopUpOKCancel.dc6", Palette: paletteFechar, Meaning: "confirmation popup", References: ref("OpenDiablo2")},
		{ID: "create-background-exp", Screen: "character-create", Path: "data/global/ui/FrontEnd/charactercreationscreenEXP.dc6", Palette: paletteSky, Meaning: "expansion character creation background", References: ref("OpenDiablo2")},
		{ID: "create-campfire", Screen: "character-create", Path: "data/global/ui/FrontEnd/fire.DC6", Palette: paletteUnits, Meaning: "character creation campfire", References: ref("OpenDiablo2")},
		{ID: "select-background-exp", Screen: "character-select", Path: "data/global/ui/CharSelect/characterselectscreenEXP.dc6", Palette: paletteSky, Meaning: "expansion saved-character list background", References: ref("OpenDiablo2")},
		{ID: "select-box", Screen: "character-select", Path: "data/global/ui/CharSelect/charselectbox.dc6", Palette: paletteSky, Meaning: "saved-character selection highlight", References: ref("OpenDiablo2")},
		{ID: "loading-global", Screen: "loading", Path: "data/global/ui/Loading/loadingscreen.dc6", Palette: paletteLoading, Meaning: "global loading progress frames", References: ref("OpenDiablo2")},
		{ID: "loading-local", Screen: "loading", Path: "data/local/ui/loadingscreen.dc6", Palette: paletteLoading, Meaning: "local loading progress frames candidate", References: ref("Community research")},
		{ID: "hud-control-panel", Screen: "game-hud", Path: "data/global/ui/PANEL/800ctrlpnl7.dc6", Palette: paletteSky, Meaning: "800x600 control panel pieces", References: ref("OpenDiablo2", "Community research")},
		{ID: "hud-overlap", Screen: "game-hud", Path: "data/global/ui/PANEL/overlap.DC6", Palette: paletteSky, Meaning: "health and mana globe overlap", References: ref("OpenDiablo2")},
		{ID: "hud-health-mana", Screen: "game-hud", Path: "data/global/ui/PANEL/hlthmana.DC6", Palette: paletteSky, Meaning: "health and mana numeric indicator", References: ref("OpenDiablo2")},
		{ID: "hud-mini-buttons", Screen: "game-hud", Path: "data/global/ui/PANEL/minipanelbtn.DC6", Palette: paletteSky, Meaning: "mini-panel button state pairs", References: ref("OpenDiablo2", "Community research")},
		{ID: "inventory-character", Screen: "inventory", Path: "data/global/ui/PANEL/invchar6.DC6", Palette: paletteSky, Meaning: "character inventory panel", References: ref("OpenDiablo2", "Community research")},
		{ID: "inventory-weapon-tabs", Screen: "inventory", Path: "data/global/ui/PANEL/invchar6Tab.DC6", Palette: paletteSky, Meaning: "alternate weapon tabs", References: ref("OpenDiablo2", "Community research")},
		{ID: "inventory-buy-sell-button", Screen: "inventory", Path: "data/global/ui/PANEL/buysellbtn.DC6", Palette: paletteSky, Meaning: "inventory and vendor button states", References: ref("OpenDiablo2", "Community research")},
		{ID: "quest-background", Screen: "quest-log", Path: "data/global/ui/MENU/questbackground.dc6", Palette: paletteSky, Meaning: "quest log panel pieces", References: ref("OpenDiablo2")},
		{ID: "quest-tabs-exp", Screen: "quest-log", Path: "data/global/ui/MENU/expquesttabs.dc6", Palette: paletteSky, Meaning: "expansion quest tab states", References: ref("OpenDiablo2")},
		{ID: "quest-sockets", Screen: "quest-log", Path: "data/global/ui/MENU/questsockets.dc6", Palette: paletteSky, Meaning: "quest socket states", References: ref("OpenDiablo2")},
		{ID: "waypoint-background", Screen: "waypoint", Path: "data/global/ui/MENU/waygatebackground.dc6", Palette: paletteSky, Meaning: "waypoint panel background", References: ref("OpenDiablo2")},
		{ID: "waypoint-tabs-exp", Screen: "waypoint", Path: "data/global/ui/MENU/expwaygatetabs.dc6", Palette: paletteSky, Meaning: "expansion waypoint tabs", References: ref("OpenDiablo2")},
		{ID: "waypoint-icons", Screen: "waypoint", Path: "data/global/ui/MENU/waygateicons.dc6", Palette: paletteSky, Meaning: "waypoint icons", References: ref("OpenDiablo2")},
	}}
	manifest.Assets = append(manifest.Assets, characterCreationHypotheses(ref)...)
	manifest.Assets = append(manifest.Assets,
		Hypothesis{ID: "inventory-armor-slot", Screen: "inventory", Path: "data/global/ui/PANEL/inv_armor.DC6", Palette: paletteSky, Meaning: "armor equipment placeholder", References: ref("OpenDiablo2", "Community research")},
		Hypothesis{ID: "inventory-belt-slot", Screen: "inventory", Path: "data/global/ui/PANEL/inv_belt.DC6", Palette: paletteSky, Meaning: "belt equipment placeholder", References: ref("OpenDiablo2", "Community research")},
		Hypothesis{ID: "inventory-boots-slot", Screen: "inventory", Path: "data/global/ui/PANEL/inv_boots.DC6", Palette: paletteSky, Meaning: "boots equipment placeholder", References: ref("OpenDiablo2", "Community research")},
		Hypothesis{ID: "inventory-helm-glove-slot", Screen: "inventory", Path: "data/global/ui/PANEL/inv_helm_glove.DC6", Palette: paletteSky, Meaning: "helm and glove equipment placeholders", References: ref("OpenDiablo2", "Community research")},
		Hypothesis{ID: "inventory-ring-amulet-slot", Screen: "inventory", Path: "data/global/ui/PANEL/inv_ring_amulet.DC6", Palette: paletteSky, Meaning: "ring and amulet equipment placeholders", References: ref("OpenDiablo2", "Community research")},
		Hypothesis{ID: "inventory-weapons-slot", Screen: "inventory", Path: "data/global/ui/PANEL/inv_weapons.DC6", Palette: paletteSky, Meaning: "weapon equipment placeholders", References: ref("OpenDiablo2", "Community research")},
		Hypothesis{ID: "records-inventory", Screen: "inventory", Path: "data/global/excel/Inventory.txt", Meaning: "class inventory and equipment geometry", References: ref("OpenDiablo2", "Community research")},
		Hypothesis{ID: "records-sounds", Screen: "audio", Path: "data/global/excel/Sounds.txt", Meaning: "sound handles, files, groups, channels, and playback metadata", References: ref("OpenDiablo2", "Community research")},
		Hypothesis{ID: "records-sound-environment", Screen: "audio", Path: "data/global/excel/SoundEnviron.txt", Meaning: "area music, ambience, and environmental sound mapping", References: ref("OpenDiablo2")},
		Hypothesis{ID: "records-skill-description", Screen: "skills", Path: "data/global/excel/SkillDesc.txt", Meaning: "skill icon cells and localized description keys", References: ref("OpenDiablo2", "Community research")},
		Hypothesis{ID: "audio-title-introedit", Screen: "main-menu", Path: "data/global/music/introedit.wav", Meaning: "title music candidate", References: ref("OpenDiablo2")},
		Hypothesis{ID: "audio-title-diablo", Screen: "main-menu", Path: "data/global/music/Act4/diablo.wav", Meaning: "title music candidate", References: ref("Community research")},
		Hypothesis{ID: "audio-button", Screen: "front-end-common", Path: "data/global/sfx/cursor/button.wav", Meaning: "button interaction sound", References: ref("Community research")},
		Hypothesis{ID: "audio-select", Screen: "front-end-common", Path: "data/global/sfx/cursor/select.wav", Meaning: "selection confirmation sound", References: ref("Community research")},
		Hypothesis{ID: "video-blizzard", Screen: "startup", Path: "data/local/video/New_Bliz640x480.bik", Meaning: "Blizzard startup video", References: ref("Community research")},
		Hypothesis{ID: "video-blizzard-north", Screen: "startup", Path: "data/local/video/BlizNorth640x480.bik", Meaning: "Blizzard North startup video", References: ref("Community research")},
	)
	return manifest
}

func characterCreationHypotheses(ref func(...string) []string) []Hypothesis {
	type asset struct{ id, class, file, meaning string }
	assets := []asset{
		{"amazon-unselected", "amazon", "AMNU1.DC6", "unselected"}, {"amazon-hover", "amazon", "AMNU2.DC6", "hover"}, {"amazon-selected", "amazon", "AMNU3.DC6", "selected"}, {"amazon-forward", "amazon", "AMFW.DC6", "forward walk"}, {"amazon-forward-overlay", "amazon", "AMFWs.DC6", "forward walk overlay"}, {"amazon-back", "amazon", "AMBW.DC6", "back walk"},
		{"sorceress-unselected", "sorceress", "SONU1.DC6", "unselected"}, {"sorceress-hover", "sorceress", "SONU2.DC6", "hover"}, {"sorceress-selected", "sorceress", "SONU3.DC6", "selected"}, {"sorceress-selected-overlay", "sorceress", "SONU3s.DC6", "selected overlay"}, {"sorceress-forward", "sorceress", "SOFW.DC6", "forward walk"}, {"sorceress-forward-overlay", "sorceress", "SOFWs.DC6", "forward walk overlay"}, {"sorceress-back", "sorceress", "SOBW.DC6", "back walk"}, {"sorceress-back-overlay", "sorceress", "SOBWs.DC6", "back walk overlay"},
		{"necromancer-unselected", "necromancer", "NENU1.DC6", "unselected"}, {"necromancer-hover", "necromancer", "NENU2.DC6", "hover"}, {"necromancer-selected", "necromancer", "NENU3.DC6", "selected"}, {"necromancer-selected-overlay", "necromancer", "NENU3s.DC6", "selected overlay"}, {"necromancer-forward", "necromancer", "NEFW.DC6", "forward walk"}, {"necromancer-forward-overlay", "necromancer", "NEFWs.DC6", "forward walk overlay"}, {"necromancer-back", "necromancer", "NEBW.DC6", "back walk"}, {"necromancer-back-overlay", "necromancer", "NEBWs.DC6", "back walk overlay"},
		{"paladin-unselected", "paladin", "PANU1.DC6", "unselected"}, {"paladin-hover", "paladin", "PANU2.DC6", "hover"}, {"paladin-selected", "paladin", "PANU3.DC6", "selected"}, {"paladin-forward", "paladin", "PAFW.DC6", "forward walk"}, {"paladin-forward-overlay", "paladin", "PAFWs.DC6", "forward walk overlay"}, {"paladin-back", "paladin", "PABW.DC6", "back walk"},
		{"barbarian-unselected", "barbarian", "banu1.DC6", "unselected"}, {"barbarian-hover", "barbarian", "banu2.DC6", "hover"}, {"barbarian-selected", "barbarian", "banu3.DC6", "selected"}, {"barbarian-forward", "barbarian", "bafw.DC6", "forward walk"}, {"barbarian-forward-overlay", "barbarian", "BAFWs.DC6", "forward walk overlay"}, {"barbarian-back", "barbarian", "babw.DC6", "back walk"},
		{"assassin-unselected", "assassin", "ASNU1.DC6", "unselected"}, {"assassin-hover", "assassin", "ASNU2.DC6", "hover"}, {"assassin-selected", "assassin", "ASNU3.DC6", "selected"}, {"assassin-forward", "assassin", "ASFW.DC6", "forward walk"}, {"assassin-back", "assassin", "ASBW.DC6", "back walk"},
		{"druid-unselected", "druid", "DZNU1.dc6", "unselected"}, {"druid-hover", "druid", "DZNU2.dc6", "hover"}, {"druid-selected", "druid", "DZNU3.DC6", "selected"}, {"druid-forward", "druid", "DZFW.DC6", "forward walk"}, {"druid-back", "druid", "DZBW.DC6", "back walk"},
	}
	result := make([]Hypothesis, 0, len(assets))
	for _, item := range assets {
		result = append(result, Hypothesis{ID: "create-" + item.id, Screen: "character-create", Path: "data/global/ui/FrontEnd/" + item.class + "/" + item.file, Palette: paletteUnits, Meaning: item.class + " " + item.meaning + " animation", References: ref("OpenDiablo2")})
	}
	return result
}
