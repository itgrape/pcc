package httpapi

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itgrape/pcc/server/pkg/types"
)

const sessionCookieName = "pcc_session"

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginAttempt struct {
	count int
	until time.Time
}
type loginLimiter struct {
	mu       sync.Mutex
	attempts map[string]loginAttempt
}

func (l *loginLimiter) blocked(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	a := l.attempts[ip]
	if time.Now().After(a.until) {
		delete(l.attempts, ip)
		return false
	}
	return a.count >= 5
}

func (l *loginLimiter) record(ip string, success bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if success {
		delete(l.attempts, ip)
		return
	}
	if len(l.attempts) > 10000 {
		for key, value := range l.attempts {
			if time.Now().After(value.until) {
				delete(l.attempts, key)
			}
		}
	}
	if len(l.attempts) > 10000 {
		return
	}
	a := l.attempts[ip]
	if time.Now().After(a.until) {
		a = loginAttempt{until: time.Now().Add(15 * time.Minute)}
	}
	a.count++
	l.attempts[ip] = a
}

func sameOrigin(c *gin.Context) bool {
	origin := c.GetHeader("Origin")
	if origin == "" {
		return false
	}
	u, err := url.Parse(origin)
	return err == nil && (u.Scheme == "https" || u.Scheme == "http") &&
		strings.EqualFold(u.Host, c.Request.Host) && u.Path == "" && u.RawQuery == ""
}

func requireSameOrigin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead && c.Request.Method != http.MethodOptions && !sameOrigin(c) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}

func (h *handlers) sessionCookie(token string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name: sessionCookieName, Value: token, Path: "/", MaxAge: maxAge,
		HttpOnly: true, Secure: h.deps.Auth.SecureCookie(), SameSite: http.SameSiteStrictMode,
	}
}

func (h *handlers) login(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	if h.deps.Auth.SecureCookie() && strings.HasPrefix(c.GetHeader("Origin"), "http://") {
		c.JSON(http.StatusBadRequest, types.Response{Code: http.StatusBadRequest, Message: "HTTPS is required for login"})
		return
	}
	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	if h.limiter.blocked(ip) {
		c.Header("Retry-After", "900")
		c.JSON(http.StatusTooManyRequests, types.Response{Code: http.StatusTooManyRequests, Message: "too many login attempts"})
		return
	}
	var req loginRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8192)
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Username) > 256 || len(req.Password) > 4096 {
		c.JSON(http.StatusBadRequest, types.Response{Code: http.StatusBadRequest, Message: "invalid login request"})
		return
	}
	token, err := h.deps.Auth.Login(req.Username, req.Password)
	if err != nil {
		h.limiter.record(ip, false)
		c.JSON(http.StatusUnauthorized, types.Response{Code: http.StatusUnauthorized, Message: "invalid username or password"})
		return
	}
	h.limiter.record(ip, true)
	http.SetCookie(c.Writer, h.sessionCookie(token, 12*60*60))
	c.JSON(http.StatusOK, resOK(true))
}

func (h *handlers) logout(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	if cookie, err := c.Request.Cookie(sessionCookieName); err == nil {
		h.deps.Auth.Logout(cookie.Value)
	}
	http.SetCookie(c.Writer, h.sessionCookie("", -1))
	c.JSON(http.StatusOK, resOK(true))
}
