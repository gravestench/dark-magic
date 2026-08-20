package clientapp

import (
	"fmt"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// registerRemoteViewSchemas installs the same allowlisted component shapes needed by connected
// projection while omitting unrelated production systems.
func registerRemoteViewSchemas(t *testing.T, engine *gameecs.Engine) {
	t.Helper()

	for _, schema := range remoteViewTestSchemas() {
		if _, err := akara.RegisterSchema(engine.World(), schema); err != nil {
			t.Fatal(err)
		}
	}
}

// remoteViewTestSchemas declares projection schemas explicitly so tests fail when production writes a
// field that was not intentionally exposed.
func remoteViewTestSchemas() []akara.Schema {
	schemas := remoteIdentityAndMonsterTestSchemas()
	schemas = append(schemas, remotePlayerAndPresentationTestSchemas()...)
	schemas = append(schemas, remoteItemAndInteractionTestSchemas()...)
	schemas = append(schemas, remoteWorldTestSchemas()...)

	return schemas
}

// remoteIdentityAndMonsterTestSchemas defines public identity, appearance, combat, and death-cue boundaries.
func remoteIdentityAndMonsterTestSchemas() []akara.Schema {
	return []akara.Schema{
		remoteTestSchema("d2legacy.player.identity",
			remoteTestFields(akara.FieldString, "character_id", "player", "name", "class")),
		remoteTestSchema("d2legacy.monster.identity",
			remoteTestFields(
				akara.FieldString,
				"spawn_id", "definition_id", "base_id", "graphics_id", "seed", "treasure_class",
			)),
		remoteTestSchema("d2legacy.monster.appearance",
			remoteTestFields(
				akara.FieldString,
				"token", "mode", "weapon_class", "name_key", "components", "death_sound",
			),
			remoteTestFields(akara.FieldInt64, "overlay_height")),
		remoteTestSchema("d2legacy.monster.stats",
			remoteTestFields(
				akara.FieldInt64,
				"level", "health", "max_health", "defense", "attack_rating",
				"physical_min", "physical_max", "experience",
			)),
		remoteTestSchema("d2legacy.monster.death_event",
			remoteTestFields(
				akara.FieldString,
				"kind", "monster_id", "killer_id", "credited_id", "loot_seed",
				"treasure_class", "drops",
			),
			remoteTestFields(
				akara.FieldInt64,
				"tick", "xp", "game_player_count", "effective_player_count",
				"nearby_party_member_count", "monster_player_count", "no_drop_player_count",
			)),
	}
}

// remotePlayerAndPresentationTestSchemas covers owner-private data and the read-only render components it may create.
func remotePlayerAndPresentationTestSchemas() []akara.Schema {
	return []akara.Schema{
		remoteTestSchema("d2legacy.player.vitals",
			remoteTestFields(
				akara.FieldInt64,
				"health", "max_health", "mana", "max_mana", "mana_raw", "max_mana_raw",
				"stamina", "max_stamina", "stamina_raw", "max_stamina_raw",
			)),
		remoteTestSchema("d2legacy.player.progress",
			remoteTestFields(akara.FieldInt64, "level", "experience", "unspent_skill_points")),
		remoteTestSchema("d2legacy.player.combat_stats",
			remoteTestFields(akara.FieldInt64, "attack_rating", "defense")),
		remoteTestSchema("d2legacy.player.animation",
			remoteTestFields(akara.FieldInt64, "direction", "start_tick"),
			remoteTestFields(akara.FieldString, "mode")),
		remoteTestSchema("d2legacy.presentation.animation_clock",
			remoteTestFields(akara.FieldFloat64, "seconds")),
		remoteTestSchema("d2legacy.presentation.overlay_anchor",
			remoteTestFields(akara.FieldInt64, "height")),
		remoteTestSchema("d2legacy.presentation.missile", remoteMissileTestFields()),
		remoteTestSchema("d2legacy.presentation.state",
			remoteTestFields(akara.FieldEntity, "target"),
			remoteTestFields(akara.FieldString, "state_id"),
			remoteTestFields(akara.FieldInt64, "period_ticks")),
		remoteTestSchema("d2legacy.presentation.effect_cue",
			remoteTestFields(akara.FieldString, "kind", "overlay_id", "sound"),
			remoteTestFields(akara.FieldInt64, "tick"),
			remoteTestFields(akara.FieldEntity, "target")),
		remoteTestSchema("d2legacy.player.appearance",
			remoteTestFields(akara.FieldString, "cof", "token", "palette", "weapon_class")),
		remoteTestSchema("d2legacy.player.movement_mode",
			remoteTestFields(akara.FieldBool, "running")),
		remoteTestSchema("d2legacy.player.movement_stats",
			remoteTestFields(
				akara.FieldInt64,
				"run_drain", "velocitypercent", "item_fastermovevelocity",
				"staminarecoverybonus", "item_staminadrainpct", "armor_run_drain",
			)),
		remoteTestSchema("d2legacy.player.skill_assignment",
			remoteTestFields(akara.FieldInt64, "left", "right")),
		remoteTestSchema("d2legacy.player.learned_skill",
			remoteTestFields(akara.FieldEntity, "owner"),
			remoteTestFields(akara.FieldInt64, "skill_id", "level", "list_row"),
			remoteTestFields(akara.FieldBool, "left_allowed", "right_allowed")),
		remoteTestSchema("d2legacy.player.belt",
			remoteTestFields(akara.FieldInt64, "capacity")),
		remoteTestSchema("d2legacy.player.party_view", remotePartyTestFields()),
	}
}

// remoteItemAndInteractionTestSchemas keeps inventory and interaction mirrors separate from gameplay authority schemas.
func remoteItemAndInteractionTestSchemas() []akara.Schema {
	return []akara.Schema{
		remoteTestSchema("d2legacy.items.layout", remoteItemLayoutTestFields()),
		remoteTestSchema("d2legacy.item.identity", remoteItemIdentityTestFields()),
		remoteTestSchema("d2legacy.item.placement", remoteItemPlacementTestFields()),
		remoteTestSchema("d2legacy.item.presentation", remoteItemPresentationTestFields()),
		remoteTestSchema("d2legacy.interaction.target", remoteInteractionTargetTestFields()),
		remoteTestSchema("d2legacy.interaction.context",
			remoteTestFields(akara.FieldString, "owner"),
			remoteTestFields(akara.FieldEntity, "target")),
		remoteTestSchema("d2legacy.interaction.null_target"),
		remoteTestSchema("d2legacy.skill.cast_cue", remoteCastCueTestFields()),
		remoteTestSchema("d2legacy.state.event", remoteStateEventTestFields()),
	}
}

// remoteWorldTestSchemas supplies only spatial and selection facts required to render connected entities.
func remoteWorldTestSchemas() []akara.Schema {
	return []akara.Schema{
		remoteTestSchema("d2legacy.world.position",
			remoteTestFields(akara.FieldFloat64, "x", "y")),
		remoteTestSchema("d2legacy.world.velocity",
			remoteTestFields(akara.FieldFloat64, "x", "y")),
		remoteTestSchema("d2legacy.world.facing",
			remoteTestFields(akara.FieldInt64, "direction", "directions")),
		remoteTestSchema("d2legacy.world.location",
			remoteTestFields(akara.FieldInt64, "act", "level_id")),
		remoteTestSchema("d2legacy.world.player_control",
			remoteTestFields(akara.FieldString, "player")),
		remoteTestSchema("d2legacy.world.bounds",
			remoteTestFields(akara.FieldFloat64, "width", "height")),
		remoteTestSchema("d2legacy.world.collider",
			remoteTestFields(akara.FieldFloat64, "radius")),
		remoteTestSchema("d2legacy.world.selectable", remoteSelectableTestFields()),
	}
}

// remoteMissileTestFields excludes collision and damage by construction, making authority leakage a
// schema failure.
func remoteMissileTestFields() []akara.Field {
	return remoteTestFieldGroups(
		remoteTestFields(akara.FieldString, "missile_id", "dcc", "palette"),
		remoteTestFields(
			akara.FieldFloat64,
			"velocity_x", "velocity_y", "offset_x", "offset_y", "offset_z",
		),
		remoteTestFields(
			akara.FieldInt64,
			"logical_direction", "directions", "frames_per_second", "transparency_mode",
		),
		remoteTestFields(akara.FieldBool, "loop"),
	)
}

// remotePartyTestFields builds every fixed roster slot so shortening a party can be tested for stale data.
func remotePartyTestFields() []akara.Field {
	fields := remoteTestFieldGroups(
		remoteTestFields(akara.FieldInt64, "schema_version", "revision", "roster_count"),
		remoteTestFields(akara.FieldString, "party_id"),
	)

	for slot := 1; slot <= playeradapter.MaxPartyViewRoster; slot++ {
		suffix := fmt.Sprintf("_%d", slot)
		fields = append(fields,
			remoteTestField("player"+suffix, akara.FieldString),
			remoteTestField("name"+suffix, akara.FieldString),
			remoteTestField("class"+suffix, akara.FieldString),
			remoteTestField("level"+suffix, akara.FieldInt64),
			remoteTestField("relationship"+suffix, akara.FieldString),
		)
	}

	return fields
}

// remoteItemLayoutTestFields covers the private graph root without adding item behavior fields.
func remoteItemLayoutTestFields() []akara.Field {
	return remoteTestFieldGroups(
		remoteTestFields(akara.FieldString, "owner"),
		remoteTestFields(
			akara.FieldInt64,
			"inventory_width", "inventory_height", "stash_width", "stash_height",
			"cube_width", "cube_height", "belt_capacity", "active_weapon_set",
			"vendor_width", "vendor_height", "carried_gold", "stashed_gold",
		),
	)
}

// remoteItemIdentityTestFields allowlists stable owner-visible item identity used by inventory UI.
func remoteItemIdentityTestFields() []akara.Field {
	return remoteTestFieldGroups(
		remoteTestFields(akara.FieldEntity, "owner"),
		remoteTestFields(akara.FieldString, "id", "code", "body_slots", "applied_services"),
		remoteTestFields(akara.FieldInt64, "width", "height", "base_cost"),
		remoteTestFields(akara.FieldBool, "belt_eligible"),
	)
}

// remoteItemPlacementTestFields covers each mutually applicable container and equipment coordinate.
func remoteItemPlacementTestFields() []akara.Field {
	return remoteTestFieldGroups(
		remoteTestFields(akara.FieldString, "container", "slot"),
		remoteTestFields(akara.FieldInt64, "x", "y", "belt_slot", "weapon_set", "page"),
	)
}

// remoteItemPresentationTestFields restricts private item projection to render asset identifiers.
func remoteItemPresentationTestFields() []akara.Field {
	return remoteTestFieldGroups(
		remoteTestFields(
			akara.FieldString,
			"inventory_dc6", "world_dc6", "composite", "weapon_class",
		),
		remoteTestFields(akara.FieldBool, "world_animated"),
	)
}

// remoteInteractionTargetTestFields includes only owner-visible target and service metadata.
func remoteInteractionTargetTestFields() []akara.Field {
	return remoteTestFieldGroups(
		remoteTestFields(akara.FieldString, "id", "npc", "vendor", "categories", "services"),
		remoteTestFields(akara.FieldFloat64, "x", "y", "radius"),
	)
}

// remoteCastCueTestFields contains presentation timing and targeting without skill execution authority.
func remoteCastCueTestFields() []akara.Field {
	return remoteTestFieldGroups(
		remoteTestFields(akara.FieldString, "kind", "player", "target_id"),
		remoteTestFields(akara.FieldInt64, "tick", "effect_tick", "skill_id"),
		remoteTestFields(akara.FieldEntity, "caster"),
		remoteTestFields(akara.FieldFloat64, "target_x", "target_y"),
	)
}

// remoteStateEventTestFields captures lifecycle cues while omitting state magnitude and effect policy.
func remoteStateEventTestFields() []akara.Field {
	return remoteTestFieldGroups(
		remoteTestFields(akara.FieldString, "kind", "state_id", "source_id", "reason"),
		remoteTestFields(akara.FieldInt64, "tick", "expires_tick"),
		remoteTestFields(akara.FieldEntity, "target"),
	)
}

// remoteSelectableTestFields defines the public pointer-selection facts shared by units and corpses.
func remoteSelectableTestFields() []akara.Field {
	return remoteTestFieldGroups(
		remoteTestFields(akara.FieldString, "id", "kind", "label", "owner"),
		remoteTestFields(akara.FieldFloat64, "radius"),
		remoteTestFields(akara.FieldInt64, "priority"),
	)
}

// remoteTestSchema assembles readable field groups into the exact dynamic schema expected by projection.
func remoteTestSchema(name string, groups ...[]akara.Field) akara.Schema {
	return akara.Schema{Name: name, Fields: remoteTestFieldGroups(groups...)}
}

// remoteTestFieldGroups preserves declaration order while allowing schemas to be documented by domain.
func remoteTestFieldGroups(groups ...[]akara.Field) []akara.Field {
	var fields []akara.Field

	for _, group := range groups {
		fields = append(fields, group...)
	}

	return fields
}

// remoteTestFields removes repetitive declarations without hiding each field's storage kind.
func remoteTestFields(kind akara.FieldKind, names ...string) []akara.Field {
	fields := make([]akara.Field, 0, len(names))

	for _, name := range names {
		fields = append(fields, remoteTestField(name, kind))
	}

	return fields
}

// remoteTestField centralizes the dynamic-field descriptor used by the projection test registry.
func remoteTestField(name string, kind akara.FieldKind) akara.Field {
	return akara.Field{Name: name, Kind: kind}
}
