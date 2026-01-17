package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const DefaultBCryptCost = 12 // reasonable range: 10-14; tune for your machine/load.

// HashPassword hashes a password with bcrypt (recommended for login passwords).
func HashPassword(plain string) (string, error) {
	return HashPasswordWithCost(plain, DefaultBCryptCost)
}

// HashPasswordWithCost hashes a password with bcrypt using the given cost.
func HashPasswordWithCost(plain string, cost int) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPasswordHash compares a plaintext password with a stored bcrypt hash.
func CheckPasswordHash(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}

func SHA1(input []byte) string {
	sum := sha1.Sum(input)
	return hex.EncodeToString(sum[:])
}

func SHA256(input []byte) string {
	sum := sha256.Sum256(input)
	return hex.EncodeToString(sum[:])
}

func SHA512(input []byte) string {
	sum := sha512.Sum512(input)
	return hex.EncodeToString(sum[:])
}

func MD5(input []byte) string {
	sum := md5.Sum(input)
	return hex.EncodeToString(sum[:])
}
