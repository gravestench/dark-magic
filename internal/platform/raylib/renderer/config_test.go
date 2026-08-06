package raylibRenderer

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestDefaultConfigDataRoundTripsCompleteConfiguration(t *testing.T) {
	want := DefaultConfig()
	service := &Service{}
	var got Config
	if err := json.Unmarshal(service.DefaultConfigData(), &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("config = %#v, want %#v", got, want)
	}
	if got.Window.Width <= 0 || got.Window.Height <= 0 || got.Resolution.Width <= 0 || got.Resolution.Height <= 0 || got.Cache.BudgetMB <= 0 {
		t.Fatal("default configuration is incomplete")
	}
}
