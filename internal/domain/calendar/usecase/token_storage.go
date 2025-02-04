package usecase

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/oauth2"
)

// EncryptedTokenStorage implements TokenManager interface with encryption
type EncryptedTokenStorage struct {
	// In production, use a secure key management service
	encryptionKey []byte
	// In production, use a proper database
	tokenStore map[string]string
}

// NewEncryptedTokenStorage creates a new instance of EncryptedTokenStorage
func NewEncryptedTokenStorage(key []byte) (*EncryptedTokenStorage, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes")
	}

	return &EncryptedTokenStorage{
		encryptionKey: key,
		tokenStore:    make(map[string]string),
	}, nil
}

func (s *EncryptedTokenStorage) GetToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	encryptedToken, exists := s.tokenStore[userID]
	if !exists {
		return nil, fmt.Errorf("no token found for user %s", userID)
	}

	// Decrypt token
	tokenBytes, err := s.decrypt(encryptedToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %v", err)
	}

	// Unmarshal token
	var token oauth2.Token
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %v", err)
	}

	return &token, nil
}

func (s *EncryptedTokenStorage) SaveToken(ctx context.Context, userID string, token *oauth2.Token) error {
	// Marshal token
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %v", err)
	}

	// Encrypt token
	encryptedToken, err := s.encrypt(tokenBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt token: %v", err)
	}

	// Store encrypted token
	s.tokenStore[userID] = encryptedToken
	return nil
}

func (s *EncryptedTokenStorage) DeleteToken(ctx context.Context, userID string) error {
	delete(s.tokenStore, userID)
	return nil
}

func (s *EncryptedTokenStorage) encrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *EncryptedTokenStorage) decrypt(encryptedData string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
