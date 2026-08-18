package main

import "log/slog"

// clientConfig collects the command-line policy consumed while assembling the client.
// Keeping it as one value makes the handoff between parsing and startup explicit.
type clientConfig struct {
	profileDirectory      string
	profileScenes         string
	captureDirectory      string
	captureScenes         string
	captureSettle         int
	startScene            string
	startOverlays         string
	fixtureCharacters     int
	fixtureWorldLevel     int
	fixtureWorldSpawn     string
	fixturePointerMove    bool
	outputPalette         string
	viewportFit           string
	fullscreen            bool
	nativeAudio           bool
	presentationProfileID string
	mods                  string
	logLevel              slog.Level
}
