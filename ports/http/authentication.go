package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/torwig/user-service/entities"
)

type contextKey int

const (
	authenticatedUserCtxKey contextKey = 1
)

var ErrNotFoundInRequest = errors.New("failed to get from request")

func BearerTokenAuthentication(authenticator UserAuthenticator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeaderValue := r.Header.Get("Authorization")
			values := strings.Split(authHeaderValue, " ")
			if len(values) != 2 || values[0] != "Bearer" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			user, err := authenticator.ParseAccessToken(values[1])
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctxWithUser := context.WithValue(r.Context(), authenticatedUserCtxKey, user)
			next.ServeHTTP(w, r.WithContext(ctxWithUser))
		})
	}
}

func AuthenticatedUserFromRequest(r *http.Request) (*entities.AuthenticatedUser, error) {
	au, ok := r.Context().Value(authenticatedUserCtxKey).(*entities.AuthenticatedUser)
	if !ok {
		return au, ErrNotFoundInRequest
	}

	return au, nil
}
