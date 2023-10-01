package hero

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gravestench/dark-magic/pkg/models"
)

func (s *Service) Slug() string {
	return "heros"
}

func (s *Service) InitRoutes(group *gin.RouterGroup) {
	group.GET("", s.handleGetHeros)
	group.GET("load", s.handleLoadHeros)
	group.GET("save", s.handleSaveHeros)
	group.GET("create/:class/:name", s.handleCreateHero)
}

func (s *Service) handleGetHeros(c *gin.Context) {
	heros := s.GetHeros()
	c.JSON(http.StatusOK, heros)
}

func (s *Service) handleLoadHeros(c *gin.Context) {
	err := s.LoadHeros()
	if err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("loading heros: %v", err))
		return
	}

	c.JSON(http.StatusOK, fmt.Sprintf("loaded %d heros", len(s.heroStates)))
}

func (s *Service) handleSaveHeros(c *gin.Context) {
	err := s.SaveHeros()
	if err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("saving heros: %v", err))
		return
	}

	c.JSON(http.StatusOK, fmt.Sprintf("saved %d heros", len(s.heroStates)))
}

func (s *Service) handleCreateHero(c *gin.Context) {
	class := c.Param("class")
	name := c.Param("name")

	if name == "" {
		c.String(http.StatusBadRequest, "invalid name")
		return
	}

	heroType := models.HeroFromString(class)

	if heroType == models.HeroNone {
		c.String(http.StatusBadRequest, "invalid hero type")
		return
	}

	s.CreateHero(name, heroType)

	c.String(http.StatusOK, fmt.Sprintf("created %q hero with name %q", heroType.String(), name))
}
