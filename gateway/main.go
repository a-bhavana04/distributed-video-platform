package main

import (
	"log"
	"net/http"

	"github.com/rs/cors"
)

func main() {
	cfg := LoadConfig()
	
	// Setup routes
	router := SetupRoutes(cfg)
	
	// Setup CORS for frontend
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Next.js default port
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	
	handler := c.Handler(router)
	
	log.Printf("API Gateway starting on port %s", cfg.Port)
	log.Printf("Monitoring nodes: %v", cfg.NodeURLs)
	
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatalf("Gateway failed to start: %v", err)
	}
}
