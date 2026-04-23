// Package crypto 提供 AES-256-GCM 加解密工具，供配置中心和 MFA 等敏感数据加密使用。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
)

const (
	// KeySize AES-256 密钥长度（32 字节）
	KeySize = 32
	// NonceSize GCM 推荐 IV 长度（12 字节）
	NonceSize = 12
)

var (
	ErrInvalidKeySize    = errors.New("密钥长度必须是 32 字节（64 字符 hex 或原始 32 字节）")
	ErrCiphertextTooShort = errors.New("密文长度不足（至少包含 12 字节 nonce + 16 字节 tag）")
	ErrDecryptionFailed   = errors.New("解密失败：密钥不匹配或数据被篡改")
)

// DeriveKey 将任意长度的输入转换为 32 字节 AES-256 密钥。
// 如果输入是 64 字符的 hex 字符串，直接解码为 32 字节。
// 否则使用 SHA-256 哈希生成 32 字节密钥。
func DeriveKey(input string) ([]byte, error) {
	key, err := hex.DecodeString(input)
	if err != nil {
		// 不是 hex，直接哈希生成
		h := sha256.Sum256([]byte(input))
		return h[:], nil
	}
	if len(key) != KeySize {
		return nil, ErrInvalidKeySize
	}
	return key, nil
}

// Encrypt 使用 AES-256-GCM 加密原文，返回 hex 编码的密文（nonce || ciphertext || tag）。
//
//	传入的 key 必须是 32 字节，不足会自动通过 DeriveKey 转换。
func Encrypt(plaintext []byte, keyInput string) (string, error) {
	key, err := DeriveKey(keyInput)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Seal 自动在密文末尾追加 16 字节 auth tag
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt 解密 hex 编码的密文（nonce || ciphertext || tag），返回原文。
//
//	传入的 key 必须是 32 字节，不足会自动通过 DeriveKey 转换。
func Decrypt(ciphertextHex string, keyInput string) ([]byte, error) {
	key, err := DeriveKey(keyInput)
	if err != nil {
		return nil, err
	}

	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < NonceSize+16 {
		return nil, ErrCiphertextTooShort
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := ciphertext[:NonceSize]
	ct := ciphertext[NonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// MustEncrypt 加密，失败时 panic（适用于启动时已知密钥固定的场景）。
func MustEncrypt(plaintext []byte, keyInput string) string {
	out, err := Encrypt(plaintext, keyInput)
	if err != nil {
		panic("crypto: MustEncrypt failed: " + err.Error())
	}
	return out
}

// MustDecrypt 解密，失败时 panic。
func MustDecrypt(ciphertextHex string, keyInput string) []byte {
	out, err := Decrypt(ciphertextHex, keyInput)
	if err != nil {
		panic("crypto: MustDecrypt failed: " + err.Error())
	}
	return out
}
