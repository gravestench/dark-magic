package input

// these are static declarations that force a
// compile-time error if the service does not
// implement them.
var _ GetsInputState = &Service{}

// this is an alias which can be used to make
// the dependency resolution methods of other
// services more coherent. It's just sugar.

type Dependency = GetsInputState

// Here is the declaration of our service as
// an interface. This is all the dependent services
// should know about this service.

type GetsInputState interface {
	KeyState(key int32) InputState
	KeyboardState() map[int32]InputState
	KeyboardModifierState() map[int32]InputState
	MouseCursorState() (x, y int)
	MouseWheelState() (x, y float32)
	MouseButtonState() map[int32]InputState
}
