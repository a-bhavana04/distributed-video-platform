package main

type VideoMeta struct {
	Bucket      string `json:"bucket"`
	Object      string `json:"object"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}
