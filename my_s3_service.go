package main

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type myS3Config struct {
	Bucket       string `yaml:"bucket"`
	EndpointURL  string `yaml:"s3_endpoint"`
	Region       string `yaml:"s3_region"`
	AccessKey    string `yaml:"access_key"`
	Secret       string `yaml:"secret"`
	UsePathStyle bool   `yaml:"use_path_style"`
}

type myS3Service struct {
	client *s3.Client
	cfg    myS3Config
}

func newMyS3Service(ctx context.Context, cfg myS3Config) (*myS3Service, error) {
	c, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey, cfg.Secret, "",
		)),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, err
	}

	svc := s3.NewFromConfig(c, func(o *s3.Options) {
		// エンドポイントの指定方法については
		// https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/endpoints/
		// 参照
		o.BaseEndpoint = aws.String(cfg.EndpointURL)
		o.UsePathStyle = cfg.UsePathStyle
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})

	return &myS3Service{client: svc, cfg: cfg}, nil
}

func (s *myS3Service) UploadBytes(ctx context.Context, content []byte, remotePath string) error {
	// PubObject内でSeekしてコンテンツのサイズを調べるらしく
	// bytes.Readerだとエラーになったのでbytes.Bufferを使う。
	r := bytes.NewBuffer(content)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(normalizeRemotePath(remotePath)),
		Body:   r,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *myS3Service) DownloadBytes(ctx context.Context, remotePath string) ([]byte, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(normalizeRemotePath(remotePath)),
	})
	if err != nil {
		return nil, err
	}

	content, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (s *myS3Service) DeleteObject(ctx context.Context, remotePath string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(normalizeRemotePath(remotePath)),
	})
	if err != nil {
		return err
	}
	return nil
}

func normalizeRemotePath(remotePath string) string {
	// NOTE: さくらのオブジェクトストレージにアップロードする場合、
	// Keyが/で始まっていると、エラーは出ないがオブジェクトが作られなかった。
	//
	// S3のPutObjectのAPIドキュメント
	// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObject.html
	// のRequest Syntaxにも
	// PUT /Key+ HTTP/1.1
	// とあるのでKeyは先頭の/を除くのが正しい。
	return strings.TrimPrefix(remotePath, "/")
}
