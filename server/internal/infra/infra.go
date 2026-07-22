package infra

import (
	"fmt"
	"slices"
	"strings"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
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

// ---- Casdoor authentication ----

type Casdoor struct {
	client *casdoorsdk.Client
}

// NewCasdoor creates the SDK client and loads the application's public
// certificate from Casdoor, so deployments do not need to carry a PEM value.
func NewCasdoor(cfg config.CasdoorConfig) (*Casdoor, error) {
	client := casdoorsdk.NewClient(
		strings.TrimRight(cfg.Endpoint, "/"),
		cfg.ClientID,
		cfg.ClientSecret,
		"",
		cfg.Organization,
		cfg.Application,
	)

	application, err := client.GetApplication(cfg.Application)
	if err != nil {
		return nil, fmt.Errorf("infra: get Casdoor application: %w", err)
	}

	certificate := application.CertPublicKey
	if certificate == "" && application.CertObj != nil {
		certificate = application.CertObj.Certificate
	}
	if certificate == "" && application.Cert != "" {
		cert, certErr := client.GetCert(application.Cert)
		if certErr != nil {
			return nil, fmt.Errorf("infra: get Casdoor certificate: %w", certErr)
		}
		certificate = cert.Certificate
	}
	if certificate == "" {
		return nil, fmt.Errorf("infra: Casdoor application %q has no public certificate", cfg.Application)
	}
	client.Certificate = certificate

	return &Casdoor{client: client}, nil
}

func (c *Casdoor) ExchangeCode(code, state string) (string, error) {
	if code == "" {
		return "", fmt.Errorf("missing authorization code")
	}
	token, err := c.client.GetOAuthToken(code, state)
	if err != nil {
		return "", fmt.Errorf("exchange OAuth code: %w", err)
	}
	if token.AccessToken == "" {
		return "", fmt.Errorf("Casdoor returned an empty access token")
	}
	return token.AccessToken, nil
}

// GetUserFromToken verifies a Casdoor access token and returns all user claims
// included in it.
func (c *Casdoor) GetUserFromToken(authHeader string) (*casdoorsdk.User, error) {
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" || token == authHeader {
		return nil, fmt.Errorf("missing or malformed authorization header")
	}

	claims, err := c.client.ParseJwtToken(token)
	if err != nil {
		return nil, fmt.Errorf("parse Casdoor token: %w", err)
	}
	if claims.TokenType != "" && claims.TokenType != "access-token" {
		return nil, fmt.Errorf("unexpected Casdoor token type %q", claims.TokenType)
	}
	if !slices.Contains(claims.Audience, c.client.ClientId) {
		return nil, fmt.Errorf("Casdoor token audience does not contain this client")
	}
	if claims.Name == "" {
		return nil, fmt.Errorf("Casdoor token has no user name")
	}
	return &claims.User, nil
}
