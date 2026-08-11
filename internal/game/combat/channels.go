package combat

import "fmt"

// Channel identifies one independently resolved quantity in a damage bundle.
// Life and mana are retained as typed channels because some skills/missiles
// carry those values separately from ordinary elemental damage.
type Channel uint8

const (
	Physical Channel = iota
	Fire
	Lightning
	Cold
	Poison
	Magic
	Life
	Mana
	channelCount
)

var channelNames = [...]string{
	Physical:  "physical",
	Fire:      "fire",
	Lightning: "lightning",
	Cold:      "cold",
	Poison:    "poison",
	Magic:     "magic",
	Life:      "life",
	Mana:      "mana",
}

func (channel Channel) valid() bool { return channel < channelCount }

func (channel Channel) String() string {
	if !channel.valid() {
		return fmt.Sprintf("channel(%d)", channel)
	}
	return channelNames[channel]
}

// ParseChannel turns a stable semantic name into a channel.
func ParseChannel(name string) (Channel, error) {
	for channel, candidate := range channelNames {
		if candidate == name {
			return Channel(channel), nil
		}
	}
	return 0, fmt.Errorf("combat: unknown damage channel %q", name)
}

// Bundle keeps channels separate while conversion and mitigation are applied.
// The fixed array has stable order and cannot inherit map iteration randomness.
type Bundle struct {
	values [channelCount]Amount
}

// NewBundle validates and constructs a bundle from semantic channel values.
func NewBundle(values map[Channel]Amount) (Bundle, error) {
	var bundle Bundle
	for channel, amount := range values {
		if !channel.valid() {
			return Bundle{}, fmt.Errorf("combat: invalid damage channel %d", channel)
		}
		bundle.values[channel] = amount
	}
	return bundle, nil
}

// Amount returns one channel; invalid channels are rejected instead of reading
// outside the fixed vocabulary.
func (bundle Bundle) Amount(channel Channel) (Amount, error) {
	if !channel.valid() {
		return 0, fmt.Errorf("combat: invalid damage channel %d", channel)
	}
	return bundle.values[channel], nil
}

// With returns a copy with one channel replaced.
func (bundle Bundle) With(channel Channel, amount Amount) (Bundle, error) {
	if !channel.valid() {
		return Bundle{}, fmt.Errorf("combat: invalid damage channel %d", channel)
	}
	bundle.values[channel] = amount
	return bundle, nil
}

// Add combines matching channels with checked fixed-point addition.
func (bundle Bundle) Add(other Bundle) (Bundle, error) {
	var result Bundle
	for channel := Channel(0); channel < channelCount; channel++ {
		amount, err := bundle.values[channel].Add(other.values[channel])
		if err != nil {
			return Bundle{}, fmt.Errorf("combat: add %s channel: %w", channel, err)
		}
		result.values[channel] = amount
	}
	return result, nil
}

// Scale applies one explicitly rounded ratio to every channel.
func (bundle Bundle) Scale(numerator, denominator int64, rounding Rounding) (Bundle, error) {
	var result Bundle
	for channel := Channel(0); channel < channelCount; channel++ {
		amount, err := bundle.values[channel].Scale(numerator, denominator, rounding)
		if err != nil {
			return Bundle{}, fmt.Errorf("combat: scale %s channel: %w", channel, err)
		}
		result.values[channel] = amount
	}
	return result, nil
}

// Entries returns every channel in stable order, including zero values. This is
// suitable for deterministic snapshots, events, and debugging output.
func (bundle Bundle) Entries() []ChannelAmount {
	entries := make([]ChannelAmount, 0, channelCount)
	for channel := Channel(0); channel < channelCount; channel++ {
		entries = append(entries, ChannelAmount{Channel: channel, Amount: bundle.values[channel]})
	}
	return entries
}

// ChannelAmount is one stable-order bundle entry.
type ChannelAmount struct {
	Channel Channel `json:"channel"`
	Amount  Amount  `json:"amount"`
}
