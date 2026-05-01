package middleware

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/gin-gonic/gin"
)

var maxManagementJSONBodyBytes int64 = 2 << 20

func RequireJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodDelete ||
			c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if !isJSONContentType(contentType) {
			resp.Error(c, http.StatusUnsupportedMediaType, resp.ErrInvalidJSON)
			c.Abort()
			return
		}

		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/") {
			body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxManagementJSONBodyBytes+1))
			if err != nil {
				resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
				c.Abort()
				return
			}
			if int64(len(body)) > maxManagementJSONBodyBytes {
				resp.Error(c, http.StatusRequestEntityTooLarge, "request body too large")
				c.Abort()
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}

		c.Next()
	}
}

func isJSONContentType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}
