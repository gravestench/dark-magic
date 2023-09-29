package mapGenerator

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Service) Slug() string {
	return "mapgen"
}

func (s *Service) InitRoutes(group *gin.RouterGroup) {
	group.GET("seed", s.handleGetSeed)
	group.PUT("seed", s.handlePutSeed)
	group.GET("create/:act/:difficulty", s.handleGenerateMap)
}

func (s *Service) handleGetSeed(c *gin.Context) {
	c.JSON(http.StatusOK, s.Seed())
}

func (s *Service) handlePutSeed(c *gin.Context) {
	type payload struct {
		Seed uint64
	}

	var data payload

	if err := c.Bind(&data); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("error binding payload: %v", err))
	}

	s.SetSeed(data.Seed)

	c.String(http.StatusOK, "success")
}

func (s *Service) handleGenerateMap(c *gin.Context) {
	paramAct, err := strconv.ParseInt(c.Param("act"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("error parsing act: %v", err))
		return
	}

	paramDifficulty, err := strconv.ParseInt(c.Param("difficulty"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("error parsing difficulty: %v", err))
		return
	}

	m, err := s.GenerateMap(uint(paramAct), uint(paramDifficulty))
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("error generating world map: %v", err))
		return
	}

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusOK, m)
}
