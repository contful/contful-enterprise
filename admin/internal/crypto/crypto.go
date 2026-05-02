// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package crypto

import "errors"

// Crypter 加解密接口
type Crypter interface {
	// Encrypt 加密原文，返回 hex 编码的密文（nonce || ciphertext || tag）
	Encrypt(plaintext []byte) (string, error)
	// Decrypt 解密 hex 编码的密文，返回原文
	Decrypt(ciphertextHex string) ([]byte, error)
}

// 算法名称常量
const (
	AlgorithmAES = "aes-256-gcm"
	AlgorithmSM4  = "sm4-gcm"
)

// NewCrypter 根据算法名称创建加密器
func NewCrypter(algorithm, secret string) (Crypter, error) {
	switch algorithm {
	case AlgorithmAES:
		return NewAESGCM(secret), nil
	case AlgorithmSM4:
		return NewSM4GCM(secret)
	default:
		return nil, errors.New("unsupported algorithm: " + algorithm)
	}
}
