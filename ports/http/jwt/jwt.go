package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/torwig/user-service/entities"
)

const maxNumberOfPermissions = 4

var (
	ErrUnexpectedIssuer        = errors.New("unexpected token issuer")
	ErrUnexpectedClaims        = errors.New("unexpected claims received")
	ErrInvalidAccessToken      = errors.New("invalid access token")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
)

type authClaims struct {
	UserID         int64 `json:"user_id,omitempty"`
	CanCreateUsers bool  `json:"can_create_users"`
	CanDeleteUsers bool  `json:"can_delete_users"`
	CanUpdateUsers bool  `json:"can_update_users"`
	CanViewUsers   bool  `json:"can_view_users"`
	jwt.RegisteredClaims
}

type Config struct {
	SecretKey string `yaml:"secret_key"`
	Issuer    string `yaml:"issuer"`
}

type Authenticator struct {
	cfg Config
}

func NewAuthenticator(cfg Config) *Authenticator {
	return &Authenticator{cfg: cfg}
}

func (a *Authenticator) ParseAccessToken(t string) (*entities.AuthenticatedUser, error) {
	token, err := jwt.ParseWithClaims(t, &authClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnexpectedSigningMethod
		}

		return []byte(a.cfg.SecretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidAccessToken
	}

	iss, err := token.Claims.GetIssuer()
	if err != nil || iss != a.cfg.Issuer {
		return nil, ErrUnexpectedIssuer
	}

	claims, ok := token.Claims.(*authClaims)
	if !ok {
		return nil, ErrUnexpectedClaims
	}

	permissions := make([]entities.UserPermission, 0, maxNumberOfPermissions)

	if claims.CanCreateUsers {
		permissions = append(permissions, entities.CreateUsersGranted())
	}
	if claims.CanViewUsers {
		permissions = append(permissions, entities.ViewUsersGranted())
	}
	if claims.CanUpdateUsers {
		permissions = append(permissions, entities.UpdateUsersGranted())
	}
	if claims.CanDeleteUsers {
		permissions = append(permissions, entities.DeleteUsersGranted())
	}

	au := entities.NewAuthenticatedUser(claims.UserID, permissions...)

	return au, nil
}
