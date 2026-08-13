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
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
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
	authority, err := gameserver.NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	if err != nil {
		t.Fatal(err)
	}
	ticket, err := authority.Issue(gameserver.Principal{ID: "account", CharacterID: "character", PlayerID: "player", CharacterRevision: 2, RuntimeIdentityHash: allocation.IdentityHash}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	hud := playeradapter.HUD{Version: playeradapter.HUDVersion, Tick: 0, Player: playeradapter.HUDIdentity{CharacterID: "character", Name: "Hero", Class: "Amazon"}}
	endpoint, err := gameserver.NewEndpoint(&gameserver.Host{Engine: engine, Session: session, Allocation: allocation}, authority,
		func(string, simulation.Checkpoint) (json.RawMessage, error) { return json.Marshal(hud) })
	if err != nil {
		t.Fatal(err)
	}
	serverTLS, clientTLS, fingerprint := connectTLS(t)
	server, err := sessionquic.Listen("127.0.0.1:0", serverTLS, endpoint)
	if err != nil {
		t.Fatal(err)
	}
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
}

func TestConnectRejectsMalformedDiscoveryBeforeDial(t *testing.T) {
	if _, err := Connect(context.Background(), realm.JoinAssignment{GameID: "game", Ticket: "ticket", Endpoint: realm.GameEndpoint{Address: "https://example", TLSFingerprint: "bad"}}, &tls.Config{}); err == nil {
		t.Fatal("malformed discovery was accepted")
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
