package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gravestench/dark-magic/internal/content"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

func main() {
	asset := flag.String("asset", "", "WAV asset path in the layered content filesystem")
	play := flag.Bool("play", false, "play with mplayer after extraction")
	output := flag.String("output", "", "optional output WAV path")
	flag.Parse()
	if *asset == "" {
		fmt.Fprintln(os.Stderr, "usage: audio_file_test -asset data/global/sfx/example.wav [-output file.wav] [-play]")
		os.Exit(2)
	}
	contentFS, err := content.FromEnvironment()
	if err != nil {
		fatal(err)
	}
	data, err := fs.ReadFile(contentFS, *asset)
	if err != nil {
		fatal(err)
	}
	fileName := *output
	if fileName != "" {
		fileName, err = darkpaths.ExpandHost(fileName)
		if err != nil {
			fatal(err)
		}
	}
	temporary := false
	if fileName == "" && *play {
		file, err := os.CreateTemp("", "dark-magic-*.wav")
		if err != nil {
			fatal(err)
		}
		fileName = file.Name()
		_ = file.Close()
		temporary = true
	}
	if fileName != "" {
		if err := os.WriteFile(fileName, data, 0o644); err != nil {
			fatal(err)
		}
		fmt.Printf("wrote %d bytes to %s\n", len(data), fileName)
	} else {
		fmt.Printf("%s: %d bytes\n", *asset, len(data))
	}
	if *play {
		if temporary {
			defer os.Remove(fileName)
		}
		if _, err := exec.LookPath("mplayer"); err != nil {
			fatal(fmt.Errorf("mplayer is required for -play: %w", err))
		}
		command := exec.Command("mplayer", filepath.Clean(fileName))
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			fatal(err)
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "audio_file_test:", err)
	os.Exit(1)
}
