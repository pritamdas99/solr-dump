package model

import "context"

type Provider string

type Blob interface {
	Get(ctx context.Context, filepath string) ([]byte, error)
	List(ctx context.Context, dir string) ([]string, error)
}

const (
	ProviderS3    Provider = "S3"
	ProviderGCS   Provider = "GCS"
	ProviderAZURE Provider = "AZURE"
)

type S3 struct {
	Bucket   string
	Region   string
	Endpoint string
	Prefix   string
}
type GCS struct {
	Bucket string
	Prefix string
}
type AZURE struct {
	Container string
	Prefix    string
}
type Storage struct {
	Provider Provider
	S3       *S3
	Gcs      *GCS
	Azure    *AZURE
}
type BackupStorage struct {
	Storage Storage
}
