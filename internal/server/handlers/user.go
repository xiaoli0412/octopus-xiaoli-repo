package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/auth"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/user").
		Use(middleware.RequireJSON()).
		AddRoute(
			router.NewRoute("/login", http.MethodPost).
				Handle(login),
		)
	router.NewGroupRouter("/api/v1/user").
		Use(middleware.Auth()).
		Use(middleware.RequireJSON()).
		AddRoute(
			router.NewRoute("/force-change-password", http.MethodPost).
				Handle(forceChangePassword),
		).
		AddRoute(
			router.NewRoute("/change-password", http.MethodPost).
				Handle(changePassword),
		).
		AddRoute(
			router.NewRoute("/change-username", http.MethodPost).
				Handle(changeUsername),
		).
		AddRoute(
			router.NewRoute("/status", http.MethodGet).
				Handle(status),
		)
}

func login(c *gin.Context) {
	var user model.UserLogin
	if err := c.ShouldBindJSON(&user); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	throttleKey := loginThrottleKey(c, user.Username)
	if retryAfter, blocked := loginThrottleBlocked(throttleKey); blocked {
		c.Header("Retry-After", formatRetryAfterSeconds(retryAfter))
		resp.Error(c, http.StatusTooManyRequests, "Too many login attempts. Try again later.")
		return
	}
	if user.Expire < -1 {
		resp.Error(c, http.StatusBadRequest, "expire must be -1, 0, or a positive number of minutes")
		return
	}
	if err := op.UserVerify(user.Username, user.Password); err != nil {
		loginThrottleRecordFailure(throttleKey)
		resp.Error(c, http.StatusUnauthorized, resp.ErrUnauthorized)
		return
	}
	loginThrottleReset(throttleKey)
	token, expire, err := auth.GenerateJWTToken(user.Expire)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrInternalServer)
		return
	}
	resp.Success(c, model.UserLoginResponse{Token: token, ExpireAt: expire, MustChangePassword: op.UserMustChangePassword()})
}

func formatRetryAfterSeconds(delay time.Duration) string {
	seconds := int(delay.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	return fmt.Sprintf("%d", seconds)
}

func changePassword(c *gin.Context) {
	var user model.UserChangePassword
	if err := c.ShouldBindJSON(&user); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := op.UserChangePassword(user.OldPassword, user.NewPassword); err != nil {
		switch {
		case errors.Is(err, op.ErrIncorrectOldPassword):
			resp.Error(c, http.StatusUnauthorized, resp.ErrUnauthorized)
			return
		case errors.Is(err, op.ErrInvalidNewPassword):
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "password changed successfully")
}

func forceChangePassword(c *gin.Context) {
	var user model.UserForceChangePassword
	if err := c.ShouldBindJSON(&user); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := op.UserForceChangePassword(user.NewUsername, user.NewPassword); err != nil {
		switch {
		case errors.Is(err, op.ErrInvalidNewPassword):
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, op.ErrForcePasswordNotRequired):
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, model.UserStatusResponse{OK: true, MustChangePassword: false})
}

func changeUsername(c *gin.Context) {
	var user model.UserChangeUsername
	if err := c.ShouldBindJSON(&user); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if err := op.UserChangeUsername(user.CurrentPassword, user.NewUsername); err != nil {
		switch {
		case errors.Is(err, op.ErrInvalidUsername), errors.Is(err, op.ErrSameUsername):
			resp.Error(c, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, op.ErrIncorrectCurrentPassword):
			resp.Error(c, http.StatusUnauthorized, resp.ErrUnauthorized)
			return
		}
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "username changed successfully")
}

func status(c *gin.Context) {
	resp.Success(c, model.UserStatusResponse{OK: true, MustChangePassword: op.UserMustChangePassword()})
}
