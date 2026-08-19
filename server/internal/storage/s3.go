package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
)

func newS3Client(ctx context.Context) (*s3.Client, string, error) {
	cfg := appconfig.Load()

	configOpts := []func(*config.LoadOptions) error{}
	if cfg.OlakeS3Region != "" {
		configOpts = append(configOpts, config.WithRegion(cfg.OlakeS3Region))
	}

	if cfg.OlakeS3AccessKeyID != "" && cfg.OlakeS3SecretAccessKey != "" {
		configOpts = append(configOpts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.OlakeS3AccessKeyID, cfg.OlakeS3SecretAccessKey, cfg.OlakeS3SessionToken),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load AWS config: %s", err)
	}

	var s3Opts []func(*s3.Options)
	if endpoint := cfg.OlakeS3Endpoint; endpoint != "" {
		// Path-style is required for MinIO and other S3-compatible endpoints (virtual-hosted
		// style would resolve bucket as a subdomain, e.g. olake.host.docker.internal).
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
	} else {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(awsCfg, s3Opts...)
	return client, cfg.OlakeS3Bucket, nil
}

func s3ObjectKey(relPath string) string {
	key := path.Clean(strings.Trim(relPath, "/"))
	prefix := strings.Trim(strings.TrimSpace(appconfig.Load().OlakeS3Prefix), "/")
	if prefix != "" {
		key = path.Join(prefix, key)
	}
	return key
}

// s3DirPrefix returns an S3 key prefix for a directory. The trailing slash is required for
// delimiter listings — path.Clean strips it, so it is re-appended after key normalization.
func s3DirPrefix(relDir string) string {
	return s3ObjectKey(relDir) + "/"
}

// keyToRelPath converts an S3 object key back to a workflow-relative path (without configured prefix).
func keyToRelPath(key string) string {
	key = strings.TrimPrefix(key, "/")
	prefix := strings.Trim(strings.TrimSpace(appconfig.Load().OlakeS3Prefix), "/")
	if prefix != "" {
		prefix = prefix + "/"
		if strings.HasPrefix(key, prefix) {
			return strings.TrimPrefix(key, prefix)
		}
	}
	return key
}

// PrefixExists checks whether any object exists under the given workflow-relative prefix.
func PrefixExists(ctx context.Context, relPrefix string) (bool, error) {
	client, bucket, err := newS3Client(ctx)
	if err != nil {
		return false, err
	}

	prefix := s3DirPrefix(relPrefix)
	out, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  &bucket,
		Prefix:  &prefix,
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return false, fmt.Errorf("failed to list s3://%s/%s: %s", bucket, prefix, err)
	}

	return len(out.Contents) > 0 || len(out.CommonPrefixes) > 0, nil
}

// ListCommonPrefixes lists immediate child "directories" under a workflow-relative prefix.
func ListCommonPrefixes(ctx context.Context, relPrefix string) ([]string, error) {
	client, bucket, err := newS3Client(ctx)
	if err != nil {
		return nil, err
	}

	prefix := s3DirPrefix(relPrefix)
	out, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:    &bucket,
		Prefix:    &prefix,
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list s3://%s/%s: %s", bucket, prefix, err)
	}

	children := make([]string, 0, len(out.CommonPrefixes))
	for _, cp := range out.CommonPrefixes {
		if cp.Prefix == nil {
			continue
		}
		rel := keyToRelPath(strings.TrimSuffix(*cp.Prefix, "/"))
		name := path.Base(rel)
		if name != "" && name != "." {
			children = append(children, name)
		}
	}

	return children, nil
}

// ListObjectNames lists object base names directly under a workflow-relative directory prefix.
func ListObjectNames(ctx context.Context, relDir string) ([]string, error) {
	client, bucket, err := newS3Client(ctx)
	if err != nil {
		return nil, err
	}

	prefix := s3DirPrefix(relDir)
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket:    &bucket,
		Prefix:    &prefix,
		Delimiter: aws.String("/"),
	})

	names := make([]string, 0)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list s3://%s/%s: %s", bucket, prefix, err)
		}

		for _, obj := range page.Contents {
			if obj.Key == nil {
				continue
			}
			rel := keyToRelPath(*obj.Key)
			name := path.Base(rel)
			if name != "" && name != "." {
				names = append(names, name)
			}
		}
	}

	return names, nil
}

// ReadObjectBytes downloads an object by workflow-relative path.
func ReadObjectBytes(ctx context.Context, relPath string) ([]byte, error) {
	return ReadFileFromS3AtRelPath(ctx, relPath)
}

// ListAllObjectRelPaths lists every object key under a prefix (recursive).
func ListAllObjectRelPaths(ctx context.Context, relPrefix string) ([]string, error) {
	client, bucket, err := newS3Client(ctx)
	if err != nil {
		return nil, err
	}

	prefix := s3DirPrefix(relPrefix)
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	})

	paths := make([]string, 0)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list s3://%s/%s: %s", bucket, prefix, err)
		}

		for _, obj := range page.Contents {
			if obj.Key == nil {
				continue
			}
			if obj.Size != nil && *obj.Size == 0 && strings.HasSuffix(*obj.Key, "/") {
				continue
			}
			paths = append(paths, keyToRelPath(*obj.Key))
		}
	}

	return paths, nil
}

func ReadFileFromS3AtRelPath(ctx context.Context, relPath string) ([]byte, error) {
	client, bucket, err := newS3Client(ctx)
	if err != nil {
		return nil, err
	}

	key := s3ObjectKey(relPath)
	out, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download s3://%s/%s: %s", bucket, key, err)
	}
	defer out.Body.Close()

	body, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read s3://%s/%s: %s", bucket, key, err)
	}

	return body, nil
}

func WriteFileToS3AtRelPath(ctx context.Context, relPath string, data []byte) error {
	client, bucket, err := newS3Client(ctx)
	if err != nil {
		return err
	}

	key := s3ObjectKey(relPath)
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to s3://%s/%s: %s", bucket, key, err)
	}

	return nil
}
