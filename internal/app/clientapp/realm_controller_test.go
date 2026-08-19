package clientapp

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/preferences"
)

// TestNormalizeRealmEndpointUsesTLSAndLegacyPort verifies the strict gateway format.
func TestNormalizeRealmEndpointUsesTLSAndLegacyPort(t *testing.T) {
	endpoint, address, err := normalizeRealmEndpoint("127.0.0.1")
	if err != nil || endpoint != "https://127.0.0.1:6112" || address != "127.0.0.1:6112" {
		t.Fatalf("endpoint=%q address=%q error=%v", endpoint, address, err)
	}

	if _, _, err := normalizeRealmEndpoint("http://realm.example"); err == nil {
		t.Fatal("plaintext realm endpoint accepted")
	}
}

// TestRealmControllerRequiresExplicitLoginAfterCompatibilityCheck guards against implicit login.
func TestRealmControllerRequiresExplicitLoginAfterCompatibilityCheck(t *testing.T) {
	control, server := newRealmTestServer(t)
	controller := newRealmTLSController(t)

	if _, err := control.CreateAccount(t.Context(), "Alice", "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}

	if err := controller.Connect(server.URL); err != nil {
		t.Fatal(err)
	}

	waitRealmPhase(t, controller, "login")

	status := controller.Status()
	if status["endpoint"] != server.URL || status["gateway"] != server.URL {
		t.Fatalf("status=%#v", status)
	}

	if account := status["account"].(map[string]any); account["id"] != "" {
		t.Fatalf("compatibility check implicitly logged in: %#v", status)
	}

	if err := controller.Login("Alice", "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}

	waitRealmPhase(t, controller, "characters")

	if account := controller.Status()["account"].(map[string]any); account["name"] != "Alice" {
		t.Fatalf("logged in account = %#v", account)
	}
}

// newRealmTestServer creates a real TLS control plane for compatibility and login tests.
func newRealmTestServer(t *testing.T) (*realm.ControlPlane, *httptest.Server) {
	t.Helper()

	control, err := realm.NewControlPlane(realm.ControlPlaneConfig{})
	if err != nil {
		t.Fatal(err)
	}

	handler, err := realm.NewHTTPHandler(control)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)

	return control, server
}

// newRealmTLSController creates a controller that trusts test-server certificates.
func newRealmTLSController(t *testing.T) *realmController {
	t.Helper()

	trust, err := networktrust.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	return newRealmController(&application{
		ctx:          context.Background(),
		networkTrust: trust,
		gameSettings: preferences.NewTransient(),
	})
}

// TestRealmControllerSignupAndRecoveryDoNotImplicitlyLogin keeps authorization explicit.
func TestRealmControllerSignupAndRecoveryDoNotImplicitlyLogin(t *testing.T) {
	controller := newRealmController(&application{ctx: context.Background()})
	controller.client = &fakeRealmAPI{}
	controller.state.Phase = "login"

	if err := controller.Signup("Alice", "alice@example.test", "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}

	waitRealmPhase(t, controller, "verification_required")

	if err := controller.RecoverPassword("alice@example.test"); err != nil {
		t.Fatal(err)
	}

	waitRealmPhase(t, controller, "recovery_sent")
}

// TestRealmControllerPersistsSelectedGateway verifies the preferred endpoint survives reload.
func TestRealmControllerPersistsSelectedGateway(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")

	settings, err := preferences.New(path)
	if err != nil {
		t.Fatal(err)
	}

	controller := newRealmController(&application{
		ctx:          context.Background(),
		gameSettings: settings,
	})
	if err := controller.SetGateway("realm.example"); err != nil {
		t.Fatal(err)
	}

	reloaded, err := preferences.New(path)
	if err != nil {
		t.Fatal(err)
	}

	if got := reloaded.Values().RealmGateway; got != "realm.example" {
		t.Fatalf("gateway=%q", got)
	}
}

// TestRealmControllerLogoutAndCloseClearLivePresence verifies both session exit paths.
func TestRealmControllerLogoutAndCloseClearLivePresence(t *testing.T) {
	api := &fakeRealmAPI{}
	controller := newRealmController(&application{ctx: context.Background()})
	controller.client = api
	controller.state = authenticatedRealmState()

	if err := controller.Logout(); err != nil {
		t.Fatal(err)
	}

	waitRealmPhase(t, controller, "login")

	if api.logouts != 1 || controller.state.Account.ID != "" {
		t.Fatalf("logouts=%d state=%#v", api.logouts, controller.state)
	}

	controller.state.Account = realm.Account{ID: "account", Name: "Alice"}
	if err := controller.Close(t.Context()); err != nil {
		t.Fatal(err)
	}

	if api.logouts != 2 || controller.state.Phase != "disconnected" {
		t.Fatalf("logouts=%d state=%#v", api.logouts, controller.state)
	}
}

// authenticatedRealmState returns the shared starting state for logout tests.
func authenticatedRealmState() realmClientState {
	return realmClientState{
		Phase:    "lobby",
		Gateway:  "realm.example",
		Endpoint: "https://realm.example:6112",
		Account:  realm.Account{ID: "account", Name: "Alice"},
	}
}

// waitRealmPhase waits for an asynchronous Realm operation to publish its final phase.
func waitRealmPhase(t *testing.T, controller *realmController, phase string) {
	t.Helper()

	// bcrypt and TLS are intentionally exercised here and are substantially
	// slower under the race detector and loaded CI hosts.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if controller.Status()["phase"] == phase {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatalf("phase=%#v, want %s", controller.Status(), phase)
}
