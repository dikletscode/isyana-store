package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dikletscode/isyana-store/pkg/httperrors"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

// Constants for context keys
const (
	userCtxKey contextKey = "user"
)

func UserFromContext(ctx context.Context) jwt.MapClaims {
	claims := ctx.Value(userCtxKey).(jwt.MapClaims)
	return claims
}
func AuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		headerAuth := r.Header.Get("Authorization")

		var httpError *httperrors.Response
		if headerAuth == "" {

			httpError = &httperrors.Response{
				Status: "failed",
				Data:   nil,
				Errors: &httperrors.Errors{Code: 401, Message: "Missing token"},
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(httpError)
			return
		}
		headerParts := strings.Split(headerAuth, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			httpError = &httperrors.Response{
				Status: "failed",
				Data:   nil,
				Errors: &httperrors.Errors{Code: 401, Message: "Invalid authorization header format"},
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(httpError)
			return

		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(headerParts[1], claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte(os.Getenv("SECRET_TOKEN")), nil
		})

		if err != nil {
			// http.Error(w, "Error parsing authorization token.", http.StatusUnauthorized)
			httpError = &httperrors.Response{
				Status: "failed",
				Data:   nil,
				Errors: &httperrors.Errors{Code: 401, Message: "Unauthorized "},
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(httpError)
			return
		}

		// if result == nil {
		if token.Valid {
			ctx := context.WithValue(r.Context(), userCtxKey, claims)

			// Access context values in handlers like this
			// props, _ := r.Context().Value("props").(jwt.MapClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		// } else {
		// 	err := json.NewEncoder(w).Encode(result)
		// 	if err != nil {
		// 		http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	}

		// }
	})

}
