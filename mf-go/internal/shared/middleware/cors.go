package middleware

import (
	"github.com/go-chi/cors"
)

// CORSOptions builds chi/cors options with safe credential handling.
func CORSOptions(origins []string) cors.Options {
	opts := cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Organization-ID", "X-App-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}

	if len(origins) == 0 {
		opts.AllowedOrigins = []string{}
		opts.AllowCredentials = false
		return opts
	}

	for _, origin := range origins {
		if origin == "*" {
			opts.AllowCredentials = false
			break
		}
	}

	return opts
}
