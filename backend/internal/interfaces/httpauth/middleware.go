package httpauth

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey int

const claimsKey ctxKey = 1

func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	c, ok := ctx.Value(claimsKey).(Claims)
	return c, ok
}

func Middleware(secret, issuer string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		cl, err := Parse(parts[1], secret, issuer)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, cl)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
