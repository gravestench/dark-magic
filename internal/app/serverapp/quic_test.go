package serverapp

import (
	"testing"
)

// TestStartQUICRejectsPartialConfiguration ensures an optional transport cannot
// silently start with only part of its TLS and admission boundary configured.
func TestStartQUICRejectsPartialConfiguration(t *testing.T) {
	if _, err := StartQUIC(QUICConfig{Address: "127.0.0.1:0", SessionID: "session"}, nil); err == nil {
		t.Fatal("partial QUIC configuration was accepted")
	}
}
