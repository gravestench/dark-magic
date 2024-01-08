paths = require("paths.lua")

-- init invokes all functionality for diablo 2
function init()
    initRecords()

    -- we must invoke everything
    initUI()
    initEvents()
end

-- initUI will create all game screens
function initUI()
    screen = api.ui.NewScreen()
    root = screen.Renderable()

    root.AddChild()
end

local recordsToLoad = {
    LevelPreset           = paths.PathLevelPreset,
    LevelType             = paths.PathLevelType,
    ObjectType            = paths.PathObjectType,
    LevelWarp             = paths.PathLevelWarp,
    LevelDetails          = paths.PathLevelDetails,
    LevelMaze             = paths.PathLevelMaze,
    LevelSubstitutions    = paths.PathLevelSubstitutions,
    ObjectDetails         = paths.PathObjectDetails,
    ObjectMode            = paths.PathObjectMode,
    ObjectGroup           = paths.PathObjectGroup,
    SoundSettings         = paths.PathSoundSettings,
    ItemStatCost          = paths.PathItemStatCost,
    ItemRatio             = paths.PathItemRatio,
    ItemTypes             = paths.PathItemTypes,
    QualityItems          = paths.PathQualityItems,
    Overlays              = paths.PathOverlays,
    Runes                 = paths.PathRunes,
    Sets                  = paths.PathSets,
    SetItems              = paths.PathSetItems,
    AutoMagic             = paths.PathAutoMagic,
    Properties            = paths.PathProperties,
    Hireling              = paths.PathHireling,
    HirelingDescription   = paths.PathHirelingDescription,
    DifficultyLevels      = paths.PathDifficultyLevels,
    AutoMap               = paths.PathAutoMap,
    CubeRecipes           = paths.PathCubeRecipes,
    Skills                = paths.PathSkills,
    SkillDesc             = paths.PathSkillDesc,
    TreasureClass         = paths.PathTreasureClass,
    TreasureClassEx       = paths.PathTreasureClassEx,
    States                = paths.PathStates,
    SoundEnvirons         = paths.PathSoundEnvirons,
    Shrines               = paths.PathShrines,
    PetType               = paths.PathPetType,
    NPC                   = paths.PathNPC,
    MonsterUniqueModifier = paths.PathMonsterUniqueModifier,
    MonsterEquipment      = paths.PathMonsterEquipment,
    UniqueAppellation     = paths.PathUniqueAppellation,
    MonsterLevel          = paths.PathMonsterLevel,
    MonsterPreset         = paths.PathMonsterPreset,
    MonsterProperties     = paths.PathMonsterProperties,
    MonsterStats          = paths.PathMonsterStats,
    MonsterStats2         = paths.PathMonsterStats2,
    MonsterSound          = paths.PathMonsterSound,
    MonsterSequence       = paths.PathMonsterSequence,
    Belts                 = paths.PathBelts,
    Gamble                = paths.PathGamble,
    Inventory             = paths.PathInventory,
    Weapons               = paths.PathWeapons,
    Armor                 = paths.PathArmor,
    WeaponClass           = paths.PathWeaponClass,
    Books                 = paths.PathBooks,
    Misc                  = paths.PathMisc,
    UniqueItems           = paths.PathUniqueItems,
    Gems                  = paths.PathGems,
    MagicPrefix           = paths.PathMagicPrefix,
    MagicSuffix           = paths.PathMagicSuffix,
    RarePrefix            = paths.PathRarePrefix,
    RareSuffix            = paths.PathRareSuffix,
    UniquePrefix          = paths.PathUniquePrefix,
    UniqueSuffix          = paths.PathUniqueSuffix,
    Experience            = paths.PathExperience,
    CharStats             = paths.PathCharStats,
    Missiles              = paths.PathMissiles,
}

-- initRecords loads generic records into the lua environment via `api.records`
-- the keys will be fields for the loaded records file, to be used elsewhere
-- like `api.records.Skills`
function initRecords()
    log("loading records")

    for key, path in recordsToLoad do
        api.records[key] = api.records.Load(key, path)
        numRecords = #api.records[key]
        msg = "loaded " + numRecords + " records", "key"
        log(msg, key, "path", path)
    end
end