package secretref

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/smithy-go"
)

// AWSConfig configures the aws-sm provider. Credentials come from the default chain (EKS Pod
// Identity, IRSA, instance profile, env) unless both static keys are set.
type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

// awsProvider reads AWS Secrets Manager. path is a secret name or a full secret ARN; an ARN's
// region overrides the configured one.
type awsProvider struct {
	client *secretsmanager.Client
	region string
}

func newAWSProvider(ctx context.Context, cfg *Config) (Provider, error) {
	var opts []func(*config.LoadOptions) error
	if cfg.AWS.Region != "" {
		opts = append(opts, config.WithRegion(cfg.AWS.Region))
	}
	if cfg.AWS.AccessKeyID != "" && cfg.AWS.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AWS.AccessKeyID, cfg.AWS.SecretAccessKey, ""),
		))
	}
	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}
	return &awsProvider{
		client: secretsmanager.NewFromConfig(awsCfg),
		region: awsCfg.Region,
	}, nil
}

func (p *awsProvider) Fetch(ctx context.Context, path string) (string, error) {
	region := awsSecretRegion(path)
	if region == "" {
		region = p.region
	}
	if region == "" {
		return "", unresolvable("AWS region is not set: set OLAKE_GITOPS_SECRETS_AWS_REGION or reference the secret by ARN")
	}

	out, err := p.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: aws.String(path)},
		func(o *secretsmanager.Options) { o.Region = region })
	if err != nil {
		return "", classifyAWSError(err)
	}
	return aws.ToString(out.SecretString), nil
}

// awsSecretRegion returns the region of a secret ARN, or "" for a secret name.
func awsSecretRegion(path string) string {
	parsed, err := arn.Parse(path)
	if err != nil {
		return ""
	}
	return parsed.Region
}

// classifyAWSError marks a 4xx response as unresolvable (not found, access denied, KMS decryption
// failure, deleted secret), except throttling. Everything else is retried: 5xx, and network
// failures such as DNS errors, which the SDK reports with status 0.
func classifyAWSError(err error) error {
	var respErr *awshttp.ResponseError
	if !errors.As(err, &respErr) {
		return err
	}
	status := respErr.HTTPStatusCode()
	if status < 400 || status >= 500 || status == 429 {
		return err
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		if _, throttled := retry.DefaultThrottleErrorCodes[apiErr.ErrorCode()]; throttled {
			return err
		}
	}
	return fmt.Errorf("%w: %w", ErrUnresolvable, err)
}
