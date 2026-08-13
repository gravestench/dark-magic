package clientapp

import (
	"context"
	"strings"
	"testing"

	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

func TestNetworkControllerRejectsHostAndJoinWithoutCrashingPresentation(t *testing.T) {
	app := &application{ctx: context.Background(), saves: d2save.New()}
	controller := newNetworkController(app)
	if err := controller.Host(); err == nil {
		t.Fatal("host without selected character was accepted")
	}
	status := controller.Status()
	if status["phase"] != "failed" || status["mode"] != "host" || !strings.Contains(status["error"].(string), "select") {
		t.Fatalf("host rejection status = %#v", status)
	}
	if err := controller.Join("127.0.0.1:4433"); err == nil {
		t.Fatal("join without trust invitation was accepted")
	}
	status = controller.Status()
	if status["phase"] != "failed" || status["mode"] != "join" || !strings.Contains(status["error"].(string), "trust") {
		t.Fatalf("join rejection status = %#v", status)
	}
}
