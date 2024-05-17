package blob

import (
	"github.com/pritamdas99/solr-dump/model"
	"strings"

	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
)

const (
	gcsPrefix   = "gs://"
	s3Prefix    = "s3://"
	azurePrefix = "azblob://"
)

func gcsBlob(bs *model.BackupStorage) (*Blob, error) {
	return &Blob{
		storageURL: strings.Join([]string{gcsPrefix, bs.Storage.Gcs.Bucket}, ""),
		prefix:     bs.Storage.Gcs.Prefix,
	}, nil
}

func azureBlob(bs *model.BackupStorage) (*Blob, error) {
	return &Blob{
		storageURL: strings.Join([]string{azurePrefix, bs.Storage.Azure.Container}, ""),
		prefix:     bs.Storage.Azure.Prefix,
	}, nil
}

func s3Blob(bs *model.BackupStorage) (*Blob, error) {
	var storageUrl string
	storageUrl = strings.Join([]string{s3Prefix, bs.Storage.S3.Bucket}, "")
	if bs.Storage.S3.Region != "" {
		storageUrl = strings.Join([]string{storageUrl, "?region=", bs.Storage.S3.Region}, "")
	}
	if bs.Storage.S3.Endpoint != "" {
		storageUrl = strings.Join([]string{storageUrl, "&endpoint=", bs.Storage.S3.Endpoint}, "")
	}
	return &Blob{
		storageURL: storageUrl,
		prefix:     bs.Storage.S3.Prefix,
	}, nil
}
