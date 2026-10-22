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

func TestOwnedInventoryContentsDrawAbovePanelAfterSceneTeardown(t *testing.T) {
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
				t.Context(), runtime, source, "lua/d2legacy/tests/support/inventory_render.lua",
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
					t.Fatal("inventory drawable is not a decoded image")
				}

				bounds := pixels.Bounds()
				if bounds.Dx()*bounds.Dy() > 100*100 {
					if drawable != backgrounds {
						t.Fatalf("inventory panel paints after foreground at draw %d", drawable)
					}

					backgrounds++
				}

				// These nodes share an untransformed root. Match backend layer/Z/slot order.
				x := int(node.X - node.OriginX*float64(bounds.Dx()))
				y := int(node.Y - node.OriginY*float64(bounds.Dy()))
				draw.Draw(frame, image.Rect(x, y, x+bounds.Dx(), y+bounds.Dy()), pixels, bounds.Min, draw.Over)

				drawable++
			}

			if backgrounds != 4 || drawable < 6 {
				t.Fatalf("inventory panel/content missing: backgrounds=%d drawable=%d", backgrounds, drawable)
			}

			frames[reused] = frame
		})
	}

	if frames[false] != nil && frames[true] != nil && !bytes.Equal(frames[false].Pix, frames[true].Pix) {
		t.Fatal("recycled retained slots changed visible inventory pixels")
	}
}
