package movement

import (
	"errors"
	"testing"
)

type catalogRecords struct {
	rows []map[string]string
	err  error
}

func (records catalogRecords) Load(string) ([]map[string]string, error) {
	return records.rows, records.err
}

func TestLoadCatalogUsesCaseInsensitivePinnedClassVelocities(t *testing.T) {
	catalog, err := LoadCatalog(catalogRecords{rows: []map[string]string{
		{"class": "Amazon", "WalkVelocity": "6", "RunVelocity": "9"},
		{"class": "Assassin", "WalkVelocity": "7", "RunVelocity": "10"},
		{"class": "Expansion"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if rates, ok := catalog.Rates("amazon"); !ok || rates != (ClassRates{Walk: 6, Run: 9}) {
		t.Fatalf("Amazon rates = %+v, %v", rates, ok)
	}
	if rates, ok := catalog.Rates(" ASSASSIN "); !ok || rates != (ClassRates{Walk: 7, Run: 10}) {
		t.Fatalf("Assassin rates = %+v, %v", rates, ok)
	}
}

func TestLoadCatalogRejectsMissingMalformedAndDuplicateRates(t *testing.T) {
	for name, records := range map[string]catalogRecords{
		"read":      {err: errors.New("missing")},
		"empty":     {},
		"malformed": {rows: []map[string]string{{"class": "Amazon", "WalkVelocity": "", "RunVelocity": "9"}}},
		"slower run": {rows: []map[string]string{{
			"class": "Amazon", "WalkVelocity": "9", "RunVelocity": "6",
		}}},
		"duplicate": {rows: []map[string]string{
			{"class": "Amazon", "WalkVelocity": "6", "RunVelocity": "9"},
			{"class": "amazon", "WalkVelocity": "6", "RunVelocity": "9"},
		}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadCatalog(records); err == nil {
				t.Fatal("invalid movement catalog was accepted")
			}
		})
	}
}
