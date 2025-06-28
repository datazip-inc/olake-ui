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

// getSecretKey returns the encryption key, KMS client (if using KMS), and any error
// If the key is empty, encryption is considered disabled
// key: encryption/decryption key (local key or KMS key ID)
// kmsClient: non-nil if using AWS KMS
// error: any error that occurred during initialization
func getSecretKey() (key []byte, kmsClient *kms.Client, err error) {
	envKey := os.Getenv("ENCRYPTION_KEY")
	if strings.TrimSpace(envKey) == "" {
		return nil, nil, nil // Encryption is disabled
	}

	if strings.HasPrefix(envKey, "arn:aws:kms:") {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load AWS config: %w", err)
		}
		return []byte(envKey), kms.NewFromConfig(cfg), nil
	}

	// Local AES-GCM Mode with SHA-256 derived key
	hash := sha256.Sum256([]byte(envKey))
	return hash[:], nil, nil
}

func Encrypt(plaintext string) ([]byte, error) {
	key, kmsClient, err := getSecretKey()
	if err != nil {
		return nil, err
	}

	// If key is empty, encryption is disabled
	if key == nil {
		return []byte(plaintext), nil
	}

	// Use KMS if client is provided
	if kmsClient != nil {
		keyID := string(key)
		result, err := kmsClient.Encrypt(context.Background(), &kms.EncryptInput{
			KeyId:     &keyID,
			Plaintext: []byte(plaintext),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt with KMS: %w", err)
		}
		return result.CiphertextBlob, nil
	}

	// Local AES-GCM encryption
	block, err := aes.NewCipher(key)
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
	key, kmsClient, err := getSecretKey()
	if err != nil {
		return "", err
	}

	// If key is empty, decryption is disabled
	if key == nil {
		return string(cipherData), nil
	}

	var plaintext []byte

	// Use KMS if client is provided
	if kmsClient != nil {
		result, err := kmsClient.Decrypt(context.Background(), &kms.DecryptInput{
			CiphertextBlob: cipherData,
		})
		if err != nil {
			return "", fmt.Errorf("failed to decrypt with KMS: %w", err)
		}
		plaintext = result.Plaintext
	} else {
		// Local AES-GCM decryption
		block, err := aes.NewCipher(key)
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
	if config == "" {
		return "", nil
	}
	encryptedBytes, err := Encrypt(config)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %v", err)
	}
	//TODO: check if this is needed
	return `"` + base64.URLEncoding.EncodeToString(encryptedBytes) + `"`, nil
}

// DecryptConfig decrypts base64 encoded encrypted data
func DecryptConfig(encryptedConfig string) (string, error) {
	// Check for empty or whitespace-only input
	if strings.TrimSpace(encryptedConfig) == "" {
		return "", fmt.Errorf("cannot decrypt empty or whitespace-only input")
	}

	// Try to unquote JSON string, if it fails use the input as is
	var unquotedString string
	err := json.Unmarshal([]byte(encryptedConfig), &unquotedString)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON string: %v", err)
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
