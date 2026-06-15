package main

import (
	"net/http"
	"strings"
)

// AuthMiddleware Bearer鉴权中间件
// 配置中api_key为空时直接放行
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := GetConfig()
		if cfg.APIKey == "" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondJSON(w, false, "", "missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondJSON(w, false, "", "invalid Authorization format", http.StatusUnauthorized)
			return
		}

		if parts[1] != cfg.APIKey {
			respondJSON(w, false, "", "invalid API key", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}
