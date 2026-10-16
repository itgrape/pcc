package infra

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/itgrape/pcc/server/internal/config"
	"github.com/itgrape/pcc/server/internal/domain"
)

const SessionDuration = 12 * time.Hour

type AdminAuth struct {
	username     string
	passwordHash [32]byte
	secureCookie bool
	mu           sync.Mutex
	sessions     map[[32]byte]time.Time
}

func NewAdminAuth(cfg config.AdminAuthConfig) (*AdminAuth, error) {
	if cfg.Username == "" || len(cfg.Password) < 16 {
		return nil, errors.New("admin username and password of at least 16 characters are required")
	}
	return &AdminAuth{
		username:     cfg.Username,
		passwordHash: sha256.Sum256([]byte(cfg.Password)),
		secureCookie: cfg.SecureCookie,
		sessions:     make(map[[32]byte]time.Time),
	}, nil
}

func (a *AdminAuth) SecureCookie() bool { return a.secureCookie }

func (a *AdminAuth) Login(username, password string) (string, error) {
	nameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(a.username))
	passwordHash := sha256.Sum256([]byte(password))
	passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], a.passwordHash[:])
	if nameMatch&passwordMatch != 1 {
		return "", domain.ErrUnauthenticated
	}
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw[:])
	key := sha256.Sum256([]byte(token))
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	for k, expires := range a.sessions {
		if !expires.After(now) {
			delete(a.sessions, k)
		}
	}
	a.sessions[key] = now.Add(SessionDuration)
	return token, nil
}

func (a *AdminAuth) Authenticate(token string) (*domain.Identity, error) {
	if len(token) != 43 {
		return nil, domain.ErrUnauthenticated
	}
	key := sha256.Sum256([]byte(token))
	a.mu.Lock()
	expires, ok := a.sessions[key]
	if ok && !expires.After(time.Now()) {
		delete(a.sessions, key)
		ok = false
	}
	a.mu.Unlock()
	if !ok {
		return nil, domain.ErrUnauthenticated
	}
	return &domain.Identity{Owner: a.username, Username: a.username, DisplayName: a.username, Roles: []string{"admin"}}, nil
}

func (a *AdminAuth) Logout(token string) {
	key := sha256.Sum256([]byte(token))
	a.mu.Lock()
	delete(a.sessions, key)
	a.mu.Unlock()
}
