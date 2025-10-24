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
	"github.com/datazip/olake-ui/server/internal/constants"
)

// utility provides encryption and decryption functionality using either AWS KMS or local AES-256-GCM.
//
// Configuration:
// - Set OLAKE_SECRET_KEY environment variable to enable encryption
// - For AWS KMS: Set OLAKE_SECRET_KEY to a KMS ARN (e.g., "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012")
// - For local AES: Set OLAKE_SECRET_KEY to any non-empty string (will be hashed to 256-bit key)
// - For no encryption: Leave OLAKE_SECRET_KEY empty (not recommended for production)

// getSecretKey returns the encryption key, KMS client (if using KMS), and error
func getSecretKey() ([]byte, *kms.Client, error) {
	// TODO: can we move this to constants and set key and kms client
	// TODO: use viper package to read environment variables
	envKey := os.Getenv(constants.EncryptionKey)
	if strings.TrimSpace(envKey) == "" {
		return []byte{}, nil, nil // Encryption is disabled
	}

	if strings.HasPrefix(envKey, "arn:aws:kms:") {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load AWS config: %s", err)
		}
		return []byte(envKey), kms.NewFromConfig(cfg), nil
	}

	// Local AES-GCM Mode with SHA-256 derived key
	hash := sha256.Sum256([]byte(envKey))
	return hash[:], nil, nil
}

func Encrypt(plaintext string) (string, error) {
	if strings.TrimSpace(plaintext) == "" {
		return plaintext, nil
	}

	key, kmsClient, err := getSecretKey()
	if err != nil || key == nil || len(key) == 0 {
		return plaintext, err
	}

	// Use KMS if client is provided
	if kmsClient != nil {
		keyID := string(key)
		result, err := kmsClient.Encrypt(context.Background(), &kms.EncryptInput{
			KeyId:     &keyID,
			Plaintext: []byte(plaintext),
		})
		if err != nil {
			return "", fmt.Errorf("failed to encrypt with KMS: %s", err)
		}
		return fmt.Sprintf("%q", base64.StdEncoding.EncodeToString(result.CiphertextBlob)), nil
	}

	// Local AES-GCM encryption
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %s", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %s", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %s", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return fmt.Sprintf("%q", base64.StdEncoding.EncodeToString(ciphertext)), nil
}

func Decrypt(encryptedText string) (string, error) {
	if strings.TrimSpace(encryptedText) == "" {
		return "", fmt.Errorf("cannot decrypt empty or whitespace-only input")
	}

	key, kmsClient, err := getSecretKey()
	if err != nil || key == nil || len(key) == 0 {
		return encryptedText, err
	}

	var config string
	err = json.Unmarshal([]byte(encryptedText), &config)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON string: %s", err)
	}

	encryptedData, err := base64.StdEncoding.DecodeString(config)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %s", err)
	}

	// Use KMS if client is provided
	if kmsClient != nil {
		result, err := kmsClient.Decrypt(context.Background(), &kms.DecryptInput{
			CiphertextBlob: encryptedData,
		})
		if err != nil {
			return "", fmt.Errorf("failed to decrypt with KMS: %s", err)
		}
		return string(result.Plaintext), nil
	}

	// Local AES-GCM decryption
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %s", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %s", err)
	}

	if len(encryptedData) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	plaintext, err := gcm.Open(nil, encryptedData[:gcm.NonceSize()], encryptedData[gcm.NonceSize():], nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %s", err)
	}
	return string(plaintext), nil
}
