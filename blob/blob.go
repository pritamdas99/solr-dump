package blob

import (
	"context"
	"fmt"
	"github.com/pritamdas99/solr-dump/model"
	"io"
	"os"
	"path"
	"strings"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
)

type Blob struct {
	prefix     string
	storageURL string
}

func NewBlob(bs *model.BackupStorage) (*Blob, error) {
	switch bs.Storage.Provider {
	case model.ProviderS3:
		return s3Blob(bs)
	case model.ProviderGCS:
		return gcsBlob(bs)
	case model.ProviderAZURE:
		return azureBlob(bs)
	default:
		return nil, fmt.Errorf("unknown provider: %s", bs.Storage.Provider)
	}
}

func (b *Blob) Get(ctx context.Context, filepath string) ([]byte, error) {
	dir, filename := path.Split(filepath)
	bucket, err := b.openBucket(ctx, dir)
	if err != nil {
		return nil, err
	}
	defer closeBucket(bucket)
	r, err := bucket.NewReader(ctx, filename, nil)
	if err != nil {
		return nil, err
	}
	defer func(r *blob.Reader) {
		closeErr := r.Close()
		if closeErr != nil {
			err := fmt.Errorf("failed to close reader: %s", closeErr)
			fmt.Print(err)
		}
	}(r)
	return io.ReadAll(r)
}

func (b *Blob) List(ctx context.Context, dir string) ([]string, error) {
	bucket, err := b.openBucket(ctx, dir)
	if err != nil {
		return nil, err
	}
	defer closeBucket(bucket)
	var objects []string
	iter := bucket.List(nil)
	for {
		obj, err := iter.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if ifFileObject(obj) {
			fName := path.Join(dir, obj.Key)
			//file, err := b.Get(ctx, fName)
			//if err != nil {
			//	return nil, err
			//}
			objects = append(objects, fName)
		}
	}
	return objects, nil
}

func (b *Blob) openBucket(ctx context.Context, dir string) (*blob.Bucket, error) {
	bucket, err := blob.OpenBucket(ctx, b.storageURL)
	if err != nil {
		return nil, err
	}
	suffix := strings.Trim(path.Join(b.prefix, dir), "/") + "/"
	if suffix == string(os.PathSeparator) {
		return bucket, nil
	}
	return blob.PrefixedBucket(bucket, suffix), nil
}

func ifFileObject(obj *blob.ListObject) bool {
	if !obj.IsDir && len(obj.Key) > 0 && obj.Key[len(obj.Key)-1] != '/' {
		return true
	}
	return false
}

func closeBucket(bucket *blob.Bucket) {
	closeErr := bucket.Close()
	if closeErr != nil {
		err := fmt.Errorf("failed to close bucket: %s", closeErr)
		fmt.Print(err)
	}
}
