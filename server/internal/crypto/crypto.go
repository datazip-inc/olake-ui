package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
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
	kmsClient *kms.Client
	keyId     string
	localKey  []byte
	useKMS    bool
	once      sync.Once
)

// InitEncryption initializes encryption based on KMS key or passphrase
func InitEncryption() error {
	key := os.Getenv("ENCRYPTION_KEY")
	var initErr error

	once.Do(func() {
		if strings.HasPrefix(key, "arn:aws:kms:") {
			cfg, err := config.LoadDefaultConfig(context.Background())
			if err != nil {
				initErr = fmt.Errorf("failed to load AWS config: %w", err)
				return
			}
			kmsClient = kms.NewFromConfig(cfg)
			keyId = key
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
	if useKMS {
		out, err := kmsClient.Encrypt(context.Background(), &kms.EncryptInput{
			KeyId:     &keyId,
			Plaintext: []byte(plaintext),
		})
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
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
