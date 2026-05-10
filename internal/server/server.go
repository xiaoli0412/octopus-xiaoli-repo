package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	_ "github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/handlers"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
	"github.com/xiaoli0412/octopus-xiaoli-repo/static"
)

var httpSrv http.Server

const (
	httpReadHeaderTimeout = 10 * time.Second
	httpReadTimeout       = 3 * time.Minute
	httpIdleTimeout       = 60 * time.Second
	httpShutdownTimeout   = 15 * time.Second
	httpMaxHeaderBytes    = 1 << 20
)

func Start() error {
	r, err := newEngine()
	if err != nil {
		return err
	}

	httpSrv = newHTTPServer(fmt.Sprintf("%s:%d", conf.AppConfig.Server.Host, conf.AppConfig.Server.Port), r)
	listener, err := net.Listen("tcp", httpSrv.Addr)
	if err != nil {
		return err
	}
	go func() {
		if serveErr := httpSrv.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			log.Errorf("http server listen and serve error: %v", serveErr)
		}
	}()
	return nil
}

func newEngine() (*gin.Engine, error) {
	if conf.IsDebug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	// Octopus currently has no explicit trusted reverse-proxy allowlist.
	// Disable forwarded-IP trust so spoofed client IP headers do not change
	// auth throttling, logging, or future ClientIP()-based checks by default.
	if err := r.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		resp.Error(c, http.StatusInternalServerError, resp.ErrInternalServer)
		c.Abort()
	}))

	if conf.IsDebug() {
		r.Use(middleware.Logger())
	}
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.Use(middleware.Cors())
	r.Use(buildStaticMiddleware())

	if err := router.RegisterAll(r); err != nil {
		return nil, err
	}
	return r, nil
}

func Close() error {
	return closeHTTPServer(&httpSrv, httpShutdownTimeout)
}

func closeHTTPServer(srv *http.Server, timeout time.Duration) error {
	if srv == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := srv.Shutdown(ctx)
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		if closeErr := srv.Close(); closeErr != nil && !errors.Is(closeErr, http.ErrServerClosed) {
			return closeErr
		}
	}

	return err
}

func newHTTPServer(addr string, handler http.Handler) http.Server {
	return http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: httpReadHeaderTimeout,
		ReadTimeout:       httpReadTimeout,
		IdleTimeout:       httpIdleTimeout,
		MaxHeaderBytes:    httpMaxHeaderBytes,
	}
}

func buildStaticMiddleware() gin.HandlerFunc {
	staticDir := conf.AppConfig.Server.StaticDir
	if staticDir != "" {
		if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
			log.Infof("serving static assets from local directory: %s", staticDir)
			return middleware.StaticLocal("/", staticDir)
		}
		log.Warnf("configured static directory unavailable, falling back to embedded assets: %s", staticDir)
	}

	log.Infof("serving embedded static assets")
	return middleware.StaticEmbed("/", static.StaticFS)
}
