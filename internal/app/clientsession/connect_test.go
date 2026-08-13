package clientsession

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/app/serverapp"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

func TestConnectVerifiesAssignmentTLSRuntimeAndHUD(t *testing.T) {
	identity := simulation.RuntimeIdentity{ModID: "d2legacy", ContractVersion: "v1", PackageHash: "package", AuthoritativeHash: "rules", ConfigurationHash: "config"}
	allocation, err := gamesession.Allocate("game", identity, gamesession.PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	if err := session.Register(playeradapter.EnterCommand, gamesession.CommandHandler{
		Validate: func(simulation.Command) error { return nil }, Apply: func(*gameecs.Engine, simulation.Command) error { return nil },
		Allowed: []simulation.Authority{simulation.AuthoritySystem},
	}); err != nil {
		t.Fatal(err)
	}
	authority, err := gameserver.NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	if err != nil {
		t.Fatal(err)
	}
	ticket, err := authority.Issue(gameserver.Principal{ID: "account", CharacterID: "character", PlayerID: "player", CharacterRevision: 2, RuntimeIdentityHash: allocation.IdentityHash}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	hud := playeradapter.HUD{Version: playeradapter.HUDVersion, Tick: 0, Player: playeradapter.HUDIdentity{CharacterID: "character", Name: "Hero", Class: "Amazon"}}
	view := playeradapter.ClientView{Version: playeradapter.ClientViewVersion, Tick: 0, HUD: hud, World: playeradapter.WorldView{Version: playeradapter.WorldViewVersion, Tick: 0, Entities: []playeradapter.WorldEntity{}}}
	endpoint, err := gameserver.NewEndpoint(&gameserver.Host{Engine: engine, Session: session, Allocation: allocation}, authority,
		func(string, simulation.Checkpoint) (json.RawMessage, error) { return json.Marshal(view) })
	if err != nil {
		t.Fatal(err)
	}
	serverTLS, clientTLS, fingerprint := connectTLS(t)
	server, err := sessionquic.Listen("127.0.0.1:0", serverTLS, endpoint)
	if err != nil {
		t.Fatal(err)
	}
	destination, _ := playeradapter.NewDestination(10, 20, 100, 100, 1, 40)
	profiles, err := serverapp.NewRemoteProfileAdmissions(&gameserver.Host{Mode: gameserver.ModeStandalone, Engine: engine, Session: session, Allocation: allocation}, authority,
		serverapp.RemoteProfileConfig{Credential: "profile-secret", PrincipalID: "self-host-user", PlayerID: "player", Destination: destination, Lifetime: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	server.SetProfileAdmissions(profiles)
	t.Cleanup(func() { _ = server.Close() })
	serveContext, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = server.Serve(serveContext) }()
	assignment := realm.JoinAssignment{GameID: "game", Endpoint: realm.GameEndpoint{Address: server.Addr(), TLSFingerprint: fingerprint}, Ticket: ticket, Runtime: identity}

	wrong := assignment
	wrong.Endpoint.TLSFingerprint = "sha256:" + strings.Repeat("0", 64)
	ctx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	if _, err := Connect(ctx, wrong, clientTLS); err == nil {
		t.Fatal("wrong TLS fingerprint was accepted")
	}
	connected, err := Connect(ctx, assignment, clientTLS)
	if err != nil {
		t.Fatal(err)
	}
	if connected.HUD.Player.Name != "Hero" || connected.Admission.Admission.IdentityHash != allocation.IdentityHash {
		t.Fatalf("session = %#v", connected)
	}
	delta, err := connected.Refresh(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(delta.Upserts) != 0 || len(delta.Removed) != 0 {
		t.Fatalf("unchanged refresh delta = %#v", delta)
	}
	watchContext, cancelWatch := context.WithCancel(ctx)
	deltas, watchErrors, err := connected.Watch(watchContext)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case delta := <-deltas:
		if len(delta.Upserts) != 0 || len(delta.Removed) != 0 {
			t.Fatalf("watch delta = %#v", delta)
		}
	case err := <-watchErrors:
		t.Fatalf("watch error = %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	cancelWatch()
	firstCredential := connected.credential
	if err := connected.Reconnect(ctx); err != nil {
		t.Fatal(err)
	}
	if connected.credential == firstCredential {
		t.Fatal("reconnect did not rotate credential")
	}
	if err := connected.Close(ctx); err != nil {
		t.Fatal(err)
	}
	profile := d2save.New(d2save.Character{ID: "character", Name: "Hero", Class: "Amazon"})
	if err := profile.Select("character"); err != nil {
		t.Fatal(err)
	}
	selfHosted, err := ConnectSelfHosted(ctx, SelfHostedAssignment{GameID: "game", Endpoint: assignment.Endpoint,
		Runtime: identity, ProfileCredential: "profile-secret"}, clientTLS, profile)
	if err != nil {
		t.Fatal(err)
	}
	if selfHosted.HUD.Player.CharacterID != "character" || selfHosted.Admission.Admission.CharacterID != "character" {
		t.Fatalf("self-hosted session = %#v", selfHosted)
	}
	if err := selfHosted.Close(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestConnectRejectsMalformedDiscoveryBeforeDial(t *testing.T) {
	if _, err := Connect(context.Background(), realm.JoinAssignment{GameID: "game", Ticket: "ticket", Endpoint: realm.GameEndpoint{Address: "https://example", TLSFingerprint: "bad"}}, &tls.Config{}); err == nil {
		t.Fatal("malformed discovery was accepted")
	}
}

func TestCorrectionRejectsStaleAndConflictingSnapshots(t *testing.T) {
	current := gameserver.Snapshot{Tick: 8, Checksum: "current"}
	if err := validateCorrection(current, gameserver.Snapshot{Tick: 7, Checksum: "old"}); err != ErrStaleCorrection {
		t.Fatalf("stale error = %v", err)
	}
	if err := validateCorrection(current, gameserver.Snapshot{Tick: 8, Checksum: "different"}); err != ErrStaleCorrection {
		t.Fatalf("conflict error = %v", err)
	}
	if err := validateCorrection(current, gameserver.Snapshot{Tick: 9, Checksum: "next"}); err != nil {
		t.Fatal(err)
	}
}

func connectTLS(t *testing.T) (*tls.Config, *tls.Config, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "localhost"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certificate, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(certificate)
	sum := sha256.Sum256(der)
	return &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}}, &tls.Config{RootCAs: pool, ServerName: "127.0.0.1"}, "sha256:" + hex.EncodeToString(sum[:])
}
