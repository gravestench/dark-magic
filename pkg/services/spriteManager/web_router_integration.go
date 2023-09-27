package spriteManager

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gravestench/dark-magic/pkg/paths"
)

func (s *Service) Slug() string {
	return "sprite"
}

func (s *Service) InitRoutes(group *gin.RouterGroup) {
	group.GET("png/:palette/*path", s.handleCreatePngSpriteAtlas)
	group.GET("gif/:palette/*path", s.createAnimatedGifFromSpriteAtlas)
}

func (s *Service) extractAndDownloadFromMpq(c *gin.Context) {
	path := c.Param("path")

	s.logger.Info().Msg(path)

	stream, err := s.mpq.Load(path)
	if err != nil {
		s.logger.Error().Msgf("loading file: %v", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Error().Msgf("loading file: %v", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	filename := filepath.Base(path)

	if strings.HasSuffix(filename, ".txt") {
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusOK, string(data))
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/octet-stream", data)
}

func (s *Service) handleCreatePngSpriteAtlas(c *gin.Context) {
	path := c.Param("path")
	fileExtension := strings.ToLower(filepath.Ext(path))

	switch fileExtension {
	case ".dc6":
		s.handleCreatePngSpriteAtlasFromDc6(c)
	case ".dcc":
		s.handleCreatePngSpriteAtlasFromDcc(c)
	case ".dt1":
		s.handleCreatePngSpriteAtlasFromDt1(c)
	default:
		c.JSON(http.StatusBadRequest, fmt.Errorf("cannot create sprite for file with extension '%s'", fileExtension))
		return
	}
}

func (s *Service) handleCreatePngSpriteAtlasFromDc6(c *gin.Context) {
	paletteKey := c.Param("palette")
	path := c.Param("path")
	palette := s.lookupPalettePathByKey(paletteKey)

	s.logger.Info().Msg(path)

	data, err := s.LoadDc6ToPngSpriteAtlas(path, palette)
	if err != nil {
		s.logger.Error().Msgf("loading file: %v", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.Header("Content-Disposition", "inline")
	c.Header("Content-Type", "image/png")
	c.Data(http.StatusOK, "application/octet-stream", data)
}

func (s *Service) handleCreatePngSpriteAtlasFromDcc(c *gin.Context) {
	paletteKey := c.Param("palette")
	path := c.Param("path")
	palette := s.lookupPalettePathByKey(paletteKey)

	s.logger.Info().Msg(path)

	data, err := s.LoadDccToPngSpriteAtlas(path, palette)
	if err != nil {
		s.logger.Error().Msgf("loading file: %v", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.Header("Content-Disposition", "inline")
	c.Header("Content-Type", "image/png")
	c.Data(http.StatusOK, "application/octet-stream", data)
}

func (s *Service) handleCreatePngSpriteAtlasFromDt1(c *gin.Context) {
	paletteKey := c.Param("palette")
	path := c.Param("path")
	palette := s.lookupPalettePathByKey(paletteKey)

	s.logger.Info().Msg(path)

	data, err := s.LoadDt1ToPngSpriteAtlas(path, palette)
	if err != nil {
		s.logger.Error().Msgf("loading file: %v", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.Header("Content-Disposition", "inline")
	c.Header("Content-Type", "image/png")
	c.Data(http.StatusOK, "application/octet-stream", data)
}

func (s *Service) createAnimatedGifFromSpriteAtlas(c *gin.Context) {
	path := c.Param("path")
	fileExtension := strings.ToLower(filepath.Ext(path))

	switch fileExtension {
	case ".dc6":
		s.handleCreateAnimatedGifFromDc6(c)
	case ".dcc":
		s.handleCreateAnimatedGifFromDcc(c)
	default:
		c.JSON(http.StatusBadRequest, fmt.Errorf("cannot create sprite for file with extension '%s'", fileExtension))
		return
	}
}

func (s *Service) handleCreateAnimatedGifFromDc6(c *gin.Context) {
	paletteKey := c.Param("palette")
	path := c.Param("path")

	pathPL2 := s.lookupPalettePathByKey(paletteKey)
	gifImage, err := s.LoadDc6ToAnimatedGif(path, pathPL2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.Header("Content-Disposition", "inline")
	c.Header("Content-Type", "image/gif")
	c.Data(http.StatusOK, "application/octet-stream", gifImage)
}

func (s *Service) handleCreateAnimatedGifFromDcc(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, "not implemented")
}

func (s *Service) lookupPalettePathByKey(key string) string {
	lookup := map[string]string{
		"act1":      paths.PaletteTransformAct1,
		"act2":      paths.PaletteTransformAct2,
		"act3":      paths.PaletteTransformAct3,
		"act4":      paths.PaletteTransformAct4,
		"act5":      paths.PaletteTransformAct5,
		"endgame":   paths.PaletteTransformEndGame,
		"endgame2":  paths.PaletteTransformEndGame2,
		"fechar":    paths.PaletteTransformFechar,
		"loading":   paths.PaletteTransformLoading,
		"menu0":     paths.PaletteTransformMenu0,
		"menu1":     paths.PaletteTransformMenu1,
		"menu2":     paths.PaletteTransformMenu2,
		"menu3":     paths.PaletteTransformMenu3,
		"menu4":     paths.PaletteTransformMenu4,
		"sky":       paths.PaletteTransformSky,
		"trademark": paths.PaletteTransformTrademark,
	}

	path, found := lookup[key]
	if !found {
		path = lookup["act1"]
	}

	return path
}

func findPngEnd(data []byte) int {
	// Check if the data starts with a valid PNG signature.
	if len(data) < 8 || string(data[:8]) != "\x89PNG\r\n\x1a\n" {
		return -1
	}

	// Loop through the chunks to find the IEND chunk.
	for i := 8; i < len(data)-12; {
		chunkLength := int(data[i+0])<<24 | int(data[i+1])<<16 | int(data[i+2])<<8 | int(data[i+3])
		chunkType := string(data[i+4 : i+8])

		if chunkType == "IEND" {
			return i + 8 // Return the position after the IEND chunk.
		}

		// Move to the next chunk.
		i += 12 + chunkLength
	}

	return -1 // IEND chunk not found.
}
