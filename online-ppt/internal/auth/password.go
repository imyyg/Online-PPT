package auth

import (
	cryptoRand "crypto/rand"
	cryptoSubtle "crypto/subtle"
	encodingBase64 "encoding/base64"
	fmt "fmt"
	strings "strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonMemory      = 64 * 1024
	argonIterations  = 3
	argonParallelism = 2
	saltLength       = 16
	hashLength       = 32
)

// HashPassword returns an encoded Argon2id password hash.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	salt := make([]byte, saltLength)
	if _, err := cryptoRand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, hashLength)

	return fmt.Sprintf("argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory,
		argonIterations,
		argonParallelism,
		encodingBase64.RawStdEncoding.EncodeToString(salt),
		encodingBase64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// VerifyPassword compares a password against its encoded hash.
func VerifyPassword(encodedHash, password string) (bool, error) {
	if encodedHash == "" {
		return false, fmt.Errorf("hash cannot be empty")
	}
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 5 {
		return false, fmt.Errorf("invalid hash format")
	}

	paramsPart := parts[2]
	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(paramsPart, "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, fmt.Errorf("parse params: %w", err)
	}

	salt, err := encodingBase64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}

	expected, err := encodingBase64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}

	recomputed := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expected)))
	if cryptoSubtle.ConstantTimeCompare(expected, recomputed) == 1 {
		return true, nil
	}
	return false, nil
}
