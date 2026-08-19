package headlessshell

import (
	"context"
	"errors"
	"io"

	tea "charm.land/bubbletea/v2"
	"github.com/gravestench/dark-magic/internal/shell"
	"github.com/gravestench/dark-magic/internal/shell/tui"
)

// runTerminal distinguishes deliberate UI exit from evaluation/runtime failure,
// preventing a normal quit key from being reported as a server error.
func runTerminal(ctx context.Context, session *shell.Session, input io.Reader, output io.Writer) error {
	err := tui.Run(ctx, session, input, output)
	if errors.Is(err, tea.ErrInterrupted) || errors.Is(err, tea.ErrProgramKilled) {
		return nil
	}

	return err
}
