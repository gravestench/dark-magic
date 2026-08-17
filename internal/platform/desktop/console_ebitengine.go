//go:build ebitengine

package desktop

import (
	"context"

	"github.com/gravestench/dark-magic/internal/inputstate"
)

// The shell session remains alive under Ebitengine, while native console text
// drawing is intentionally deferred until the font adapter is backend-neutral.
type headlessConsole struct{}

func NewConsole(ConsoleOptions) Console                               { return headlessConsole{} }
func (headlessConsole) LoadFont() error                               { return nil }
func (headlessConsole) Handle(context.Context, inputstate.Frame) bool { return false }
func (headlessConsole) Draw(int, int)                                 {}
func (headlessConsole) Close()                                        {}

var _ Console = headlessConsole{}
