package serverapp

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/app/realm"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const workerControlTestToken = "0123456789abcdef0123456789abcdef"

type workerControlFixture struct {
	directory   string
	clientTLS   *tls.Config
	fingerprint string
	host        *gameserver.Host
	authority   *gameserver.TicketAuthority
	drain       chan struct{}
}

// TestWorkerControlUsesPinnedTLSAndSharedTicketAuthority verifies the complete
// allocator channel: certificate pinning, runtime description, shared ticket
// revocation, drain signaling, and cancellation-driven shutdown.
func TestWorkerControlUsesPinnedTLSAndSharedTicketAuthority(t *testing.T) {
	fixture := newWorkerControlFixture(t)

	control := startWorkerControlForTest(t, fixture)
	if control.TLSFingerprint() != fixture.fingerprint {
		t.Fatalf("fingerprint = %q, want %q", control.TLSFingerprint(), fixture.fingerprint)
	}

	serveContext, cancelServe := context.WithCancel(t.Context())
	t.Cleanup(cancelServe)

	serveDone := make(chan error, 1)
	go func() {
		serveDone <- control.Serve(serveContext)
	}()

	client := newWorkerControlClientForTest(t, control, fixture)

	description, err := client.Describe(t.Context())
	if err != nil || description.IdentityHash != fixture.host.Allocation.IdentityHash {
		t.Fatalf("description=%#v error=%v", description, err)
	}

	ticket, err := client.Issue(t.Context(), realm.AdmissionPrincipal{
		AccountID:           "account",
		CharacterID:         "character",
		PlayerID:            "player",
		CharacterRevision:   1,
		RuntimeIdentityHash: fixture.host.Allocation.IdentityHash,
	}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.Revoke(t.Context(), ticket); err != nil {
		t.Fatal(err)
	}

	if _, err := fixture.authority.Authenticate(t.Context(), ticket); err == nil {
		t.Fatal("ticket revoked over worker control remained valid in QUIC authority")
	}

	if err := client.Close(t.Context()); err != nil {
		t.Fatal(err)
	}

	waitForWorkerDrain(t, fixture.drain)

	cancelServe()
	waitForWorkerControlStop(t, serveDone)
}

// newWorkerControlFixture creates a pinned TLS identity and runtime whose
// cleanup order releases the session before its ECS storage.
func newWorkerControlFixture(t *testing.T) workerControlFixture {
	t.Helper()

	directory := t.TempDir()

	trust, err := networktrust.New(directory)
	if err != nil {
		t.Fatal(err)
	}

	_, clientTLS, fingerprint, err := trust.HostTLS()
	if err != nil {
		t.Fatal(err)
	}

	tokenPath := filepath.Join(directory, "control-token")
	if err := os.WriteFile(tokenPath, []byte(workerControlTestToken), 0o600); err != nil {
		t.Fatal(err)
	}

	engine := gameecs.New()

	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = session.Close()
		_ = engine.Close()
	})

	allocation, err := gamesession.Allocate("game", workerControlIdentity(), gamesession.PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}

	host := &gameserver.Host{Engine: engine, Session: session, Allocation: allocation}

	authority, err := gameserver.NewTicketAuthority([]byte("abcdef0123456789abcdef0123456789"), "game")
	if err != nil {
		t.Fatal(err)
	}

	return workerControlFixture{
		directory:   directory,
		clientTLS:   clientTLS,
		fingerprint: fingerprint,
		host:        host,
		authority:   authority,
		drain:       make(chan struct{}, 1),
	}
}

// startWorkerControlForTest binds the private listener and registers a fallback
// shutdown so an assertion failure cannot leave a socket or goroutine behind.
func startWorkerControlForTest(t *testing.T, fixture workerControlFixture) *WorkerControlServer {
	t.Helper()

	control, err := StartWorkerControl(WorkerControlConfig{
		Address:         "127.0.0.1:0",
		CertificatePath: filepath.Join(fixture.directory, "host-certificate.pem"),
		PrivateKeyPath:  filepath.Join(fixture.directory, "host-identity.pem"),
		TokenPath:       filepath.Join(fixture.directory, "control-token"),
		Tickets:         fixture.authority,
		Memberships:     realm.NewWorkerMemberships(),
		Drain: func() {
			fixture.drain <- struct{}{}
		},
	}, fixture.host)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_ = control.Close(shutdownContext)
	})

	return control
}

// newWorkerControlClientForTest pins the generated host certificate and bounds
// every request so transport regressions fail instead of hanging the test.
func newWorkerControlClientForTest(
	t *testing.T,
	control *WorkerControlServer,
	fixture workerControlFixture,
) *realm.WorkerHTTPClient {
	t.Helper()

	transport := &http.Transport{TLSClientConfig: fixture.clientTLS}
	t.Cleanup(transport.CloseIdleConnections)

	client, err := realm.NewWorkerHTTPClient(
		"https://"+control.Addr().String(),
		workerControlTestToken,
		&http.Client{Transport: transport, Timeout: 2 * time.Second},
	)
	if err != nil {
		t.Fatal(err)
	}

	return client
}

// waitForWorkerDrain gives the asynchronous HTTP handler a bounded opportunity
// to fence new admissions before the test cancels the server.
func waitForWorkerDrain(t *testing.T, drain <-chan struct{}) {
	t.Helper()

	select {
	case <-drain:
	case <-time.After(time.Second):
		t.Fatal("worker control did not signal drain")
	}
}

// waitForWorkerControlStop joins the Serve goroutine so shutdown errors cannot
// be lost after the test returns.
func waitForWorkerControlStop(t *testing.T, serveDone <-chan error) {
	t.Helper()

	select {
	case err := <-serveDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("worker control did not stop")
	}
}

// workerControlIdentity keeps the control-plane description deterministic
// without loading content or starting gameplay systems.
func workerControlIdentity() simulation.RuntimeIdentity {
	const packageDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	return simulation.RuntimeIdentity{
		Recipe: simulation.RuntimeRecipe{
			Schema:               simulation.RuntimeRecipeSchema,
			EngineAPI:            "v1",
			NetworkProtocol:      "test/v1",
			AssetSetID:           simulation.EmptyAssetSetID,
			GameDataGenerationID: simulation.GameDataGenerationIDForAssetSet(simulation.EmptyAssetSetID),
			Packages: simulation.RuntimePackageSet{
				Base: simulation.RuntimePackage{
					ID:      "d2legacy",
					Version: "1.0.0",
					Digest:  packageDigest,
					Size:    1,
				},
			},
			AuthoritativeHash: "rules",
			ConfigurationHash: "config",
		},
	}
}

// TestWorkerControlRejectsPublicListenAddress ensures the allocator channel
// cannot be exposed on wildcard or non-loopback interfaces.
func TestWorkerControlRejectsPublicListenAddress(t *testing.T) {
	config := WorkerControlConfig{
		Address:         "0.0.0.0:1234",
		CertificatePath: "cert",
		PrivateKeyPath:  "key",
		TokenPath:       "token",
		Tickets:         &gameserver.TicketAuthority{},
	}
	if _, err := StartWorkerControl(config, &gameserver.Host{}); err == nil {
		t.Fatal("public worker-control address was accepted")
	}
}
