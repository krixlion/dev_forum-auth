package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-lib/logging"
)

// Type created only to avoid collisions when using context.WithValue().
type contextKey string

// TokenKey to use to extract translated token from request context.
const TokenKey contextKey = "token-key"

// Auth returns Middleware which validates the Bearer token extracted from
// the Authorization header. If the token is missing or invalid, it responds with 401.
// If the token is valid, the translated token is added to the request's context
// using context.WithValue().
// Use r.Context().Value(middleware.TokenKey) to extract the token.
func Auth(translator tokens.Translator, logger logging.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			bearer, ok := r.Header["Authorization"]
			if !ok {
				respond(ctx, w, http.StatusUnauthorized, "Authorization header is missing", logger)
				return
			}

			if len(bearer) <= 0 || bearer[0] == "" {
				respond(ctx, w, http.StatusUnauthorized, "Bearer token is missing", logger)
				return
			}

			opaqueToken, found := strings.CutPrefix(bearer[0], "Bearer ")
			if !found {
				respond(ctx, w, http.StatusUnauthorized, "Bearer token is malformed", logger)
				return
			}

			token, err := translator.TranslateAccessToken(opaqueToken)
			if err != nil {
				respond(ctx, w, http.StatusUnauthorized, "Bearer token is invalid", logger)
				return
			}

			h.ServeHTTP(w, r.WithContext(context.WithValue(ctx, TokenKey, token)))
		})
	}
}

func respond(ctx context.Context, w http.ResponseWriter, status int, body string, logger logging.Logger) {
	w.WriteHeader(status)
	if _, err := w.Write([]byte(body)); err != nil {
		logger.Log(ctx, "failed to write response: %v", err)
	}
}
