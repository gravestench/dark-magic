package sessionquic

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/json"
	"errors"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

type authenticator struct{}

type profileAdmissions struct{}

type packageProvider struct{ recipe simulation.RuntimeRecipe }

func (provider packageProvider) Recipe() simulation.RuntimeRecipe { return provider.recipe }
func (packageProvider) ReadChunk(_ context.Context, request PackageRequest) (PackageChunk, error) {
	return PackageChunk{ID: request.ID, Digest: request.Digest, Offset: request.Offset, Total: 3, Data: []byte("abc")}, nil
}

func (profileAdmissions) Admit(_ context.Context, credential string, offer []byte) (string, error) {
	if credential != "profile-secret" || string(offer) != `{"version":1}` {
		return "", ErrWire
	}
	return "session-ticket", nil
}

func (authenticator) Authenticate(_ context.Context, credential string) (gameserver.Principal, error) {
	if credential != "realm-ticket" {
		return gameserver.Principal{}, gameserver.ErrAuthentication
	}
	return gameserver.Principal{ID: "account", CharacterID: "character", PlayerID: "player"}, nil
}

func TestQUICJoinCommandAndReconnect(t *testing.T) {
	identity := testRuntimeIdentity()
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
	if err := session.Register("move", gamesession.CommandHandler{
		Validate: func(simulation.Command) error { return nil },
		Apply:    func(*gameecs.Engine, simulation.Command) error { return nil },
	}); err != nil {
		t.Fatal(err)
	}
	endpoint, err := gameserver.NewEndpoint(
		&gameserver.Host{Engine: engine, Session: session, Allocation: allocation},
		authenticator{},
		func(player string, _ simulation.Checkpoint) (json.RawMessage, error) {
			return json.Marshal(map[string]string{"player": player})
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	serverTLS, clientTLS := testTLS(t)
	server, err := Listen("127.0.0.1:0", serverTLS, endpoint)
	if err != nil {
		t.Fatal(err)
	}
	server.SetProfileAdmissions(profileAdmissions{})
	t.Cleanup(func() { _ = server.Close() })
	serveContext, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = server.Serve(serveContext) }()

	ctx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	client, err := Dial(ctx, server.Addr(), clientTLS)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	ticket, err := client.AdmitProfile(ctx, "profile-secret", []byte(`{"version":1}`))
	if err != nil || ticket != "session-ticket" {
		t.Fatalf("profile ticket=%q error=%v", ticket, err)
	}
	for attempt := 0; attempt < 256; attempt++ {
		stream, err := client.connection.OpenStreamSync(ctx)
		if err != nil {
			t.Fatal(err)
		}
		malformed := []byte(`{"operation":"leave","credential":"forged","unknown":true}`)
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], uint32(len(malformed)))
		if err := writeAll(stream, append(size[:], malformed...)); err != nil {
			t.Fatal(err)
		}
		var rejected response
		if err := readFrame(stream, &rejected); err != nil {
			t.Fatal(err)
		}
		if rejected.Error == "" {
			t.Fatalf("malformed stream %d was accepted", attempt)
		}
		stream.CancelRead(0)
		_ = stream.Close()
	}
	joined, err := client.Join(ctx, gameserver.JoinRequest{Version: gameserver.SessionProtocolVersion, Credential: "realm-ticket", Identity: identity})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Submit(ctx, joined.Credential, gameserver.CommandIntent{TargetTick: 1, Sequence: 1, Kind: "move", Payload: json.RawMessage(`{}`)}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Refresh(ctx, joined.Credential); err != nil {
		t.Fatal(err)
	}
	watchContext, cancelWatch := context.WithCancel(ctx)
	snapshots, watchErrors, err := client.Watch(watchContext, joined.Credential)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case snapshot := <-snapshots:
		if snapshot.Version != gameserver.SessionProtocolVersion {
			t.Fatalf("watch snapshot = %#v", snapshot)
		}
	case err := <-watchErrors:
		t.Fatalf("watch error = %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	cancelWatch()
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	reconnected, err := client.Reconnect(ctx, gameserver.ReconnectRequest{Credential: joined.Credential, Identity: identity, Nonce: "0123456789abcdef0123456789abcdef"})
	if err != nil {
		t.Fatal(err)
	}
	if reconnected.Snapshot.Tick != 1 || reconnected.Credential == joined.Credential {
		t.Fatalf("reconnect = %#v", reconnected)
	}
	if err := client.Leave(ctx, reconnected.Credential); err != nil {
		t.Fatal(err)
	}
}

func testRuntimeIdentity() simulation.RuntimeIdentity {
	const packageDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	return simulation.RuntimeIdentity{Recipe: simulation.RuntimeRecipe{
		Schema: simulation.RuntimeRecipeSchema, EngineAPI: "v1", NetworkProtocol: "test/v1", AssetSetID: simulation.EmptyAssetSetID,
		Packages:          simulation.RuntimePackageSet{Base: simulation.RuntimePackage{ID: "d2legacy", Version: "1.0.0", Digest: packageDigest, Size: 1, Redistributable: true}},
		AuthoritativeHash: "rules", ConfigurationHash: "config",
	}}
}

func TestConnectionMembershipsTrackRotatedReconnectCredential(t *testing.T) {
	memberships := &connectionMemberships{credentials: map[gameserver.SessionCredential]struct{}{
		"before": {},
	}}
	memberships.observe(request{
		Operation: operationReconnect,
		Reconnect: &gameserver.ReconnectRequest{Credential: "before"},
	}, response{Join: &gameserver.JoinResponse{Credential: "after"}})

	tracked := memberships.snapshot()
	if len(tracked) != 1 || tracked[0] != "after" {
		t.Fatalf("tracked reconnect memberships = %q", tracked)
	}
}

func TestFramesRejectOversizeAndUnknownFields(t *testing.T) {
	var oversized bytes.Buffer
	var size [4]byte
	binary.BigEndian.PutUint32(size[:], MaxFrameBytes+1)
	oversized.Write(size[:])
	if err := readFrame(&oversized, &request{}); err != ErrWire {
		t.Fatalf("oversized frame error = %v", err)
	}

	data := []byte(`{"operation":"leave","surprise":true}`)
	var unknown bytes.Buffer
	binary.BigEndian.PutUint32(size[:], uint32(len(data)))
	unknown.Write(size[:])
	unknown.Write(data)
	if err := readFrame(&unknown, &request{}); err == nil {
		t.Fatal("unknown field was accepted")
	}

	data = []byte(`{"operation":"leave"}{"operation":"join"}`)
	var trailing bytes.Buffer
	binary.BigEndian.PutUint32(size[:], uint32(len(data)))
	trailing.Write(size[:])
	trailing.Write(data)
	if err := readFrame(&trailing, &request{}); err == nil {
		t.Fatal("trailing message was accepted")
	}
}

func TestRemoteErrorDistinguishesSemanticRejectionFromTransportFailure(t *testing.T) {
	err := remoteError(gameserver.ErrRateLimit.Error())
	var remote *RemoteError
	if !errors.As(err, &remote) || remote.Message != gameserver.ErrRateLimit.Error() {
		t.Fatalf("remote error = %#v", err)
	}
}

func TestWireOperationsRejectAmbiguousShapes(t *testing.T) {
	server := &Server{}
	result := server.dispatch(context.Background(), request{Operation: operationLeave, Credential: "credential", Command: &gameserver.CommandIntent{}})
	if result.Error != ErrWire.Error() {
		t.Fatalf("ambiguous result = %#v", result)
	}
	if validShape(request{Operation: operationSubmit, Command: &gameserver.CommandIntent{}}) {
		t.Fatal("submit without credential was accepted")
	}
	result = server.dispatch(context.Background(), request{Operation: operationProfileAdmit, Credential: "secret", Offer: json.RawMessage(`{}`)})
	if result.Error != ErrWire.Error() {
		t.Fatalf("realm/default server enabled profile admission: %#v", result)
	}
}

func TestRecipeAndPackageChunkDispatchAreExactAndBounded(t *testing.T) {
	recipe := testRuntimeIdentity().Recipe
	server := &Server{packages: packageProvider{recipe: recipe}}
	recipeResult := server.dispatch(t.Context(), request{Operation: operationRecipe})
	if recipeResult.Error != "" || recipeResult.Recipe == nil || recipeResult.Recipe.Packages.Base != recipe.Packages.Base {
		t.Fatalf("recipe result = %#v", recipeResult)
	}
	chunkResult := server.dispatch(t.Context(), request{Operation: operationPackageChunk,
		Package: &PackageRequest{ID: "extension", Digest: "digest", Limit: 3}})
	if chunkResult.Error != "" || chunkResult.Package == nil || string(chunkResult.Package.Data) != "abc" {
		t.Fatalf("chunk result = %#v", chunkResult)
	}
	oversized := server.dispatch(t.Context(), request{Operation: operationPackageChunk,
		Package: &PackageRequest{ID: "extension", Digest: "digest", Limit: MaxPackageChunkBytes + 1}})
	if oversized.Error != ErrWire.Error() {
		t.Fatalf("oversized chunk result = %#v", oversized)
	}
}

func TestWireOperationShapesAreExhaustive(t *testing.T) {
	operations := []operation{operationJoin, operationSubmit, operationRefresh, operationWatch, operationReconnect, operationLeave, operationProfileAdmit, operationRecipe, operationPackageChunk, "unknown"}
	for _, candidate := range operations {
		for mask := 0; mask < 16; mask++ {
			message := request{Operation: candidate}
			if mask&1 != 0 {
				message.Credential = "credential"
			}
			if mask&2 != 0 {
				message.Join = &gameserver.JoinRequest{}
			}
			if mask&4 != 0 {
				message.Command = &gameserver.CommandIntent{}
			}
			if mask&8 != 0 {
				message.Offer = json.RawMessage(`{}`)
			}
			// Reconnect is tested separately because it is mutually exclusive
			// with every field represented by the mask.
			valid := validShape(message)
			expected := (candidate == operationJoin && mask == 2) ||
				(candidate == operationSubmit && mask == 5) ||
				((candidate == operationRefresh || candidate == operationWatch || candidate == operationLeave) && mask == 1) ||
				(candidate == operationProfileAdmit && (mask == 8 || mask == 9)) ||
				(candidate == operationRecipe && mask == 0)
			if valid != expected {
				t.Fatalf("operation=%q mask=%03b valid=%t want=%t", candidate, mask, valid, expected)
			}
			message.Reconnect = &gameserver.ReconnectRequest{}
			if got := validShape(message); got != (candidate == operationReconnect && mask == 0) {
				t.Fatalf("reconnect operation=%q mask=%03b valid=%t", candidate, mask, got)
			}
		}
	}
	packageMessage := request{Operation: operationPackageChunk, Package: &PackageRequest{ID: "extension", Digest: "digest", Limit: MaxPackageChunkBytes}}
	if !validShape(packageMessage) {
		t.Fatal("bounded package request was rejected")
	}
	packageMessage.Credential = "mixed"
	if validShape(packageMessage) {
		t.Fatal("ambiguous package request was accepted")
	}
}

func TestPackageRateLimiterBoundsBurstAndRefills(t *testing.T) {
	limiter := newPackageRateLimiter()
	now := time.Unix(100, 0)
	for consumed := 0; consumed < packageBurstBytes; consumed += MaxPackageChunkBytes {
		if !limiter.Allow(MaxPackageChunkBytes, now) {
			t.Fatalf("initial burst rejected at %d bytes", consumed)
		}
	}
	if limiter.Allow(1, now) {
		t.Fatal("package burst limit was not enforced")
	}
	if !limiter.Allow(packageBytesPerSecond, now.Add(time.Second)) {
		t.Fatal("package rate limiter did not refill")
	}
}

func TestPackageAndRecipeResponsesRejectMixedOrOutOfBoundsShapes(t *testing.T) {
	recipe := testRuntimeIdentity().Recipe
	if !validRecipeResponse(response{Recipe: &recipe}) {
		t.Fatal("canonical recipe response was rejected")
	}
	if validRecipeResponse(response{Recipe: &recipe, Ticket: "smuggled"}) {
		t.Fatal("mixed recipe response was accepted")
	}
	request := PackageRequest{ID: "extension", Digest: recipe.Packages.Base.Digest, Offset: 4, Limit: 8}
	chunk := PackageChunk{ID: request.ID, Digest: request.Digest, Offset: request.Offset, Total: 12, Data: []byte("payload")}
	if !validPackageResponse(response{Package: &chunk}, request) {
		t.Fatal("canonical package response was rejected")
	}
	chunk.Data = []byte("too-large")
	if validPackageResponse(response{Package: &chunk}, request) {
		t.Fatal("package response crossing total size was accepted")
	}
	chunk.Data = []byte("payload")
	if validPackageResponse(response{Package: &chunk, Snapshot: &gameserver.Snapshot{}}, request) {
		t.Fatal("mixed package response was accepted")
	}
}

func TestQUICConfigurationUsesConservativeInitialPacketSize(t *testing.T) {
	config := quicConfig()
	if config.InitialPacketSize != 1200 || config.DisablePathMTUDiscovery || !config.EnableDatagrams ||
		config.MaxIncomingStreams != 16 || config.MaxIncomingUniStreams != -1 ||
		config.MaxStreamReceiveWindow != MaxFrameBytes || config.MaxConnectionReceiveWindow != 2*MaxFrameBytes {
		t.Fatalf("unsafe QUIC configuration: %#v", config)
	}
}

func TestTypicalCommandEncodingFitsReservedDatagramBudget(t *testing.T) {
	message := request{
		Operation:  operationSubmit,
		Credential: gameserver.SessionCredential("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"),
		Command: &gameserver.CommandIntent{
			ObservedServerTick: 118, TargetTick: 120, Sequence: 41, Kind: "player.move",
			Payload: json.RawMessage(`{"destination":{"x":123.25,"y":94.5}}`),
		},
	}
	encoded, err := json.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) > MaxDatagramPayloadBytes {
		t.Fatalf("representative command encoding is %d bytes, budget is %d", len(encoded), MaxDatagramPayloadBytes)
	}
}

func FuzzReadFrame(f *testing.F) {
	f.Add([]byte{0, 0, 0, 2, '{', '}'})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})
	f.Fuzz(func(t *testing.T, data []byte) {
		var message request
		_ = readFrame(bytes.NewReader(data), &message)
	})
}

func testTLS(t *testing.T) (*tls.Config, *tls.Config) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "localhost"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	certificate, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	pool.AddCert(certificate)
	return &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}},
		&tls.Config{RootCAs: pool, ServerName: "127.0.0.1"}
}
