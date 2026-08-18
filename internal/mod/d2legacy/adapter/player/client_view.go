package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const (
	ClientViewVersion    uint32 = 8
	MaxHUDLearnedSkills         = 256
	MaxHUDBeltSlots             = 16
	MaxPrivateItems             = 1024
	maxViewIdentityBytes        = 128
	maxViewLabelBytes           = 256
	maxViewTokenBytes           = 256
	maxItemGridDimension        = 256
	maxItemDimension            = 16
)

var ErrClientView = errors.New("client view: invalid projection")

// ClientView is the complete initial/correction projection envelope. Owner
// private and nearby public schemas remain independently versioned within it.
type ClientView struct {
	Version uint32      `json:"version"`
	Tick    uint64      `json:"tick"`
	HUD     HUD         `json:"hud"`
	World   WorldView   `json:"world"`
	Private PrivateView `json:"private"`
	Party   PartyView   `json:"party"`
	Events  EventView   `json:"events"`
}

// ValidateClientView treats every network projection as untrusted input. It
// enforces the same collection, identity, and numeric bounds used while
// projecting canonical state so a compromised host cannot turn a valid framed
// message into unbounded client allocation or invalid presentation state.
func ValidateClientView(view ClientView, tick uint64) error {
	if view.Version != ClientViewVersion || view.Tick != tick ||
		view.HUD.Version != HUDVersion || view.HUD.Tick != tick ||
		view.World.Version != WorldViewVersion || view.World.Tick != tick ||
		view.Private.Version != PrivateViewVersion || view.Private.Tick != tick ||
		view.Party.Version != PartyViewVersion || view.Party.Tick != tick ||
		view.Events.Version != EventViewVersion || view.Events.Tick != tick {
		return ErrClientView
	}
	if err := validateHUDView(view.HUD); err != nil {
		return err
	}
	if err := validateDecodedWorldView(view.World); err != nil {
		return err
	}
	if err := validatePrivateView(view.Private); err != nil {
		return err
	}
	if err := validateEventView(view.Events, tick); err != nil {
		return err
	}
	return validatePartyView(view.Party, view.HUD.Player.PlayerID)
}

func validatePartyView(party PartyView, owner string) error {
	if len(party.Roster) < 1 || len(party.Roster) > MaxPartyViewRoster ||
		!bounded(party.PartyID, maxViewIdentityBytes) || party.Revision > uint64(1<<63-1) ||
		party.Roster[0].PlayerID != owner || party.Roster[0].Relationship != "self" {
		return ErrClientView
	}
	seen, ownerFound, memberFound := make(map[string]struct{}, len(party.Roster)), false, false
	for _, entry := range party.Roster {
		if !boundedRequired(entry.PlayerID, maxViewIdentityBytes) ||
			!boundedRequired(entry.Name, maxViewLabelBytes) || !boundedRequired(entry.Class, maxWorldKindBytes) ||
			entry.Level < 1 || !validPartyRelationship(entry.Relationship) {
			return ErrClientView
		}
		if _, duplicate := seen[entry.PlayerID]; duplicate {
			return ErrClientView
		}
		seen[entry.PlayerID] = struct{}{}
		if entry.PlayerID == owner {
			ownerFound = entry.Relationship == "self"
		}
		memberFound = memberFound || entry.Relationship == "party"
	}
	if !ownerFound || (party.PartyID == "") != !memberFound {
		return ErrClientView
	}
	return nil
}

func validPartyRelationship(value string) bool {
	switch value {
	case "self", "party", "invited_you", "invited", "unavailable", "available":
		return true
	default:
		return false
	}
}

func validateHUDView(hud HUD) error {
	identity := hud.Player
	if !boundedRequired(identity.PlayerID, maxViewIdentityBytes) ||
		!boundedRequired(identity.CharacterID, maxViewIdentityBytes) ||
		!boundedRequired(identity.Name, maxViewLabelBytes) || !boundedRequired(identity.Class, maxWorldKindBytes) ||
		hud.Vitals.MaxHealth < 0 || hud.Vitals.Health < 0 || hud.Vitals.Health > hud.Vitals.MaxHealth ||
		hud.Vitals.MaxMana < 0 || hud.Vitals.Mana < 0 || hud.Vitals.Mana > hud.Vitals.MaxMana ||
		hud.Vitals.MaxStamina < 0 || hud.Vitals.Stamina < 0 || hud.Vitals.Stamina > hud.Vitals.MaxStamina ||
		hud.Vitals.MaxStaminaRaw < 0 || hud.Vitals.StaminaRaw < 0 || hud.Vitals.StaminaRaw > hud.Vitals.MaxStaminaRaw ||
		hud.Vitals.Stamina != hud.Vitals.StaminaRaw/256 || hud.Vitals.MaxStamina != hud.Vitals.MaxStaminaRaw/256 ||
		!finiteView(hud.Position.X, hud.Position.Y, hud.Movement.Velocity.X, hud.Movement.Velocity.Y,
			hud.Movement.Bounds.X, hud.Movement.Bounds.Y, hud.Movement.Radius) ||
		hud.Movement.Bounds.X < 0 || hud.Movement.Bounds.Y < 0 || hud.Movement.Radius < 0 ||
		!bounded(hud.Animation.Mode, maxWorldKindBytes) || len(hud.Skills.Learned) > MaxHUDLearnedSkills ||
		len(hud.Belt.Slots) > MaxHUDBeltSlots || hud.Belt.Capacity < 0 || hud.Belt.Capacity > MaxHUDBeltSlots {
		return ErrClientView
	}
	for _, slot := range hud.Belt.Slots {
		if !bounded(slot, maxViewIdentityBytes) {
			return ErrClientView
		}
	}
	for _, skill := range hud.Skills.Learned {
		if skill.SkillID < 0 || skill.Level < 0 || skill.ListRow < 0 {
			return ErrClientView
		}
	}
	return nil
}

func validateDecodedWorldView(world WorldView) error {
	if len(world.Entities) > MaxWorldViewEntities || !finiteView(world.Origin.X, world.Origin.Y) {
		return ErrClientView
	}
	seen := make(map[string]struct{}, len(world.Entities))
	for _, entity := range world.Entities {
		if err := validateWorldEntity(entity); err != nil ||
			!bounded(entity.Class, maxWorldKindBytes) || !bounded(entity.Token, maxViewTokenBytes) ||
			!bounded(entity.Mode, maxWorldKindBytes) {
			return ErrClientView
		}
		if (entity.Health == nil) != (entity.MaxHealth == nil) ||
			(entity.Health != nil && (*entity.Health < 0 || *entity.MaxHealth < 0 || *entity.Health > *entity.MaxHealth)) {
			return ErrClientView
		}
		if _, duplicate := seen[entity.ID]; duplicate {
			return ErrClientView
		}
		seen[entity.ID] = struct{}{}
	}
	return nil
}

func validatePrivateView(private PrivateView) error {
	layout := private.Items.Layout
	if len(private.Items.Items) > MaxPrivateItems ||
		!boundedDimension(layout.InventoryWidth) || !boundedDimension(layout.InventoryHeight) ||
		!boundedDimension(layout.StashWidth) || !boundedDimension(layout.StashHeight) ||
		!boundedDimension(layout.CubeWidth) || !boundedDimension(layout.CubeHeight) ||
		layout.BeltCapacity < 0 || layout.BeltCapacity > MaxHUDBeltSlots ||
		!boundedDimension(layout.VendorWidth) || !boundedDimension(layout.VendorHeight) ||
		layout.CarriedGold < 0 || layout.StashedGold < 0 {
		return ErrClientView
	}
	seen := make(map[string]struct{}, len(private.Items.Items))
	for _, item := range private.Items.Items {
		if !boundedRequired(item.ID, maxViewIdentityBytes) || !boundedRequired(item.Code, maxWorldKindBytes) ||
			!bounded(item.BodySlots, maxViewTokenBytes) || !bounded(item.AppliedServices, maxViewTokenBytes) ||
			!bounded(item.Container, maxWorldKindBytes) || !bounded(item.Slot, maxWorldKindBytes) ||
			!bounded(item.InventoryDC6, maxViewTokenBytes) || !bounded(item.WorldDC6, maxViewTokenBytes) ||
			!bounded(item.Composite, maxViewTokenBytes) || !bounded(item.WeaponClass, maxWorldKindBytes) ||
			item.Width < 0 || item.Width > maxItemDimension || item.Height < 0 || item.Height > maxItemDimension || item.BaseCost < 0 {
			return ErrClientView
		}
		if _, duplicate := seen[item.ID]; duplicate {
			return ErrClientView
		}
		seen[item.ID] = struct{}{}
	}
	target := private.Interaction.Target
	if private.Interaction.Active != (target != nil) {
		return ErrClientView
	}
	if target != nil && (!boundedRequired(target.ID, maxViewIdentityBytes) ||
		!bounded(target.NPC, maxWorldKindBytes) || !bounded(target.Vendor, maxWorldKindBytes) ||
		!bounded(target.Categories, maxViewTokenBytes) || !bounded(target.Services, maxViewTokenBytes) ||
		!finiteView(target.X, target.Y, target.Radius) || target.Radius < 0) {
		return ErrClientView
	}
	return nil
}

func bounded(value string, maximum int) bool { return len(value) <= maximum }

func boundedRequired(value string, maximum int) bool {
	return strings.TrimSpace(value) != "" && bounded(value, maximum)
}

func boundedDimension(value int64) bool { return value >= 0 && value <= maxItemGridDimension }

func finiteView(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}

func ProjectClientView(playerID string, checkpoint simulation.Checkpoint) (json.RawMessage, error) {
	hudPayload, err := ProjectHUD(playerID, checkpoint)
	if err != nil {
		return nil, err
	}
	worldPayload, err := ProjectWorldView(playerID, checkpoint)
	if err != nil {
		return nil, err
	}
	private, err := ProjectPrivateView(playerID, checkpoint)
	if err != nil {
		return nil, err
	}
	party, err := ProjectPartyView(playerID, checkpoint)
	if err != nil {
		return nil, err
	}
	events, err := ProjectEventView(playerID, checkpoint)
	if err != nil {
		return nil, err
	}
	var hud HUD
	var world WorldView
	if err := json.Unmarshal(hudPayload, &hud); err != nil {
		return nil, fmt.Errorf("client view: HUD: %w", err)
	}
	if err := json.Unmarshal(worldPayload, &world); err != nil {
		return nil, fmt.Errorf("client view: world: %w", err)
	}
	return json.Marshal(ClientView{Version: ClientViewVersion, Tick: checkpoint.Tick, HUD: hud, World: world, Private: private, Party: party, Events: events})
}
