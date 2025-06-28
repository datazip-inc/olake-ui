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

func getEncryptionConfig() (kmsClient *kms.Client, keyID string, localKey []byte, useKMS, disabled bool, err error) {
	key := os.Getenv("ENCRYPTION_KEY")

	if strings.TrimSpace(key) == "" {
		return nil, "", nil, false, true, nil
	}

	if strings.HasPrefix(key, "arn:aws:kms:") {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, "", nil, false, false, fmt.Errorf("failed to load AWS config: %w", err)
		}
		client := kms.NewFromConfig(cfg)
		return client, key, nil, true, false, nil
	}

	// Local AES-GCM Mode with SHA-256 derived key
	hash := sha256.Sum256([]byte(key))
	return nil, "", hash[:], false, false, nil
}

func Encrypt(plaintext string) ([]byte, error) {
	kmsClient, keyID, localKey, useKMS, disabled, err := getEncryptionConfig()
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	if disabled {
		return []byte(plaintext), nil
	}

	if useKMS {
		out, err := kmsClient.Encrypt(context.Background(), &kms.EncryptInput{
			KeyId:     &keyID,
			Plaintext: []byte(plaintext),
		})
		if err != nil {
			return nil, fmt.Errorf("KMS encryption failed: %w", err)
		}
		return out.CiphertextBlob, nil
	}

	block, err := aes.NewCipher(localKey)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertext, nil
}

func Decrypt(cipherData []byte) (string, error) {
	kmsClient, _, localKey, useKMS, disabled, err := getEncryptionConfig()
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	if disabled {
		return string(cipherData), nil
	}

	if useKMS {
		out, err := kmsClient.Decrypt(context.Background(), &kms.DecryptInput{
			CiphertextBlob: cipherData,
		})
		if err != nil {
			return "", fmt.Errorf("decryption failed: %w", err)
		}
		return string(out.Plaintext), nil
	}

	block, err := aes.NewCipher(localKey)
	if err != nil {
		return "", err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aead.NonceSize()
	if len(cipherData) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := cipherData[:nonceSize], cipherData[nonceSize:]

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
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
