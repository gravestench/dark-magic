// Command bik_view validates and plays a BIK from a file, directory, ZIP, MPQ,
// or standard Diablo II MPQ directory without permanently extracting it.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gravestench/dark-magic/internal/content"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
	"github.com/gravestench/dark-magic/pkg/assetdecode"
)

var mpqPriority = []string{
	"patch_d2.mpq", "d2exp.mpq", "d2data.mpq", "d2char.mpq",
	"d2music.mpq", "d2sfx.mpq", "d2speech.mpq", "d2video.mpq",
	"d2xmusic.mpq", "d2xtalk.mpq", "d2xvideo.mpq",
}

func main() {
	fileName := flag.String("file", "", "standalone BIK file")
	sourceName := flag.String("source", "", "directory, ZIP, MPQ, or MPQ directory")
	assetName := flag.String("asset", "", "BIK path inside the source")
	playerName := flag.String("player", "ffplay", "ffplay-compatible executable")
	flag.Parse()
	if (*fileName == "") == (*sourceName == "" || *assetName == "") {
		fatal(errors.New("use either -file <movie.bik> or -source <source> -asset <path>"))
	}

	playable := *fileName
	cleanup := func() {}
	if playable == "" {
		var err error
		playable, cleanup, err = extract(*sourceName, *assetName)
		if err != nil {
			fatal(err)
		}
	}
	defer cleanup()
	expanded, err := darkpaths.ExpandHost(playable)
	if err != nil {
		fatal(err)
	}
	data, err := os.ReadFile(expanded)
	if err != nil {
		fatal(fmt.Errorf("reading BIK: %w", err))
	}
	metadata, err := assetdecode.BIK(data)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("%s: %dx%d, %d frames, %.3fs, %d audio track(s)\n",
		metadata.Version, metadata.Width, metadata.Height, metadata.Frames,
		float64(metadata.DurationMillis)/1000, len(metadata.AudioTracks))

	player, err := exec.LookPath(*playerName)
	if err != nil {
		fatal(fmt.Errorf("locating %q: %w (install FFmpeg or pass -player)", *playerName, err))
	}
	command := exec.Command(player, playerArguments(expanded)...)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := command.Run(); err != nil {
		fatal(fmt.Errorf("running player: %w", err))
	}
}

func extract(sourceName, assetName string) (string, func(), error) {
	source, err := openSource(sourceName)
	if err != nil {
		return "", func() {}, err
	}
	data, err := fs.ReadFile(source, assetName)
	if err != nil {
		return "", func() {}, fmt.Errorf("reading %q: %w", assetName, err)
	}
	file, err := os.CreateTemp("", "dark-magic-*.bik")
	if err != nil {
		return "", func() {}, fmt.Errorf("creating temporary BIK: %w", err)
	}
	name := file.Name()
	cleanup := func() { _ = os.Remove(name) }
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		cleanup()
		return "", func() {}, err
	}
	if err := file.Close(); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return name, cleanup, nil
}

func openSource(sourceName string) (fs.FS, error) {
	expanded, err := darkpaths.ExpandHost(sourceName)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(expanded)
	if err != nil || !info.IsDir() {
		return content.OpenSource(expanded)
	}
	layers := make([]content.Layer, 0, len(mpqPriority))
	for _, name := range mpqPriority {
		path := filepath.Join(expanded, name)
		if _, err := os.Stat(path); err != nil {
			continue
		}
		archive, err := content.MPQ(path)
		if err != nil {
			return nil, err
		}
		layers = append(layers, content.Layer{Name: name, FS: archive})
	}
	if len(layers) == 0 {
		return content.OpenSource(expanded)
	}
	return content.New(layers...)
}

func playerArguments(fileName string) []string {
	return []string{"-autoexit", "-window_title", "Dark Magic BIK Viewer", "-loglevel", "warning", fileName}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "bik_view:", strings.TrimSpace(err.Error()))
	os.Exit(1)
}
