package serverapp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

const remoteProfileTestSecret = "0123456789abcdef0123456789abcdef"

type remoteProfileFixture struct {
	session     *gamesession.Session
	host        *gameserver.Host
	tickets     *gameserver.TicketAuthority
	destination playeradapter.Destination
	offer       []byte
}

// TestRemoteProfileAdmissionAuthenticatesQueuesAndIssuesOneUseTicket verifies
// that host policy controls authentication, ticket identity, command authority,
// and deterministic spawn spacing for consecutive remote players.
func TestRemoteProfileAdmissionAuthenticatesQueuesAndIssuesOneUseTicket(t *testing.T) {
	fixture := newRemoteProfileFixture(t, 100)
	admissions := newRemoteProfileAdmissionsForTest(t, fixture, RemoteProfileConfig{
		Credential:  "host-password",
		PrincipalID: "self-host-user",
		PlayerID:    "player",
		Destination: fixture.destination,
		Lifetime:    time.Minute,
	})

	if _, err := admissions.Admit(context.Background(), "wrong", fixture.offer); !errors.Is(
		err,
		ErrRemoteProfileAdmission,
	) {
		t.Fatalf("credential error = %v", err)
	}

	ticket, err := admissions.Admit(context.Background(), "host-password", fixture.offer)
	if err != nil {
		t.Fatal(err)
	}

	assertRemoteProfilePrincipal(t, fixture, ticket)

	if _, err := admissions.Admit(context.Background(), "host-password", fixture.offer); err != nil {
		t.Fatalf("second player admission: %v", err)
	}

	if err := fixture.session.Step(); err != nil {
		t.Fatal(err)
	}

	assertRemoteProfileSpawns(t, fixture.session)
}

// TestRemoteProfileAdmissionIsUnavailableToRealmHosts protects the ownership
// boundary that reserves remote profile admission for standalone/listen hosts.
func TestRemoteProfileAdmissionIsUnavailableToRealmHosts(t *testing.T) {
	tickets, err := gameserver.NewTicketAuthority([]byte(remoteProfileTestSecret), "game")
	if err != nil {
		t.Fatal(err)
	}

	host := &gameserver.Host{Mode: gameserver.ModeRealm, Session: &gamesession.Session{}}
	config := RemoteProfileConfig{
		Credential:  "secret",
		PrincipalID: "user",
		PlayerID:    "player",
		Lifetime:    time.Minute,
	}

	_, err = NewRemoteProfileAdmissions(host, tickets, config)
	if !errors.Is(err, ErrRemoteProfileAdmission) {
		t.Fatalf("realm error = %v", err)
	}
}

// newRemoteProfileFixture creates isolated runtime ownership and a valid player
// offer while leaving each test's admission policy explicit at the call site.
func newRemoteProfileFixture(t *testing.T, destinationWidth float64) remoteProfileFixture {
	t.Helper()

	engine := gameecs.New()

	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	// The session must release its work before the ECS storage it references.
	t.Cleanup(func() {
		_ = session.Close()
		_ = engine.Close()
	})

	if err := session.Register(playeradapter.EnterCommand, gamesession.CommandHandler{
		Validate: func(simulation.Command) error { return nil },
		Apply:    func(*gameecs.Engine, simulation.Command) error { return nil },
		Allowed:  []simulation.Authority{simulation.AuthoritySystem},
	}); err != nil {
		t.Fatal(err)
	}

	allocation, err := gamesession.Allocate("game", remoteProfileIdentity(), gamesession.PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}

	host := &gameserver.Host{
		Mode:       gameserver.ModeStandalone,
		Engine:     engine,
		Session:    session,
		Allocation: allocation,
	}

	tickets, err := gameserver.NewTicketAuthority([]byte(remoteProfileTestSecret), "game")
	if err != nil {
		t.Fatal(err)
	}

	destination, err := playeradapter.NewDestination(10, 20, destinationWidth, 100, 1, 40)
	if err != nil {
		t.Fatal(err)
	}

	offer, err := d2save.EncodeCharacterOffer(d2save.Character{
		ID:    "hero",
		Name:  "Hero",
		Class: "Amazon",
	})
	if err != nil {
		t.Fatal(err)
	}

	return remoteProfileFixture{
		session:     session,
		host:        host,
		tickets:     tickets,
		destination: destination,
		offer:       offer,
	}
}

// newRemoteProfileAdmissionsForTest fails at fixture construction so scenario
// tests can focus on request ordering and observable admission results.
func newRemoteProfileAdmissionsForTest(
	t *testing.T,
	fixture remoteProfileFixture,
	config RemoteProfileConfig,
) *RemoteProfileAdmissions {
	t.Helper()

	admissions, err := NewRemoteProfileAdmissions(fixture.host, fixture.tickets, config)
	if err != nil {
		t.Fatal(err)
	}

	return admissions
}

// assertRemoteProfilePrincipal checks that ticket identity is derived from host
// configuration and bound to the exact runtime allocation being served.
func assertRemoteProfilePrincipal(t *testing.T, fixture remoteProfileFixture, ticket string) {
	t.Helper()

	principal, err := fixture.tickets.Authenticate(context.Background(), ticket)
	if err != nil || principal.CharacterID != "hero" || principal.PlayerID != "player-1" ||
		principal.RuntimeIdentityHash != fixture.host.Allocation.IdentityHash {
		t.Fatalf("principal=%#v error=%v", principal, err)
	}
}

// assertRemoteProfileSpawns verifies command provenance and the deterministic
// separation applied to the first two admitted players.
func assertRemoteProfileSpawns(t *testing.T, session *gamesession.Session) {
	t.Helper()

	replay, err := session.Replay()
	if err != nil || len(replay.Commands) != 2 || replay.Commands[0].Player != "self-host:remote-profile" {
		t.Fatalf("replay=%#v error=%v", replay, err)
	}

	var first playeradapter.Entry
	if err := json.Unmarshal(replay.Commands[0].Payload, &first); err != nil {
		t.Fatal(err)
	}

	var second playeradapter.Entry
	if err := json.Unmarshal(replay.Commands[1].Payload, &second); err != nil {
		t.Fatal(err)
	}

	if first.X != 10 || second.X != 18 || first.Y != second.Y {
		t.Fatalf("direct player spawns = (%v,%v), (%v,%v)", first.X, first.Y, second.X, second.Y)
	}
}

// remoteProfileIdentity supplies a valid deterministic recipe so tickets can
// assert their binding without involving package or content loading.
func remoteProfileIdentity() simulation.RuntimeIdentity {
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
					ID:              "d2legacy",
					Version:         "1.0.0",
					Digest:          packageDigest,
					Size:            1,
					Redistributable: true,
				},
			},
			AuthoritativeHash: "rules",
			ConfigurationHash: "config",
		},
	}
}
