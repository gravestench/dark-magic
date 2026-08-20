package sessionquic

import (
	"encoding/binary"
	"encoding/json"
	"hash/fnv"
	"math"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

const (
	transformMagic uint16 = 0x444d
	// TransformFrameVersion permits strict rejection if the compact binary layout changes.
	TransformFrameVersion uint8 = 1
	transformHeaderBytes  int   = 45
	transformEntityBytes  int   = 23
)

// TransformEntity is the compact unreliable motion state for an ID authenticated on the reliable projection.
type TransformEntity struct {
	IDHash             uint64
	X                  float64
	Y                  float64
	Direction          uint8
	Mode               [2]byte
	AnimationStartTick uint64
}

// TransformFrame carries latest-wins presentation state that can be discarded without losing gameplay truth.
type TransformFrame struct {
	Tick                    uint64
	OwnerX                  float64
	OwnerY                  float64
	VelocityX               float64
	VelocityY               float64
	OwnerDirection          uint8
	OwnerMode               [2]byte
	OwnerAnimationStartTick uint64
	Truncated               bool
	Entities                []TransformEntity
}

// encodeTransformFrame derives a credential-bound MTU-safe sample from the already authorized reliable view.
func encodeTransformFrame(
	credential gameserver.SessionCredential,
	snapshot gameserver.Snapshot,
) ([]byte, error) {
	view, err := transformClientView(snapshot)
	if err != nil {
		return nil, err
	}

	capacity := (MaxDatagramPayloadBytes - transformHeaderBytes) / transformEntityBytes
	total := len(view.World.Entities) + len(view.World.Missiles)
	count := min(total, capacity)
	data := make([]byte, transformHeaderBytes+count*transformEntityBytes)

	encodeTransformHeader(data, credential, snapshot.Tick, view, count, total)
	encodeNearestTransforms(data, view, snapshot.Tick, count)

	if len(data) > MaxDatagramPayloadBytes {
		return nil, ErrWire
	}

	return data, nil
}

// transformClientView rejects snapshots whose semantic version or tick disagrees with the reliable envelope.
func transformClientView(snapshot gameserver.Snapshot) (playeradapter.ClientView, error) {
	var view playeradapter.ClientView
	if err := json.Unmarshal(snapshot.Payload, &view); err != nil ||
		view.Version != playeradapter.ClientViewVersion ||
		view.Tick != snapshot.Tick {
		return playeradapter.ClientView{}, ErrWire
	}

	return view, nil
}

// encodeTransformHeader writes owner motion and the credential hash before any variable entity records.
func encodeTransformHeader(
	data []byte,
	credential gameserver.SessionCredential,
	tick uint64,
	view playeradapter.ClientView,
	count int,
	total int,
) {
	binary.BigEndian.PutUint16(data[0:2], transformMagic)

	data[2] = TransformFrameVersion
	if count < total || view.World.Truncated {
		data[3] = 1
	}

	binary.BigEndian.PutUint64(data[4:12], credentialHash(credential))
	binary.BigEndian.PutUint64(data[12:20], tick)
	putFloat32(data[20:24], view.HUD.Position.X)
	putFloat32(data[24:28], view.HUD.Position.Y)
	putFloat32(data[28:32], view.HUD.Movement.Velocity.X)
	putFloat32(data[32:36], view.HUD.Movement.Velocity.Y)
	binary.BigEndian.PutUint32(data[36:40], tickAge(tick, view.HUD.Animation.StartTick))
	data[40] = uint8(view.HUD.Animation.Direction)
	copy(data[41:43], []byte(view.HUD.Animation.Mode))
	binary.BigEndian.PutUint16(data[43:45], uint16(count))
}

// encodeNearestTransforms merges two nearest-first reliable collections so neither can starve the other.
func encodeNearestTransforms(
	data []byte,
	view playeradapter.ClientView,
	tick uint64,
	count int,
) {
	entityIndex, missileIndex := 0, 0

	for index := range count {
		offset := transformHeaderBytes + index*transformEntityBytes

		useMissile := missileIndex < len(view.World.Missiles) &&
			(entityIndex >= len(view.World.Entities) ||
				transformDistance2(view.World.Missiles[missileIndex].Position, view.World.Origin) <
					transformDistance2(view.World.Entities[entityIndex].Position, view.World.Origin))
		if useMissile {
			encodeTransformMissile(data[offset:offset+transformEntityBytes], view.World.Missiles[missileIndex])
			missileIndex++

			continue
		}

		encodeTransformEntity(data[offset:offset+transformEntityBytes], view.World.Entities[entityIndex], tick)
		entityIndex++
	}
}

// encodeTransformMissile omits actor-only animation fields while retaining reliable missile identity and position.
func encodeTransformMissile(data []byte, missile playeradapter.WorldMissile) {
	binary.BigEndian.PutUint64(data[0:8], PublicIDHash(missile.ID))
	putFloat32(data[8:12], missile.Position.X)
	putFloat32(data[12:16], missile.Position.Y)
	// Max age decodes to an unknown start tick; missile visual metadata remains on the reliable projection.
	binary.BigEndian.PutUint32(data[19:23], math.MaxUint32)
}

// encodeTransformEntity writes one actor record whose animation age remains bounded across long-running sessions.
func encodeTransformEntity(data []byte, entity playeradapter.WorldEntity, tick uint64) {
	binary.BigEndian.PutUint64(data[0:8], PublicIDHash(entity.ID))
	putFloat32(data[8:12], entity.Position.X)
	putFloat32(data[12:16], entity.Position.Y)
	data[16] = uint8(entity.Direction)
	copy(data[17:19], []byte(entity.Mode))
	binary.BigEndian.PutUint32(data[19:23], tickAge(tick, entity.AnimationStartTick))
}

// transformDistance2 compares nearest-first collections without the unnecessary square root of Euclidean distance.
func transformDistance2(position, origin playeradapter.HUDPosition) float64 {
	dx, dy := position.X-origin.X, position.Y-origin.Y

	return dx*dx + dy*dy
}

// decodeTransformFrame authenticates the credential hash and exact layout before exposing any presentation sample.
func decodeTransformFrame(
	credential gameserver.SessionCredential,
	data []byte,
) (TransformFrame, error) {
	if !validTransformEnvelope(credential, data) {
		return TransformFrame{}, ErrWire
	}

	count := int(binary.BigEndian.Uint16(data[43:45]))
	if len(data) != transformHeaderBytes+count*transformEntityBytes {
		return TransformFrame{}, ErrWire
	}

	frame := decodeTransformHeader(data, count)
	if frame.Tick == 0 || !finiteFrame(frame.OwnerX, frame.OwnerY, frame.VelocityX, frame.VelocityY) {
		return TransformFrame{}, ErrWire
	}

	if err := decodeTransformEntities(data, &frame); err != nil {
		return TransformFrame{}, err
	}

	return frame, nil
}

// validTransformEnvelope performs fixed-header checks before reading the variable entity count.
func validTransformEnvelope(credential gameserver.SessionCredential, data []byte) bool {
	return len(data) >= transformHeaderBytes &&
		len(data) <= MaxDatagramPayloadBytes &&
		binary.BigEndian.Uint16(data[0:2]) == transformMagic &&
		data[2] == TransformFrameVersion &&
		binary.BigEndian.Uint64(data[4:12]) == credentialHash(credential)
}

// decodeTransformHeader reconstructs owner motion and allocates exactly the validated entity count.
func decodeTransformHeader(data []byte, count int) TransformFrame {
	frame := TransformFrame{
		Tick:           binary.BigEndian.Uint64(data[12:20]),
		OwnerX:         readFloat32(data[20:24]),
		OwnerY:         readFloat32(data[24:28]),
		VelocityX:      readFloat32(data[28:32]),
		VelocityY:      readFloat32(data[32:36]),
		OwnerDirection: data[40],
		OwnerMode:      [2]byte{data[41], data[42]},
		Truncated:      data[3]&1 != 0,
		Entities:       make([]TransformEntity, count),
	}
	frame.OwnerAnimationStartTick = startTick(frame.Tick, binary.BigEndian.Uint32(data[36:40]))

	return frame
}

// decodeTransformEntities rejects zero identities and non-finite positions before presentation can consume them.
func decodeTransformEntities(data []byte, frame *TransformFrame) error {
	for index := range frame.Entities {
		offset := transformHeaderBytes + index*transformEntityBytes
		entityData := data[offset : offset+transformEntityBytes]
		frame.Entities[index] = TransformEntity{
			IDHash:             binary.BigEndian.Uint64(entityData[0:8]),
			X:                  readFloat32(entityData[8:12]),
			Y:                  readFloat32(entityData[12:16]),
			Direction:          entityData[16],
			Mode:               [2]byte{entityData[17], entityData[18]},
			AnimationStartTick: startTick(frame.Tick, binary.BigEndian.Uint32(entityData[19:23])),
		}

		if frame.Entities[index].IDHash == 0 ||
			!finiteFrame(frame.Entities[index].X, frame.Entities[index].Y) {
			return ErrWire
		}
	}

	return nil
}

// tickAge reserves MaxUint32 for unknown starts and saturates other ages without wrapping.
func tickAge(tick, start uint64) uint32 {
	if start == 0 {
		return math.MaxUint32
	}

	if start >= tick {
		return 0
	}

	return uint32(min(tick-start, uint64(math.MaxUint32-1)))
}

// startTick restores a known start when representable and maps unknown or impossible ages to zero.
func startTick(tick uint64, age uint32) uint64 {
	if age == math.MaxUint32 {
		return 0
	}

	if uint64(age) >= tick {
		return 0
	}

	return tick - uint64(age)
}

// PublicIDHash is a stable compact join key for IDs authenticated on the reliable projection stream.
func PublicIDHash(id string) uint64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(id))

	return hash.Sum64()
}

// credentialHash prevents one membership from accepting transform datagrams addressed to another.
func credentialHash(credential gameserver.SessionCredential) uint64 {
	return PublicIDHash(credential.String())
}

// putFloat32 trades precision for datagram capacity using a fixed big-endian representation.
func putFloat32(destination []byte, value float64) {
	binary.BigEndian.PutUint32(destination, math.Float32bits(float32(value)))
}

// readFloat32 widens the compact wire value for the presentation-facing API.
func readFloat32(source []byte) float64 {
	return float64(math.Float32frombits(binary.BigEndian.Uint32(source)))
}

// finiteFrame rejects NaN and infinity because they poison interpolation and spatial ordering.
func finiteFrame(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}

	return true
}
