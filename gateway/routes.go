package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRoutes configures all API routes
func SetupRoutes(cfg Config) *mux.Router {
	r := mux.NewRouter()
	
	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}).Methods("GET")
	
	// Cluster status endpoint
	r.HandleFunc("/cluster/status", func(w http.ResponseWriter, r *http.Request) {
		status := cfg.GetClusterStatus()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}).Methods("GET")
	
	// Video upload - proxy to leader
	r.HandleFunc("/upload", cfg.ProxyToLeader).Methods("POST")
	
	// Video metadata operations - proxy to leader
	r.HandleFunc("/videos", cfg.ProxyToLeader).Methods("GET", "POST")
	r.HandleFunc("/videos/{id}", cfg.ProxyToLeader).Methods("GET", "PUT", "DELETE")
	
	// Video streaming - can be served from any node
	r.HandleFunc("/videos/{id}/stream", cfg.ProxyToLeader).Methods("GET")
	r.HandleFunc("/videos/{id}/thumbnail", cfg.ProxyToLeader).Methods("GET")
	
	// RAFT admin endpoints - proxy to specific nodes
	r.PathPrefix("/raft/").HandlerFunc(cfg.ProxyToLeader)
	
	return r
}
