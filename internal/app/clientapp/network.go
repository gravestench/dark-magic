package clientapp

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
	"errors"
	"math/big"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/app/serverapp"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

type networkController struct {
	mu                            sync.Mutex
	app                           *application
	phase, mode, address, failure string
	host                          *gameserver.Host
	server                        *sessionquic.Server
	client                        *clientsession.Session
	cancel                        context.CancelFunc
}

func newNetworkController(app *application) *networkController {
	return &networkController{app: app, phase: "idle"}
}

func (controller *networkController) Host() error {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if controller.phase != "idle" && controller.phase != "failed" {
		return errors.New("network operation already active")
	}
	if _, selected := controller.app.saves.Selected(); !selected {
		return errors.New("select a character before hosting")
	}
	controller.phase, controller.mode, controller.failure = "starting", "host", ""
	go controller.startHost()
	return nil
}

func (controller *networkController) Join(address string) error {
	if strings.TrimSpace(address) == "" {
		return errors.New("server address is required")
	}
	return errors.New("direct join requires the host trust invitation; use the in-client Host flow in this slice")
}

func (controller *networkController) Status() map[string]any {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	return map[string]any{"phase": controller.phase, "mode": controller.mode, "address": controller.address, "error": controller.failure}
}

func (controller *networkController) startHost() {
	ctx, cancel := context.WithCancel(controller.app.ctx)
	fail := func(err error) {
		cancel()
		controller.mu.Lock()
		controller.phase, controller.failure = "failed", err.Error()
		controller.mu.Unlock()
	}
	host, err := gameserver.Start(ctx, controller.app.options.Content, recordstore.New(controller.app.options.Content), gameserver.Config{
		Mode: gameserver.ModeListen, SessionID: "listen-local", Prediction: gamesession.PredictionLimited,
	})
	if err != nil {
		fail(err)
		return
	}
	secret, err := randomBytes(32)
	if err != nil {
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	tickets, err := gameserver.NewTicketAuthority(secret, "listen-local")
	if err != nil {
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	endpoint, err := gameserver.NewEndpoint(host, tickets, playeradapter.ProjectClientView)
	if err != nil {
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	endpoint.SetSnapshotPending(func(err error) bool { return errors.Is(err, playeradapter.ErrHUDPlayer) })
	serverTLS, clientTLS, fingerprint, err := listenTLS()
	if err != nil {
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	server, err := sessionquic.Listen("127.0.0.1:0", serverTLS, endpoint)
	if err != nil {
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	level := controller.app.activeWorldLevel
	world, zone := controller.app.gameWorlds[level], controller.app.gameWorldZones[level]
	spawn, found := controller.app.gameWorldSpawns[level]
	if world == nil || zone == nil || !found {
		_ = server.Close()
		_ = host.Close(context.Background())
		fail(errors.New("active world has no trusted host destination"))
		return
	}
	request := zone.Request()
	destination, err := playeradapter.NewDestination(spawn[0], spawn[1], float64(world.WidthSubtiles), float64(world.HeightSubtiles), int64(request.Act), int64(request.LevelID))
	if err != nil {
		_ = server.Close()
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	credentialBytes, err := randomBytes(32)
	if err != nil {
		_ = server.Close()
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	profileCredential := hex.EncodeToString(credentialBytes)
	profiles, err := serverapp.NewRemoteProfileAdmissions(host, tickets, serverapp.RemoteProfileConfig{
		Credential: profileCredential, PrincipalID: "listen-local-user", PlayerID: "local-player", Destination: destination, Lifetime: time.Minute,
	})
	if err != nil {
		_ = server.Close()
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	server.SetProfileAdmissions(profiles)
	go host.Session.Run(ctx)
	go server.Serve(ctx)
	client, err := clientsession.ConnectSelfHosted(ctx, clientsession.SelfHostedAssignment{
		GameID: "listen-local", Endpoint: realm.GameEndpoint{Address: server.Addr(), TLSFingerprint: fingerprint},
		Runtime: host.Authority.Identity, ProfileCredential: profileCredential,
	}, clientTLS, controller.app.saves)
	if err != nil {
		_ = server.Close()
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	controller.mu.Lock()
	controller.host, controller.server, controller.client, controller.cancel = host, server, client, cancel
	controller.phase, controller.address = "connected", server.Addr()
	controller.mu.Unlock()
}

func (controller *networkController) Close() error {
	controller.mu.Lock()
	cancel, client, server, host := controller.cancel, controller.client, controller.server, controller.host
	controller.cancel, controller.client, controller.server, controller.host = nil, nil, nil, nil
	controller.phase = "closed"
	controller.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	ctx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	var err error
	if client != nil {
		err = errors.Join(err, client.Close(ctx))
	}
	if server != nil {
		err = errors.Join(err, server.Close())
	}
	if host != nil {
		err = errors.Join(err, host.Close(ctx))
	}
	return err
}

func randomBytes(size int) ([]byte, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return nil, err
	}
	return value, nil
}

func listenTLS() (*tls.Config, *tls.Config, string, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, "", err
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "localhost"},
		NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(24 * time.Hour), KeyUsage: x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, "", err
	}
	certificate, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, "", err
	}
	pool := x509.NewCertPool()
	pool.AddCert(certificate)
	sum := sha256.Sum256(der)
	return &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}},
		&tls.Config{RootCAs: pool, ServerName: "127.0.0.1"}, "sha256:" + hex.EncodeToString(sum[:]), nil
}
