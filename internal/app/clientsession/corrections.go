package clientsession

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/networkclock"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// Refresh fetches and installs one complete reliable correction under a stable credential.
func (session *Session) Refresh(ctx context.Context) (playeradapter.WorldDelta, error) {
	return session.refresh(ctx)
}

// Watch combines reliable lifecycle corrections and lossy transform frames. Its one-element output
// buffer intentionally drops obsolete notifications while retaining the newest installed revision.
func (session *Session) Watch(
	ctx context.Context,
) (<-chan playeradapter.WorldDelta, <-chan error, error) {
	return session.watch(ctx)
}

// refresh fetches one reliable correction under a stable credential, then installs it only while the
// session remains open. reconnectMu prevents the transport from rotating during the request.
func (session *Session) refresh(ctx context.Context) (playeradapter.WorldDelta, error) {
	session.reconnectMu.RLock()
	defer session.reconnectMu.RUnlock()

	transport, credential, err := session.correctionTarget()
	if err != nil {
		return playeradapter.WorldDelta{}, err
	}

	snapshot, err := transport.Refresh(ctx, credential)
	if err != nil {
		return playeradapter.WorldDelta{}, err
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	if session.closed {
		return playeradapter.WorldDelta{}, errors.New("client session: closed")
	}

	return session.applyCorrection(snapshot)
}

// correctionTarget snapshots the transport and credential used by reliable correction requests.
func (session *Session) correctionTarget() (
	*sessionquic.Client,
	gameserver.SessionCredential,
	error,
) {
	session.mu.Lock()
	defer session.mu.Unlock()

	if session.closed {
		return nil, "", errors.New("client session: closed")
	}

	return session.transport, session.credential, nil
}

// correctionStreams groups the reliable lifecycle stream and lossy transform stream opened against
// one credential. Nil channels are retired independently as transports close them.
type correctionStreams struct {
	snapshots       <-chan gameserver.Snapshot
	reliableErrors  <-chan error
	transforms      <-chan sessionquic.TransformFrame
	transformErrors <-chan error
}

// watch opens both correction channels before releasing credential stability, then consumes them in
// one worker so reliable lifecycle and transform revisions remain ordered under session.mu.
func (session *Session) watch(
	ctx context.Context,
) (<-chan playeradapter.WorldDelta, <-chan error, error) {
	watchContext, cancelWatch := context.WithCancel(ctx)

	streams, err := session.openCorrectionStreams(watchContext)
	if err != nil {
		cancelWatch()

		return nil, nil, err
	}

	deltas := make(chan playeradapter.WorldDelta, 1)
	errorsOut := make(chan error, 1)

	go session.consumeCorrectionStreams(ctx, cancelWatch, streams, deltas, errorsOut)

	return deltas, errorsOut, nil
}

// openCorrectionStreams ensures credential rotation cannot split reliable and transform watches across
// two transports.
func (session *Session) openCorrectionStreams(ctx context.Context) (correctionStreams, error) {
	session.reconnectMu.RLock()
	defer session.reconnectMu.RUnlock()

	transport, credential, err := session.correctionTarget()
	if err != nil {
		return correctionStreams{}, err
	}

	snapshots, reliableErrors, err := transport.Watch(ctx, credential)
	if err != nil {
		return correctionStreams{}, err
	}

	transforms, transformErrors, err := transport.WatchTransforms(ctx, credential)
	if err != nil {
		return correctionStreams{}, err
	}

	return correctionStreams{
		snapshots:       snapshots,
		reliableErrors:  reliableErrors,
		transforms:      transforms,
		transformErrors: transformErrors,
	}, nil
}

// consumeCorrectionStreams publishes only the newest pending delta while delivering the first stream
// failure reliably. Closing one channel does not discard updates still arriving on the others.
func (session *Session) consumeCorrectionStreams(
	ctx context.Context,
	cancel context.CancelFunc,
	streams correctionStreams,
	deltas chan playeradapter.WorldDelta,
	errorsOut chan error,
) {
	defer cancel()
	defer close(deltas)
	defer close(errorsOut)

	for streams.active() {
		select {
		case snapshot, open := <-streams.snapshots:
			if !open {
				streams.snapshots = nil

				continue
			}

			delta, err := session.applyWatchedSnapshot(snapshot)
			if err != nil {
				errorsOut <- err

				return
			}

			publishLatestDelta(ctx, deltas, delta)
		case streamErr, open := <-streams.reliableErrors:
			if !open {
				streams.reliableErrors = nil

				continue
			}

			if streamErr != nil {
				errorsOut <- streamErr

				return
			}
		case frame, open := <-streams.transforms:
			if !open {
				streams.transforms = nil

				continue
			}

			delta, err := session.applyWatchedTransform(frame)
			if err != nil {
				errorsOut <- err

				return
			}

			publishLatestDelta(ctx, deltas, delta)
		case streamErr, open := <-streams.transformErrors:
			if !open {
				streams.transformErrors = nil

				continue
			}

			if streamErr != nil {
				errorsOut <- streamErr

				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// active reports whether any correction channel can still produce a value or terminal error.
func (streams correctionStreams) active() bool {
	return streams.snapshots != nil || streams.reliableErrors != nil ||
		streams.transforms != nil || streams.transformErrors != nil
}

// applyWatchedSnapshot decodes outside the session lock, then installs the complete reliable view.
func (session *Session) applyWatchedSnapshot(
	snapshot gameserver.Snapshot,
) (playeradapter.WorldDelta, error) {
	view, err := decodeView(snapshot)
	if err != nil {
		return playeradapter.WorldDelta{}, err
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	return session.applyDecodedCorrection(snapshot, view)
}

// applyWatchedTransform serializes lossy transform installation with reliable correction commits.
func (session *Session) applyWatchedTransform(
	frame sessionquic.TransformFrame,
) (playeradapter.WorldDelta, error) {
	session.mu.Lock()
	defer session.mu.Unlock()

	return session.applyTransform(frame)
}

// applyCorrection decodes a reliable snapshot before applying it; callers hold session.mu.
func (session *Session) applyCorrection(
	snapshot gameserver.Snapshot,
) (playeradapter.WorldDelta, error) {
	view, err := decodeView(snapshot)
	if err != nil {
		return playeradapter.WorldDelta{}, err
	}

	return session.applyDecodedCorrection(snapshot, view)
}

// applyDecodedCorrection validates monotonicity and owner identity before replacing canonical state.
func (session *Session) applyDecodedCorrection(
	snapshot gameserver.Snapshot,
	view playeradapter.ClientView,
) (playeradapter.WorldDelta, error) {
	if err := validateCorrection(session.Admission.Snapshot, snapshot); err != nil {
		return playeradapter.WorldDelta{}, err
	}

	if err := validateOwnerIdentity(session.reliableHUD.Player, view.HUD.Player); err != nil {
		return playeradapter.WorldDelta{}, err
	}

	previousReliable := session.reliableWorld
	if previousReliable.Version == 0 {
		previousReliable = session.World
	}

	delta := playeradapter.DiffWorldView(previousReliable, view.World)
	session.Admission.Snapshot = snapshot
	session.Private = view.Private
	session.Party = view.Party
	session.Events = view.Events
	session.reliableHUD = view.HUD
	session.reliableWorld = view.World
	session.HUD, session.World = mergeReliablePresentation(view.HUD, view.World, session.HUD, session.World)
	session.viewRevision++
	session.observeSnapshotLocked(snapshot, time.Now())
	session.discardAcknowledgedLocked(snapshot.AcknowledgedInput)
	session.publishPresentationLocked()

	return delta, nil
}

// validateOwnerIdentity prevents correction and recovery data from changing the admitted player.
func validateOwnerIdentity(expected, actual playeradapter.HUDIdentity) error {
	if expected.PlayerID == "" || expected.CharacterID == "" ||
		actual.PlayerID != expected.PlayerID || actual.CharacterID != expected.CharacterID {
		return ErrAssignment
	}

	return nil
}

// applyTransform updates only already-known entity transforms. Reliable snapshots exclusively own
// lifecycle; unknown hashes wait, and hash collisions fail instead of aliasing identities.
func (session *Session) applyTransform(
	frame sessionquic.TransformFrame,
) (playeradapter.WorldDelta, error) {
	if frame.Tick <= session.World.Tick {
		return unchangedTransformDelta(session.World.Tick), nil
	}

	before := session.World
	world := cloneTransformWorld(session.World)

	entities, missiles, err := transformIndexes(world)
	if err != nil {
		return playeradapter.WorldDelta{}, err
	}

	applyFrameTransforms(&world, frame, entities, missiles)
	hud := transformedOwnerHUD(session.HUD, frame, world.Origin)
	session.HUD, session.World = hud, world
	session.viewRevision++
	session.publishPresentationLocked()
	session.observeTransformLocked(frame)

	return playeradapter.DiffWorldView(before, world), nil
}

// unchangedTransformDelta acknowledges an obsolete datagram without creating lifecycle changes.
func unchangedTransformDelta(tick uint64) playeradapter.WorldDelta {
	return playeradapter.WorldDelta{
		Version:  playeradapter.WorldDeltaVersion,
		BaseTick: tick,
		Tick:     tick,
	}
}

// cloneTransformWorld detaches the two collections a transform frame may mutate.
func cloneTransformWorld(world playeradapter.WorldView) playeradapter.WorldView {
	world.Entities = append([]playeradapter.WorldEntity(nil), world.Entities...)
	world.Missiles = append([]playeradapter.WorldMissile(nil), world.Missiles...)

	return world
}

// transformIndexes builds collision-checked public-ID hash lookups for entities and missiles.
func transformIndexes(
	world playeradapter.WorldView,
) (map[uint64]int, map[uint64]int, error) {
	entities := make(map[uint64]int, len(world.Entities))
	for index, entity := range world.Entities {
		hash := sessionquic.PublicIDHash(entity.ID)
		if _, duplicate := entities[hash]; duplicate {
			return nil, nil, ErrStaleCorrection
		}

		entities[hash] = index
	}

	missiles := make(map[uint64]int, len(world.Missiles))
	for index, missile := range world.Missiles {
		hash := sessionquic.PublicIDHash(missile.ID)
		if _, duplicate := entities[hash]; duplicate {
			return nil, nil, ErrStaleCorrection
		}

		if _, duplicate := missiles[hash]; duplicate {
			return nil, nil, ErrStaleCorrection
		}

		missiles[hash] = index
	}

	return entities, missiles, nil
}

// applyFrameTransforms updates recognized public identities and ignores unknown lifecycle safely.
func applyFrameTransforms(
	world *playeradapter.WorldView,
	frame sessionquic.TransformFrame,
	entities map[uint64]int,
	missiles map[uint64]int,
) {
	for _, transform := range frame.Entities {
		if index, found := entities[transform.IDHash]; found {
			world.Entities[index].Position = playeradapter.HUDPosition{X: transform.X, Y: transform.Y}
			world.Entities[index].Direction = int64(transform.Direction)
			world.Entities[index].Mode = strings.TrimRight(string(transform.Mode[:]), "\x00")
			world.Entities[index].AnimationStartTick = transform.AnimationStartTick

			continue
		}

		if index, found := missiles[transform.IDHash]; found {
			world.Missiles[index].Position = playeradapter.HUDPosition{X: transform.X, Y: transform.Y}
		}
	}

	world.Tick = frame.Tick
	world.Origin = playeradapter.HUDPosition{X: frame.OwnerX, Y: frame.OwnerY}
}

// transformedOwnerHUD applies the owner fields carried by the low-latency datagram.
func transformedOwnerHUD(
	hud playeradapter.HUD,
	frame sessionquic.TransformFrame,
	origin playeradapter.HUDPosition,
) playeradapter.HUD {
	hud.Tick = frame.Tick
	hud.Position = origin
	hud.Movement.Velocity = playeradapter.HUDPosition{X: frame.VelocityX, Y: frame.VelocityY}
	hud.Animation.Direction = int64(frame.OwnerDirection)
	hud.Animation.Mode = strings.TrimRight(string(frame.OwnerMode[:]), "\x00")
	hud.Animation.StartTick = frame.OwnerAnimationStartTick

	return hud
}

// observeTransformLocked updates the clock from datagram timing and transport latency.
func (session *Session) observeTransformLocked(frame sessionquic.TransformFrame) {
	if session.clock == nil {
		session.clock = networkclock.New(networkclock.Config{UpdateInterval: sessionquic.TransformInterval})
	}

	stats := sessionquic.NetworkStats{}
	if session.transport != nil {
		stats = session.transport.NetworkStats()
	}

	session.clock.Observe(networkclock.Sample{
		Tick:         frame.Tick,
		Step:         time.Duration(session.Admission.Snapshot.StepNanos),
		ReceivedAt:   time.Now(),
		SmoothedRTT:  stats.SmoothedRTT,
		RTTVariation: stats.RTTVariation,
	})
}

// mergeReliablePresentation installs reliable structure while preserving newer datagram transforms.
func mergeReliablePresentation(
	reliableHUD playeradapter.HUD,
	reliableWorld playeradapter.WorldView,
	currentHUD playeradapter.HUD,
	currentWorld playeradapter.WorldView,
) (playeradapter.HUD, playeradapter.WorldView) {
	if currentWorld.Tick <= reliableWorld.Tick {
		return reliableHUD, reliableWorld
	}

	mergeEntityTransforms(&reliableWorld, currentWorld)
	mergeMissileTransforms(&reliableWorld, currentWorld)
	reliableWorld.Tick = currentWorld.Tick
	reliableWorld.Origin = currentWorld.Origin
	reliableHUD.Tick = currentHUD.Tick
	reliableHUD.Position = currentHUD.Position
	reliableHUD.Movement.Velocity = currentHUD.Movement.Velocity
	reliableHUD.Animation = currentHUD.Animation

	return reliableHUD, reliableWorld
}

// mergeEntityTransforms copies only transform fields for identities retained by reliable lifecycle.
func mergeEntityTransforms(reliable *playeradapter.WorldView, current playeradapter.WorldView) {
	transforms := make(map[string]playeradapter.WorldEntity, len(current.Entities))
	for _, entity := range current.Entities {
		transforms[entity.ID] = entity
	}

	reliable.Entities = append([]playeradapter.WorldEntity(nil), reliable.Entities...)
	for index := range reliable.Entities {
		if currentEntity, found := transforms[reliable.Entities[index].ID]; found {
			reliable.Entities[index].Position = currentEntity.Position
			reliable.Entities[index].Direction = currentEntity.Direction
			reliable.Entities[index].Mode = currentEntity.Mode
			reliable.Entities[index].AnimationStartTick = currentEntity.AnimationStartTick
		}
	}
}

// mergeMissileTransforms preserves newer positions only for missiles retained by reliable lifecycle.
func mergeMissileTransforms(reliable *playeradapter.WorldView, current playeradapter.WorldView) {
	positions := make(map[string]playeradapter.HUDPosition, len(current.Missiles))
	for _, missile := range current.Missiles {
		positions[missile.ID] = missile.Position
	}

	reliable.Missiles = append([]playeradapter.WorldMissile(nil), reliable.Missiles...)
	for index := range reliable.Missiles {
		if position, found := positions[reliable.Missiles[index].ID]; found {
			reliable.Missiles[index].Position = position
		}
	}
}

// publishLatestDelta replaces an obsolete unread notification because session state is already newer.
func publishLatestDelta(
	ctx context.Context,
	deltas chan playeradapter.WorldDelta,
	delta playeradapter.WorldDelta,
) {
	select {
	case deltas <- delta:
		return
	default:
	}

	select {
	case <-deltas:
	default:
	}

	select {
	case deltas <- delta:
	case <-ctx.Done():
	}
}

// validateCorrection permits idempotent same-tick replay but rejects regression or conflicting history.
func validateCorrection(previous, next gameserver.Snapshot) error {
	if next.Tick < previous.Tick ||
		next.Tick == previous.Tick && next.Checksum != previous.Checksum {
		return ErrStaleCorrection
	}

	return nil
}
