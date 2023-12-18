package hero

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gravestench/dark-magic/pkg/models"
)

func (s *Service) Slug() string {
	return "heroes"
}

func (s *Service) InitRoutes(group *gin.RouterGroup) {
	group.GET("", s.handleGetHeroes)
	group.GET("load", s.handleLoadHeroes)
	group.GET("save", s.handleSaveHeroes)
	group.GET("create/:class/:name", s.handleCreateHero)
}

func (s *Service) handleGetHeroes(c *gin.Context) {
	heroes := s.GetHeroes()
	c.JSON(http.StatusOK, heroes)
}

func (s *Service) handleLoadHeroes(c *gin.Context) {
	err := s.LoadHeroes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("loading heroes", "error", err))
		return
	}

	c.JSON(http.StatusOK, fmt.Sprintf("loaded %d heroes", len(s.heroStates)))
}

func (s *Service) handleSaveHeroes(c *gin.Context) {
	err := s.SaveHeroes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("saving heroes", "error", err))
		return
	}

	c.JSON(http.StatusOK, fmt.Sprintf("saved %d heroes", len(s.heroStates)))
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
