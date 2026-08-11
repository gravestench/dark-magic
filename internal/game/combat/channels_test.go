package combat

import (
	"reflect"
	"testing"
)

func TestDamageBundleKeepsChannelsSeparateAndOrdered(t *testing.T) {
	bundle, err := NewBundle(map[Channel]Amount{Fire: MustWhole(3), Physical: MustWhole(5), Mana: FromRaw(128)})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := bundle.Amount(Fire); got != MustWhole(3) {
		t.Fatalf("fire = %d", got)
	}
	entries := bundle.Entries()
	channels := make([]Channel, len(entries))
	for index, entry := range entries {
		channels[index] = entry.Channel
	}
	want := []Channel{Physical, Fire, Lightning, Cold, Poison, Magic, Life, Mana}
	if !reflect.DeepEqual(channels, want) {
		t.Fatalf("channel order = %v", channels)
	}
}

func TestDamageBundleAddAndScale(t *testing.T) {
	left, _ := NewBundle(map[Channel]Amount{Physical: MustWhole(4), Cold: MustWhole(2)})
	right, _ := NewBundle(map[Channel]Amount{Physical: MustWhole(1), Fire: MustWhole(3)})
	total, err := left.Add(right)
	if err != nil {
		t.Fatal(err)
	}
	scaled, err := total.Scale(1, 2, RoundTowardZero)
	if err != nil {
		t.Fatal(err)
	}
	assertChannel(t, scaled, Physical, FromRaw(640))
	assertChannel(t, scaled, Fire, FromRaw(384))
	assertChannel(t, scaled, Cold, MustWhole(1))
}

func TestChannelNamesAreStable(t *testing.T) {
	for channel := Channel(0); channel < channelCount; channel++ {
		parsed, err := ParseChannel(channel.String())
		if err != nil || parsed != channel {
			t.Fatalf("round trip %v = %v, %v", channel, parsed, err)
		}
	}
	if _, err := ParseChannel("shadow"); err == nil {
		t.Fatal("expected unknown channel error")
	}
}

func assertChannel(t *testing.T, bundle Bundle, channel Channel, want Amount) {
	t.Helper()
	got, err := bundle.Amount(channel)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("%s = %d, want %d", channel, got, want)
	}
}
