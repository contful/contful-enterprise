// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package crypto

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"

	"github.com/tjfoc/gmsm/sm4"
)

const (
	// KeySizeSM4 SM4 密钥长度（16 字节）
	KeySizeSM4 = 16
)

var (
	ErrInvalidKeySizeSM4     = errors.New("SM4 密钥长度必须是 16 字节")
	ErrCiphertextTooShortSM4 = errors.New("SM4 密文长度不足")
	ErrDecryptionFailedSM4   = errors.New("SM4 解密失败：密钥不匹配或数据被篡改")
)

// SM4GCM SM4-GCM 加密器（国密算法）
type SM4GCM struct {
	key []byte
}

// NewSM4GCM 创建 SM4-GCM 加密器
func NewSM4GCM(secret string) (*SM4GCM, error) {
	key, err := deriveKeySM4(secret)
	if err != nil {
		return nil, err
	}
	return &SM4GCM{key: key}, nil
}

// deriveKeySM4 将任意长度的输入转换为 16 字节 SM4 密钥
func deriveKeySM4(input string) ([]byte, error) {
	key, err := hex.DecodeString(input)
	if err != nil {
		if len(input) < KeySizeSM4 {
			return nil, ErrInvalidKeySizeSM4
		}
		return []byte(input[:KeySizeSM4]), nil
	}
	if len(key) != KeySizeSM4 {
		return nil, ErrInvalidKeySizeSM4
	}
	return key, nil
}

// Encrypt 使用 SM4-GCM 加密原文，返回 hex 编码的密文
func (s *SM4GCM) Encrypt(plaintext []byte) (string, error) {
	block, err := sm4.NewCipher(s.key)
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

// Decrypt 解密 hex 编码的密文
func (s *SM4GCM) Decrypt(ciphertextHex string) ([]byte, error) {
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < NonceSizeGCM+16 {
		return nil, ErrCiphertextTooShortSM4
	}

	block, err := sm4.NewCipher(s.key)
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
		return nil, ErrDecryptionFailedSM4
	}

	return plaintext, nil
}
