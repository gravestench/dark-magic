package clientapp

import (
	"context"
	"strings"
	"testing"

	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

func TestNetworkControllerDefersHostUntilCharacterSelection(t *testing.T) {
	app := &application{ctx: context.Background(), saves: d2save.New()}
	controller := newNetworkController(app)
	if err := controller.Host(); err != nil {
		t.Fatalf("begin host: %v", err)
	}
	status := controller.Status()
	if status["phase"] != "selecting" || status["mode"] != "host" {
		t.Fatalf("pending host status = %#v", status)
	}
	controller.Cancel()
	if status = controller.Status(); status["phase"] != "idle" || status["mode"] != "" {
		t.Fatalf("cancelled host status = %#v", status)
	}
}

func TestNetworkControllerKeepsStartFailuresAndNormalizesDirectJoin(t *testing.T) {
	app := &application{ctx: context.Background(), saves: d2save.New()}
	controller := newNetworkController(app)
	if err := controller.Host(); err != nil {
		t.Fatal(err)
	}
	if err := controller.StartSelected(); err == nil {
		t.Fatal("starting without selected character was accepted")
	}
	status := controller.Status()
	if status["phase"] != "failed" || status["mode"] != "host" || !strings.Contains(status["error"].(string), "select") {
		t.Fatalf("start rejection status = %#v", status)
	}
	if err := controller.Join("127.0.0.1"); err != nil {
		t.Fatalf("direct join: %v", err)
	}
	status = controller.Status()
	if status["phase"] != "selecting" || status["mode"] != "join" || status["address"] != "127.0.0.1:6112" {
		t.Fatalf("join selection status = %#v", status)
	}
}
