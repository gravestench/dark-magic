package main

import (
	"flag"
	"fmt"
	"os"
)

const audioFileUsage = "usage: audio_file_test -asset data/global/sfx/example.wav [-output file.wav] [-play]"

type commandOptions struct {
	assetPath  string
	outputPath string
	play       bool
}

// main keeps the command's phases and fatal exits visible so errors retain their established messages, codes, and
// temporary-file ownership.
func main() {
	options := parseCommandOptions()
	if options.assetPath == "" {
		exitWithAudioFileUsage()
	}

	audioData, err := readLayeredAudioAsset(options.assetPath)
	if err != nil {
		exitWithAudioFileError(err)
	}

	destination, err := resolveAudioDestination(options.outputPath, options.play)
	if err != nil {
		exitWithAudioFileError(err)
	}

	if err := writeOrDescribeAudio(os.Stdout, options.assetPath, audioData, destination); err != nil {
		exitWithAudioFileError(err)
	}

	if !options.play {
		return
	}

	// A successful command removes tool-owned output. Fatal playback errors still use os.Exit and therefore preserve
	// the command's existing behavior of bypassing deferred cleanup.
	if destination.temporary {
		defer removeTemporaryAudio(destination.path)
	}

	if err := playAudioFile(destination.path); err != nil {
		exitWithAudioFileError(err)
	}
}

// parseCommandOptions uses the process FlagSet so help text, parse failures, and exit behavior stay compatible with
// the original command-line interface.
func parseCommandOptions() commandOptions {
	assetPath := flag.String("asset", "", "WAV asset path in the layered content filesystem")
	play := flag.Bool("play", false, "play with mplayer after extraction")
	outputPath := flag.String("output", "", "optional output WAV path")

	flag.Parse()

	return commandOptions{
		assetPath:  *assetPath,
		outputPath: *outputPath,
		play:       *play,
	}
}

// exitWithAudioFileUsage distinguishes missing required input from extraction failures by retaining exit status 2.
func exitWithAudioFileUsage() {
	fmt.Fprintln(os.Stderr, audioFileUsage)
	os.Exit(2)
}

// exitWithAudioFileError gives every operational failure the command prefix and exit status callers already expect.
func exitWithAudioFileError(err error) {
	fmt.Fprintln(os.Stderr, "audio_file_test:", err)
	os.Exit(1)
}
