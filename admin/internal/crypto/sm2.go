// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package crypto

import (
	"crypto/rand"
	"encoding/asn1"
	"math/big"

	gmsmX509 "github.com/tjfoc/gmsm/x509"

	"github.com/tjfoc/gmsm/sm2"
)

// SM2Crypter SM2 椭圆曲线非对称加密器（国密）
type SM2Crypter struct{}

// NewSM2Crypter 创建 SM2 非对称加密器
func NewSM2Crypter() *SM2Crypter {
	return &SM2Crypter{}
}

// sm2Signature ASN.1 编码的 SM2 签名结构
type sm2RawSignature struct {
	R, S *big.Int
}

// AsymEncrypt 使用公钥加密明文，返回 C1C3C2 格式密文
func (c *SM2Crypter) AsymEncrypt(pubKeyPEM, plaintext []byte) ([]byte, error) {
	pub, err := gmsmX509.ReadPublicKeyFromPem(pubKeyPEM)
	if err != nil {
		return nil, err
	}
	return sm2.Encrypt(pub, plaintext, rand.Reader, sm2.C1C3C2)
}

// AsymDecrypt 使用私钥解密 C1C3C2 格式密文，返回明文
func (c *SM2Crypter) AsymDecrypt(privKeyPEM, ciphertext []byte) ([]byte, error) {
	priv, err := gmsmX509.ReadPrivateKeyFromPem(privKeyPEM, nil)
	if err != nil {
		return nil, err
	}
	return sm2.Decrypt(priv, ciphertext, sm2.C1C3C2)
}

// Sign 使用私钥对数据进行 SM2 签名，返回 ASN.1 DER 编码签名
func (c *SM2Crypter) Sign(privKeyPEM, data []byte) ([]byte, error) {
	priv, err := gmsmX509.ReadPrivateKeyFromPem(privKeyPEM, nil)
	if err != nil {
		return nil, err
	}
	r, s, err := sm2.Sm2Sign(priv, data, nil, rand.Reader)
	if err != nil {
		return nil, err
	}
	return asn1.Marshal(sm2RawSignature{R: r, S: s})
}

// Verify 使用公钥验签
func (c *SM2Crypter) Verify(pubKeyPEM, data, sig []byte) (bool, error) {
	pub, err := gmsmX509.ReadPublicKeyFromPem(pubKeyPEM)
	if err != nil {
		return false, err
	}

	var sm2Sig sm2RawSignature
	_, err = asn1.Unmarshal(sig, &sm2Sig)
	if err != nil {
		return false, err
	}

	return sm2.Sm2Verify(pub, data, nil, sm2Sig.R, sm2Sig.S), nil
}

// GenerateKeyPair 生成 SM2 密钥对，返回 PEM 编码字符串
func (c *SM2Crypter) GenerateKeyPair() (pubPEM, privPEM string, err error) {
	priv, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}

	privPEMBytes, err := gmsmX509.WritePrivateKeyToPem(priv, nil)
	if err != nil {
		return "", "", err
	}

	pubPEMBytes, err := gmsmX509.WritePublicKeyToPem(&priv.PublicKey)
	if err != nil {
		return "", "", err
	}

	return string(pubPEMBytes), string(privPEMBytes), nil
}
