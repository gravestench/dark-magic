package raylibRenderer

import "github.com/gravestench/dark-magic/internal/presentation/render"

// these are static declarations that force a
// compile-time error if the service does not
// implement them.
var _ IsRenderer = &Service{}

// this is an alias which can be used to make
// the dependency resolution methods of other
// services more coherent. It's just sugar.

type Dependency = IsRenderer

// Here is the declaration of our service as
// an interface. This is all the dependent services
// should know about this service.

type IsRenderer interface {
	OnFrame(func())
	SubscribeFrame(func()) func()
	ManagesWindow
	AttachComposer(*render.Composer) error
}

type ManagesWindow interface {
	SetWindowTitle(string)
	WindowSize() (width, height int)
	Resolution() (width, height int)
	ScreenToGame(x, y int) (gameX, gameY int, inside bool)
}
