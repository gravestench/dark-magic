package item

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const (
	MoveCommand       = "item.move"
	WeaponSetCommand  = "item.weapon_set"
	VendorSellCommand = "item.vendor_sell"
	VendorBuyCommand  = "item.vendor_buy"
	ServiceCommand    = "item.service_complete"
)

// MovePayload describes intent, not trusted results. In particular, Displaced
// is computed by authority during a held-item swap and never supplied here.
type MovePayload struct {
	Owner       string    `json:"owner,omitempty"`
	ItemID      string    `json:"item_id"`
	Destination Placement `json:"destination"`
	PlaceHeld   bool      `json:"place_held,omitempty"`
}

type WeaponSetPayload struct {
	Owner string `json:"owner,omitempty"`
	Set   int    `json:"set"`
}

type VendorPayload struct {
	Owner    string `json:"owner,omitempty"`
	ItemID   string `json:"item_id"`
	Vendor   string `json:"vendor"`
	Category string `json:"category,omitempty"`
}

type ServicePayload struct {
	Owner   string `json:"owner,omitempty"`
	Service string `json:"service"`
}

func RegisterCommands(session *gamesession.Session, authority *Authority) error {
	if session == nil || authority == nil {
		return fmt.Errorf("item: session and authority are required")
	}
	if err := session.Register(MoveCommand, gamesession.CommandHandler{
		Validate: validateMoveCommand,
		Apply: func(_ *gameecs.Engine, command simulation.Command) error {
			payload, err := decodeMove(command.Payload)
			if err != nil {
				return err
			}
			owner := payload.Owner
			if owner == "" {
				owner = command.Player
			}
			_, err = authority.move(owner, payload.ItemID, payload.Destination, payload.PlaceHeld)
			return err
		},
		Allowed: []simulation.Authority{simulation.AuthorityPlayer, simulation.AuthorityAdmin},
	}); err != nil {
		return err
	}
	if err := session.Register(WeaponSetCommand, gamesession.CommandHandler{
		Validate: validateWeaponSetCommand,
		Apply: func(_ *gameecs.Engine, command simulation.Command) error {
			payload, err := decodeWeaponSet(command.Payload)
			if err != nil {
				return err
			}
			owner := payload.Owner
			if owner == "" {
				owner = command.Player
			}
			return authority.selectWeaponSet(owner, payload.Set)
		},
		Allowed: []simulation.Authority{simulation.AuthorityPlayer, simulation.AuthorityAdmin},
	}); err != nil {
		return err
	}
	if err := registerVendorCommands(session, authority); err != nil {
		return err
	}
	return session.Register(ServiceCommand, gamesession.CommandHandler{
		Validate: validateServiceCommand,
		Apply: func(_ *gameecs.Engine, command simulation.Command) error {
			payload, err := decodeService(command.Payload)
			if err != nil {
				return err
			}
			owner := payload.Owner
			if owner == "" {
				owner = command.Player
			}
			return authority.completeService(owner, payload.Service)
		},
		Allowed: []simulation.Authority{simulation.AuthorityPlayer, simulation.AuthorityAdmin},
	})
}

func validateServiceCommand(command simulation.Command) error {
	payload, err := decodeService(command.Payload)
	if err != nil {
		return err
	}
	if command.Authority == simulation.AuthorityPlayer && payload.Owner != "" && payload.Owner != command.Player {
		return fmt.Errorf("item: player cannot complete another owner's service")
	}
	return nil
}

func decodeService(encoded []byte) (ServicePayload, error) {
	var payload ServicePayload
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return ServicePayload{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return ServicePayload{}, fmt.Errorf("item: service payload has trailing data")
	}
	payload.Owner, payload.Service = strings.TrimSpace(payload.Owner), strings.TrimSpace(payload.Service)
	if payload.Service == "" {
		return ServicePayload{}, fmt.Errorf("item: service identity is required")
	}
	return payload, nil
}

func registerVendorCommands(session *gamesession.Session, authority *Authority) error {
	for _, definition := range []struct {
		kind string
		sell bool
	}{{VendorSellCommand, true}, {VendorBuyCommand, false}} {
		kind, isSell := definition.kind, definition.sell
		if err := session.Register(kind, gamesession.CommandHandler{
			Validate: func(command simulation.Command) error {
				_, err := decodeVendor(command, isSell)
				return err
			},
			Apply: func(_ *gameecs.Engine, command simulation.Command) error {
				payload, err := decodeVendor(command, isSell)
				if err != nil {
					return err
				}
				owner := payload.Owner
				if owner == "" {
					owner = command.Player
				}
				if isSell {
					return authority.sellHeld(owner, payload.ItemID, payload.Vendor, payload.Category)
				}
				return authority.buyToHeld(owner, payload.ItemID, payload.Vendor)
			},
			Allowed: []simulation.Authority{simulation.AuthorityPlayer, simulation.AuthorityAdmin},
		}); err != nil {
			return err
		}
	}
	return nil
}

func decodeVendor(command simulation.Command, sell bool) (VendorPayload, error) {
	var payload VendorPayload
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return VendorPayload{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return VendorPayload{}, fmt.Errorf("item: vendor payload has trailing data")
	}
	payload.Owner = strings.TrimSpace(payload.Owner)
	payload.ItemID = strings.TrimSpace(payload.ItemID)
	payload.Vendor = strings.TrimSpace(payload.Vendor)
	payload.Category = strings.TrimSpace(payload.Category)
	if payload.ItemID == "" || payload.Vendor == "" {
		return VendorPayload{}, fmt.Errorf("item: item identity and vendor are required")
	}
	if sell && (payload.Category == "" || strings.Contains(payload.Category, "/")) {
		return VendorPayload{}, fmt.Errorf("item: valid vendor category is required")
	}
	if !sell && payload.Category != "" {
		return VendorPayload{}, fmt.Errorf("item: vendor purchase does not accept a category")
	}
	if command.Authority == simulation.AuthorityPlayer && payload.Owner != "" && payload.Owner != command.Player {
		return VendorPayload{}, fmt.Errorf("item: player cannot transact another owner's items")
	}
	return payload, nil
}

func validateMoveCommand(command simulation.Command) error {
	payload, err := decodeMove(command.Payload)
	if err != nil {
		return err
	}
	if command.Authority == simulation.AuthorityPlayer && payload.Owner != "" && payload.Owner != command.Player {
		return fmt.Errorf("item: player cannot move another owner's items")
	}
	if payload.Destination.Container == ContainerVendor {
		return fmt.Errorf("item: vendor stock changes require a transaction command")
	}
	if payload.PlaceHeld && !isHeldDestination(payload.Destination.Container) {
		return fmt.Errorf("item: held placement requires a grid or named-slot destination")
	}
	return nil
}

func isHeldDestination(container Container) bool {
	return isGrid(container) || isHeldSlot(container)
}

func validateWeaponSetCommand(command simulation.Command) error {
	payload, err := decodeWeaponSet(command.Payload)
	if err != nil {
		return err
	}
	if command.Authority == simulation.AuthorityPlayer && payload.Owner != "" && payload.Owner != command.Player {
		return fmt.Errorf("item: player cannot change another owner's weapon set")
	}
	return nil
}

func decodeWeaponSet(encoded []byte) (WeaponSetPayload, error) {
	var payload WeaponSetPayload
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return WeaponSetPayload{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return WeaponSetPayload{}, fmt.Errorf("item: weapon-set payload has trailing data")
	}
	payload.Owner = strings.TrimSpace(payload.Owner)
	if !validWeaponSet(payload.Set) {
		return WeaponSetPayload{}, fmt.Errorf("item: weapon set must be 0 or 1")
	}
	return payload, nil
}

func decodeMove(encoded []byte) (MovePayload, error) {
	var payload MovePayload
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return MovePayload{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return MovePayload{}, fmt.Errorf("item: move payload has trailing data")
	}
	payload.Owner = strings.TrimSpace(payload.Owner)
	payload.ItemID = strings.TrimSpace(payload.ItemID)
	if payload.ItemID == "" {
		return MovePayload{}, fmt.Errorf("item: item identity is required")
	}
	return payload, nil
}

func Command(payload MovePayload, actor string, sequence, tick uint64, authority simulation.Authority) (simulation.Command, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return simulation.Command{}, err
	}
	return simulation.Command{Tick: tick, Player: actor, Authority: authority, Sequence: sequence, Kind: MoveCommand, Payload: encoded}, nil
}

func WeaponSetSelectionCommand(payload WeaponSetPayload, actor string, sequence, tick uint64, authority simulation.Authority) (simulation.Command, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return simulation.Command{}, err
	}
	return simulation.Command{Tick: tick, Player: actor, Authority: authority, Sequence: sequence, Kind: WeaponSetCommand, Payload: encoded}, nil
}

func VendorCommand(kind string, payload VendorPayload, actor string, sequence, tick uint64, authority simulation.Authority) (simulation.Command, error) {
	if kind != VendorSellCommand && kind != VendorBuyCommand {
		return simulation.Command{}, fmt.Errorf("item: unsupported vendor command %q", kind)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return simulation.Command{}, err
	}
	return simulation.Command{Tick: tick, Player: actor, Authority: authority, Sequence: sequence, Kind: kind, Payload: encoded}, nil
}

func ServiceCompletionCommand(payload ServicePayload, actor string, sequence, tick uint64, authority simulation.Authority) (simulation.Command, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return simulation.Command{}, err
	}
	return simulation.Command{Tick: tick, Player: actor, Authority: authority, Sequence: sequence, Kind: ServiceCommand, Payload: encoded}, nil
}
