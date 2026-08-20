package clientsession

import (
	"context"
	"crypto/tls"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/networkclock"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

var ErrAssignment = errors.New("client session: invalid server assignment")
var ErrStaleCorrection = errors.New("client session: stale or conflicting correction")

// Session owns one authenticated transport and the latest canonical projections received through it.
// reconnectMu prevents credential rotation while requests use the old credential; mu protects the
// mutable state consumed concurrently by transport workers and presentation.
type Session struct {
	mu             sync.Mutex
	reconnectMu    sync.RWMutex
	transport      *sessionquic.Client
	credential     gameserver.SessionCredential
	identity       simulation.RuntimeIdentity
	closed         bool
	Admission      gameserver.JoinResponse
	HUD            playeradapter.HUD
	World          playeradapter.WorldView
	Private        playeradapter.PrivateView
	Party          playeradapter.PartyView
	Events         playeradapter.EventView
	reliableHUD    playeradapter.HUD
	reliableWorld  playeradapter.WorldView
	viewRevision   uint64
	eventEpoch     uint64
	presentation   atomic.Pointer[PresentationSnapshot]
	pending        map[uint64]gameserver.CommandIntent
	clock          *networkclock.Clock
	gameID         string
	endpoint       realm.GameEndpoint
	tlsConfig      *tls.Config
	reconnectNonce string
}

// PresentationSnapshot is immutable after publication. The network worker builds and atomically
// swaps it, allowing the render thread to retain a revision without copying the world every frame.
type PresentationSnapshot struct {
	HUD        playeradapter.HUD
	World      playeradapter.WorldView
	Private    playeradapter.PrivateView
	Party      playeradapter.PartyView
	Events     playeradapter.EventView
	Revision   uint64
	EventEpoch uint64
}

// SelfHostedAssignment contains the public runtime identity plus the private credential needed to
// admit one selected local profile. It is never a substitute for the one-use gameplay ticket.
type SelfHostedAssignment struct {
	GameID            string
	Endpoint          realm.GameEndpoint
	Runtime           simulation.RuntimeIdentity
	ProfileCredential string
}

// View returns a coherent defensive public projection while correction workers may install newer
// state. Callers must not read the exported initial-view fields concurrently with Watch.
func (session *Session) View() (playeradapter.HUD, playeradapter.WorldView) {
	session.mu.Lock()
	defer session.mu.Unlock()

	return clonePresentationView(session.HUD, session.World)
}

// PresentationView adds the monotonic content revision used to distinguish lifecycle changes from
// transform-only updates at the same simulation tick.
func (session *Session) PresentationView() (playeradapter.HUD, playeradapter.WorldView, uint64) {
	session.mu.Lock()
	defer session.mu.Unlock()

	hud, world := clonePresentationView(session.HUD, session.World)

	return hud, world, session.viewRevision
}

// PresentationSnapshot returns the lock-free immutable projection published for rendering. The slow
// path initializes publication for sessions constructed before their first correction.
func (session *Session) PresentationSnapshot() *PresentationSnapshot {
	if session == nil {
		return nil
	}

	if snapshot := session.presentation.Load(); snapshot != nil {
		return snapshot
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	session.publishPresentationLocked()

	return session.presentation.Load()
}

// publishPresentationLocked atomically exposes one coherent revision; callers must hold session.mu
// so public, private, party, and event projections cannot come from different corrections.
func (session *Session) publishPresentationLocked() {
	session.presentation.Store(&PresentationSnapshot{
		HUD:        session.HUD,
		World:      session.World,
		Private:    session.Private,
		Party:      clonePartyView(session.Party),
		Events:     cloneEventView(session.Events),
		Revision:   session.viewRevision,
		EventEpoch: session.eventEpoch,
	})
}

// clonePartyView detaches the variable roster from the correction decoder's backing storage.
func clonePartyView(view playeradapter.PartyView) playeradapter.PartyView {
	view.Roster = append([]playeradapter.PartyRosterEntry(nil), view.Roster...)

	return view
}

// cloneEventView detaches event slices and optional payloads so retained snapshots remain immutable.
func cloneEventView(view playeradapter.EventView) playeradapter.EventView {
	view.Events = append([]playeradapter.SemanticEvent(nil), view.Events...)

	for index := range view.Events {
		if view.Events[index].Cast != nil {
			value := *view.Events[index].Cast
			view.Events[index].Cast = &value
		}

		if view.Events[index].State != nil {
			value := *view.Events[index].State
			view.Events[index].State = &value
		}
	}

	return view
}

// clonePresentationView recursively detaches every mutable public slice and optional health value.
func clonePresentationView(
	hud playeradapter.HUD,
	world playeradapter.WorldView,
) (playeradapter.HUD, playeradapter.WorldView) {
	hud.Belt.Slots = append([]string(nil), hud.Belt.Slots...)
	hud.Skills.Learned = append([]playeradapter.HUDLearnedSkill(nil), hud.Skills.Learned...)
	world.Entities = append([]playeradapter.WorldEntity(nil), world.Entities...)
	world.Missiles = append([]playeradapter.WorldMissile(nil), world.Missiles...)
	world.States = append([]playeradapter.WorldState(nil), world.States...)

	for index := range world.Entities {
		if health := world.Entities[index].Health; health != nil {
			value := *health
			world.Entities[index].Health = &value
		}

		if maximum := world.Entities[index].MaxHealth; maximum != nil {
			value := *maximum
			world.Entities[index].MaxHealth = &value
		}
	}

	return hud, world
}

// PrivateView returns an owner-only defensive projection through a separate API, preventing public
// world consumers from accidentally depending on inventory or interaction state.
func (session *Session) PrivateView() playeradapter.PrivateView {
	session.mu.Lock()
	defer session.mu.Unlock()

	view := session.Private
	view.Items.Items = append([]playeradapter.ItemEntityView(nil), view.Items.Items...)

	if view.Interaction.Target != nil {
		target := *view.Interaction.Target
		view.Interaction.Target = &target
	}

	return view
}

// NextInputTick remains inside authority's two-tick admission lead even when presentation extrapolates
// farther during a stall. Late admissible input can use rollback; future input would be rejected.
func (session *Session) NextInputTick(now time.Time) uint64 {
	session.mu.Lock()
	defer session.mu.Unlock()

	timeline := session.timelineLocked(now)

	return timeline.LatestServerTick + 2
}

// NetworkTimeline exposes prediction and interpolation moments without exposing the mutable clock.
func (session *Session) NetworkTimeline(now time.Time) networkclock.Timeline {
	session.mu.Lock()
	defer session.mu.Unlock()

	return session.timelineLocked(now)
}

// StepDuration returns the authority-advertised fixed simulation step for animation and prediction.
func (session *Session) StepDuration() time.Duration {
	session.mu.Lock()
	defer session.mu.Unlock()

	return time.Duration(session.Admission.Snapshot.StepNanos)
}

// timelineLocked initializes the clock from admission before returning a derived timeline.
func (session *Session) timelineLocked(now time.Time) networkclock.Timeline {
	if session.clock == nil {
		session.observeSnapshotLocked(session.Admission.Snapshot, now)
	}

	return session.clock.Timeline(now)
}

// observeSnapshotLocked combines authority time with current transport latency. Callers hold mu so
// reliable corrections and transform datagrams update one ordered clock.
func (session *Session) observeSnapshotLocked(snapshot gameserver.Snapshot, receivedAt time.Time) {
	if session.clock == nil {
		session.clock = networkclock.New(networkclock.Config{UpdateInterval: sessionquic.TransformInterval})
	}

	stats := sessionquic.NetworkStats{}
	if session.transport != nil {
		stats = session.transport.NetworkStats()
	}

	session.clock.Observe(networkclock.Sample{
		Tick:         snapshot.Tick,
		Step:         time.Duration(snapshot.StepNanos),
		ReceivedAt:   receivedAt,
		SmoothedRTT:  stats.SmoothedRTT,
		RTTVariation: stats.RTTVariation,
	})
}

// Close marks the session closed before attempting a bounded graceful leave, then always releases the
// transport. Repeated calls are harmless and cannot send a second leave.
func (session *Session) Close(ctx context.Context) error {
	if session == nil || session.transport == nil {
		return nil
	}

	session.mu.Lock()
	if session.closed {
		session.mu.Unlock()

		return nil
	}

	session.closed = true
	transport, credential := session.transport, session.credential
	session.mu.Unlock()

	var leaveErr error

	if ctx.Err() == nil {
		leaveContext, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
		leaveErr = transport.Leave(leaveContext, credential)

		cancel()
	}

	return errors.Join(leaveErr, transport.Close())
}
