package raylibRenderer

import (
	"fmt"
	"image"
	"image/color"
	"io/fs"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/gravestench/dark-magic/internal/assets/decode"
)

const (
	paletteCubeSize  = 32
	paletteLUTWidth  = paletteCubeSize * paletteCubeSize
	paletteLUTHeight = paletteCubeSize
)

const paletteFragmentShader = `#version 330
in vec2 fragTexCoord;
in vec4 fragColor;
uniform sampler2D texture0;
uniform sampler2D paletteLUT;
out vec4 finalColor;

void main() {
    vec4 source = texture(texture0, fragTexCoord) * fragColor;
    vec3 cell = floor(clamp(source.rgb, 0.0, 1.0) * 31.0 + 0.5);
    float x = cell.r + cell.g * 32.0;
    float y = cell.b;
    vec2 uv = vec2((x + 0.5) / 1024.0, (y + 0.5) / 32.0);
    vec3 quantized = texture(paletteLUT, uv).rgb;
    finalColor = vec4(quantized, source.a);
}`

type paletteQuantizer struct {
	path            string
	lut             *image.RGBA
	target          rl.RenderTexture2D
	shader          rl.Shader
	texture         rl.Texture2D
	textureLocation int32
}

type gpuPaletteEffect struct {
	shader          rl.Shader
	texture         rl.Texture2D
	textureLocation int32
}

func newGPUPaletteEffect(palette color.Palette) (*gpuPaletteEffect, error) {
	lut := buildPaletteLUT(palette)
	image := rl.NewImageFromImage(lut)
	texture := rl.LoadTextureFromImage(image)
	rl.UnloadImage(image)
	if !rl.IsTextureValid(texture) {
		return nil, fmt.Errorf("renderer: upload palette lookup texture")
	}
	rl.SetTextureFilter(texture, rl.FilterPoint)
	shader := rl.LoadShaderFromMemory("", paletteFragmentShader)
	if !rl.IsShaderValid(shader) {
		rl.UnloadTexture(texture)
		return nil, fmt.Errorf("renderer: compile palette quantization shader")
	}
	location := rl.GetShaderLocation(shader, "paletteLUT")
	return &gpuPaletteEffect{shader: shader, texture: texture, textureLocation: location}, nil
}

func (effect *gpuPaletteEffect) close() {
	if effect == nil || effect.shader.ID == 0 {
		return
	}
	rl.UnloadShader(effect.shader)
	rl.UnloadTexture(effect.texture)
	effect.shader = rl.Shader{}
}

// ConfigurePaletteQuantization selects a Diablo pal.dat as an optional final
// display transform. GPU resources are created later on the renderer owner
// thread during Start.
func (s *Service) ConfigurePaletteQuantization(source fs.FS, palettePath string) error {
	if palettePath == "" {
		s.paletteQuantizer = nil
		return nil
	}
	palette, err := assetdecode.DisplayPalette(source, palettePath)
	if err != nil {
		return fmt.Errorf("renderer: output palette: %w", err)
	}
	s.paletteQuantizer = &paletteQuantizer{path: palettePath, lut: buildPaletteLUT(palette)}
	return nil
}

func buildPaletteLUT(palette color.Palette) *image.RGBA {
	lut := image.NewRGBA(image.Rect(0, 0, paletteLUTWidth, paletteLUTHeight))
	colors := make([]color.RGBA, len(palette))
	for index, entry := range palette {
		colors[index] = color.RGBAModel.Convert(entry).(color.RGBA)
		colors[index].A = 0xff
	}
	for blue := 0; blue < paletteCubeSize; blue++ {
		for green := 0; green < paletteCubeSize; green++ {
			for red := 0; red < paletteCubeSize; red++ {
				sample := color.RGBA{
					R: uint8(red * 255 / (paletteCubeSize - 1)),
					G: uint8(green * 255 / (paletteCubeSize - 1)),
					B: uint8(blue * 255 / (paletteCubeSize - 1)),
					A: 0xff,
				}
				lut.SetRGBA(red+green*paletteCubeSize, blue, nearestPaletteColor(sample, colors))
			}
		}
	}
	return lut
}

func nearestPaletteColor(sample color.RGBA, palette []color.RGBA) color.RGBA {
	best, bestDistance := color.RGBA{A: 0xff}, int(^uint(0)>>1)
	for _, candidate := range palette {
		red := int(sample.R) - int(candidate.R)
		green := int(sample.G) - int(candidate.G)
		blue := int(sample.B) - int(candidate.B)
		distance := red*red + green*green + blue*blue
		if distance < bestDistance {
			best, bestDistance = candidate, distance
		}
	}
	return best
}

func (s *Service) startPaletteQuantizer() error {
	quantizer := s.paletteQuantizer
	if quantizer == nil {
		return nil
	}
	image := rl.NewImageFromImage(quantizer.lut)
	quantizer.texture = rl.LoadTextureFromImage(image)
	rl.UnloadImage(image)
	if !rl.IsTextureValid(quantizer.texture) {
		return fmt.Errorf("renderer: upload palette lookup texture")
	}
	rl.SetTextureFilter(quantizer.texture, rl.FilterPoint)
	quantizer.shader = rl.LoadShaderFromMemory("", paletteFragmentShader)
	if !rl.IsShaderValid(quantizer.shader) {
		rl.UnloadTexture(quantizer.texture)
		return fmt.Errorf("renderer: compile palette quantization shader")
	}
	location := rl.GetShaderLocation(quantizer.shader, "paletteLUT")
	quantizer.textureLocation = location
	if err := s.resizePaletteTarget(rl.GetRenderWidth(), rl.GetRenderHeight()); err != nil {
		rl.UnloadShader(quantizer.shader)
		rl.UnloadTexture(quantizer.texture)
		return err
	}
	s.logger.Info("enabled display palette quantization", "palette", quantizer.path, "cube", paletteCubeSize)
	return nil
}

func (s *Service) stopPaletteQuantizer() {
	if quantizer := s.paletteQuantizer; quantizer != nil && quantizer.shader.ID != 0 {
		rl.UnloadShader(quantizer.shader)
		rl.UnloadTexture(quantizer.texture)
		rl.UnloadRenderTexture(quantizer.target)
		quantizer.shader = rl.Shader{}
	}
}

func (s *Service) resizePaletteTarget(width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("renderer: palette target requires positive dimensions, got %dx%d", width, height)
	}
	quantizer := s.paletteQuantizer
	if int(quantizer.target.Texture.Width) == width && int(quantizer.target.Texture.Height) == height {
		return nil
	}
	if rl.IsRenderTextureValid(quantizer.target) {
		rl.UnloadRenderTexture(quantizer.target)
	}
	quantizer.target = rl.LoadRenderTexture(int32(width), int32(height))
	if !rl.IsRenderTextureValid(quantizer.target) {
		return fmt.Errorf("renderer: create palette render target %dx%d", width, height)
	}
	rl.SetTextureFilter(quantizer.target.Texture, rl.FilterPoint)
	return nil
}

func (s *Service) renderQuantizedFrame() error {
	quantizer := s.paletteQuantizer
	if err := s.resizePaletteTarget(rl.GetRenderWidth(), rl.GetRenderHeight()); err != nil {
		return err
	}
	rl.BeginTextureMode(quantizer.target)
	rl.ClearBackground(rl.Black)
	rl.BeginMode2D(*s.GetDefaultCamera())
	s.update()
	s.render()
	rl.EndMode2D()
	rl.EndTextureMode()

	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)
	rl.BeginShaderMode(quantizer.shader)
	// Raylib clears auxiliary sampler registrations after every render-batch
	// flush, so the LUT must be rebound for the draw that consumes it.
	rl.SetShaderValueTexture(quantizer.shader, quantizer.textureLocation, quantizer.texture)
	source := rl.NewRectangle(0, 0, float32(quantizer.target.Texture.Width), -float32(quantizer.target.Texture.Height))
	destination := rl.NewRectangle(0, 0, float32(rl.GetRenderWidth()), float32(rl.GetRenderHeight()))
	rl.DrawTexturePro(quantizer.target.Texture, source, destination, rl.Vector2{}, 0, rl.White)
	rl.EndShaderMode()
	rl.EndDrawing()
	return nil
}
