package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port      string
	NodeURLs  []string // List of all RAFT nodes
}

type NodeStatus struct {
	ID       string `json:"id"`
	IsLeader bool   `json:"is_leader"`
	Status   string `json:"status"`
	URL      string `json:"url"`
}

type ClusterStatus struct {
	Leader    *NodeStatus   `json:"leader"`
	Followers []NodeStatus  `json:"followers"`
	Healthy   bool          `json:"healthy"`
}

func LoadConfig() Config {
	nodeURLsStr := os.Getenv("NODE_URLS")
	if nodeURLsStr == "" {
		nodeURLsStr = "http://node-1:9000,http://node-2:9000,http://node-3:9000"
	}
	
	return Config{
		Port:     getEnv("PORT", "8080"),
		NodeURLs: strings.Split(nodeURLsStr, ","),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// DiscoverLeader queries all nodes to find the current RAFT leader
func (cfg Config) DiscoverLeader() (*NodeStatus, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	
	for _, nodeURL := range cfg.NodeURLs {
		resp, err := client.Get(nodeURL + "/raft/status")
		if err != nil {
			continue // Try next node
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			continue
		}
		
		var status NodeStatus
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			continue
		}
		
		if status.IsLeader {
			status.URL = nodeURL
			return &status, nil
		}
	}
	
	return nil, fmt.Errorf("no leader found among nodes: %v", cfg.NodeURLs)
}

// GetClusterStatus returns the status of all nodes in the cluster
func (cfg Config) GetClusterStatus() ClusterStatus {
	client := &http.Client{Timeout: 3 * time.Second}
	var leader *NodeStatus
	var followers []NodeStatus
	
	for _, nodeURL := range cfg.NodeURLs {
		resp, err := client.Get(nodeURL + "/raft/status")
		if err != nil {
			// Node is down
			followers = append(followers, NodeStatus{
				URL:    nodeURL,
				Status: "down",
			})
			continue
		}
		defer resp.Body.Close()
		
		var status NodeStatus
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			followers = append(followers, NodeStatus{
				URL:    nodeURL,
				Status: "unhealthy",
			})
			continue
		}
		
		status.URL = nodeURL
		status.Status = "healthy"
		
		if status.IsLeader {
			leader = &status
		} else {
			followers = append(followers, status)
		}
	}
	
	return ClusterStatus{
		Leader:    leader,
		Followers: followers,
		Healthy:   leader != nil,
	}
}

// ProxyToLeader forwards requests to the current RAFT leader
func (cfg Config) ProxyToLeader(w http.ResponseWriter, r *http.Request) {
	leader, err := cfg.DiscoverLeader()
	if err != nil {
		http.Error(w, "No leader available: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	
	// Create proxy request
	targetURL := leader.URL + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}
	
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}
	
	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}
	
	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Proxy request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	
	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	
	// Copy status code and body
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
