package static_assets

import (
	"net/http"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func (m *Middleware) staticWebUIHandler(c *gin.Context) {
	file := c.Request.RequestURI

	_, baseFilename := path.Split(file)
	ext := filepath.Ext(baseFilename)

	if file == "" || ext == "" {
		file = "index.html"
	}

	data, readErr := m.ReadFile(file)
	if readErr != nil {
		m.log.Warn().Msgf("reading file: %v", readErr)
		c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
		return
	}

	contentTypeToUse := getMIMEFromFileExtension(ext)
	if contentTypeToUse == "" {
		contentTypeToUse = http.DetectContentType(data)
	}

	c.Data(http.StatusOK, contentTypeToUse, data)

	return
}

func (m *Middleware) isRouterChanged() bool {
	if m.lastRouter == nil {
		m.lastRouter = m.router.RouteRoot()
	}

	if m.lastRouter == m.router.RouteRoot() {
		return false
	}

	m.log.Info().Msg("router change detected")

	m.lastRouter = m.router.RouteRoot()

	return true
}

func getMIMEFromFileExtension(ext string) (result string) {
	return map[string]string{
		".js":  "text/javascript",
		".css": "text/css",
	}[ext]
}
