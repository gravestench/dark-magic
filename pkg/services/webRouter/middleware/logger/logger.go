package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ginHands struct {
	*slog.Logger
	SerName    string
	Path       string
	Latency    time.Duration
	Method     string
	StatusCode int
	ClientIP   string
	MsgStr     string
}

func Logger(serName string, l *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		// before request
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// after request
		// latency := time.Since(t)
		// clientIP := c.ClientIP()
		// method := c.Request.Method
		// statusCode := c.Writer.Status()
		if raw != "" {
			path = path + "?" + raw
		}
		msg := c.Errors.String()
		if msg == "" {
			msg = "Request"
		}
		cData := &ginHands{
			Logger:     l,
			SerName:    serName,
			Path:       path,
			Latency:    time.Since(t),
			Method:     c.Request.Method,
			StatusCode: c.Writer.Status(),
			ClientIP:   c.ClientIP(),
			MsgStr:     msg,
		}

		defer func() {
			_ = recover()
		}()

		c.Next()

		logSwitch(cData)
	}
}

func logSwitch(data *ginHands) {
	switch {
	case data.StatusCode >= http.StatusBadRequest && data.StatusCode <= http.StatusInternalServerError:
		data.Warn(data.Method, "status", data.StatusCode, "path", data.Path, "latency", data.Latency)
	default:
		data.Info(data.Method, "status", data.StatusCode, "path", data.Path, "latency", data.Latency)
	}
}
