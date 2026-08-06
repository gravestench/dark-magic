package models

// ObjectClientFunction represents the various functions an object can have.
type ObjectClientFunction int

const (
	// DoNothing The object does nothing.
	DoNothing ObjectClientFunction = iota

	// AmbientSoundAlwaysTrue Always return true for an ambient sound.
	AmbientSoundAlwaysTrue

	// Ripple Randomly plays between the Operating animation and loops back to the Neutral animation.
	Ripple

	// HellFire Same as function Ripple, but sound will also be processed.
	HellFire

	// Drinker Randomly plays between the Special 0 animation and loops back to the Neutral animation. Also processes sound.
	Drinker

	// Gesturer Randomly plays between the Special 0 / Special 1 animation and loops back to the Neutral animation. Also processes sound.
	Gesturer

	// Turner Randomly plays between the Special 0 animation and loops back to the Neutral animation. Uses different tick counts than function Drinker. Also processes sound.
	Turner

	// Skeleton Randomly plays between the Operating animation and loops back to the Neutral animation.
	Skeleton

	// DurielEntrance If the object is not in Neutral mode, then preload the Duriel monster.
	DurielEntrance

	// ClientSmoke Controls how the object can be removed from the client based on distance to a player and if the object has a specific tick count.
	ClientSmoke

	// Bubbles Randomly plays between the Operating animation and loops back to the Neutral animation. Uses different tick counts than function Skeleton.
	Bubbles

	// Floaters Always return true.
	Floaters

	// Altar If the object is not in Neutral mode, then preload the Ancients statues.
	Altar

	// InvisibleAncient If the object is in its Neutral mode and the player operating the object has not completed the Rite of Passage quest, then handle the control of operating the object.
	InvisibleAncient

	// Bonfire Updates the object’s animation modes based on the time of day.
	Bonfire

	// FrozenAnya If the object is in Neutral mode, then play the “npcalert” overlay.
	FrozenAnya

	// LastExit If the object is in its Operating mode, then modify the animation frames.
	LastExit

	// Zoo Handle the creation of monsters if monsters need to be created.
	Zoo

	// Keeper Randomly plays the “barbarian_grunt_small_1” sound.
	Keeper
)

// ObjectInitFunction represents the various functions an object can have.
type ObjectInitFunction int

const (
	// InitDoNothing The object does nothing.
	InitDoNothing ObjectInitFunction = iota

	// ObjectInitShrine General function for determining which type of Shrine function to pick for the Shrine object.
	// This also uses the “Parm0” field to define the Shrine Type:
	// · If equals 0, default to health shrine
	// · If equals 1, then use Health Shrine
	// · If equals 2, then use Mana Shrine
	// · If equals 3, then pick a random stats shrine with a 10% chance to spawn a surprise shrine
	ObjectInitShrine

	// ObjectInitTrappable Handle a random chance to give the object 1 of the 9 random traps.
	// This random chance depends on the area level’s monster level.
	ObjectInitTrappable

	// ObjectInitChest Run function 1, and also determine if the object should be Locked or not.
	// The random chance to make the object Locked depends on the area level’s monster level.
	ObjectInitChest

	// QuestObjectTowerTomeInit If The Forgotten Tower quest is active, then set the object to run in Special 0 Mode.
	QuestObjectTowerTomeInit

	_

	// QuestObjectStoneInit Sets the object’s mode to be Opened or Neutral, depending on the progress with the Portal to Tristram for the Search for Cain quest.
	QuestObjectStoneInit

	// QuestObjectGibbetInit Sets the object’s mode, depending on the progress with Cain’s Cage for the Search for Cain quest.
	QuestObjectGibbetInit

	// ObjectInitDungeonTorch Sets the object’s mode to Opened.
	ObjectInitDungeonTorch

	// Quest Object Inifuss Init Sets the object’s mode, depending on the progress with the Tree for the Search for Cain quest.
	QuestObjectInifussInit

	// ObjectInitBonfire If the current level is Act 1 Rogue Encampment, then tell the object to do a periodic skill, otherwise set the object mode to Opened.
	ObjectInitBonfire

	// ObjectInitTownPortal Initializes the object’s mode and adds the level ID as an attribute to keep track of.
	ObjectInitTownPortal

	// ObjectInitPermanentPortal Handles specific level transitions for permanent portals found throughout the game.
	ObjectInitPermanentPortal

	// QuestObjectStoneSoundInit Attaches the object to the Search for Cain quest functions.
	QuestObjectStoneSoundInit

	// ObjectInitDungeonTorch2 Sets the object’s mode to Operating.
	ObjectInitDungeonTorch2

	// QuestObjectMalusInit Attaches the object to the Tools of the Trade quest functions.
	QuestObjectMalusInit

	// ObjectInitWell Sets the object’s attributes for a well including the amount of charges.
	// This also uses the “Parm2” field to define the amount of Life healed.
	ObjectInitWell

	// ObjectInitWaypoint Handles setting up the waypoint mechanic to the object for the current area level.
	ObjectInitWaypoint

	// QuestObjectJerhyn1Init Handle where to place Jerhyn (near the palace entrance) based on Arcane Sanctuary quest progress.
	QuestObjectJerhyn1Init

	// QuestObjectJerhyn2Init Handle where to place Jerhyn (inside the palace) based on The Seven Tombs quest progress.
	QuestObjectJerhyn2Init

	// QuestObjectTaintedSunAltarInit Attaches the object to the Tainted Sun quest functions.
	QuestObjectTaintedSunAltarInit

	// QuestObjectSevenTombsReceptacleInit Setup the object to be a receptacle for the Horadric Staff, based on The Seven Tombs quest progress.
	QuestObjectSevenTombsReceptacleInit

	// ObjectInitFire Setup the object to act as fire.
	ObjectInitFire

	// QuestObjectLamEsensTomeInit Attaches the object to the Lam Esen’s Tome quest functions.
	QuestObjectLamEsensTomeInit

	// ObjectInitTrap1 Handle setting up the object frame count and making sure it has full stats.
	ObjectInitTrap1

	// QuestObjectGidbinnInit Attaches the object to the Blade of the Old Religion quest functions.
	QuestObjectGidbinnInit

	// TestObjectInit Sets the object’s mode to Operating.
	TestObjectInit

	// ObjectInitTrappablePoison Sets up the random chance of 333/1000 for the object to have a trap that creates a poison nova.
	ObjectInitTrappablePoison

	// ObjectInitGold Create a random amount of gold piles (between 1 to 10) in random locations around the object.
	ObjectInitGold

	// QuestObjectInitArcanePortal Setup the object to link area levels between the Palace Cellar Level 3 and the Arcane Sanctuary.
	QuestObjectInitArcanePortal

	// QuestObjectHaremBlockerInit Setup the object’s collision based on the Arcane Sanctuary quest progress.
	QuestObjectHaremBlockerInit

	// QuestObjectHoradricCubeChestInit Sets up information about the object.
	QuestObjectHoradricCubeChestInit

	// QuestObjectHoradricScrollChestInit Sets up information about the object.
	QuestObjectHoradricScrollChestInit

	// QuestObjectStaffOfKingsChestInit Sets up information about the object.
	QuestObjectStaffOfKingsChestInit

	// ObjectInitHellTorch Randomly set the object’s mode to Operating.
	ObjectInitHellTorch

	// Return null A placeholder value that does nothing.
	_

	// Return null A placeholder value that does nothing.
	_

	// QuestObjectDurielPassagewayInit Decide between setting the object’s mode to Opened or Neutral, based on the progress of the The Seven Tombs quest.
	QuestObjectDurielPassagewayInit

	// QuestObjectTyraelDoorInit Decide between setting the object’s mode to Opened or Neutral, based on the progress of the The Seven Tombs quest.
	QuestObjectTyraelDoorInit

	// QuestObjectGidbinnTownAltarInit Decide between setting the object’s mode to Opened or Neutral, based on the progress of the The Blade of the Old Religion quest.
	QuestObjectGidbinnTownAltarInit

	// Return null A placeholder value that does nothing.
	ReturnNull

	// QuestObjectBeneathTheCityStairsInit Decide between setting the object’s mode to Opened or Neutral, based on the progress of the Khalim’s Flail quest.
	QuestObjectBeneathTheCityStairsInit

	// QuestObjectBeneathTheCityLeverInit If the Khalim’s Flail quest is complete, then set the object’s mode to Opened.
	QuestObjectBeneathTheCityLeverInit

	// QuestObjectDarkWandererInit Create the “darkwanderer” monster and order to walk to the object’s location.
	// This depends on the players character save from having witnessed this event before.
	QuestObjectDarkWandererInit

	// QuestObjectInitHellGate Decide between setting the object’s mode to Opened or Neutral, based on the progress of the The Guardian quest.
	QuestObjectInitHellGate

	// QuestObjectMephistoBridgeInit Decide between setting the object’s mode to Opened or Neutral, based on the progress of the The Guardian quest.
	// If the object is not Opened, then also tell it to do its unique event.
	QuestObjectMephistoBridgeInit

	// ObjectTrappedSoulInit Determine where to spawn the “trappedsoul1” and “trappedsoul2” monster classes in the area level.
	ObjectTrappedSoulInit

	// QuestObjectForgottenTowerChestInit Decide between setting up the chest object, relying on the Forgotten Tower quest being in progress.
	QuestObjectForgottenTowerChestInit

	// QuestObjectSoulstoneForgeInit Decide between setting the object’s mode to Opened or Neutral, based on the progress of the Hell’s Forge quest.
	QuestObjectSoulstoneForgeInit

	// QuestObjectHratliStartInit Handle placing Hratli near the starting point of Act 3, based on the player’s Act 3 prologue progress.
	QuestObjectHratliStartInit

	// QuestObjectHratliEndInit Handle placing Hratli near his forge, if the player has progressed past the Act 3 prologue.
	QuestObjectHratliEndInit

	// ObjectJackInTheBoxInit If the object is in Opened or Opening mode, then tell the object to do a periodic item skill event.
	ObjectJackInTheBoxInit

	// QuestObjectNatalyaInit Handle placing Natalya at her location based on the player’s progress of The Guardian quest.
	QuestObjectNatalyaInit

	// QuestObjectMephistoDoorInit Handle setting the object to Opened mode based on the player’s progress of destroying the orb for The Blackened Temple quest.
	QuestObjectMephistoDoorInit

	// QuestObjectCainStartInit Handle creating the Cain unit in the Rogue Encampment based on the player’s progress of The Search for Cain quest.
	QuestObjectCainStartInit

	// QuestObjectDiabloStartInit Handle the spawning event of Diablo based on the player’s progress of activating the seal objects in the Chaos Sanctuary.
	QuestObjectDiabloStartInit

	// QuestObjectDiabloSealInit Do nothing.
	QuestObjectDiabloSealInit

	// ObjectInitBetterChest Initialize the chest object and give it the special magical property.
	ObjectInitBetterChest

	// ObjectInitFissure Tell the object to do a periodic skill event at random times.
	ObjectInitFissure

	// ObjectVileDoggieInit If the object is in Neutral mode, then set the object to Operating mode and tell it to do a unique event.
	ObjectVileDoggieInit

	// QuestObjectCompellingOrbInit Set the object to Opened based on the progress of The Blackened Temple quest.
	QuestObjectCompellingOrbInit

	// QuestObjectCainPortalInit Set the object to Operating mode and tell it to do a unique event.
	QuestObjectCainPortalInit

	// QuestCagedWussie1Init Spawn the “act5pow” units based on the player’s progress of the Rescue on Mount Arreat quest.
	QuestCagedWussie1Init

	// QuestMoeInit Setup the Korlic statue object with quest data based on the Right of Passage quest progress.
	QuestMoeInit

	// QuestLarryInit Setup the Madawc statue object with quest data based on the Right of Passage quest progress.
	QuestLarryInit

	// QuestCurlyInit Setup the Talic statue object with quest data based on the Right of Passage quest progress.
	QuestCurlyInit

	// QuestAnyaInsideTownInit Handle for creating the Anya NPC in town, based on the progress of the Prison of Ice quest.
	QuestAnyaInsideTownInit

	// QuestAnyaOutsideTownInit Handle this object during the progress of the Prison of Ice quest and tell it to do its unique event.
	QuestAnyaOutsideTownInit

	// QuestNihlathakInsideTownInit Create the Nihlathak NPC in town, based on the progress of the Prison of Ice quest.
	QuestNihlathakInsideTownInit

	// QuestNihlathakOutsideTownInit Create the “Nihlathak Boss” super unique monster, based on the progress of the Prison of Ice quest.
	QuestNihlathakOutsideTownInit

	// QuestLarzukStartInit Do nothing.
	QuestLarzukStartInit

	// QuestLarzukEndInit Object placeholder to create the “Larzuk” NPC in town.
	QuestLarzukEndInit

	// QuestAncientTomeInit Set the tome object mode to Opened or Neutral based on the progress of The Rite of Passage quest.
	QuestAncientTomeInit

	// QuestAncientGatewayInit Set the door object mode to Opened or Neutral based on the progress of The Rite of Passage quest.
	QuestAncientGatewayInit

	// QuestFrozenAnyaInit Handle this object during the progress of the Prison of Ice quest and tell it to do its unique event.
	QuestFrozenAnyaInit

	// QuestLastExitInit Set the Throne of Destruction exit object mode to Operating or Opened based on the progress of the Eve of Destruction quest.
	QuestLastExitInit

	// QuestSummitDoorInit Set this door object mode to Operating or Opened based on the progress of the Rite of Passage quest.
	QuestSummitDoorInit

	// QuestPlayerLastPortalInit Set the last portal object mode to Operating or Opened based on the progress of the Eve of Destruction quest.
	QuestPlayerLastPortalInit

	// QuestTyraelPortalToExpansionInit Set this object mode to Operating or Opened based on the progress of the Terror’s End quest.
	QuestTyraelPortalToExpansionInit

	// QuestZooInit Attempt a random chance based on successfully selecting a “zoo” type monster from the entire list of possible monsters (See monstats.txt).
	// If selected, then send the quest update command to all players, based on the Eve of Destruction quest.
	QuestZooInit
)

// ObjectPopulateFunction represents the various functions the game will use to spawn objects.
type ObjectPopulateFunction int

const (
	// DoNotSpawn Do not spawn the object.
	DoNotSpawn ObjectPopulateFunction = iota

	// AddClumpedGroup Handles creating multiple of these objects randomly in a room,
	// based on the object’s size and Class. This function only handles specific object classes
	// such as caskets, urns, and baskets.
	AddClumpedGroup

	// AddSingleShrine Handles the creation of a shrine object.
	AddSingleShrine

	// AddSimpleObjects Handles randomly spawning the object in a room, based on the object’s size.
	AddSimpleObjects

	// AddBarrels Handle creating multiple barrel or exploding barrel Class objects in a room.
	AddBarrels

	// AddCrates Handle creating multiple crate or urn Class objects in a room.
	AddCrates

	// AddCorpse Use function 3 to handle spawning the object.
	// Also call a random chance to spawn the “Flies” object class on top of the objects that spawn.
	AddCorpse

	// AddStakedCorpses Handles how to specifically spawn the “RogueCorpse1” and “RogueCorpse2” objects,
	// based on their sizes and the locations in the room.
	// Also call a random chance to spawn the “Flies” object class on top of the objects that spawn.
	AddStakedCorpses

	// AddWell Handles the creation of one of these objects randomly in a room based on the object’s size.
	// A level can have a maximum of 4 these objects that spawn.
	AddWell

	// AddOne Handles the creation of one of these objects randomly in a room based on the object’s size.
	AddOne
)

// ObjectOperateFunction represents the various functions that the game will use when the player clicks on an object.
type ObjectOperateFunction int

const (
	// DoNotOperate Do nothing.
	DoNotOperate ObjectOperateFunction = iota

	// SpawnItemAndMaybeMonster General function to operate an object and spawn items.
	// Also can randomly spawn a monster and/or trigger a trap.
	SpawnItemAndMaybeMonster

	// Shrine General function for Shrine objects. Uses fields from the shrines.txt file for determining specific Shrine functions.
	ShrineOperate

	// SpawnItemSometimes General function to operate the object and spawn random items.
	// Has a 20% chance to spawn a random item. Can also randomly trigger a trap.
	SpawnItemSometimes

	// ChestOperate General function for opening chest objects and spawning random items.
	// Handles key interaction functionality if the chest object is locked.
	ChestOperate

	// BarrelOperate General function for breaking barrel objects and randomly spawning items or possibly a monster.
	BarrelOperate

	// QuestTomeOperate Handles updating The Forgotten Tower quest progress.
	QuestTomeOperate

	// BarrelExplodingOperate Explode the object and also explode adjacent Exploding Barrel object Classes.
	BarrelExplodingOperate

	// DoorOperate General function for opening and closing door objects.
	DoorOperate

	// QuestCairnStoneOperate Handle operating the 5 Cairn Stone objects based on the player’s progress for the Search for Cain quest
	// and if the player has the deciphered Scroll of Inifuss item. Also removes the Scroll of Inifuss item once successfully operated.
	QuestCairnStoneOperate

	// QuestGibbetOperate Handle operating the object and updating the player’s progress for the Search for Cain quest.
	// This is used for the cage object that Deckard Cain is trapped in.
	QuestGibbetOperate

	// BrazierOperate Switch the object from Neutral mode to Operating/Opened mode, or vice versa.
	BrazierOperate

	// QuestInifussOperate Handle dropping the Bark Scroll item, based on the player’s progress for the Search for Cain quest.
	QuestInifussOperate

	// TikiOperate Switch the object from Neutral mode to Operating mode, or vice versa.
	TikiOperate

	// SpawnItem General function to operate an object and have it spawn random items.
	// Can also remove the object’s collision and randomly trigger a trap.
	SpawnItem

	// TownPortalOperate Controls the Town Portal functionalities, including how to teleport players back to town or to the current level,
	// and handling how players interact with other player Town Portals.
	TownPortalOperate

	// TrapDoorOperate Open a door type object and then control its level warp capabilities.
	TrapDoorOperate

	// Obelisk1 Use the transaction UI if the player has a gem in their inventory, and operate the object.
	Obelisk1

	// SecretDoorOperate Handle operating an object and removing its collision.
	SecretDoorOperate

	// ArmorRackOperate Activate the object to spawn a random armor item.
	ArmorRackOperate

	// WeaponRackOperate Activate the object to spawn a random weapon item.
	WeaponRackOperate

	// QuestMalusOperate Handle dropping the Horadric Malus item, based on the player’s progress for the Tools of the Trade quest.
	QuestMalusOperate

	// WellOperate Handle healing the player and keeping track of the charges and regeneration of charges for the well object.
	WellOperate

	// WaypointOperate Handle activating a waypoint object and using the Waypoint UI when clicking on an activated waypoint object.
	WaypointOperate

	// QuestTaintedSunAltarOperate Create the Amulet of the Viper item and other treasure items based on The Horadric Staff quest progress
	// and the number of players in the game. Also update the progress for the Tainted Sun quest.
	QuestTaintedSunAltarOperate

	// QuestSevenTombsReceptacleOperate Handle using the Horadric Staff item with the transaction UI to operate the object.
	QuestSevenTombsReceptacleOperate

	// BookshelfOperate Randomly create either tomes or scrolls of Identify or Town Portal.
	BookshelfOperate

	// TeleportPadOperate Teleport the player to another part of the room.
	TeleportPadOperate

	// QuestLamEsenTomeOperate Handle dropping the Lam Esen’s Tome item, based on the player’s progress for the Lam Esen’s Tome quest.
	QuestLamEsenTomeOperate

	// BreakableOperate Animate the object and remove its collision.
	BreakableOperate

	// Exploding Create an explosion around the object.
	Exploding

	// QuestGidbinnOperate Handle dropping the Decoy Gidbinn item, based on the player’s progress for the Blade of the Old Religion quest.
	QuestGidbinnOperate

	// PlayerBankOperate Control accessing the Stash UI while in town for the Bank object Class.
	PlayerBankOperate

	// WirtSpurt Create the Wirt’s leg item and animate the object.
	WirtSpurt

	// ArcanePortal Control how the warp object transitions the player from the Palace Cellar Level 3 to the Arcane Sanctuary.
	ArcanePortal

	// ReturnNull Placeholder to return null.
	_

	// QuestHoradricCubeChestOperate Create the Horadric Cube item and other treasure items based on The Horadric Staff quest progress
	// and the number of players in the game.
	QuestHoradricCubeChestOperate

	// QuestHoradricScrollChestOperate Create the Horadric Scroll item and other treasure items based on The Horadric Staff quest progress
	// and the number of players in the game.
	QuestHoradricScrollChestOperate

	// QuestStaffOfKingsChestOperate Create the Staff of Kings item and other treasure items based on The Horadric Staff quest progress
	// and the number of players in the game.
	QuestStaffOfKingsChestOperate

	// QuestArcaneTomeOperate Handles updating The Arcane Sanctuary quest progress.
	QuestArcaneTomeOperate

	// OneWayPortalOperate Controls the functionalities of the “DurielPortal” one way warp object.
	OneWayPortalOperate

	// QuestBeneathTheCityStairsOperate Handles warp object operates based on the Khalim’s Flail quest progress.
	QuestBeneathTheCityStairsOperate

	// QuestBeneathTheCityLeverOperate Handles operating an object based on the Khalim’s Flail quest progress.
	QuestBeneathTheCityLeverOperate

	// HellGateOperate Handles how to transition the player to Act 4 based on The Guardian quest progress.
	HellGateOperate

	// StairsOperate Handles how the stairs object opens or warp the player to another level.
	StairsOperate

	// JackInTheBoxOperate Handles the operating the object and having it spawn items and set its mode to Special 2.
	JackInTheBoxOperate

	// QuestSoulstoneForgeOperate Handle operating the object based on The Hellforge quest progress and how it spawns items.
	// Also remove the Hellforge Hammer weapon from the player.
	QuestSoulstoneForgeOperate

	// QuestMephistoDoorOperate Handles how the stairs object opens or warp the player to another level.
	QuestMephistoDoorOperate

	// DelaySpawnOperate Waits until the object is done operating before updating events.
	DelaySpawnOperate

	// QuestDiabloSealOperate Handle operating a Diablo Seal object while also tracking the progress on the other related Diablo Seal objects (5 in total).
	QuestDiabloSealOperate

	// QuestCompellingOrbOperate Handle operating the object based on the Khalim’s Flail quest progress and The Blackened Temple progress.
	// Also remove the Khalim’s Flail weapon from the player.
	QuestCompellingOrbOperate

	// QuestDiabloSeal1Operate Handle operating a Diablo Seal object Class and getting a spawn point for monsters. Also calls function 52.
	QuestDiabloSeal1Operate

	// QuestDiabloSeal3Operate Handle operating a Diablo Seal object Class and getting a spawn point for monsters. Also calls function 52.
	QuestDiabloSeal3Operate

	// QuestDiabloSeal5Operate Handle operating a Diablo Seal object Class and getting a spawn point for monsters. Also calls function 52.
	QuestDiabloSeal5Operate

	// QuestKhalimHeartChestOperate Create the Khalim’s Heart item and other treasure items based on the Khalim’s Flail quest progress and the number of players in the game.
	QuestKhalimHeartChestOperate

	// QuestKhalimEyeChestOperate Create the Khalim’s Eye item and other treasure items based on the Khalim’s Flail quest progress and the number of players in the game.
	QuestKhalimEyeChestOperate

	// QuestKhalimBrainChestOperate Create the Khalim’s Brain item and other treasure items based on the Khalim’s Flail quest progress and the number of players in the game.
	QuestKhalimBrainChestOperate

	// ReturnNull Placeholder to return null.
	ReturnNull2

	// TownGate Handles how the gate object opens and closes.
	TownGate

	// AncientStatue1Operate Handles the modes of one of the Ancient’s statues based on the player’s progress of the Rite of Passage quest.
	AncientStatue1Operate

	// AncientStatue2Operate Same as function 62. Handles the modes of one of the Ancient’s statues based on the player’s progress of the Rite of Passage quest.
	AncientStatue2Operate

	// AncientStatue3Operate Same as function 62. Handles the modes of one of the Ancient’s statues based on the player’s progress of the Rite of Passage quest.
	AncientStatue3Operate

	// QuestAncientAltarOperate Handle displaying quest text and disabling the player’s town portals,
	// based on the player’s progress of the Rite of Passage quest.
	QuestAncientAltarOperate

	// QuestAncientGatewayOperate Handle opening the door object based on the player’s progress of the Rite of Passage quest.
	QuestAncientGatewayOperate

	// QuestFrozenAnyaOperate Handles the object displaying quest text or validating that the player has the Malah’s Potion item
	// and updating the Prison of Ice quest.
	QuestFrozenAnyaOperate

	// EvilUrn Handle triggering a trap from the object.
	EvilUrn

	// QuestAncientInvisibleOperate Handle displaying the A5Q6InitAncients string conversation text based on the player’s progress of the Rite of Passage quest.
	QuestAncientInvisibleOperate

	// QuestLastExitOperate Handle transitioning the player to the from the Throne of Destruction level to the Worldstone Chamber level.
	QuestLastExitOperate

	// QuestSummitDoorOperate Handle opening the door object based on the player’s progress of the Rite of Passage quest.
	QuestSummitDoorOperate

	// QuestPlayerLastPortalOperate Handle transitioning the player to completing the game after completing the Destruction’s End quest.
	QuestPlayerLastPortalOperate

	// QuestTyraelPortalToExpansionOperate Handle transitioning the player to Act 5 after completing the Act 4 Terror’s End quest.
	QuestTyraelPortalToExpansionOperate
)
