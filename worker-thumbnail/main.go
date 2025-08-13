package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type VideoUploaded struct {
	Bucket string `json:"bucket"`
	Object string `json:"object"`
}

func loadConfig() Config {
	cfg := Config{
		RabbitURL:      os.Getenv("RABBIT_URL"),
		MinIOEndpoint:  os.Getenv("MINIO_ENDPOINT"),
		MinIOAccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		MinIOSecretKey: os.Getenv("MINIO_SECRET_KEY"),
		MinIOBucket:    os.Getenv("MINIO_BUCKET"),
	}
	if cfg.MinIOBucket == "" {
		cfg.MinIOBucket = "videos"
	}
	return cfg
}

func main() {
	cfg := loadConfig()

	if err := InitMinIO(cfg); err != nil {
		log.Fatalf("init MinIO: %v", err)
	}
	if err := InitRabbit(cfg); err != nil {
		log.Fatalf("init RabbitMQ: %v", err)
	}
	log.Println("worker-thumbnail: connected to RabbitMQ & MinIO")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err := ConsumeQueue("video_uploaded", func(body []byte) error {
		var msg VideoUploaded
		if err := json.Unmarshal(body, &msg); err != nil {
			return err
		}

		bucket := msg.Bucket
		if bucket == "" {
			bucket = cfg.MinIOBucket
		}

		
		base := filepath.Base(msg.Object)
		base = strings.TrimSuffix(base, filepath.Ext(base))
		thumbKey := "thumbnails/" + base + ".jpg"

		log.Printf("creating thumbnail for s3://%s/%s -> %s", bucket, msg.Object, thumbKey)

		ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		if err := CreateAndUploadThumbnail(ctxTimeout, bucket, msg.Object, thumbKey, 1); err != nil {
			return err
		}

		log.Printf("thumbnail uploaded: s3://%s/%s", bucket, thumbKey)
		return nil
	})
	if err != nil {
		log.Fatalf("consume: %v", err)
	}

	<-ctx.Done()
	log.Println("worker-thumbnail: shutting down")
	CloseRabbit()
}
