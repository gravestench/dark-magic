package serverapp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadAdmissionKeyIsProtectedAndBounded(t *testing.T) {
	directory := t.TempDir()
	valid := filepath.Join(directory, "valid.key")
	if err := os.WriteFile(valid, []byte("0123456789abcdef0123456789abcdef\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	key, err := ReadAdmissionKey(valid)
	if err != nil {
		t.Fatal(err)
	}
	if string(key) != "0123456789abcdef0123456789abcdef" {
		t.Fatalf("key = %q", key)
	}
	for name, data := range map[string]struct {
		data []byte
		mode os.FileMode
	}{
		"short": {[]byte("short"), 0o600}, "large": {make([]byte, 4097), 0o600},
		"exposed": {[]byte("0123456789abcdef0123456789abcdef"), 0o644},
	} {
		path := filepath.Join(directory, name+".key")
		if err := os.WriteFile(path, data.data, data.mode); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadAdmissionKey(path); err == nil {
			t.Fatalf("%s key was accepted", name)
		}
	}
}

func TestStartQUICRejectsPartialConfiguration(t *testing.T) {
	if _, err := StartQUIC(QUICConfig{Address: "127.0.0.1:0", SessionID: "session"}, nil); err == nil {
		t.Fatal("partial QUIC configuration was accepted")
	}
}
