package locale

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Service) Slug() string {
	return "locale"
}

func (s *Service) InitRoutes(group *gin.RouterGroup) {
	group.GET("", func(c *gin.Context) {
		serveJSONData(c, s.compositeStringTable)
	})

	group.GET("lookup/:key", s.handleLookupString)
}

func (s *Service) handleLookupString(c *gin.Context) {
	key := c.Param("key")

	val, err := s.GetTextByKey(key)
	if err != nil {
		val = key
	}

	c.String(http.StatusInternalServerError, val)
}

func serveJSONData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}
