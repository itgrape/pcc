package infra

import (
	"fmt"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v4"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/itgrape/pcc/server/internal/config"
	"github.com/itgrape/pcc/server/internal/query"
	"github.com/itgrape/pcc/server/pkg/idgen"
)

// OpenDB 初始化 GORM、连接池、雪花 ID 生成器与 gorm-gen 全局查询单例。
func OpenDB(dbCfg config.DBConfig, nodeID int64) (*gorm.DB, error) {
	if err := idgen.Init(nodeID); err != nil {
		return nil, err
	}

	db, err := gorm.Open(mysql.Open(dbCfg.DSN), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Warn),
		TranslateError:         true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("infra: open db: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("infra: get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(dbCfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(dbCfg.MaxIdleConns)
	if dbCfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(dbCfg.ConnMaxLifetime)
	}

	query.SetDefault(db)
	return db, nil
}

// ---- Casdoor JWT 解析 ----

type casdoorClaims struct {
	User
	jwt.RegisteredClaims
	TokenType string `json:"tokenType"`
}

type User struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
	Email string `json:"email"`
}

var (
	certMu  sync.RWMutex
	certPEM string
)

// InitCasdoor 初始化 Casdoor 公钥证书。必须在 GetUserFromToken 前调用。
func InitCasdoor(certificate string) {
	certMu.Lock()
	certPEM = certificate
	certMu.Unlock()
}

// GetUserFromToken 从 Authorization header 解析 Casdoor JWT，返回用户名。
func GetUserFromToken(authHeader string) (string, error) {
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" || token == authHeader {
		return "", fmt.Errorf("missing or malformed authorization header")
	}

	certMu.RLock()
	cert := certPEM
	certMu.RUnlock()
	if cert == "" {
		return "", fmt.Errorf("casdoor certificate not initialized")
	}

	claims, err := verifyToken(token, cert)
	if err != nil {
		return "", err
	}
	return claims.Name, nil
}

func verifyToken(tokenStr, cert string) (*casdoorClaims, error) {
	claims := &casdoorClaims{}
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		switch token.Method.Alg() {
		case jwt.SigningMethodES256.Alg():
			return jwt.ParseECPublicKeyFromPEM([]byte(cert))
		case jwt.SigningMethodES512.Alg():
			return jwt.ParseECPublicKeyFromPEM([]byte(cert))
		case jwt.SigningMethodRS256.Alg():
			return jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
		case jwt.SigningMethodRS512.Alg():
			return jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
		default:
			return nil, fmt.Errorf("unsupported jwt signing method: %v", token.Header["alg"])
		}
	}

	t, err := jwt.ParseWithClaims(tokenStr, claims, keyFunc)
	if err != nil {
		return nil, fmt.Errorf("jwt parse: %w", err)
	}
	if !t.Valid {
		return nil, fmt.Errorf("jwt token invalid")
	}
	return claims, nil
}
