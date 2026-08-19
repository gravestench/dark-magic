package assetdecode

import "testing"

// TestTextDecodesUTF8AndUTF16 covers the supported encodings and proves an
// incomplete UTF-16 code unit fails instead of being silently discarded.
func TestTextDecodesUTF8AndUTF16(t *testing.T) {
	if got, err := Text([]byte("hello")); err != nil || got != "hello" {
		t.Fatalf("UTF-8 = %q, %v", got, err)
	}

	got, err := Text([]byte{0xff, 0xfe, 'h', 0, 'i', 0, 0x3d, 0xd8, 0x00, 0xde})
	if err != nil || got != "hi😀" {
		t.Fatalf("UTF-16 = %q, %v", got, err)
	}

	if _, err := Text([]byte{0xff, 0xfe, 1}); err == nil {
		t.Fatal("expected malformed UTF-16 error")
	}
}
