package record_manager

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed diabloiidatafileguide.html
var guide []byte

func (s *Service) Slug() string {
	return "records"
}

func (s *Service) InitRoutes(group *gin.RouterGroup) {
	group.GET("", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", guide)
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
