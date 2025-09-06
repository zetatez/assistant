package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 哈希密码(推荐用于用户登录密码)
func HashPassword(plain string) (string, error) {
	const cost = 12 // 10-14 合理值，根据你的机器/负载调整
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPasswordHash 比较明文密码与存储的 hash
func CheckPasswordHash(hash, plain string) error {
	// bcrypt.CompareFromPassword 在成功时返回 nil
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}

func Sha1(input string) string {
	h := sha1.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

func Sha256(input string) string {
	h := sha256.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

func Sha512(input string) string {
	h := sha512.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

func MD5(input string) string {
	h := md5.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}
