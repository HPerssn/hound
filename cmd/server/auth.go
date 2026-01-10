package main

import (
	"context"
	"log"
	"net/http"
)

type contextKey string

const UserIDKey contextKey = "userId"

func ExtractUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Traefik BasicAuth sets this header
		userId := r.Header.Get("X-Auth-User")

		// Also check common alternatives
		if userId == "" {
			userId = r.Header.Get("X-Forwarded-User")
		}
		if userId == "" {
			userId = r.Header.Get("Remote-User")
		}

		// Development mode - uncomment to bypass auth locally
		if userId == "" {
			userId = "dev-user"
			log.Println("Warning: No auth header, using dev-user")
		}

		if userId == "" {
			log.Printf("Authentication failed: no user header found")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Printf("Authenticated request from user: %s", userId)

		ctx := context.WithValue(r.Context(), UserIDKey, userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserId(r *http.Request) string {
	userId, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		return ""
	}
	return userId
}
