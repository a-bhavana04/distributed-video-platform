package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type RaftState int

const (
	Follower RaftState = iota
	Candidate
	Leader
)

type RaftNode struct {
	mu           sync.RWMutex
	id           string
	state        RaftState
	currentTerm  int
	votedFor     string
	lastHeartbeat time.Time
	peers        []string
	
	// Video metadata storage (in production, use a real database)
	videos       map[string]VideoMetadata
}

type VideoMetadata struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Bucket      string    `json:"bucket"`
	Object      string    `json:"object"`
	ThumbnailURL string   `json:"thumbnail_url"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	UploadedAt  time.Time `json:"uploaded_at"`
	Resolutions []string  `json:"resolutions"`
}

type RaftStatus struct {
	ID       string    `json:"id"`
	IsLeader bool      `json:"is_leader"`
	State    string    `json:"state"`
	Term     int       `json:"term"`
	Peers    []string  `json:"peers"`
}

var raftNode *RaftNode

func InitRaft() {
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "node-1"
	}
	
	peersStr := os.Getenv("RAFT_PEERS")
	peers := []string{}
	if peersStr != "" {
		// Parse peers from comma-separated string
		// peers = strings.Split(peersStr, ",")
	}
	
	raftNode = &RaftNode{
		id:           nodeID,
		state:        Follower,
		currentTerm:  0,
		lastHeartbeat: time.Now(),
		peers:        peers,
		videos:       make(map[string]VideoMetadata),
	}
	
	// Start RAFT consensus protocol
	go raftNode.Run()
	
	log.Printf("RAFT node %s initialized", nodeID)
}

func (r *RaftNode) Run() {
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			r.mu.Lock()
			switch r.state {
			case Follower:
				if time.Since(r.lastHeartbeat) > 500*time.Millisecond {
					r.becomeCandidate()
				}
			case Candidate:
				r.startElection()
			case Leader:
				r.sendHeartbeats()
			}
			r.mu.Unlock()
		}
	}
}

func (r *RaftNode) becomeCandidate() {
	r.state = Candidate
	r.currentTerm++
	r.votedFor = r.id
	log.Printf("Node %s became candidate for term %d", r.id, r.currentTerm)
}

func (r *RaftNode) startElection() {
	// Simplified election - in a real implementation, you'd send vote requests
	// For now, just become leader if no peers or first node
	if len(r.peers) == 0 || r.id == "node-1" {
		r.becomeLeader()
	}
}

func (r *RaftNode) becomeLeader() {
	if r.state != Leader {
		r.state = Leader
		log.Printf("Node %s became LEADER for term %d", r.id, r.currentTerm)
	}
}

func (r *RaftNode) sendHeartbeats() {
	r.lastHeartbeat = time.Now()
	// In a real implementation, send heartbeats to followers
}

func (r *RaftNode) IsLeader() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state == Leader
}

func (r *RaftNode) GetStatus() RaftStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	stateStr := "follower"
	switch r.state {
	case Candidate:
		stateStr = "candidate"
	case Leader:
		stateStr = "leader"
	}
	
	return RaftStatus{
		ID:       r.id,
		IsLeader: r.state == Leader,
		State:    stateStr,
		Term:     r.currentTerm,
		Peers:    r.peers,
	}
}

// Video metadata operations (only allowed on leader)
func (r *RaftNode) StoreVideoMetadata(meta VideoMetadata) error {
	if !r.IsLeader() {
		return fmt.Errorf("not the leader")
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.videos[meta.ID] = meta
	return nil
}

func (r *RaftNode) GetVideoMetadata(id string) (VideoMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	meta, exists := r.videos[id]
	if !exists {
		return VideoMetadata{}, fmt.Errorf("video not found")
	}
	
	return meta, nil
}

func (r *RaftNode) ListVideos() []VideoMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	videos := make([]VideoMetadata, 0, len(r.videos))
	for _, video := range r.videos {
		videos = append(videos, video)
	}
	
	return videos
}

// HTTP handlers for RAFT endpoints
func RaftStatusHandler(w http.ResponseWriter, r *http.Request) {
	status := raftNode.GetStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func VideosListHandler(w http.ResponseWriter, r *http.Request) {
	videos := raftNode.ListVideos()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(videos)
}
