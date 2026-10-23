package d2legacy

import (
	"bytes"
	"context"
	"image"
	"image/draw"
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// TestOwnedVendorContentsDrawAbovePanelAfterSceneTeardown executes the shipped
// overlay, text and button helpers through real decoding and retained rendering.
func TestOwnedVendorContentsDrawAbovePanelAfterSceneTeardown(t *testing.T) {
	testOwnedOverlayAfterSceneTeardown(t, "vendor", 1, 12)
}

func testOwnedOverlayAfterSceneTeardown(t *testing.T, overlay string, panels, minimumDrawables int) {
	t.Helper()
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the owned expansion MPQ directory")
	}

	t.Setenv("MPQ_DIRECTORY", directory)

	assets, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = assets.Close() })

	frames := map[bool]*image.RGBA{}

	for _, reused := range []bool{false, true} {
		name := "fresh"
		if reused {
			name = "recycled"
		}

		t.Run(name, func(t *testing.T) {
			runtime := modruntime.New()

			var composer render.Composer

			streams, err := NewRandomStreams(1)
			if err != nil {
				t.Fatal(err)
			}

			source := content.D2Legacy()
			if err := ConfigureModuleRuntime(runtime, source, recordstore.New(assets), streams,
				map[string]any{"test.reuse_nodes": reused}); err != nil {
				t.Fatal(err)
			}

			if err := runtime.RegisterModule(modruntime.RenderModuleWithAssets(runtime, &composer, assets)); err != nil {
				t.Fatal(err)
			}

			if err := runtime.Start(t.Context()); err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() { _ = runtime.Stop(context.Background()) })

			definition, err := modruntime.LoadDefinition(
				t.Context(), runtime, source, "lua/d2legacy/tests/support/"+overlay+"_render.lua",
			)
			if err != nil {
				t.Fatal(err)
			}

			component, err := definition.Managed().New(t.Context())
			if err != nil {
				t.Fatal(err)
			}

			if err := component.Start(t.Context()); err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() { _ = component.Stop(context.Background()) })

			backgrounds, drawable := 0, 0
			frame := image.NewRGBA(image.Rect(0, 0, 800, 600))

			for _, node := range composer.Snapshot() {
				if node.Layer != render.LayerModal || node.Resource == (render.ResourceID{}) {
					continue
				}

				resource, err := composer.ResourceSnapshot(node.Resource)
				if err != nil {
					t.Fatal(err)
				}

				pixels, ok := resource.Payload.(image.Image)
				if !ok {
					t.Fatal("overlay drawable is not a decoded image")
				}

				bounds := pixels.Bounds()
				area := bounds.Dx() * bounds.Dy()
				t.Logf("draw%d slot%d generation%d Z%d size%v position%g,%g",
					drawable, node.ID.Slot, node.ID.Generation, node.Z, bounds.Size(), node.X, node.Y)

				if area > 100*100 {
					if drawable != backgrounds {
						t.Fatalf("%s panel %d paints after foreground at draw %d", overlay, backgrounds, drawable)
					}

					backgrounds++
				}
				// These overlay nodes share an untransformed root. Composite
				// their real decoded pixels in the backend's layer/Z/slot order.
				x := int(node.X - node.OriginX*float64(bounds.Dx()))
				y := int(node.Y - node.OriginY*float64(bounds.Dy()))
				draw.Draw(frame, image.Rect(x, y, x+bounds.Dx(), y+bounds.Dy()), pixels, bounds.Min, draw.Over)

				drawable++
			}

			if backgrounds != panels || drawable < minimumDrawables {
				t.Fatalf("panels must precede all content: backgrounds=%d drawable=%d", backgrounds, drawable)
			}

			frames[reused] = frame
		})
	}

	if frames[false] != nil && frames[true] != nil && !bytes.Equal(frames[false].Pix, frames[true].Pix) {
		t.Fatal("recycled retained slots changed visible overlay pixels")
	}
}
