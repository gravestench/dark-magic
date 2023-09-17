package record_manager

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed diabloiidatafileguide.html.gz
var guide []byte

func (s *Service) Slug() string {
	return "records"
}

func (s *Service) InitRoutes(group *gin.RouterGroup) {
	group.GET("", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", s.extractGzip(guide))
	})

	group.GET("Inventory", func(c *gin.Context) {
		serveJSONData(c, s.inventory)
	})

	group.GET("ItemArmor", func(c *gin.Context) {
		serveJSONData(c, s.itemArmor)
	})

	group.GET("ObjectTypes", func(c *gin.Context) {
		serveJSONData(c, s.objectTypes)
	})

	group.GET("LevelPresets", func(c *gin.Context) {
		serveJSONData(c, s.levelPresets)
	})

	group.GET("LevelType", func(c *gin.Context) {
		serveJSONData(c, s.levelType)
	})

	group.GET("LevelWarp", func(c *gin.Context) {
		serveJSONData(c, s.levelWarp)
	})

	group.GET("LevelDetails", func(c *gin.Context) {
		serveJSONData(c, s.levelDetails)
	})

	group.GET("LevelMaze", func(c *gin.Context) {
		serveJSONData(c, s.levelMaze)
	})
}

func serveJSONData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func (s *Service) extractGzip(data []byte) []byte {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		s.logger.Fatal().Msg("ExtractTarGz: NewReader failed")
	}

	out := bytes.NewBuffer([]byte{})

	// Copy the decompressed content to the output file
	_, err = io.Copy(out, r)
	if err != nil {
		s.logger.Fatal().Msgf("extracting file: %v", err)
	}

	return out.Bytes()
}
