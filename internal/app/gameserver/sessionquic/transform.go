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
	transformMagic        uint16 = 0x444d
	TransformFrameVersion uint8  = 1
	transformHeaderBytes         = 45
	transformEntityBytes         = 23
)

type TransformEntity struct {
	IDHash             uint64
	X                  float64
	Y                  float64
	Direction          uint8
	Mode               [2]byte
	AnimationStartTick uint64
}

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

func encodeTransformFrame(credential gameserver.SessionCredential, snapshot gameserver.Snapshot) ([]byte, error) {
	var view playeradapter.ClientView
	if err := json.Unmarshal(snapshot.Payload, &view); err != nil || view.Version != playeradapter.ClientViewVersion || view.Tick != snapshot.Tick {
		return nil, ErrWire
	}
	capacity := (MaxDatagramPayloadBytes - transformHeaderBytes) / transformEntityBytes
	count := min(len(view.World.Entities), capacity)
	data := make([]byte, transformHeaderBytes+count*transformEntityBytes)
	binary.BigEndian.PutUint16(data[0:2], transformMagic)
	data[2] = TransformFrameVersion
	if count < len(view.World.Entities) || view.World.Truncated {
		data[3] = 1
	}
	binary.BigEndian.PutUint64(data[4:12], credentialHash(credential))
	binary.BigEndian.PutUint64(data[12:20], snapshot.Tick)
	putFloat32(data[20:24], view.HUD.Position.X)
	putFloat32(data[24:28], view.HUD.Position.Y)
	putFloat32(data[28:32], view.HUD.Movement.Velocity.X)
	putFloat32(data[32:36], view.HUD.Movement.Velocity.Y)
	binary.BigEndian.PutUint32(data[36:40], tickAge(snapshot.Tick, view.HUD.Animation.StartTick))
	data[40] = uint8(view.HUD.Animation.Direction)
	copy(data[41:43], []byte(view.HUD.Animation.Mode))
	binary.BigEndian.PutUint16(data[43:45], uint16(count))
	for index := 0; index < count; index++ {
		offset := transformHeaderBytes + index*transformEntityBytes
		entity := view.World.Entities[index]
		binary.BigEndian.PutUint64(data[offset:offset+8], PublicIDHash(entity.ID))
		putFloat32(data[offset+8:offset+12], entity.Position.X)
		putFloat32(data[offset+12:offset+16], entity.Position.Y)
		data[offset+16] = uint8(entity.Direction)
		copy(data[offset+17:offset+19], []byte(entity.Mode))
		binary.BigEndian.PutUint32(data[offset+19:offset+23], tickAge(snapshot.Tick, entity.AnimationStartTick))
	}
	if len(data) > MaxDatagramPayloadBytes {
		return nil, ErrWire
	}
	return data, nil
}

func decodeTransformFrame(credential gameserver.SessionCredential, data []byte) (TransformFrame, error) {
	if len(data) < transformHeaderBytes || len(data) > MaxDatagramPayloadBytes ||
		binary.BigEndian.Uint16(data[0:2]) != transformMagic || data[2] != TransformFrameVersion ||
		binary.BigEndian.Uint64(data[4:12]) != credentialHash(credential) {
		return TransformFrame{}, ErrWire
	}
	count := int(binary.BigEndian.Uint16(data[43:45]))
	if len(data) != transformHeaderBytes+count*transformEntityBytes {
		return TransformFrame{}, ErrWire
	}
	frame := TransformFrame{
		Tick: binary.BigEndian.Uint64(data[12:20]), Truncated: data[3]&1 != 0,
		OwnerX: readFloat32(data[20:24]), OwnerY: readFloat32(data[24:28]),
		VelocityX: readFloat32(data[28:32]), VelocityY: readFloat32(data[32:36]),
		OwnerDirection: data[40], OwnerMode: [2]byte{data[41], data[42]},
		Entities: make([]TransformEntity, count),
	}
	frame.OwnerAnimationStartTick = startTick(frame.Tick, binary.BigEndian.Uint32(data[36:40]))
	if frame.Tick == 0 || !finiteFrame(frame.OwnerX, frame.OwnerY, frame.VelocityX, frame.VelocityY) {
		return TransformFrame{}, ErrWire
	}
	for index := range frame.Entities {
		offset := transformHeaderBytes + index*transformEntityBytes
		frame.Entities[index] = TransformEntity{
			IDHash: binary.BigEndian.Uint64(data[offset : offset+8]),
			X:      readFloat32(data[offset+8 : offset+12]), Y: readFloat32(data[offset+12 : offset+16]),
			Direction: data[offset+16], Mode: [2]byte{data[offset+17], data[offset+18]},
			AnimationStartTick: startTick(frame.Tick, binary.BigEndian.Uint32(data[offset+19:offset+23])),
		}
		if frame.Entities[index].IDHash == 0 || !finiteFrame(frame.Entities[index].X, frame.Entities[index].Y) {
			return TransformFrame{}, ErrWire
		}
	}
	return frame, nil
}

func tickAge(tick, start uint64) uint32 {
	if start == 0 {
		return math.MaxUint32
	}
	if start >= tick {
		return 0
	}
	return uint32(min(tick-start, uint64(math.MaxUint32-1)))
}

func startTick(tick uint64, age uint32) uint64 {
	if age == math.MaxUint32 {
		return 0
	}
	if uint64(age) >= tick {
		return 0
	}
	return tick - uint64(age)
}

// PublicIDHash is the stable compact identity used only to join disposable
// transforms to IDs authenticated on the reliable projection stream.
func PublicIDHash(id string) uint64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(id))
	return hash.Sum64()
}

func credentialHash(credential gameserver.SessionCredential) uint64 {
	return PublicIDHash(credential.String())
}

func putFloat32(destination []byte, value float64) {
	binary.BigEndian.PutUint32(destination, math.Float32bits(float32(value)))
}

func readFloat32(source []byte) float64 {
	return float64(math.Float32frombits(binary.BigEndian.Uint32(source)))
}

func finiteFrame(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}
