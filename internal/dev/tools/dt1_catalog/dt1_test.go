package main

import (
	"bytes"
	"testing"

	"github.com/gravestench/dt1"
)

// TestWriteDT1Header distinguishes safely decodable modern files from layouts that must remain probe-only.
func TestWriteDT1Header(t *testing.T) {
	tests := []struct {
		name       string
		header     dt1.Header
		wantOutput string
		wantDecode bool
	}{
		{
			name:       "modern",
			header:     dt1.Header{Version: 7, SubVersion: 6, Layout: dt1.LayoutModern},
			wantOutput: "foo.dt1 (header 7.6, modern-7.6)",
			wantDecode: true,
		},
		{
			name:       "unknown",
			header:     dt1.Header{Version: 1, SubVersion: 0, Layout: dt1.Layout("unknown")},
			wantOutput: "foo.dt1 (header 1.0, unknown) -- body intentionally not decoded\n",
			wantDecode: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer

			gotDecode := writeDT1Header(&output, "foo.dt1", test.header)

			if gotDecode != test.wantDecode {
				t.Fatalf("writeDT1Header() decode = %t, want %t", gotDecode, test.wantDecode)
			}

			if got := output.String(); got != test.wantOutput {
				t.Fatalf("writeDT1Header() output = %q, want %q", got, test.wantOutput)
			}
		})
	}
}
