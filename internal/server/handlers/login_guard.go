package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

const (
	loginFailureWindow     = 15 * time.Minute
	loginFailureLimit      = 5
	loginBlockDuration     = 15 * time.Minute
	loginAttemptMaxEntries = 4096
	loginUnknownClientIP   = "unknown"
)

type loginAttemptState struct {
	firstFailure time.Time
	blockedUntil time.Time
	failures     int
	sequence     uint64
}

var (
	loginAttemptNow = time.Now
	loginAttemptMu  sync.Mutex
	loginAttempts   = map[string]loginAttemptState{}
	loginAttemptSeq uint64
)

func pruneExpiredLoginAttemptsLocked(now time.Time) {
	for key, state := range loginAttempts {
		if !state.blockedUntil.IsZero() {
			if now.Before(state.blockedUntil) {
				continue
			}
			delete(loginAttempts, key)
			continue
		}
		if state.firstFailure.IsZero() || now.Sub(state.firstFailure) > loginFailureWindow {
			delete(loginAttempts, key)
		}
	}
}

func loginThrottleKey(c *gin.Context, username string) string {
	clientIP := loginThrottleClientIP(c)
	if clientIP == "" {
		clientIP = loginUnknownClientIP
	}
	return loginThrottleUsernameComponent(username) + "|" + clientIP
}

func loginThrottleUsernameComponent(username string) string {
	normalized := strings.ToLower(strings.TrimSpace(username))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

func loginThrottleClientIP(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	return op.ClientIPFromRequest(c.Request)
}

func loginThrottleBlocked(key string) (time.Duration, bool) {
	now := loginAttemptNow()

	loginAttemptMu.Lock()
	defer loginAttemptMu.Unlock()
	pruneExpiredLoginAttemptsLocked(now)

	state, ok := loginAttempts[key]
	if !ok {
		return 0, false
	}
	if !state.blockedUntil.IsZero() && now.Before(state.blockedUntil) {
		return state.blockedUntil.Sub(now), true
	}
	return 0, false
}

func loginThrottleRecordFailure(key string) {
	now := loginAttemptNow()

	loginAttemptMu.Lock()
	defer loginAttemptMu.Unlock()
	pruneExpiredLoginAttemptsLocked(now)

	state := loginAttempts[key]
	if state.firstFailure.IsZero() || now.Sub(state.firstFailure) > loginFailureWindow || (!state.blockedUntil.IsZero() && !now.Before(state.blockedUntil)) {
		loginAttemptSeq++
		state = loginAttemptState{firstFailure: now, sequence: loginAttemptSeq}
	}
	state.failures++
	if state.failures >= loginFailureLimit {
		state.blockedUntil = now.Add(loginBlockDuration)
	}
	loginAttempts[key] = state
	trimLoginAttemptsLocked(key)
}

func trimLoginAttemptsLocked(currentKey string) {
	for len(loginAttempts) > loginAttemptMaxEntries {
		oldestKey := ""
		var oldestTime time.Time
		var oldestSequence uint64
		for key, state := range loginAttempts {
			if key == currentKey && len(loginAttempts) > 1 {
				continue
			}
			candidate := state.firstFailure
			if oldestKey == "" || candidate.Before(oldestTime) || (candidate.Equal(oldestTime) && state.sequence < oldestSequence) {
				oldestKey = key
				oldestTime = candidate
				oldestSequence = state.sequence
			}
		}
		if oldestKey == "" {
			break
		}
		delete(loginAttempts, oldestKey)
	}
}

func loginThrottleReset(key string) {
	loginAttemptMu.Lock()
	defer loginAttemptMu.Unlock()
	delete(loginAttempts, key)
}

func resetLoginThrottleState() {
	loginAttemptMu.Lock()
	defer loginAttemptMu.Unlock()
	loginAttemptNow = time.Now
	loginAttempts = map[string]loginAttemptState{}
	loginAttemptSeq = 0
}
