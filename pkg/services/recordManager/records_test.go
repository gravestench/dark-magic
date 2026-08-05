package recordManager

import "testing"

type testTSVLoader struct {
	data  []byte
	loads int
}

func (l *testTSVLoader) LoadTsv(string) ([]byte, error) {
	l.loads++
	return l.data, nil
}

func TestParseTSV(t *testing.T) {
	records, err := parseTSV([]byte("ID\tName\tValue\r\n1\tAlpha\t10\r\n2\tBeta\t20\r\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || records[1]["Name"] != "Beta" || records[0]["Value"] != "10" {
		t.Fatalf("records = %#v", records)
	}
}

func TestLoadGenericRecordsCachesByPath(t *testing.T) {
	loader := &testTSVLoader{data: []byte("ID\tName\n1\tAlpha\n")}
	service := &Service{assets: loader}
	for range 2 {
		records, err := service.loadGenericRecords("items.txt")
		if err != nil {
			t.Fatal(err)
		}
		if len(records) != 1 {
			t.Fatalf("record count = %d", len(records))
		}
	}
	if loader.loads != 1 {
		t.Fatalf("loads = %d, want 1", loader.loads)
	}
}

func TestReloadGenericRecordsInvalidatesCache(t *testing.T) {
	loader := &testTSVLoader{data: []byte("ID\tName\n1\tAlpha\n")}
	service := &Service{assets: loader}
	if _, err := service.loadGenericRecords("items.txt"); err != nil {
		t.Fatal(err)
	}
	loader.data = []byte("ID\tName\n2\tBeta\n")
	records, err := service.reloadGenericRecords("items.txt")
	if err != nil {
		t.Fatal(err)
	}
	if records[0]["Name"] != "Beta" || loader.loads != 2 {
		t.Fatalf("records = %#v, loads = %d", records, loader.loads)
	}
}

func TestParseTSVAllowsShortRows(t *testing.T) {
	records, err := parseTSV([]byte("ID\tName\n1\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0]["ID"] != "1" {
		t.Fatalf("records = %#v", records)
	}
}
