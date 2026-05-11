package middleware

import (
	"context"
	"net/http"
	"strings"

	"sistema-balance/response"
	"sistema-balance/criptografia"
)

const sesion = "sesion"

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            response.ResponseError(w, http.StatusUnauthorized, 3001, "Token requerido")
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 {
            response.ResponseError(w, http.StatusUnauthorized, 3002, "Formato inválido")
            return
        }

        claims, err := criptografia.ParseToken(parts[1])
        if err != nil {
            response.ResponseError(w, http.StatusUnauthorized, 3003, "Token inválido")
            return
        }

        ctx := r.Context()
        ctx = context.WithValue(ctx, sesion, claims)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
