package main

import (
	"crypto/rand"
	"log"
	"os"
	"testing"
)

const endpointURL = "https://s3.isk01.sakurastorage.jp"
const region = "jp-north-1"

var bucket string
var accessToken string
var accessTokenSecret string

func init() {
	bucket = os.Getenv("BUCKET")
	if bucket == "" {
		log.Fatalf("Please set BUCKET environment variable")
	}

	accessToken = os.Getenv("ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatalf("Please set ACCESS_TOKEN environment variable")
	}

	accessTokenSecret = os.Getenv("ACCESS_TOKEN_SECRET")
	if accessToken == "" {
		log.Fatalf("Please set ACCESS_TOKEN_SECRET environment variable")
	}
}

func TestUploadDownload(t *testing.T) {
	cfg := myS3Config{
		Bucket:      bucket,
		EndpointURL: endpointURL,
		Region:      region,
		AccessKey:   accessToken,
		Secret:      accessTokenSecret,
	}
	svc, err := newMyS3Service(t.Context(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	key := "test-" + rand.Text()
	content := "content-" + rand.Text()

	if err := svc.UploadBytes(t.Context(), []byte(content), key); err != nil {
		t.Fatal(err)
	}

	gotContentBytes, err := svc.DownloadBytes(t.Context(), key)
	if err != nil {
		t.Fatal(err)
	}

	if err := svc.DeleteObject(t.Context(), key); err != nil {
		t.Fatal(err)
	}

	gotContent := string(gotContentBytes)
	if gotContent != content {
		t.Errorf("content mismatch,\n got=%s\nwant=%s", gotContent, content)
	}
}
