// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package crypto

import "errors"

// Crypter 对称加解密接口
type Crypter interface {
	// Encrypt 加密原文，返回 hex 编码的密文（nonce || ciphertext || tag）
	Encrypt(plaintext []byte) (string, error)
	// Decrypt 解密 hex 编码的密文，返回原文
	Decrypt(ciphertextHex string) ([]byte, error)
}

// AsymmetricCrypter 非对称加解密接口（RSA / SM2）
// 注意：方法名用 AsymEncrypt/AsymDecrypt 避免与 Crypter 的 Encrypt/Decrypt 冲突
type AsymmetricCrypter interface {
	// AsymEncrypt 公钥加密
	AsymEncrypt(pubKeyPEM, plaintext []byte) ([]byte, error)
	// AsymDecrypt 私钥解密
	AsymDecrypt(privKeyPEM, ciphertext []byte) ([]byte, error)
	// Sign 私钥签名
	Sign(privKeyPEM, data []byte) ([]byte, error)
	// Verify 公钥验签
	Verify(pubKeyPEM, data, sig []byte) (bool, error)
	// GenerateKeyPair 生成密钥对，返回 PEM 编码的公钥和私钥
	GenerateKeyPair() (pubPEM, privPEM string, err error)
}

// Hasher 哈希接口（SHA-256 / SM3）
type Hasher interface {
	// Sum 计算哈希值，返回字节数组
	Sum(data []byte) []byte
	// HMAC 计算 HMAC 值
	HMAC(key, data []byte) []byte
}

// CryptoProvider 统一加密服务接口（对称 + 非对称 + 哈希一站式）
type CryptoProvider interface {
	Crypter              // 对称加解密（AES-GCM / SM4-GCM）
	AsymmetricCrypter    // 非对称加解密/签名（RSA / SM2）
	Hasher               // 哈希/HMAC（SHA-256 / SM3）
}

const (
	AlgorithmAES = "aes-256-gcm"
	AlgorithmSM4 = "sm4-gcm"
)

// CryptoMode 加密模式
const (
	ModeRSA = "rsa" // 国际算法：RSA + SHA-256 + AES-256-GCM（默认）
	ModeSM  = "sm"  // 国密算法：SM2 + SM3 + SM4-GCM
)

// NewCrypter 根据算法名称创建对称加密器
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

// NewAsymmetricCrypter 根据模式创建非对称加密器
func NewAsymmetricCrypter(mode string) (AsymmetricCrypter, error) {
	switch mode {
	case ModeSM:
		return NewSM2Crypter(), nil
	default:
		return NewRSACrypter(), nil
	}
}

// NewHasher 根据模式创建哈希器
func NewHasher(mode string) Hasher {
	if mode == ModeSM {
		return &SM3Hasher{}
	}
	return &SHA256Hasher{}
}

// NewProvider 根据模式创建完整加密 Provider
func NewProvider(mode, secret string) (CryptoProvider, error) {
	asym, err := NewAsymmetricCrypter(mode)
	if err != nil {
		return nil, err
	}

	sym, err := NewCrypter(symAlgorithm(mode), secret)
	if err != nil {
		return nil, err
	}

	h := NewHasher(mode)

	return &provider{
		Crypter:           sym,
		AsymmetricCrypter: asym,
		Hasher:            h,
	}, nil
}

// symAlgorithm 根据模式返回对称加密算法
func symAlgorithm(mode string) string {
	if mode == ModeSM {
		return AlgorithmSM4
	}
	return AlgorithmAES
}

// provider 内部实现
type provider struct {
	Crypter
	AsymmetricCrypter
	Hasher
}
