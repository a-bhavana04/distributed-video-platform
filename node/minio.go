package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client

func InitMinIO(cfg Config) error {
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: false, // in-compose, no TLS
	})
	if err != nil {
		return err
	}
	minioClient = client

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, cfg.MinIOBucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := minioClient.MakeBucket(ctx, cfg.MinIOBucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}
	return nil
}

// UploadToMinIO parses the multipart request, uploads the file, and returns metadata.
func UploadToMinIO(bucket string, r *http.Request) (VideoMeta, error) {
	var meta VideoMeta

	// Accept up to ~100MB in memory before spilling to disk (adjust as needed)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		return meta, fmt.Errorf("parse multipart form: %w", err)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return meta, fmt.Errorf("missing form file 'file': %w", err)
	}
	defer file.Close()

	// Buffer file (you can stream if you prefer)
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return meta, fmt.Errorf("read upload: %w", err)
	}

	// Choose object name (keep original filename with a timestamp prefix)
	objectName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(header.Filename))
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	info, err := minioClient.PutObject(
		context.Background(),
		bucket,
		objectName,
		bytes.NewReader(buf.Bytes()),
		int64(buf.Len()),
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return meta, fmt.Errorf("put object: %w", err)
	}

	meta = VideoMeta{
		Bucket:      bucket,
		Object:      objectName,
		Size:        info.Size,
		ContentType: contentType,
	}
	return meta, nil
}
