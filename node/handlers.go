package main

import (
	"encoding/json"
	"net/http"
	"time"
	"path/filepath"
	"strings"
	"fmt"
)

func UploadHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if this node is the leader
		if !raftNode.IsLeader() {
			http.Error(w, "Not the leader - please route through gateway", http.StatusServiceUnavailable)
			return
		}
		
		meta, err := UploadToMinIO(cfg.MinIOBucket, r)
		if err != nil {
			http.Error(w, "Upload failed: "+err.Error(), http.StatusBadRequest)
			return
		}
		
		// Generate video ID and create metadata
		videoID := generateVideoID(meta.Object)
		videoMeta := VideoMetadata{
			ID:          videoID,
			Title:       extractTitle(meta.Object),
			Bucket:      meta.Bucket,
			Object:      meta.Object,
			ThumbnailURL: fmt.Sprintf("/videos/%s/thumbnail", videoID),
			Size:        meta.Size,
			ContentType: meta.ContentType,
			UploadedAt:  time.Now(),
			Resolutions: []string{"original"}, // Will be updated by processor
		}
		
		// Store metadata in RAFT cluster
		if err := raftNode.StoreVideoMetadata(videoMeta); err != nil {
			http.Error(w, "Failed to store metadata: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Publish to message queue for processing
		body, err := json.Marshal(meta)
		if err != nil {
			http.Error(w, "Failed to encode message: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := PublishMessage("video_uploaded", body); err != nil {
			http.Error(w, "Failed to publish event: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(videoMeta)
	}
}

func generateVideoID(objectName string) string {
	// Extract base name and use timestamp
	base := filepath.Base(objectName)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	return fmt.Sprintf("%d_%s", time.Now().Unix(), base)
}

func extractTitle(objectName string) string {
	// Extract a clean title from the object name
	base := filepath.Base(objectName)
	title := strings.TrimSuffix(base, filepath.Ext(base))
	
	// Remove timestamp prefix if present
	parts := strings.SplitN(title, "_", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	
	return title
}
