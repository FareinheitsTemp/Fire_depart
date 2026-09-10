package main

import (
	"log"
	"net/http"

	"github.com/FareinheitsTemp/fire_depart/internal/api"
	"github.com/FareinheitsTemp/fire_depart/internal/config"
)

func main() {
	cfg := config.Load()

	handler := withCORS(cfg.AllowedOrigins, api.NewRouter())

	log.Printf("listening on %s", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, handler))
}

// withCORS дозволяє запити з dev-сервера фронтенду (Vite, :5173).
func withCORS(origins []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			for _, allowed := range origins {
				if allowed == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
					break
				}
			}
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
