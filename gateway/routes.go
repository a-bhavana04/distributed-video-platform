package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupRoutes(cfg Config) *mux.Router {
	r := mux.NewRouter()
	
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}).Methods("GET")
	
	r.HandleFunc("/cluster/status", func(w http.ResponseWriter, r *http.Request) {
		status := cfg.GetClusterStatus()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}).Methods("GET")
	
	r.HandleFunc("/upload", cfg.ProxyToLeader).Methods("POST")
	
	r.HandleFunc("/videos", cfg.ProxyToLeader).Methods("GET", "POST")
	r.HandleFunc("/videos/{id}", cfg.ProxyToLeader).Methods("GET", "PUT", "DELETE")
	
	r.HandleFunc("/videos/{id}/stream", cfg.ProxyToLeader).Methods("GET")
	r.HandleFunc("/videos/{id}/thumbnail", cfg.ProxyToLeader).Methods("GET")
	
	r.PathPrefix("/raft/").HandlerFunc(cfg.ProxyToLeader)
	
	return r
}
