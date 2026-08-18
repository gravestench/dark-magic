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
	identity := clientSessionIdentity()
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
	hud := playeradapter.HUD{Version: playeradapter.HUDVersion, Tick: 0, Player: playeradapter.HUDIdentity{PlayerID: "player", CharacterID: "character", Name: "Hero", Class: "Amazon"}}
	view := playeradapter.ClientView{Version: playeradapter.ClientViewVersion, Tick: 0, HUD: hud,
		World:   playeradapter.WorldView{Version: playeradapter.WorldViewVersion, Tick: 0, Entities: []playeradapter.WorldEntity{}},
		Private: playeradapter.PrivateView{Version: playeradapter.PrivateViewVersion, Tick: 0},
		Party: playeradapter.PartyView{Version: playeradapter.PartyViewVersion, Tick: 0,
			Roster: []playeradapter.PartyRosterEntry{{PlayerID: "player", Name: "Hero", Class: "Amazon", Level: 1, Relationship: "self"}}},
		Events: playeradapter.EventView{Version: playeradapter.EventViewVersion, Tick: 0, Events: []playeradapter.SemanticEvent{}}}
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
	if connected.PresentationSnapshot().EventEpoch != 1 {
		t.Fatalf("initial semantic event epoch = %d, want 1", connected.PresentationSnapshot().EventEpoch)
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
	if connected.PresentationSnapshot().EventEpoch != 2 {
		t.Fatalf("reconnect semantic event epoch = %d, want 2", connected.PresentationSnapshot().EventEpoch)
	}
	firstCredential = connected.credential
	if err := connected.transport.Close(); err != nil {
		t.Fatal(err)
	}
	if err := connected.Reconnect(ctx); err != nil {
		t.Fatalf("redial reconnect: %v", err)
	}
	if connected.credential == firstCredential {
		t.Fatal("redial reconnect did not rotate credential")
	}
	if connected.PresentationSnapshot().EventEpoch != 3 {
		t.Fatalf("redial semantic event epoch = %d, want 3", connected.PresentationSnapshot().EventEpoch)
	}
	replacementTicket, err := authority.Issue(gameserver.Principal{ID: "account", CharacterID: "character",
		PlayerID: "player", CharacterRevision: 2, RuntimeIdentityHash: allocation.IdentityHash}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	beforeReassignment := connected.credential
	if err := connected.Reassign(ctx, realm.JoinAssignment{GameID: "game", Endpoint: assignment.Endpoint,
		Ticket: replacementTicket, Runtime: identity}, clientTLS); err != nil {
		t.Fatal(err)
	}
	if connected.credential == beforeReassignment || connected.HUD.Player.PlayerID != "player" ||
		connected.Admission.Admission.CharacterID != "character" {
		t.Fatalf("reassigned session = %#v", connected)
	}
	if connected.PresentationSnapshot().EventEpoch != 4 {
		t.Fatalf("reassignment semantic event epoch = %d, want 4", connected.PresentationSnapshot().EventEpoch)
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

func clientSessionIdentity() simulation.RuntimeIdentity {
	const packageDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	return simulation.RuntimeIdentity{Recipe: simulation.RuntimeRecipe{
		Schema: simulation.RuntimeRecipeSchema, EngineAPI: "v1", NetworkProtocol: "test/v1", AssetSetID: simulation.EmptyAssetSetID,
		GameDataGenerationID: simulation.GameDataGenerationIDForAssetSet(simulation.EmptyAssetSetID),
		Packages:             simulation.RuntimePackageSet{Base: simulation.RuntimePackage{ID: "d2legacy", Version: "1.0.0", Digest: packageDigest, Size: 1, Redistributable: true}},
		AuthoritativeHash:    "rules", ConfigurationHash: "config",
	}}
}

func TestSessionExposesDistinctNetworkTimelines(t *testing.T) {
	now := time.Unix(600, 0)
	session := &Session{Admission: gameserver.JoinResponse{Snapshot: gameserver.Snapshot{Tick: 20, StepNanos: int64(40 * time.Millisecond)}}}
	timeline := session.NetworkTimeline(now)
	if !timeline.Ready || timeline.Prediction.Tick != 20 || timeline.Interpolation.Tick != 18 {
		t.Fatalf("network timeline = %#v", timeline)
	}
	if got := session.NextInputTick(now); got != 22 {
		t.Fatalf("next input tick = %d, want 22", got)
	}
	if got := session.NextInputTick(now.Add(time.Second)); got != 22 {
		t.Fatalf("next input tick after stale extrapolation = %d, want 22", got)
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

func TestCorrectionCannotReplaceAuthenticatedOwnerIdentity(t *testing.T) {
	owner := playeradapter.HUDIdentity{PlayerID: "player", CharacterID: "hero"}
	session := &Session{
		Admission:   gameserver.JoinResponse{Snapshot: gameserver.Snapshot{Tick: 1, Checksum: "before"}},
		reliableHUD: playeradapter.HUD{Player: owner},
		reliableWorld: playeradapter.WorldView{
			Version: playeradapter.WorldViewVersion, Tick: 1, Entities: []playeradapter.WorldEntity{},
		},
		pending: make(map[uint64]gameserver.CommandIntent),
	}
	view := validNetworkView(2)
	view.HUD.Player = playeradapter.HUDIdentity{PlayerID: "attacker", CharacterID: "hero"}
	if _, err := session.applyDecodedCorrection(gameserver.Snapshot{Tick: 2, Checksum: "after"}, view); err != ErrAssignment {
		t.Fatalf("owner replacement error = %v", err)
	}
	if session.reliableHUD.Player != owner {
		t.Fatalf("owner changed after rejected correction: %#v", session.reliableHUD.Player)
	}
}

func TestViewReturnsDefensiveWorldEntityState(t *testing.T) {
	health, maximum := int64(3), int64(5)
	session := &Session{HUD: playeradapter.HUD{Tick: 7}, World: playeradapter.WorldView{Tick: 7,
		Entities: []playeradapter.WorldEntity{{ID: "hostile", Health: &health, MaxHealth: &maximum}}}}
	_, view := session.View()
	view.Entities[0].ID = "changed"
	*view.Entities[0].Health = 0
	_, unchanged := session.View()
	if unchanged.Entities[0].ID != "hostile" || *unchanged.Entities[0].Health != 3 {
		t.Fatalf("session view was mutated through returned copy: %#v", unchanged)
	}
}

func TestCorrectionAcknowledgesOnlyContiguousInputHistory(t *testing.T) {
	session := &Session{pending: map[uint64]gameserver.CommandIntent{
		1: {Sequence: 1, Payload: json.RawMessage(`{"x":1}`)},
		2: {Sequence: 2, Payload: json.RawMessage(`{"x":2}`)},
		3: {Sequence: 3, Payload: json.RawMessage(`{"x":3}`)},
	}}
	session.discardAcknowledgedLocked(2)
	pending := session.PendingInputs()
	if len(pending) != 1 || pending[0].Sequence != 3 {
		t.Fatalf("pending inputs = %#v", pending)
	}
	pending[0].Payload[0] = '['
	if string(session.pending[3].Payload) != `{"x":3}` {
		t.Fatal("pending input payload was not defensively copied")
	}
}

func TestStagedInputIsImmediatelyPredictableAndDefensive(t *testing.T) {
	session := &Session{pending: make(map[uint64]gameserver.CommandIntent)}
	payload := json.RawMessage(`{"x":1}`)
	intent := gameserver.CommandIntent{TargetTick: 12, Sequence: 1, Kind: "player.move", Payload: payload}
	if err := session.StageInput(intent); err != nil {
		t.Fatal(err)
	}
	payload[5] = '9'
	pending := session.PendingInputs()
	if len(pending) != 1 || string(pending[0].Payload) != `{"x":1}` {
		t.Fatalf("pending inputs = %#v", pending)
	}
	conflict := intent
	conflict.TargetTick++
	if err := session.StageInput(conflict); err == nil {
		t.Fatal("conflicting staged sequence was accepted")
	}
	session.DiscardInput(intent.Sequence)
	if pending = session.PendingInputs(); len(pending) != 0 {
		t.Fatalf("discarded input remains pending: %#v", pending)
	}
}

func TestTransformFrameUpdatesKnownEntitiesWithoutInventingLifecycle(t *testing.T) {
	session := &Session{
		HUD: playeradapter.HUD{Version: playeradapter.HUDVersion, Tick: 10},
		World: playeradapter.WorldView{Version: playeradapter.WorldViewVersion, Tick: 10,
			Entities: []playeradapter.WorldEntity{{ID: "known", Position: playeradapter.HUDPosition{X: 1, Y: 2}}},
			Missiles: []playeradapter.WorldMissile{{ID: "missile:1", Position: playeradapter.HUDPosition{X: 2, Y: 3}}}},
		Admission: gameserver.JoinResponse{Snapshot: gameserver.Snapshot{Tick: 10, StepNanos: int64(40 * time.Millisecond)}},
	}
	delta, err := session.applyTransform(sessionquic.TransformFrame{
		Tick: 11, OwnerX: 4, OwnerY: 5, VelocityX: 6, VelocityY: 7,
		Entities: []sessionquic.TransformEntity{
			{IDHash: sessionquic.PublicIDHash("known"), X: 8, Y: 9, Direction: 3, Mode: [2]byte{'W', 'L'}},
			{IDHash: sessionquic.PublicIDHash("missile:1"), X: 10, Y: 11},
			{IDHash: sessionquic.PublicIDHash("unknown"), X: 99, Y: 99},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if session.World.Tick != 11 || len(session.World.Entities) != 1 ||
		session.World.Entities[0].Position != (playeradapter.HUDPosition{X: 8, Y: 9}) ||
		session.World.Entities[0].Direction != 3 || session.World.Entities[0].Mode != "WL" {
		t.Fatalf("transformed world = %#v", session.World)
	}
	if len(session.World.Missiles) != 1 || session.World.Missiles[0].Position != (playeradapter.HUDPosition{X: 10, Y: 11}) {
		t.Fatalf("transformed missiles = %#v", session.World.Missiles)
	}
	if session.HUD.Position != (playeradapter.HUDPosition{X: 4, Y: 5}) || session.HUD.Movement.Velocity != (playeradapter.HUDPosition{X: 6, Y: 7}) {
		t.Fatalf("transformed HUD = %#v", session.HUD)
	}
	if len(delta.Upserts) != 1 || delta.Upserts[0].ID != "known" {
		t.Fatalf("transform delta = %#v", delta)
	}
}

func TestPresentationSnapshotsAreImmutableAtomicRevisions(t *testing.T) {
	session := &Session{
		HUD:       playeradapter.HUD{Version: playeradapter.HUDVersion, Tick: 10},
		World:     playeradapter.WorldView{Version: playeradapter.WorldViewVersion, Tick: 10, Entities: []playeradapter.WorldEntity{{ID: "known", Position: playeradapter.HUDPosition{X: 1}}}},
		Events:    playeradapter.EventView{Version: playeradapter.EventViewVersion, Tick: 10, Events: []playeradapter.SemanticEvent{{ID: 1, Type: "cast", Tick: 10, Cast: &playeradapter.SemanticCastCue{Kind: "cast_started", Player: "alice"}}}},
		Admission: gameserver.JoinResponse{Snapshot: gameserver.Snapshot{Tick: 10, StepNanos: int64(40 * time.Millisecond)}},
	}
	before := session.PresentationSnapshot()
	if again := session.PresentationSnapshot(); again != before {
		t.Fatal("unchanged presentation allocated a new snapshot")
	}
	session.Events.Events[0].Cast.Kind = "mutated"
	if before.Events.Events[0].Cast.Kind != "cast_started" {
		t.Fatal("published semantic event aliases mutable session state")
	}
	if _, err := session.applyTransform(sessionquic.TransformFrame{Tick: 11, Entities: []sessionquic.TransformEntity{{IDHash: sessionquic.PublicIDHash("known"), X: 2}}}); err != nil {
		t.Fatal(err)
	}
	after := session.PresentationSnapshot()
	if after == before || after.Revision <= before.Revision {
		t.Fatalf("presentation revisions before=%p/%d after=%p/%d", before, before.Revision, after, after.Revision)
	}
	if before.World.Tick != 10 || before.World.Entities[0].Position.X != 1 || after.World.Tick != 11 || after.World.Entities[0].Position.X != 2 {
		t.Fatalf("immutable snapshots before=%#v after=%#v", before.World, after.World)
	}
}

func TestReliableMergePreservesCanonicalInputAndNewerTransforms(t *testing.T) {
	reliable := playeradapter.WorldView{Tick: 10,
		Entities: []playeradapter.WorldEntity{{ID: "known", Position: playeradapter.HUDPosition{X: 1}}},
		Missiles: []playeradapter.WorldMissile{{ID: "missile:1", Position: playeradapter.HUDPosition{X: 4}}}}
	current := playeradapter.WorldView{Tick: 11, Origin: playeradapter.HUDPosition{X: 3},
		Entities: []playeradapter.WorldEntity{{ID: "known", Position: playeradapter.HUDPosition{X: 2}}},
		Missiles: []playeradapter.WorldMissile{{ID: "missile:1", Position: playeradapter.HUDPosition{X: 5}}}}
	_, merged := mergeReliablePresentation(playeradapter.HUD{Tick: 10}, reliable, playeradapter.HUD{Tick: 11}, current)
	if reliable.Entities[0].Position.X != 1 {
		t.Fatalf("canonical reliable projection was mutated: %#v", reliable)
	}
	if merged.Tick != 11 || merged.Entities[0].Position.X != 2 || merged.Missiles[0].Position.X != 5 {
		t.Fatalf("merged presentation = %#v", merged)
	}
}

func TestLatestDeltaPublicationDropsObsoleteNotification(t *testing.T) {
	deltas := make(chan playeradapter.WorldDelta, 1)
	publishLatestDelta(context.Background(), deltas, playeradapter.WorldDelta{Tick: 1})
	publishLatestDelta(context.Background(), deltas, playeradapter.WorldDelta{Tick: 2})
	if got := <-deltas; got.Tick != 2 {
		t.Fatalf("published delta tick = %d, want 2", got.Tick)
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
