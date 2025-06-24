package kms

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

var (
	kmsClient *kms.Client
	keyId     string
	once      sync.Once
)

func initKMS() {
	once.Do(func() {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			panic(fmt.Sprintf("Unable to load AWS config: %v", err))
		}
		kmsClient = kms.NewFromConfig(cfg)

		keyId = "arn:aws:kms:ap-south-1:672919669757:key/bade09eb-5f0c-45bc-b404-fb2e623dc1b1"
		if keyId == "" {
			panic("KMS_KEY_ID not set in environment variables")
		}
	})
}

func Encrypt(text string) ([]byte, error) {
	initKMS()
	out, err := kmsClient.Encrypt(context.Background(), &kms.EncryptInput{
		KeyId:     &keyId,
		Plaintext: []byte(text),
	})
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}
	return out.CiphertextBlob, nil
}

func Decrypt(cipher []byte) (string, error) {
	initKMS()
	out, err := kmsClient.Decrypt(context.Background(), &kms.DecryptInput{
		CiphertextBlob: cipher,
	})
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}
	return string(out.Plaintext), nil
}
