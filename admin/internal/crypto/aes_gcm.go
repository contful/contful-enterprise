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
	// KeySizeAES AES-256 密钥长度（32 字节）
	KeySizeAES = 32
	// NonceSizeGCM GCM 推荐 IV 长度（12 字节）
	NonceSizeGCM = 12
)

var (
	ErrInvalidKeySizeAES     = errors.New("密钥长度必须是 32 字节")
	ErrCiphertextTooShortAES = errors.New("密文长度不足（至少包含 12 字节 nonce + 16 字节 tag）")
	ErrDecryptionFailedAES   = errors.New("解密失败：密钥不匹配或数据被篡改")
)

// deriveKey 将任意长度的输入转换为 32 字节密钥
func deriveKey(input string) ([]byte, error) {
	key, err := hex.DecodeString(input)
	if err != nil {
		h := sha256.Sum256([]byte(input))
		return h[:], nil
	}
	if len(key) != KeySizeAES {
		return nil, ErrInvalidKeySizeAES
	}
	return key, nil
}

// AESGCM AES-256-GCM 加密器
type AESGCM struct {
	key []byte
}

// NewAESGCM 创建 AES-256-GCM 加密器
func NewAESGCM(secret string) *AESGCM {
	key, _ := deriveKey(secret) // 错误已在上层处理
	return &AESGCM{key: key}
}

// Encrypt 使用 AES-256-GCM 加密原文，返回 hex 编码的密文（nonce || ciphertext || tag）
func (a *AESGCM) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, NonceSizeGCM)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt 解密 hex 编码的密文（nonce || ciphertext || tag），返回原文
func (a *AESGCM) Decrypt(ciphertextHex string) ([]byte, error) {
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < NonceSizeGCM+16 {
		return nil, ErrCiphertextTooShortAES
	}

	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := ciphertext[:NonceSizeGCM]
	ct := ciphertext[NonceSizeGCM:]

	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, ErrDecryptionFailedAES
	}

	return plaintext, nil
}
