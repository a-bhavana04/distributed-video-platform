package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client

func InitMinIO(cfg Config) error {
	cl, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: false, // using http in docker-compose
	})
	if err != nil {
		return err
	}
	minioClient = cl

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

// CreateAndUploadThumbnail downloads `srcKey` to a temp file, runs ffmpeg to make a JPG
// at `second` seconds, and uploads to `dstKey` in the same bucket.
func CreateAndUploadThumbnail(ctx context.Context, bucket, srcKey, dstKey string, second int) error {
	// 1) Download original to temp
	inFile, err := os.CreateTemp("", "video-in-*")
	if err != nil {
		return err
	}
	defer os.Remove(inFile.Name())
	defer inFile.Close()

	obj, err := minioClient.GetObject(ctx, bucket, srcKey, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	if _, err := io.Copy(inFile, obj); err != nil {
		return err
	}
	if _, err := inFile.Seek(0, 0); err != nil {
		return err
	}

	// 2) Run ffmpeg to produce a single-frame JPG
	outFile, err := os.CreateTemp("", "thumb-*.jpg")
	if err != nil {
		return err
	}
	outPath := outFile.Name()
	outFile.Close() // ffmpeg will write it
	defer os.Remove(outPath)

	// ffmpeg -ss <second> -i input -frames:v 1 -q:v 2 output.jpg
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-ss", fmt.Sprintf("00:00:%02d", second),
		"-i", inFile.Name(),
		"-frames:v", "1",
		"-q:v", "2",
		outPath,
	)
	// silence ffmpeg noise unless there's an error
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %w", err)
	}

	// 3) Upload to MinIO
	fh, err := os.Open(outPath)
	if err != nil {
		return err
	}
	defer fh.Close()

	stat, err := fh.Stat()
	if err != nil {
		return err
	}

	_, err = minioClient.PutObject(
		ctx,
		bucket,
		dstKey,
		fh,
		stat.Size(),
		minio.PutObjectOptions{ContentType: "image/jpeg"},
	)
	return err
}
