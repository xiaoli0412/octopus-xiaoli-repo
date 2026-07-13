package handlers

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
)

func init() {
	router.NewGroupRouter("").
		AddRoute(
			router.NewRoute("/version", http.MethodGet).
				Handle(version),
		)
}

// versionInfo 返回服务版本元信息，无需鉴权。
type versionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
	GoVersion string `json:"goVersion"`
}

func version(c *gin.Context) {
	c.JSON(http.StatusOK, versionInfo{
		Version:   conf.Version,
		Commit:    conf.Commit,
		BuildTime: conf.BuildTime,
		GoVersion: runtime.Version(),
	})
}
