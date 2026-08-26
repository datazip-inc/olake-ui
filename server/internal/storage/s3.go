package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/storagemode"
	"github.com/spf13/viper"
)

// JobConfig is a named blob written to S3.
type JobConfig struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

var (
	s3Client *s3.Client
	s3Bucket string
)

// InitStorage initializes the shared S3 client when storage mode is S3. No-op for NFS.
func InitStorage(ctx context.Context) error {
	if storagemode.Get() != constants.StorageModeS3 {
		return nil
	}

	configOpts := []func(*config.LoadOptions) error{}
	if region := viper.GetString(constants.EnvS3Region); region != "" {
		configOpts = append(configOpts, config.WithRegion(region))
	}

	accessKey := viper.GetString(constants.EnvS3AccessKeyID)
	secretKey := viper.GetString(constants.EnvS3SecretAccessKey)
	if accessKey != "" && secretKey != "" {
		configOpts = append(configOpts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, viper.GetString(constants.EnvS3SessionToken)),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %s", err)
	}

	var s3Opts []func(*s3.Options)
	if endpoint := viper.GetString(constants.EnvS3Endpoint); endpoint != "" {
		// Path-style is required for MinIO and other S3-compatible endpoints (virtual-hosted
		// style would resolve bucket as a subdomain, e.g. olake.host.docker.internal).
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
	}

	s3Client = s3.NewFromConfig(awsCfg, s3Opts...)
	s3Bucket = viper.GetString(constants.EnvS3Bucket)
	if s3Bucket == "" {
		return fmt.Errorf("s3 bucket is required when storage mode is s3")
	}
	return nil
}

func getS3Client() (*s3.Client, string, error) {
	if s3Client == nil {
		return nil, "", fmt.Errorf("s3 storage not initialized")
	}
	return s3Client, s3Bucket, nil
}

// configStorageKey mirrors the NFS layout as an S3 object key.
// With isDirectory true, returns a directory prefix ending with "/".
// Otherwise returns <prefix>/<workflow-dir>/<relativePath> as an object key without a trailing slash.
func configStorageKey(workDir, relativePath string, isDirectory bool) (string, error) {
	workRel, err := filepath.Rel(constants.DefaultConfigDir, workDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve storage path for %s: %s", workDir, err)
	}

	prefix := strings.Trim(viper.GetString(constants.EnvS3Prefix), "/")
	key := path.Join(prefix, workRel, relativePath)
	if isDirectory {
		return strings.TrimSuffix(key, "/") + "/", nil
	}
	return key, nil
}

// WriteFilesToS3 writes config files to the S3 bucket under workDir.
func WriteFilesToS3(ctx context.Context, workDir string, configs []JobConfig) error {
	client, bucket, err := getS3Client()
	if err != nil {
		return err
	}

	for _, jobConfig := range configs {
		key, err := configStorageKey(workDir, jobConfig.Name, false)
		if err != nil {
			return err
		}

		_, err = client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &bucket,
			Key:    &key,
			Body:   strings.NewReader(jobConfig.Data),
		})
		if err != nil {
			return fmt.Errorf("failed to upload %s to s3://%s/%s: %s", jobConfig.Name, bucket, key, err)
		}
	}

	return nil
}

// ReadFileFromS3 reads a file from the S3 bucket.
// When workDir is empty, relativePath is treated as a key under the configured prefix.
func ReadFileFromS3(ctx context.Context, workDir, relativePath string, validateJSON bool) (string, error) {
	var key string
	var err error
	if workDir == "" {
		key = s3ObjectKey(relativePath, false)
	} else {
		key, err = configStorageKey(workDir, relativePath, false)
		if err != nil {
			return "", err
		}
	}

	client, bucket, err := getS3Client()
	if err != nil {
		return "", err
	}

	out, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return "", fmt.Errorf("failed to download %s from s3://%s/%s: %s", relativePath, bucket, key, err)
	}
	defer out.Body.Close()

	body, err := io.ReadAll(out.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read %s from s3://%s/%s: %s", relativePath, bucket, key, err)
	}

	if validateJSON {
		ref := fmt.Sprintf("s3://%s/%s", bucket, key)
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return "", fmt.Errorf("failed to read %s: failed to parse JSON from %s: %s", relativePath, ref, err)
		}
	}

	return string(body), nil
}

// PrefixExists checks whether any object exists under the given workflow-relative prefix.
func PrefixExists(ctx context.Context, relPrefix string) (bool, error) {
	client, bucket, err := getS3Client()
	if err != nil {
		return false, err
	}

	prefix := s3ObjectKey(relPrefix, true)
	out, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  &bucket,
		Prefix:  &prefix,
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return false, fmt.Errorf("failed to list s3://%s/%s: %s", bucket, prefix, err)
	}

	return len(out.Contents) > 0, nil
}

func listS3Objects(ctx context.Context, relPrefix string, delimiter *string, isDirectory bool) ([]types.Object, []types.CommonPrefix, error) {
	client, bucket, err := getS3Client()
	if err != nil {
		return nil, nil, err
	}

	prefix := s3ObjectKey(relPrefix, isDirectory)
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket:    &bucket,
		Prefix:    &prefix,
		Delimiter: delimiter,
	})

	var objects []types.Object
	var commonPrefixes []types.CommonPrefix

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list s3://%s/%s: %s", bucket, prefix, err)
		}

		objects = append(objects, page.Contents...)
		commonPrefixes = append(commonPrefixes, page.CommonPrefixes...)
	}

	return objects, commonPrefixes, nil
}

// ListFolderNamesWithPrefix lists one-level folder names whose keys start with relPrefix.
// relPrefix is not a complete directory (no trailing slash), so "<hash>/logs/sync_"
// matches "sync_2026-08-25_18-10-00" and not "worker".
func ListFolderNamesWithPrefix(ctx context.Context, relPrefix string) ([]string, error) {
	_, prefixes, err := listS3Objects(ctx, relPrefix, aws.String("/"), false)
	if err != nil {
		return nil, err
	}

	folderNames := make([]string, 0, len(prefixes))
	for _, cp := range prefixes {
		if cp.Prefix == nil {
			continue
		}
		folderNames = append(folderNames, path.Base(keyToRelPath(*cp.Prefix)))
	}

	return folderNames, nil
}

// ListObjectNames lists object base names directly under a workflow-relative directory prefix.
func ListObjectNames(ctx context.Context, relDir string) ([]string, error) {
	objects, _, err := listS3Objects(ctx, relDir, aws.String("/"), true)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(objects))
	for _, obj := range objects {
		if obj.Key == nil {
			continue
		}
		// Appends the file name to the list
		names = append(names, path.Base(keyToRelPath(*obj.Key)))
	}

	return names, nil
}

// ListAllObjectRelPaths lists every object key under a prefix (recursive).
func ListAllObjectRelPaths(ctx context.Context, relPrefix string) ([]string, error) {
	objects, _, err := listS3Objects(ctx, relPrefix, nil, true)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(objects))
	for _, obj := range objects {
		if obj.Key == nil {
			continue
		}
		// Appends entire path to the list
		paths = append(paths, keyToRelPath(*obj.Key))
	}

	return paths, nil
}

// s3ObjectKey turns a path inside the workflow folder into the S3 object key
// (OLAKE_S3_PREFIX + that path). Example prefix "olake":
//
//	"<hash>/logs/sync_20260825T181000Z" → "olake/<hash>/logs/sync_20260825T181000Z"
//
// Pass isDirectory true when listing a folder. That adds a trailing slash so S3
// only matches keys under that path, not a sibling with the same name prefix:
//
//	"…/sync_20260825T181000Z/"
func s3ObjectKey(relPath string, isDirectory bool) string {
	prefix := strings.Trim(viper.GetString(constants.EnvS3Prefix), "/")
	key := path.Join(prefix, path.Clean(strings.Trim(relPath, "/")))
	if isDirectory {
		return strings.TrimSuffix(key, "/") + "/"
	}
	return key
}

// keyToRelPath strips OLAKE_S3_PREFIX from a bucket key (inverse of s3ObjectKey).
//
//	"olake/<hash>/logs/sync_20260825T181000Z/connector-000001-….log"
//	→ "<hash>/logs/sync_20260825T181000Z/connector-000001-….log"
func keyToRelPath(key string) string {
	key = strings.TrimPrefix(key, "/")
	prefix := strings.Trim(viper.GetString(constants.EnvS3Prefix), "/")
	if prefix != "" {
		prefix = prefix + "/"
		if strings.HasPrefix(key, prefix) {
			return strings.TrimPrefix(key, prefix)
		}
	}
	return key
}
