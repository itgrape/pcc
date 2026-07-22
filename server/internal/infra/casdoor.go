package infra

import (
	"fmt"
	"slices"
	"strings"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"

	"github.com/itgrape/pcc/server/internal/config"
	"github.com/itgrape/pcc/server/internal/domain"
)

type Casdoor struct {
	client *casdoorsdk.Client
}

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

func (c *Casdoor) Authenticate(authHeader string) (*domain.Identity, error) {
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
	roles := make([]string, 0, len(claims.Roles))
	for _, role := range claims.Roles {
		if role != nil && role.Name != "" {
			roles = append(roles, role.Name)
		}
	}

	return &domain.Identity{
		Owner:       claims.Owner,
		Username:    claims.Name,
		DisplayName: claims.DisplayName,
		Email:       claims.Email,
		Phone:       claims.Phone,
		Avatar:      claims.Avatar,
		Roles:       roles,
		Groups:      claims.Groups,
		Properties:  claims.Properties,
	}, nil
}
