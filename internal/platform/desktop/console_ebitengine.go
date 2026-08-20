//go:build ebitengine

package desktop

import (
	"context"

	"github.com/gravestench/dark-magic/internal/inputstate"
)

// The shell session remains alive under Ebitengine, while native console text
// drawing is intentionally deferred until the font adapter is backend-neutral.
type headlessConsole struct{}

// NewConsole returns a no-drawing adapter so the shared shell can remain enabled in Ebitengine builds without leaking
// backend-specific font or overlay types into the desktop contract.
func NewConsole(ConsoleOptions) Console { return headlessConsole{} }

// LoadFont succeeds because the headless adapter owns no font resources; startup therefore follows the Console
// lifecycle even though native text rendering is unavailable.
func (headlessConsole) LoadFont() error { return nil }

// Handle never captures input, leaving every frame available to the game while Ebitengine lacks a console overlay.
func (headlessConsole) Handle(context.Context, inputstate.Frame) bool { return false }

// Draw intentionally emits nothing because Ebitengine has no backend-neutral console font adapter yet.
func (headlessConsole) Draw(int, int) {}

// Close is a no-op because the headless adapter never acquires native resources.
func (headlessConsole) Close() {}

var _ Console = headlessConsole{}
