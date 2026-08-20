package main

import "testing"

// TestValidateCommandOptions locks the historical mode-selection predicate around complete and partial source inputs.
// In particular, a selected file continues to ignore one stray source flag for command-line compatibility.
func TestValidateCommandOptions(t *testing.T) {
	tests := []struct {
		name    string
		options commandOptions
		wantErr bool
	}{
		{name: "file", options: commandOptions{fileName: "movie.bik"}},
		{
			name:    "source asset",
			options: commandOptions{sourceName: "d2video.mpq", assetName: "data/local/video/intro.bik"},
		},
		{
			name:    "file with incomplete source",
			options: commandOptions{fileName: "movie.bik", sourceName: "ignored"},
		},
		{name: "neither", wantErr: true},
		{
			name:    "source without asset",
			options: commandOptions{sourceName: "d2video.mpq"},
			wantErr: true,
		},
		{
			name:    "asset without source",
			options: commandOptions{assetName: "data/local/video/intro.bik"},
			wantErr: true,
		},
		{
			name: "both complete modes",
			options: commandOptions{
				fileName:   "movie.bik",
				sourceName: "d2video.mpq",
				assetName:  "data/local/video/intro.bik",
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateCommandOptions(test.options)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateCommandOptions(%+v) error = %v, wantErr %t", test.options, err, test.wantErr)
			}

			if err != nil {
				const expected = "use either -file <movie.bik> or -source <source> -asset <path>"
				if err.Error() != expected {
					t.Fatalf("error = %q, want %q", err, expected)
				}
			}
		})
	}
}
