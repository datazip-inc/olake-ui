package utils

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

// this file provides encryption and decryption functionality using either AWS KMS or local AES-256-GCM.
//
// Configuration:
// - Set ENCRYPTION_KEY environment variable to enable encryption
// - For AWS KMS: Set ENCRYPTION_KEY to a KMS ARN (e.g., "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012")
// - For local AES: Set ENCRYPTION_KEY to any non-empty string (will be hashed to 256-bit key)
// - For no encryption: Leave ENCRYPTION_KEY empty (not recommended for production)

// EncryptionConfig holds the configuration for encryption operations
type EncryptionConfig struct {
	KMSClient *kms.Client
	KeyID     string
	LocalKey  []byte
	UseKMS    bool
	Disabled  bool
}

func getEncryptionConfig() (*EncryptionConfig, error) {
	key := os.Getenv("ENCRYPTION_KEY")

	if strings.TrimSpace(key) == "" {
		return &EncryptionConfig{
			Disabled: true,
		}, nil
	}

	if strings.HasPrefix(key, "arn:aws:kms:") {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}
		return &EncryptionConfig{
			KMSClient: kms.NewFromConfig(cfg),
			KeyID:     key,
			UseKMS:    true,
		}, nil
	}

	// Local AES-GCM Mode with SHA-256 derived key
	hash := sha256.Sum256([]byte(key))
	return &EncryptionConfig{
		LocalKey: hash[:],
	}, nil
}

func Encrypt(plaintext string) ([]byte, error) {
	config, err := getEncryptionConfig()
	if err != nil {
		return nil, err
	}

	if config.Disabled {
		return []byte(plaintext), nil
	}

	if config.UseKMS {
		result, err := config.KMSClient.Encrypt(context.Background(), &kms.EncryptInput{
			KeyId:     &config.KeyID,
			Plaintext: []byte(plaintext),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt with KMS: %w", err)
		}
		return result.CiphertextBlob, nil
	}

	// Local AES-GCM encryption
	block, err := aes.NewCipher(config.LocalKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertext, nil
}

func Decrypt(cipherData []byte) (string, error) {
	config, err := getEncryptionConfig()
	if err != nil {
		return "", err
	}

	if config.Disabled {
		return string(cipherData), nil
	}

	var plaintext []byte

	if config.UseKMS {
		result, err := config.KMSClient.Decrypt(context.Background(), &kms.DecryptInput{
			CiphertextBlob: cipherData,
		})
		if err != nil {
			return "", fmt.Errorf("failed to decrypt with KMS: %w", err)
		}
		plaintext = result.Plaintext
	} else {
		// Local AES-GCM decryption
		block, err := aes.NewCipher(config.LocalKey)
		if err != nil {
			return "", fmt.Errorf("failed to create cipher: %w", err)
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return "", fmt.Errorf("failed to create GCM: %w", err)
		}

		if len(cipherData) < gcm.NonceSize() {
			return "", errors.New("ciphertext too short")
		}

		plaintext, err = gcm.Open(nil, cipherData[:gcm.NonceSize()], cipherData[gcm.NonceSize():], nil)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt: %w", err)
		}
	}

	return string(plaintext), nil
}

// EncryptConfig encrypts data and returns base64 encoded string for direct DB storage
func EncryptConfig(config string) (string, error) {
	encryptedBytes, err := Encrypt(config)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %v", err)
	}
	return `"` + base64.URLEncoding.EncodeToString(encryptedBytes) + `"`, nil
}

// DecryptConfig decrypts base64 encoded encrypted data
func DecryptConfig(encryptedConfig string) (string, error) {
	// Use json.Unmarshal to properly handle JSON string unquoting
	var unquotedString string
	if err := json.Unmarshal([]byte(encryptedConfig), &unquotedString); err != nil {
		// If unmarshal fails, assume it's already unquoted
		unquotedString = encryptedConfig
	}

	encryptedData, err := base64.URLEncoding.DecodeString(unquotedString)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %v", err)
	}

	decrypted, err := Decrypt(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %v", err)
	}

	return decrypted, nil
}
