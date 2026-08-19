package clientsession

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// decodeView accepts exactly one bounded ClientView document. Unknown or trailing data fails closed
// so protocol additions require an explicit compatibility decision rather than silent acceptance.
func decodeView(snapshot gameserver.Snapshot) (playeradapter.ClientView, error) {
	var view playeradapter.ClientView

	decoder := json.NewDecoder(bytes.NewReader(snapshot.Payload))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&view); err != nil {
		return playeradapter.ClientView{}, invalidClientViewError()
	}

	var trailing any

	trailingErr := decoder.Decode(&trailing)
	validationErr := playeradapter.ValidateClientView(view, snapshot.Tick)

	if !errors.Is(trailingErr, io.EOF) || validationErr != nil {
		return playeradapter.ClientView{}, invalidClientViewError()
	}

	return view, nil
}

// invalidClientViewError identifies the expected wire schema without returning untrusted decoder text.
func invalidClientViewError() error {
	return fmt.Errorf(
		"%w: invalid ClientView/v%d",
		ErrAssignment,
		playeradapter.ClientViewVersion,
	)
}

// parseFingerprint decodes the only supported pin format and requires a complete SHA-256 digest.
func parseFingerprint(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "sha256:") {
		return nil, ErrAssignment
	}

	decoded, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	if err != nil || len(decoded) != sha256.Size {
		return nil, ErrAssignment
	}

	return decoded, nil
}
