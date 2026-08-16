package realm

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

const workerTestToken = "worker-control-token-0123456789abcdef"

type workerHTTPFixture struct {
	description WorkerDescription
	checkpoint  gamesession.RecoveryCheckpoint
	admitted    WorkerAdmission
	admitErr    error
	removed     string
	projected   bool
}

func (worker *workerHTTPFixture) RemoveCharacter(_ context.Context, playerID string) error {
	worker.removed = playerID
	return nil
}

func (worker *workerHTTPFixture) Describe(context.Context) (WorkerDescription, error) {
	return worker.description, nil
}
func (*workerHTTPFixture) Status(context.Context) (WorkerStatus, error) {
	return WorkerStatus{Ready: true, Tick: 9, ActivePlayers: 1, ExpiredPlayers: []string{"expired-player"}}, nil
}
func (worker *workerHTTPFixture) Checkpoint(context.Context) (gamesession.RecoveryCheckpoint, error) {
	return worker.checkpoint, nil
}
func (worker *workerHTTPFixture) AdmitCharacter(_ context.Context, admission WorkerAdmission) error {
	worker.admitted = admission
	return worker.admitErr
}
func (worker *workerHTTPFixture) ProjectCharacter(_ context.Context, _ string, baseline d2save.Character) (d2save.Character, error) {
	worker.projected = true
	baseline.Name = "Canonical"
	return baseline, nil
}
func (*workerHTTPFixture) Close(context.Context) error { return nil }

type ticketHTTPFixture struct {
	principal AdmissionPrincipal
	revoked   string
}

func (tickets *ticketHTTPFixture) Issue(_ context.Context, principal AdmissionPrincipal, _ time.Duration) (string, error) {
	tickets.principal = principal
	return "issued-ticket", nil
}
func (tickets *ticketHTTPFixture) Revoke(_ context.Context, ticket string) error {
	tickets.revoked = ticket
	return nil
}

func TestWorkerHTTPClientCompletesPrivateControlFlow(t *testing.T) {
	identity := workerHTTPIdentity()
	digest, err := identity.Digest()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err := testCheckpoint(identity, 9)
	if err != nil {
		t.Fatal(err)
	}
	worker := &workerHTTPFixture{description: WorkerDescription{Runtime: identity, IdentityHash: digest}, checkpoint: checkpoint}
	tickets := &ticketHTTPFixture{}
	drained := make(chan struct{}, 1)
	handler, err := NewWorkerHTTPHandler(worker, tickets, workerTestToken, func() { drained <- struct{}{} })
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewTLSServer(handler)
	defer server.Close()
	client, err := NewWorkerHTTPClient(server.URL, workerTestToken, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	description, err := client.Describe(t.Context())
	if err != nil || description.IdentityHash != digest {
		t.Fatalf("description=%#v error=%v", description, err)
	}
	status, err := client.Status(t.Context())
	if err != nil || !status.Ready || status.Tick != 9 || status.ActivePlayers != 1 ||
		len(status.ExpiredPlayers) != 1 || status.ExpiredPlayers[0] != "expired-player" {
		t.Fatalf("status=%#v error=%v", status, err)
	}
	gotCheckpoint, err := client.Checkpoint(t.Context())
	if err != nil || gotCheckpoint.State.Tick != checkpoint.State.Tick || gotCheckpoint.Checksum != checkpoint.Checksum {
		t.Fatalf("checkpoint=%#v error=%v", gotCheckpoint, err)
	}
	admission := WorkerAdmission{Character: d2save.Character{ID: "character", Name: "Hero"}, PlayerID: "player", Actor: "realm:entry:player", Sequence: 1}
	if err := client.AdmitCharacter(t.Context(), admission); err != nil {
		t.Fatal(err)
	}
	if worker.admitted.PlayerID != "player" || worker.admitted.Character.ID != "character" {
		t.Fatalf("admission = %#v", worker.admitted)
	}
	if err := client.RemoveCharacter(t.Context(), "player"); err != nil || worker.removed != "player" {
		t.Fatalf("removed=%q error=%v", worker.removed, err)
	}
	projected, err := client.ProjectCharacter(t.Context(), "player", admission.Character)
	if err != nil || projected.Name != "Canonical" || !worker.projected {
		t.Fatalf("projected=%#v called=%v error=%v", projected, worker.projected, err)
	}
	principal := AdmissionPrincipal{AccountID: "account", CharacterID: "character", PlayerID: "player", CharacterRevision: 4, RuntimeIdentityHash: digest}
	ticket, err := client.Issue(t.Context(), principal, 10*time.Second)
	if err != nil || ticket != "issued-ticket" || tickets.principal.CharacterRevision != 4 {
		t.Fatalf("ticket=%q principal=%#v error=%v", ticket, tickets.principal, err)
	}
	if err := client.Revoke(t.Context(), ticket); err != nil || tickets.revoked != ticket {
		t.Fatalf("revoked=%q error=%v", tickets.revoked, err)
	}
	if err := client.Close(t.Context()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-drained:
	case <-time.After(time.Second):
		t.Fatal("worker drain was not requested")
	}
}

func TestWorkerMembershipsReportReconnectExpiryUntilRealmRemoval(t *testing.T) {
	memberships := NewWorkerMemberships()
	deadline := time.Now().Add(time.Minute)
	memberships.Admit("second", deadline)
	memberships.Admit("first", deadline)
	memberships.Expire("unknown")
	memberships.Expire("second")
	memberships.Expire("first")
	active, expired := memberships.Status()
	if active != 2 || len(expired) != 2 || expired[0] != "first" || expired[1] != "second" {
		t.Fatalf("active=%d expired=%#v", active, expired)
	}
	memberships.Remove("first")
	active, expired = memberships.Status()
	if active != 1 || len(expired) != 1 || expired[0] != "second" {
		t.Fatalf("after removal active=%d expired=%#v", active, expired)
	}
	memberships.Admit("second", deadline)
	memberships.Connect("second")
	active, expired = memberships.Status()
	if active != 1 || len(expired) != 0 {
		t.Fatalf("after readmission active=%d expired=%#v", active, expired)
	}
	memberships.Admit("unclaimed", deadline)
	memberships.now = func() time.Time { return deadline }
	active, expired = memberships.Status()
	if active != 2 || len(expired) != 1 || expired[0] != "unclaimed" {
		t.Fatalf("unclaimed admission active=%d expired=%#v", active, expired)
	}
}

func TestWorkerHTTPRejectsAuthenticationUnknownFieldsAndVersion(t *testing.T) {
	handler, err := NewWorkerHTTPHandler(&workerHTTPFixture{}, &ticketHTTPFixture{}, workerTestToken, func() {})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewTLSServer(handler)
	defer server.Close()
	badClient, err := NewWorkerHTTPClient(server.URL, "different-worker-control-token-123456", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := badClient.Describe(t.Context()); !errors.Is(err, ErrWorkerAuthentication) {
		t.Fatalf("authentication error = %v", err)
	}
	for _, body := range []string{
		`{"version":"RealmWorkerControl/v0"}`,
		`{"version":"RealmWorkerControl/v1","unknown":true}`,
		`{"version":"RealmWorkerControl/v1"}{}`,
	} {
		request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, server.URL+"/v1/describe", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Authorization", "Bearer "+workerTestToken)
		response, err := server.Client().Do(request)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("body %q status = %d", body, response.StatusCode)
		}
	}
}

func TestWorkerHTTPBoundsTicketLifetime(t *testing.T) {
	handler, err := NewWorkerHTTPHandler(&workerHTTPFixture{}, &ticketHTTPFixture{}, workerTestToken, func() {})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewTLSServer(handler)
	defer server.Close()
	client, err := NewWorkerHTTPClient(server.URL, workerTestToken, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Issue(t.Context(), AdmissionPrincipal{AccountID: "a"}, maximumWorkerTicketLifetime+time.Millisecond); !errors.Is(err, ErrWorkerProtocol) {
		t.Fatalf("ticket lifetime error = %v", err)
	}
}

func TestWorkerHTTPPreservesRuntimeCompatibilityFailure(t *testing.T) {
	worker := &workerHTTPFixture{admitErr: gamesession.ErrCompatibility}
	handler, err := NewWorkerHTTPHandler(worker, &ticketHTTPFixture{}, workerTestToken, func() {})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewTLSServer(handler)
	defer server.Close()
	client, err := NewWorkerHTTPClient(server.URL, workerTestToken, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if err := client.AdmitCharacter(t.Context(), WorkerAdmission{
		Character: d2save.Character{ID: "character"}, PlayerID: "player", Actor: "realm:entry:player", Sequence: 1,
	}); !errors.Is(err, gamesession.ErrCompatibility) {
		t.Fatalf("compatibility error = %v", err)
	}
}

func TestWorkerHTTPRejectsOversizedCheckpointResponse(t *testing.T) {
	worker := &workerHTTPFixture{checkpoint: gamesession.RecoveryCheckpoint{Version: gamesession.RecoveryCheckpointVersion,
		State: simulation.Checkpoint{Participants: []simulation.ParticipantState{{ID: "oversized",
			Data: make([]byte, maximumGameCheckpointBytes)}}}}}
	handler, err := NewWorkerHTTPHandler(worker, &ticketHTTPFixture{}, workerTestToken, func() {})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewTLSServer(handler)
	defer server.Close()
	client, err := NewWorkerHTTPClient(server.URL, workerTestToken, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Checkpoint(t.Context()); !errors.Is(err, ErrWorkerProtocol) {
		t.Fatalf("oversized checkpoint error = %v", err)
	}
}

func workerHTTPIdentity() simulation.RuntimeIdentity {
	return simulation.RuntimeIdentity{Recipe: simulation.RuntimeRecipe{
		Schema: simulation.RuntimeRecipeSchema, EngineAPI: "v1", NetworkProtocol: "test/v1", AssetSetID: simulation.EmptyAssetSetID,
		Packages: simulation.RuntimePackageSet{Base: simulation.RuntimePackage{ID: "d2legacy", Version: "1.0.0",
			Digest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Size: 1}},
		AuthoritativeHash: "rules", ConfigurationHash: "config",
	}}
}
