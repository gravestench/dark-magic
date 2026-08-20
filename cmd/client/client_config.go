package main

import "log/slog"

// clientConfig is the immutable process policy produced after flag parsing.
// Keeping raw flag pointers out of startup code prevents later reads from
// observing mutable global flag state and makes the command/application handoff explicit.
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
	nativeResolution      bool
	windowSize            string
	fullscreen            bool
	nativeAudio           bool
	presentationProfileID string
	mapEditorOutput       string
	mods                  string
	logLevel              slog.Level
}
