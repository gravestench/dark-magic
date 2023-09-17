package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gravestench/runtime"
	"github.com/hajimehoshi/oto"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/models"
	"github.com/gravestench/dark-magic/pkg/services/loaders/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/tsvLoader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/wavLoader"
	"github.com/gravestench/dark-magic/pkg/services/record_manager"
)

type audioFileTestService struct {
	logger *zerolog.Logger

	mpq mpqLoader.Dependency
	tsv tsvLoader.Dependency
	wav wavLoader.Dependency

	ctx *oto.Context

	Sounds []models.SoundEntry
}

func (s *audioFileTestService) Name() string {
	return "music loader test"
}

func (s *audioFileTestService) DependenciesResolved() bool {
	if s.tsv == nil {
		return false
	}

	if s.wav == nil {
		return false
	}

	if s.mpq == nil {
		return false
	}

	if !s.mpq.(runtime.HasDependencies).DependenciesResolved() {
		return false
	}

	if !s.wav.(runtime.HasDependencies).DependenciesResolved() {
		return false
	}

	if !s.tsv.(runtime.HasDependencies).DependenciesResolved() {
		return false
	}

	return true
}

func (s *audioFileTestService) ResolveDependencies(rt runtime.R) {
	for _, service := range rt.Services() {
		switch candidate := service.(type) {
		case tsvLoader.Dependency:
			s.tsv = candidate
		case wavLoader.Dependency:
			s.wav = candidate
		case mpqLoader.Dependency:
			s.mpq = candidate
		}
	}
}

func (s *audioFileTestService) Init(rt runtime.R) {
	// set a random seed
	rand.Seed(time.Now().UnixNano())

	// load the tsv file into our array of record models
	err := s.tsv.Load(record_manager.PathSoundSettings, &s.Sounds)
	if err != nil {
		s.logger.Fatal().Msgf("loading sound records: %v", err)
	}

	s.logger.Info().Msgf("%d sound records loaded", len(s.Sounds))

	//target := "button.wav"
	//for _, record := range s.Sounds {
	//	if strings.Contains(record.FileName, target) {
	//		s.logger.Info().Msgf("playing %s", record.FileName)
	//
	//		if err = s.playAudioFromRecord(record); err != nil {
	//			s.logger.Debug().Msgf("couldn't play sound file: %v", err)
	//		}
	//	}
	//}

	for { // forever play random sound files
		idx := rand.Intn(len(s.Sounds)) // random index
		record := s.Sounds[idx]         // random record

		s.logger.Info().Msgf("[%d] %s", idx, record.FileName)

		if err = s.playAudioFromRecord(record); err != nil {
			s.logger.Debug().Msgf("couldn't play sound file: %v", err)
		}

		time.Sleep(time.Second)
	}
}

func (s *audioFileTestService) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *audioFileTestService) Logger() *zerolog.Logger {
	return s.logger
}

func (s *audioFileTestService) playAudioFromRecord(record models.SoundEntry) error {
	const (
		sfx   = "data/global/sfx/"
		music = "data/global/music/"
	)

	pathSfx := fmt.Sprintf("%s%s", sfx, record.FileName)
	pathMusic := fmt.Sprintf("%s%s", music, record.FileName)

	pathSfx = filepath.Clean(pathSfx)
	pathSfx = strings.ReplaceAll(pathSfx, "\\", "/")

	pathMusic = filepath.Clean(pathMusic)
	pathMusic = strings.ReplaceAll(pathMusic, "\\", "/")

	actualPath := pathSfx

	data, err := s.mpq.Load(pathSfx) // try loading as sfx
	if err != nil {
		actualPath = pathMusic
		data, err = s.mpq.Load(pathMusic) // try loading as music
		if err != nil {
			return err
		}
	}

	dataBytes, _ := io.ReadAll(data)

	return playWavWithMPlayer(filepath.Base(actualPath), dataBytes)
}

func playWavWithMPlayer(filename string, data []byte) error {
	// Check if mplayer is installed on the system.
	_, err := exec.LookPath("mplayer")
	if err != nil {
		return fmt.Errorf("mplayer not found on the system")
	}

	// Create a temporary WAV file.
	tmpFile, err := ioutil.TempFile("", "*.wav")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name()) // Clean up the temporary file after playing.
	}()

	// Write the WAV data to the temporary file.
	_, err = tmpFile.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data to temporary file: %w", err)
	}

	// Prepare the command to run mplayer with the temporary WAV file.
	cmd := exec.Command("mplayer", tmpFile.Name())

	//// Set up stdout and stderr to be redirected to the terminal.
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr

	// Start the mplayer process.
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start mplayer: %w", err)
	}

	// Wait for the mplayer process to finish.
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("mplayer execution failed: %w", err)
	}

	return nil
}
