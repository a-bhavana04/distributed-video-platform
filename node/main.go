package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	cfg := LoadConfig()

	if err := InitMinIO(cfg); err != nil {
		log.Fatalf("Failed to init MinIO: %v", err)
	}
	if err := InitRabbit(cfg); err != nil {
		log.Fatalf("Failed to init RabbitMQ: %v", err)
	}
	defer func() {
		if rabbitCh != nil {
			_ = rabbitCh.Close()
		}
		if rabbitConn != nil {
			_ = rabbitConn.Close()
		}
	}()

	// Initialize RAFT consensus
	InitRaft()

	r := mux.NewRouter()
	
	// Video operations
	r.HandleFunc("/upload", UploadHandler(cfg)).Methods("POST")
	r.HandleFunc("/videos", VideosListHandler).Methods("GET")
	
	// RAFT endpoints
	r.HandleFunc("/raft/status", RaftStatusHandler).Methods("GET")
	
	// Health check
	r.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	log.Printf("Node service listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
