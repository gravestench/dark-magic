package clientapp

import (
	"image"
	"testing"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	"github.com/gravestench/dark-magic/internal/video"
)

func TestProductionVideoBackendNeverLaunchesExternalPlayer(t *testing.T) {
	backend := newClientVideoBackend(&render.Composer{}, &audio.Mixer{}, image.Pt(800, 600))
	if _, external := backend.(video.FFplay); external {
		t.Fatal("production client selected the external ffplay diagnostic backend")
	}
}
