package loot

import (
	"strings"
	"testing"
)

const itemRatioHeader = "Version\tUber\tClass Specific\tUnique\tUniqueDivisor\tUniqueMin\tSet\tSetDivisor\tSetMin\tRare\tRareDivisor\tRareMin\tMagic\tMagicDivisor\tMagicMin\tHiQuality\tHiQualityDivisor\tNormal\tNormalDivisor\n"

func TestParseAndSelectItemRatio(t *testing.T) {
	input := itemRatioHeader +
		"0\t0\t0\t400\t2\t6400\t300\t2\t5600\t200\t2\t4800\t100\t2\t3200\t20\t2\t2\t1\n" +
		"100\t1\t0\t300\t2\t6400\t250\t2\t5600\t150\t2\t4800\t80\t2\t3200\t15\t2\t2\t1\n"
	ratios, err := ParseItemRatiosTSV(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(ratios) != 2 || ratios[1].Unique.Base != 300 || !ratios[1].Uber {
		t.Fatalf("ratios = %#v", ratios)
	}
	selected, err := SelectQualityRatio(ratios, 100, true, false)
	if err != nil || selected.Magic.Base != 80 {
		t.Fatalf("selected = %#v, error = %v", selected, err)
	}
	if _, err := SelectQualityRatio(ratios, 100, false, false); err == nil || !strings.Contains(err.Error(), "no item ratio") {
		t.Fatalf("missing selection error = %v", err)
	}
}

func TestParseItemRatioRequiresColumns(t *testing.T) {
	_, err := ParseItemRatiosTSV(strings.NewReader("Version\tUber\n"))
	if err == nil || !strings.Contains(err.Error(), "Class Specific") {
		t.Fatalf("error = %v", err)
	}
}
