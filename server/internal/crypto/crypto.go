package crypto

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
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

var (
	kmsClient          *kms.Client
	keyID              string
	localKey           []byte
	useKMS             bool
	once               sync.Once
	encryptionDisabled bool
)

// Package crypto provides encryption and decryption functionality using either AWS KMS or local AES-256-GCM.
//
// Configuration:
// - Set ENCRYPTION_KEY environment variable to enable encryption
// - For AWS KMS: Set ENCRYPTION_KEY to a KMS ARN (e.g., "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012")
// - For local AES: Set ENCRYPTION_KEY to any non-empty string (will be hashed to 256-bit key)
// - For no encryption: Leave ENCRYPTION_KEY empty (not recommended for production)
//
// Data Format:
// - Encrypted data is stored as JSON: {"encrypted_data": "base64-encoded-encrypted-data"}
// - Supports backward compatibility with unencrypted JSON data
// InitEncryption initializes encryption based on KMS key or passphrase
func InitEncryption() error {
	key := os.Getenv("ENCRYPTION_KEY")
	var initErr error

	once.Do(func() {
		if strings.TrimSpace(key) == "" {
			encryptionDisabled = true
			return
		}
		if strings.HasPrefix(key, "arn:aws:kms:") {
			cfg, err := config.LoadDefaultConfig(context.Background())
			if err != nil {
				initErr = fmt.Errorf("failed to load AWS config: %w", err)
				return
			}
			kmsClient = kms.NewFromConfig(cfg)
			keyID = key
			useKMS = true
		} else {
			// Local AES-GCM Mode with SHA-256 derived key
			hash := sha256.Sum256([]byte(key))
			localKey = hash[:]
			useKMS = false
		}
	})

	return initErr
}

func Encrypt(plaintext string) ([]byte, error) {
	if err := InitEncryption(); err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}
	if encryptionDisabled {
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
	if err := InitEncryption(); err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}
	if encryptionDisabled {
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

type cryptoObj struct {
	EncryptedData string `json:"encrypted_data"`
}

// EncryptJSONString encrypts the entire JSON string as a single value
func EncryptJSONString(rawConfig string) (string, error) {
	// Encrypt the entire config string
	encryptedBytes, err := Encrypt(rawConfig)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %v", err)
	}
	cryptoObj := cryptoObj{
		EncryptedData: base64.StdEncoding.EncodeToString(encryptedBytes),
	}
	// Marshal to JSON
	encryptedJSON, err := json.Marshal(cryptoObj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal encrypted data: %v", err)
	}

	return string(encryptedJSON), nil
}

// DecryptJSONObject decrypts a JSON object in the format {"encrypted_data": "base64-encoded-encrypted-json"}
// and returns the original JSON string
func DecryptJSONString(encryptedObjStr string) (string, error) {
	// Unmarshal the encrypted object
	cryptoObj := cryptoObj{}
	if err := json.Unmarshal([]byte(encryptedObjStr), &cryptoObj); err != nil {
		return "", fmt.Errorf("failed to unmarshal encrypted data: %v", err)
	}
	// Decode the base64-encoded encrypted data
	encryptedData, err := base64.StdEncoding.DecodeString(cryptoObj.EncryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %v", err)
	}
	// Decrypt the data
	decrypted, err := Decrypt(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %v", err)
	}
	return string(decrypted), nil
}
