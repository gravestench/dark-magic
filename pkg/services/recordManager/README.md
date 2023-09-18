# Record Manager Service

The purpose of this [runtime](https://github.com/gravestench/runtime) service is to provide a single point of
access for all records, which are instances of models that are marshalled from
the diablo2 excel files (which are tab-separated value format).

This runtime service also exports the records into the lua environment
for scripting access.

## Dependencies

This service has dependencies on MPQ and TSV file loaders:

* [mpq loader service](../mpqLoader)
* [tsv loader service](../tsvLoader)

## Integration with other services

This service integrates with the following services:

* [lua service](../lua)
* [web router service](../webRouter)

The integration is optional; if either of the lua or web router services are
omitted from the runtime then the integration methods will never be called.

This service exports an integration interface `LoadsDiablo2Records` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see runtime.HasDependencies), and expose just the methods which
other services should use.

this is a rather large interface, but there are getters for all of the record
types, which are numerous.

```golang
type Dependency = LoadsDiablo2Records

type LoadsDiablo2Records interface {
    LoadRecords() error
    Belts() []models.BeltData
    CharStartingAttributes() []models.CharStats
    Inventory() []models.InventoryData
    Overlays() []models.Overlay
    PetTypes() []models.PetType
    AutoMapEntries() []models.AutoMapEntry
    States() []models.State
    Hirelings() []models.Hireling
    HirelingDescriptions() []models.HirelingDescription
    Missiles() []models.Missile
    DifficultyLevels() []models.Difficultylevel
    Shrines() []models.Shrine
    GambleRecords() []models.GambleRecord
    NpcTradeRecords() []models.NPCTrade
    ExperienceBreakpoints() []models.ExperienceData
    ItemArmor() []models.ItemArmor
    ItemWeapon() []models.ItemWeapon
    ItemWeaponClass() []models.WeaponClass
    ItemMisc() []models.MiscItem
    ItemTypes() []models.ItemType
    ItemAutoMagic() []models.AutoMagicData
    ItemStatCost() []models.ItemStatCost
    ItemRatio() []models.ItemRatio
    ItemUnique() []models.ItemUnique
    ItemHiQualityMods() []models.ItemHighQualityModifiers
    ItemProperties() []models.ItemProperty
    CubeRecipes() []models.CubeRecipe
    Books() []models.Book
    Gems() []models.GemData
    Runes() []models.RuneWordData
    SetItems() []models.SetItemData
    SetBonuses() []models.SetBonusData
    Skills() []models.SkillData
    SkillDesc() []models.SkillDescData
    Treasures() []models.TreasureClass
    TreasuresExpansion() []models.TreasureClassEx
    MagicPrefixes() []models.MagicPrefix
    MagicSuffixes() []models.MagicSuffix
    RarePrefixes() []models.RarePrefix
    RareSuffixes() []models.RareSuffix
    UniquePrefixes() []models.UniquePrefix
    UniqueSuffixes() []models.UniqueSuffix
    Objects() []models.Object
    ObjectTypes() []models.ObjectType
    ObjectGroups() []models.ObjectGroup
    ObjectModes() []models.ObjectMode
    Sounds() []models.SoundEntry
    SoundEnvironments() []models.SoundEnvironment
    LevelPresets() []models.LevelPreset
    LevelType() []models.LevelType
    LevelWarp() []models.LevelWarp
    LevelDetails() []models.LevelData
    LevelMaze() []models.LevelMazeData
    LevelSubstitutions() []models.LevelSubstitutionData
    MonsterUniqueModifiers() []models.MonsterUniqueModifier
    MonsterEquipment() []models.MonsterEquipment
    MonsterLevelStats() []models.MonsterLevelStats
    MonsterPresets() []models.MonsterPreset
    MonsterProperties() []models.MonsterProp
    MonsterSequences() []models.MonsterSequence
    MonsterStats() []models.MonsterStats
    MonsterStats2() []models.MonsterStats2
    MonsterSounds() []models.MonsterSounds
    MonsterUniqueNames() []models.MonsterUniqueName
}
```

## Lua service integration

A global `records` table is exported to lua. All of the getter-methods
which yield the array of records are available from within the lua environment.

Here is an example of its usage:

```lua
for i, sound in ipairs(records.Sounds) do
    print("Sound " .. i .. ": Sound=" .. sound.Sound .. " FileName=" .. sound.FileName)
end
```

## Web router service integration

If the [web router service](../webRouter) is present at runtime, this service will
register routes for retrieving data.

The route slug for this service is `records`, so all routes defined will be under
that route group.

| route                            | method | purpose                                                           |
|----------------------------------|--------|-------------------------------------------------------------------|
| `records/`                       | GET    | serves a web page with information about all of the records files |
| `records/Belts`                  | GET    | yields the Belts records array                                    |
| `records/CharStartingAttributes` | GET    | yields the CharStartingAttributes records array                   |
| `records/Inventory`              | GET    | yields the Inventory records array                                |
| `records/Overlays`               | GET    | yields the Overlays records array                                 |
| `records/PetTypes`               | GET    | yields the PetTypes records array                                 |
| `records/AutoMapEntries`         | GET    | yields the AutoMapEntries records array                           |
| `records/States`                 | GET    | yields the States records array                                   |
| `records/Hirelings`              | GET    | yields the Hirelings records array                                |
| `records/HirelingDescriptions`   | GET    | yields the HirelingDescriptions records array                     |
| `records/Missiles`               | GET    | yields the Missiles records array                                 |
| `records/DifficultyLevels`       | GET    | yields the DifficultyLevels records array                         |
| `records/Shrines`                | GET    | yields the Shrines records array                                  |
| `records/GambleRecords`          | GET    | yields the GambleRecords records array                            |
| `records/NpcTradeRecords`        | GET    | yields the NpcTradeRecords records array                          |
| `records/ExperienceBreakpoints`  | GET    | yields the ExperienceBreakpoints records array                    |
| `records/ItemArmor`              | GET    | yields the ItemArmor records array                                |
| `records/ItemWeapon`             | GET    | yields the ItemWeapon records array                               |
| `records/ItemWeaponClass`        | GET    | yields the ItemWeaponClass records array                          |
| `records/ItemMisc`               | GET    | yields the ItemMisc records array                                 |
| `records/ItemTypes`              | GET    | yields the ItemTypes records array                                |
| `records/ItemAutoMagic`          | GET    | yields the ItemAutoMagic records array                            |
| `records/ItemStatCost`           | GET    | yields the ItemStatCost records array                             |
| `records/ItemRatio`              | GET    | yields the ItemRatio records array                                |
| `records/ItemUnique`             | GET    | yields the ItemUnique records array                               |
| `records/ItemHiQualityMods`      | GET    | yields the ItemHiQualityMods records array                        |
| `records/ItemProperties`         | GET    | yields the ItemProperties records array                           |
| `records/CubeRecipes`            | GET    | yields the CubeRecipes records array                              |
| `records/Books`                  | GET    | yields the Books records array                                    |
| `records/Gems`                   | GET    | yields the Gems records array                                     |
| `records/Runes`                  | GET    | yields the Runes records array                                    |
| `records/SetItems`               | GET    | yields the SetItems records array                                 |
| `records/SetBonuses`             | GET    | yields the SetBonuses records array                               |
| `records/Skills`                 | GET    | yields the Skills records array                                   |
| `records/SkillDesc`              | GET    | yields the SkillDesc records array                                |
| `records/Treasures`              | GET    | yields the Treasures records array                                |
| `records/TreasuresEx`            | GET    | yields the TreasuresEx records array                              |
| `records/MagicPrefixes`          | GET    | yields the MagicPrefixes records array                            |
| `records/MagicSuffixes`          | GET    | yields the MagicSuffixes records array                            |
| `records/RarePrefixes`           | GET    | yields the RarePrefixes records array                             |
| `records/RareSuffixes`           | GET    | yields the RareSuffixes records array                             |
| `records/UniquePrefixes`         | GET    | yields the UniquePrefixes records array                           |
| `records/UniqueSuffixes`         | GET    | yields the UniqueSuffixes records array                           |
| `records/Objects`                | GET    | yields the Objects records array                                  |
| `records/ObjectTypes`            | GET    | yields the ObjectTypes records array                              |
| `records/ObjectGroups`           | GET    | yields the ObjectGroups records array                             |
| `records/ObjectModes`            | GET    | yields the ObjectModes records array                              |
| `records/Sounds`                 | GET    | yields the Sounds records array                                   |
| `records/SoundEnvironments`      | GET    | yields the SoundEnvironments records array                        |
| `records/LevelPresets`           | GET    | yields the LevelPresets records array                             |
| `records/LevelType`              | GET    | yields the LevelType records array                                |
| `records/LevelWarp`              | GET    | yields the LevelWarp records array                                |
| `records/LevelDetails`           | GET    | yields the LevelDetails records array                             |
| `records/LevelMaze`              | GET    | yields the LevelMaze records array                                |
| `records/LevelSubstitutions`     | GET    | yields the LevelSubstitutions records array                       |
| `records/MonsterUniqueModifiers` | GET    | yields the MonsterUniqueModifiers records array                   |
| `records/MonsterEquipment`       | GET    | yields the MonsterEquipment records array                         |
| `records/MonsterLevelStats`      | GET    | yields the MonsterLevelStats records array                        |
| `records/MonsterPresets`         | GET    | yields the MonsterPresets records array                           |
| `records/MonsterProperties`      | GET    | yields the MonsterProperties records array                        |
| `records/MonsterSequences`       | GET    | yields the MonsterSequences records array                         |
| `records/MonsterStats`           | GET    | yields the MonsterStats records array                             |
| `records/MonsterStats2`          | GET    | yields the MonsterStats2 records array                            |
| `records/MonsterSounds`          | GET    | yields the MonsterSounds records array                            |
| `records/MonsterUniqueNames`     | GET    | yields the MonsterUniqueNames records array                       | 