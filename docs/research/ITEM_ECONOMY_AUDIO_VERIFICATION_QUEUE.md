# Item/economy/audio verification queue

This queue consolidates empirical probes for item generation/qualities, sockets/runewords, container effects, Cube recipes, vendor economy, quest items, and gameplay/environment audio.

## P0: architecture-shaping probes

- Build one end-to-end owned-data monster drop trace from TC entry through quality/concrete special/affixes/properties/item seed into ground item state.
- Pin TreasureClass NoDrop scaling for representative player/party counts and minion/hireling credited kills.
- Compare current Dark Magic M6 TC quality denominators/roll order against expansion 1.14d TreasureClass drop behavior across Unique/Set/Rare/Magic boundary vectors.
- Pin one direct/non-TC item creation path showing its distinct Unique->Rare->Set ordering and forced quality flags.
- Capture one socketed item save/runtime specimen proving socket capacity, child ordering, filler identities and runeword identity.
- Verify one known runeword recognition vector plus wrong-order/wrong-host/extra-filler holdouts.
- Capture exact charm source activation while moving the same charm among inventory/Cube/stash/held contexts.
- Extract enough CubeMain rows from mounted data to cover exact item, type, quality, quantity, grade, socket, ethereal, stat predicate and use-item output forms.
- Execute one Cube recipe in original/owned runtime and record exact consumed IDs/fields preserved/output ilvl/seed/quality.
- Trace one vendor item's generation, quote, purchase, sale/buyback, refresh and identity lifecycle.
- Trace one repair/recharge quote/result including durability/charges/reduced-prices state.
- Pin `Levels.SoundEnv` meaning for at least Rogue Encampment, Blood Moor, one indoor dungeon and one rainy outdoor level.
- Capture one `Sounds.txt` grouped variation including volume/pitch/delay/falloff/duplicate behavior with original client output observable enough to constrain units.

## P1: loot and qualities

- TC nested-level selection and quality-modifier inheritance.
- Negative Picks behavior and quantity semantics.
- MF diminishing-return integer rounding for Unique/Set/Rare and Magic.
- Owner + minion/hireling MF and Gold Find attribution.
- Unique no-limit/per-game occurrence behavior and fallback.
- Set concrete record election/fallback.
- Magic/rare affix count, group exclusion, automagic and staffmods.
- Crafted quality/affix count/output ilvl.
- Superior and low-quality modifier selection/ranges.
- Ethereal 5% generation, restrictions, damage/defense and durability transformation.
- Ground item placement and cleanup timing by quality/type/value.
- Multiplayer item visibility/ownership/pickup rules.

## P1: sockets, fillers and runewords

- Socket maximum by base item/ilvl and type.
- Random generation socket chance/count/difficulty caps.
- Larzuk socket count by item quality/ilvl/base.
- Cube socket recipe count/randomness.
- Gem weapon/armor/helm/shield property-set mapping.
- Jewel property transfer/restrictions.
- Rune sequence exact-length recognition.
- RuneWords allowed/excluded item-type inheritance.
- Runeword quality/version/ladder restrictions.
- Runeword property roll timing/seed.
- Individual rune source + runeword source stacking/removal.
- Unsocket child destruction/removal and preserved host fields.
- Legacy save/network nested-item order.

## P1: container-dependent effects

- Exact `ITEMS_IsCharmUsable` behavior by container/page/state.
- Active weapon set contribution changes.
- Broken/no-equip/requirements source suppression.
- Requirement dependency ordering/cycles.
- Max HP/mana/stamina current-resource adjustment when item sources change.
- Item-granted skills/tabs/all-skills source activation/removal.
- Charged-skill availability when source item is inactive/moved.
- Item aura activation/deactivation timing.
- Item event proc registration/unregistration.
- Set/socket/runeword dependent source activation.
- Unique charm possession/pickup restrictions in expansion 1.14d.

## P1: Cube

- Full CubeMain text grammar and compiled flags.
- Recipe authored-row matching order.
- Stable matching when duplicate eligible inputs exist.
- Stack quantity accounting and partial consumption.
- ladder/version/difficulty/class gates.
- operation codes and version-specific additions.
- ItemStatCost stat/base/bonus comparison operations.
- output level/plvl/ilvl calculation/rounding.
- copy-mods exact preserved fields.
- crafted output fixed/random properties.
- magic/rare reroll preservation/new seed behavior.
- normal/exceptional/elite upgrades.
- socket/unsocket/repair/recharge output flags.
- output property chance/min/max rolls.
- Cow Portal/quest world operation conditions.
- quest item difficulty checks.

## P1: vendors/economy

- exact transaction-cost formula and rounding for buy/sell/repair;
- quality/affix/set/unique/socket/stat contribution to price;
- NPC buy/sell multipliers and max-buy cap;
- reduced-prices stat and quest discounts;
- ethereal/indestructible repair behavior;
- broken and partially damaged repair costs;
- charged-skill recharge costs;
- throwable quantity/durability repair;
- vendor stock generation level/quality/affix policies;
- stock refresh triggers/timing/seed;
- sold-item buyback lifecycle;
- ordinary vendor inventory shared/per-player behavior;
- gamble offer generation/price/hidden quality and refresh;
- identify/heal service conditions/costs;
- gold limits/overflow/death/trade interactions;
- simultaneous multiplayer purchase/stale quote behavior.

## P1: quest items

- generation quality/ethereal/socket/vendor restrictions;
- quest-item difficulty provenance/checks;
- possession search containers by quest;
- duplicate/pickup prevention;
- lost/dropped item regeneration;
- ground cleanup exemption;
- party credit on pickup/consume/object use;
- vendor/hand-in restrictions;
- Cube quest recipe conditions;
- corpse/death/save/exit behavior;
- imported old-difficulty quest items;
- quest reward generation/placement;
- representative item drop/pickup/consume callbacks for each act.

## P1: gameplay/environment audio

- exact `Levels.SoundEnv` reference target and environment table/profile shape;
- `Sounds.txt` Block1/2/3 units/loop/environment semantics;
- pitch range units and selection;
- fade in/out units;
- Delay/Duration units and boundaries;
- Defer Inst/Stop Inst duplicate key and behavior;
- Compound window/volume behavior;
- Priority and voice-limit rules;
- Falloff distance units/curve;
- `Is2D`, `3dSpread`, Tracking and LFE behavior;
- Solo and MusicVol ducking/routing;
- AmbientScene versus AmbientEvent scheduling;
- music selection by level/act/scene and crossfade rules;
- rain/inside/outside environmental transitions;
- MonSounds fields and event mapping;
- UMonSound unique override behavior;
- object ambient/operate/destroy sound mapping;
- skill cast/effect and missile travel/impact cues;
- COF sound-key event timing;
- player footsteps and surface-material mapping;
- NPC localized speech interruption/subtitle behavior;
- active-room/culling behavior for tracking loops;
- remote multiplayer cue derivation/replication;
- deterministic audio capture seed strategy independent of gameplay RNG.

## P2: modern audio extension probes

These are optional engine-enhancement experiments, not Diablo compatibility blockers:

- semantic collision-based occlusion;
- indoor/outdoor reverb/low-pass profiles;
- room-connected acoustic propagation;
- richer 3D/HRTF backend behind the same audio cue boundary;
- adaptive layerable music/soundscapes for mods.

Do not let these delay reproducing the simpler legacy cue/environment behavior.

## Probe artifact standard

Each completed probe should record:

```text
probe ID/title
target runtime/content version
owned inputs/assets required
exact reproduction steps
raw capture/log/byte differences
normalized input/output vector
RNG/timing state where relevant
conflicts with older secondary sources, if relevant
confidence upgrade
primary research document updated
```

Keep proprietary captures outside Git. Commit normalized/synthetic vectors, hashes, inspectors, and analysis artifacts that are safe to redistribute.
