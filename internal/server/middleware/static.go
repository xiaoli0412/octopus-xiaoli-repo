package middleware

import (
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func StaticEmbed(urlPrefix string, embedFS fs.FS) gin.HandlerFunc {
	fs := http.FS(embedFS)
	return static(urlPrefix, fs)
}

func StaticLocal(urlPrefix string, localPath string) gin.HandlerFunc {
	fs := http.Dir(localPath)
	return static(urlPrefix, fs)
}

func static(urlPrefix string, fileSystem http.FileSystem) gin.HandlerFunc {
	fileserver := http.FileServer(fileSystem)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}
		openPath, ok := staticOpenPath(urlPrefix, c.Request.URL.Path)
		if !ok {
			c.Next()
			return
		}
		f, err := fileSystem.Open(openPath)
		if err != nil {
			if strings.Contains(c.GetHeader("Accept"), "text/html") {
				idx, idxErr := fileSystem.Open("/index.html")
				if idxErr == nil {
					defer idx.Close()
					c.Header("Content-Type", "text/html; charset=utf-8")
					c.Header("Cache-Control", "no-cache")
					c.Status(http.StatusOK)
					io.Copy(c.Writer, idx)
					c.Abort()
					return
				}
			}
			c.Next()
			return
		}
		_ = f.Close()

		c.Header("Cache-Control", cacheControlForStaticPath(openPath))
		fileserver.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}

func staticOpenPath(urlPrefix, requestPath string) (string, bool) {
	path := requestPath
	if urlPrefix != "" {
		if !strings.HasPrefix(path, urlPrefix) {
			return "", false
		}
		path = strings.TrimPrefix(path, urlPrefix)
		if path == "" {
			path = "/"
		}
	}
	return path, true
}

func cacheControlForStaticPath(path string) string {
	normalized := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(path)), "/")
	if strings.HasPrefix(normalized, "_next/static/") {
		return "public, max-age=31536000, immutable"
	}
	return "no-cache"
}
