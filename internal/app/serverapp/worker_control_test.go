package serverapp

import (
	"context"
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

func TestWorkerControlUsesPinnedTLSAndSharedTicketAuthority(t *testing.T) {
	directory := t.TempDir()
	trust, err := networktrust.New(directory)
	if err != nil {
		t.Fatal(err)
	}
	_, clientTLS, fingerprint, err := trust.HostTLS()
	if err != nil {
		t.Fatal(err)
	}
	token := "0123456789abcdef0123456789abcdef"
	tokenPath := filepath.Join(directory, "control-token")
	if err := os.WriteFile(tokenPath, []byte(token), 0o600); err != nil {
		t.Fatal(err)
	}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	identity := simulation.RuntimeIdentity{Recipe: simulation.RuntimeRecipe{Schema: simulation.RuntimeRecipeSchema,
		EngineAPI: "v1", NetworkProtocol: "test/v1", AssetSetID: simulation.EmptyAssetSetID, Packages: simulation.RuntimePackageSet{Base: simulation.RuntimePackage{
			ID: "d2legacy", Version: "1.0.0", Digest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Size: 1}},
		AuthoritativeHash: "rules", ConfigurationHash: "config"}}
	allocation, err := gamesession.Allocate("game", identity, gamesession.PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}
	host := &gameserver.Host{Engine: engine, Session: session, Allocation: allocation}
	authority, err := gameserver.NewTicketAuthority([]byte("abcdef0123456789abcdef0123456789"), "game")
	if err != nil {
		t.Fatal(err)
	}
	drain := make(chan struct{}, 1)
	control, err := StartWorkerControl(WorkerControlConfig{Address: "127.0.0.1:0",
		CertificatePath: filepath.Join(directory, "host-certificate.pem"), PrivateKeyPath: filepath.Join(directory, "host-identity.pem"),
		TokenPath: tokenPath, Tickets: authority, Memberships: realm.NewWorkerMemberships(), Drain: func() { drain <- struct{}{} }}, host)
	if err != nil {
		t.Fatal(err)
	}
	if control.TLSFingerprint() != fingerprint {
		t.Fatalf("fingerprint = %q, want %q", control.TLSFingerprint(), fingerprint)
	}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- control.Serve(ctx) }()
	transport := &http.Transport{TLSClientConfig: clientTLS}
	client, err := realm.NewWorkerHTTPClient("https://"+control.Addr().String(), token, &http.Client{Transport: transport, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	description, err := client.Describe(t.Context())
	if err != nil || description.IdentityHash != allocation.IdentityHash {
		t.Fatalf("description=%#v error=%v", description, err)
	}
	ticket, err := client.Issue(t.Context(), realm.AdmissionPrincipal{AccountID: "account", CharacterID: "character",
		PlayerID: "player", CharacterRevision: 1, RuntimeIdentityHash: allocation.IdentityHash}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Revoke(t.Context(), ticket); err != nil {
		t.Fatal(err)
	}
	if _, err := authority.Authenticate(t.Context(), ticket); err == nil {
		t.Fatal("ticket revoked over worker control remained valid in QUIC authority")
	}
	if err := client.Close(t.Context()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-drain:
	case <-time.After(time.Second):
		t.Fatal("worker control did not signal drain")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("worker control did not stop")
	}
}

func TestWorkerControlRejectsPublicListenAddress(t *testing.T) {
	if _, err := StartWorkerControl(WorkerControlConfig{Address: "0.0.0.0:1234", CertificatePath: "cert",
		PrivateKeyPath: "key", TokenPath: "token", Tickets: &gameserver.TicketAuthority{}}, &gameserver.Host{}); err == nil {
		t.Fatal("public worker-control address was accepted")
	}
}
