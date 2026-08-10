package item

import (
	"fmt"
	"math"
	"strings"
)

const priceScale int64 = 1024

// TradeTerms are recovered NPC.txt rules for one vendor and difficulty.
// BuyMultiplier is what the NPC pays; SellMultiplier is what the NPC charges.
type TradeTerms struct {
	BuyMultiplier  int64
	SellMultiplier int64
	MaxBuy         int64
}

type TradeCatalog map[string]TradeTerms

func (catalog TradeCatalog) Terms(vendor string) (TradeTerms, error) {
	terms, found := catalog[strings.ToLower(strings.TrimSpace(vendor))]
	if !found {
		return TradeTerms{}, fmt.Errorf("item: unknown vendor %q", vendor)
	}
	if terms.BuyMultiplier < 0 || terms.SellMultiplier < 0 || terms.MaxBuy < 0 {
		return TradeTerms{}, fmt.Errorf("item: vendor %q has invalid trade terms", vendor)
	}
	return terms, nil
}

func (state *State) sellHeldForGold(id, category string, terms TradeTerms) (int64, error) {
	candidate, found := state.items[id]
	if !found {
		return 0, fmt.Errorf("item: unknown item %q", id)
	}
	price, err := scaledPrice(candidate.BaseCost, terms.BuyMultiplier)
	if err != nil {
		return 0, err
	}
	if terms.MaxBuy > 0 && price > terms.MaxBuy {
		price = terms.MaxBuy
	}
	if price > math.MaxInt64-state.layout.Gold.Carried {
		return 0, fmt.Errorf("item: carried gold overflow")
	}
	if _, err := state.sellHeld(id, category); err != nil {
		return 0, err
	}
	state.layout.Gold.Carried += price
	return price, nil
}

func (state *State) buyToHeldForGold(id string, terms TradeTerms) (int64, error) {
	candidate, found := state.items[id]
	if !found {
		return 0, fmt.Errorf("item: unknown item %q", id)
	}
	price, err := scaledPrice(candidate.BaseCost, terms.SellMultiplier)
	if err != nil {
		return 0, err
	}
	if state.layout.Gold.Carried < price {
		return 0, fmt.Errorf("item: insufficient carried gold")
	}
	if err := state.buyToHeld(id); err != nil {
		return 0, err
	}
	state.layout.Gold.Carried -= price
	return price, nil
}

func scaledPrice(base, multiplier int64) (int64, error) {
	if base < 0 || multiplier < 0 {
		return 0, fmt.Errorf("item: price inputs cannot be negative")
	}
	if multiplier != 0 && base > math.MaxInt64/multiplier {
		return 0, fmt.Errorf("item: price overflow")
	}
	return base * multiplier / priceScale, nil
}
