// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

// RSACrypter RSA 非对称加密器（将现有 rsa.go 函数适配为 AsymmetricCrypter 接口）
type RSACrypter struct{}

// NewRSACrypter 创建 RSA 非对称加密器
func NewRSACrypter() *RSACrypter {
	return &RSACrypter{}
}

// AsymEncrypt 使用公钥加密明文
func (c *RSACrypter) AsymEncrypt(pubKeyPEM, plaintext []byte) ([]byte, error) {
	pub, err := ParseRSAPublicKey(string(pubKeyPEM))
	if err != nil {
		return nil, err
	}
	cipherB64, err := RSAEncrypt(pub, plaintext)
	if err != nil {
		return nil, err
	}
	return []byte(cipherB64), nil
}

// AsymDecrypt 使用私钥解密密文
func (c *RSACrypter) AsymDecrypt(privKeyPEM, ciphertext []byte) ([]byte, error) {
	priv, err := ParseRSAPrivateKey(string(privKeyPEM))
	if err != nil {
		return nil, err
	}
	return RSADecrypt(priv, string(ciphertext))
}

// Sign 使用私钥对数据签名（SHA-256 + PKCS#1 v1.5）
func (c *RSACrypter) Sign(privKeyPEM, data []byte) ([]byte, error) {
	priv, err := ParseRSAPrivateKey(string(privKeyPEM))
	if err != nil {
		return nil, err
	}

	hashed := sha256.Sum256(data)
	return rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed[:])
}

// Verify 使用公钥验签（SHA-256 + PKCS#1 v1.5）
func (c *RSACrypter) Verify(pubKeyPEM, data, sig []byte) (bool, error) {
	pub, err := ParseRSAPublicKey(string(pubKeyPEM))
	if err != nil {
		return false, err
	}

	hashed := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], sig)
	return err == nil, nil
}

// GenerateKeyPair 生成 2048 位 RSA 密钥对
func (c *RSACrypter) GenerateKeyPair() (pubPEM, privPEM string, err error) {
	return GenerateRSAKeyPair()
}
