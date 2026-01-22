package main

import (
	"net/http"
	"reviewExplorer/backend/db"
	"reviewExplorer/backend/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	db.Connect()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/schools", handlers.GetSchools)
	r.Post("/analyze", handlers.Analyze)
	r.Post("/refresh", handlers.Refresh)

	fs := http.FileServer(http.Dir("./frontend"))
	r.Handle("/*", http.StripPrefix("/", fs))

	http.ListenAndServe(":8081", r)
}
