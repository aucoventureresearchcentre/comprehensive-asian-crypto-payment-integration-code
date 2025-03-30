// Security package for Asian Cryptocurrency Payment System
// Provides security features including encryption, authentication, and fraud detection

package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

// Common errors
var (
	ErrInvalidKey       = errors.New("invalid encryption key")
	ErrInvalidData      = errors.New("invalid data for encryption/decryption")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidToken     = errors.New("invalid token")
)

// SecurityService provides security-related functionality
type SecurityService struct {
	encryptionKey []byte
	jwtSecret     []byte
	tokenExpiry   time.Duration
}

// NewSecurityService creates a new security service
func NewSecurityService(encryptionKey, jwtSecret string, tokenExpiry time.Duration) (*SecurityService, error) {
	if len(encryptionKey) < 16 {
		return nil, ErrInvalidKey
	}

	if len(jwtSecret) < 16 {
		return nil, ErrInvalidKey
	}

	// Use SHA-256 to get a fixed-size key from the provided string
	encKey := sha256.Sum256([]byte(encryptionKey))
	jwtKey := sha256.Sum256([]byte(jwtSecret))

	return &SecurityService{
		encryptionKey: encKey[:],
		jwtSecret:     jwtKey[:],
		tokenExpiry:   tokenExpiry,
	}, nil
}

// EncryptData encrypts data using AES-256-GCM
func (s *SecurityService) EncryptData(plaintext string) (string, error) {
	if plaintext == "" {
		return "", ErrInvalidData
	}

	// Create cipher block
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to create nonce: %w", err)
	}

	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	return encoded, nil
}

// DecryptData decrypts data using AES-256-GCM
func (s *SecurityService) DecryptData(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", ErrInvalidData
	}

	// Decode from base64
	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check if ciphertext is valid
	if len(decoded) < gcm.NonceSize() {
		return "", ErrInvalidData
	}

	// Extract nonce and ciphertext
	nonce, ciphertextBytes := decoded[:gcm.NonceSize()], decoded[gcm.NonceSize():]

	// Decrypt data
	plaintextBytes, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %w", err)
	}

	return string(plaintextBytes), nil
}

// HashPassword hashes a password using bcrypt
func (s *SecurityService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrInvalidData
	}

	// Hash password
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against a hash
func (s *SecurityService) VerifyPassword(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}

	// Verify password
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateAPIKey generates a new API key
func (s *SecurityService) GenerateAPIKey() (string, error) {
	// Generate random bytes
	bytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to hex
	return hex.EncodeToString(bytes), nil
}

// GenerateAPISecret generates a new API secret
func (s *SecurityService) GenerateAPISecret() (string, error) {
	// Generate random bytes
	bytes := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to hex
	return hex.EncodeToString(bytes), nil
}

// DeriveKey derives a key from a password and salt using PBKDF2
func (s *SecurityService) DeriveKey(password, salt string, iterations, keyLen int) (string, error) {
	if password == "" || salt == "" {
		return "", ErrInvalidData
	}

	// Derive key
	key := pbkdf2.Key([]byte(password), []byte(salt), iterations, keyLen, sha256.New)

	// Encode to hex
	return hex.EncodeToString(key), nil
}

// GenerateJWT generates a JWT token
func (s *SecurityService) GenerateJWT(claims map[string]interface{}) (string, error) {
	if claims == nil {
		return "", ErrInvalidData
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	tokenClaims := token.Claims.(jwt.MapClaims)
	for key, value := range claims {
		tokenClaims[key] = value
	}

	// Set expiry
	tokenClaims["exp"] = time.Now().Add(s.tokenExpiry).Unix()

	// Sign token
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// VerifyJWT verifies a JWT token
func (s *SecurityService) VerifyJWT(tokenString string) (map[string]interface{}, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Validate token
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Check expiry
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, ErrTokenExpired
		}
	}

	// Convert claims to map
	result := make(map[string]interface{})
	for key, value := range claims {
		result[key] = value
	}

	return result, nil
}

// GenerateSignature generates an HMAC signature for data
func (s *SecurityService) GenerateSignature(data string) (string, error) {
	if data == "" {
		return "", ErrInvalidData
	}

	// Create HMAC
	h := hmac.New(sha256.New, s.jwtSecret)
	h.Write([]byte(data))

	// Get signature
	signature := h.Sum(nil)

	// Encode to hex
	return hex.EncodeToString(signature), nil
}

// VerifySignature verifies an HMAC signature for data
func (s *SecurityService) VerifySignature(data, signature string) bool {
	if data == "" || signature == "" {
		return false
	}

	// Generate expected signature
	expectedSignature, err := s.GenerateSignature(data)
	if err != nil {
		return false
	}

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
