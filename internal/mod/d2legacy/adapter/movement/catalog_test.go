package movement

import (
	"errors"
	"testing"
)

// catalogRecords is a deterministic record gateway that can return either fixture rows or a storage failure.
type catalogRecords struct {
	rows []map[string]string
	err  error
}

// Load returns the configured response without interpreting the path, keeping tests focused on catalog policy.
func (records catalogRecords) Load(string) ([]map[string]string, error) {
	return records.rows, records.err
}

// movementRecord constructs a complete playable-class row so malformed cases can override only relevant fields.
func movementRecord(class, walk, run string) map[string]string {
	return map[string]string{
		"class":              class,
		"vit":                "20",
		"WalkVelocity":       walk,
		"RunVelocity":        run,
		"stamina":            "84",
		"RunDrain":           "20",
		"StaminaPerLevel":    "4",
		"StaminaPerVitality": "4",
	}
}

// TestLoadCatalogUsesCaseInsensitivePinnedClassVelocities verifies normalized lookup and Expansion sentinel removal.
func TestLoadCatalogUsesCaseInsensitivePinnedClassVelocities(t *testing.T) {
	assassin := movementRecord("Assassin", "7", "10")
	assassin["stamina"] = "95"
	assassin["RunDrain"] = "15"
	assassin["StaminaPerLevel"] = "5"
	assassin["StaminaPerVitality"] = "5"

	catalog, err := LoadCatalog(catalogRecords{rows: []map[string]string{
		movementRecord("Amazon", "6", "9"),
		assassin,
		{"class": "Expansion"},
	}})
	if err != nil {
		t.Fatal(err)
	}

	wantAmazon := ClassRates{
		Walk:               6,
		Run:                9,
		StartingVitality:   20,
		StartingStamina:    84,
		RunDrain:           20,
		StaminaPerLevel:    4,
		StaminaPerVitality: 4,
	}
	if rates, found := catalog.Rates("amazon"); !found || rates != wantAmazon {
		t.Fatalf("Amazon rates = %+v, %v", rates, found)
	}

	wantAssassin := ClassRates{
		Walk:               7,
		Run:                10,
		StartingVitality:   20,
		StartingStamina:    95,
		RunDrain:           15,
		StaminaPerLevel:    5,
		StaminaPerVitality: 5,
	}
	if rates, found := catalog.Rates(" ASSASSIN "); !found || rates != wantAssassin {
		t.Fatalf("Assassin rates = %+v, %v", rates, found)
	}
}

// TestLoadCatalogRejectsMissingMalformedAndDuplicateRates protects all-or-nothing authoritative catalog loading.
func TestLoadCatalogRejectsMissingMalformedAndDuplicateRates(t *testing.T) {
	malformed := movementRecord("Amazon", "", "9")
	slowerRun := movementRecord("Amazon", "9", "6")

	for name, records := range map[string]catalogRecords{
		"read":       {err: errors.New("missing")},
		"empty":      {},
		"malformed":  {rows: []map[string]string{malformed}},
		"slower run": {rows: []map[string]string{slowerRun}},
		"duplicate": {rows: []map[string]string{
			movementRecord("Amazon", "6", "9"),
			movementRecord("amazon", "6", "9"),
		}},
	} {
		// A named subtest preserves the rejected input category in any failure report.
		t.Run(name, func(t *testing.T) {
			if _, err := LoadCatalog(records); err == nil {
				t.Fatal("invalid movement catalog was accepted")
			}
		})
	}
}
