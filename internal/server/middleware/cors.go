package middleware

import (
	"net"
	"net/url"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func isLocalDevOrigin(origin string) bool {
	host := originHostFromOrigin(origin)
	if host == "" {
		return false
	}

	if host == "localhost" || host == "127.0.0.1" {
		return true
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func originHostFromOrigin(origin string) string {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return ""
	}

	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" {
		return strings.Trim(origin, "/")
	}

	host := strings.TrimSpace(parsed.Host)
	if extracted, _, err := net.SplitHostPort(host); err == nil {
		host = extracted
	}
	host = strings.Trim(host, "[]")
	return strings.TrimRight(host, "/")
}

func Cors() gin.HandlerFunc {
	config := cors.DefaultConfig()
	// Octopus uses explicit Authorization headers instead of browser cookies.
	// Keep credentialed cross-origin requests disabled to avoid widening trust
	// boundaries when allow-origins is configured too broadly (for example `*`).
	config.AllowCredentials = false
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"*"}
	config.ExposeHeaders = []string{"Content-Disposition"}
	config.AllowOriginFunc = func(origin string) bool {
		if conf.IsDebug() && isLocalDevOrigin(origin) {
			return true
		}

		allowed, err := op.SettingGetString(model.SettingKeyCORSAllowOrigins)
		if err != nil {
			return false
		}
		allowed = strings.TrimSpace(allowed)
		if allowed == "" {
			return false
		}
		if allowed == "*" {
			return true
		}

		origin = strings.TrimSpace(origin)
		if origin == "" {
			return false
		}

		originHost := originHostFromOrigin(origin)

		for _, item := range strings.Split(allowed, ",") {
			item = strings.TrimSpace(item)
			item = strings.TrimRight(item, "/")
			if item == "" {
				continue
			}
			if item == origin || item == originHost {
				return true
			}
		}
		return false
	}
	return cors.New(config)
}
