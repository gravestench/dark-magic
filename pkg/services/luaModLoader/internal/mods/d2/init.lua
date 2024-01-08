paths = require("d2/paths.lua")

function init()
    initRecords()

    -- we must invoke everything
    initUI()
end

function initUI()
    root = api.renderer.NewRenderable()

    root.AddChild()
end

local recordMap = {
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

function initRecords()
    -- load all the data records
    for key, path in recordMap do
        api.records[key] = api.records.Load(key, path)
    end
    print("loaded d2 records")
end