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
		serveJSONData(c, s.Inventory)
	})

	group.GET("ItemArmor", func(c *gin.Context) {
		serveJSONData(c, s.ItemArmor)
	})

	group.GET("ObjectTypes", func(c *gin.Context) {
		serveJSONData(c, s.ObjectTypes)
	})

	group.GET("LevelPresets", func(c *gin.Context) {
		serveJSONData(c, s.LevelPresets)
	})

	group.GET("LevelType", func(c *gin.Context) {
		serveJSONData(c, s.LevelType)
	})

	group.GET("LevelWarp", func(c *gin.Context) {
		serveJSONData(c, s.LevelWarp)
	})

	group.GET("LevelDetails", func(c *gin.Context) {
		serveJSONData(c, s.LevelDetails)
	})

	group.GET("LevelMaze", func(c *gin.Context) {
		serveJSONData(c, s.LevelMaze)
	})
}

func serveJSONData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}
