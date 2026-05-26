// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package crypto

import (
	"encoding/hex"
	"strings"
	"testing"
)

// TestRSAFullChain 验证 crypto_mode=rsa 下完整加密链路
// 模拟前端登录流程：获取公钥 → RSA 加密密码 → 私钥解密 → HMAC-SHA256 数据签名
func TestRSAFullChain(t *testing.T) {
	// ── 阶段 1：初始化 Provider ──
	secret := "ef3594a7274376527e4a31fc99341a27e372e7e7f3295c796684ef6a0149afea"
	provider, err := NewProvider(ModeRSA, secret)
	if err != nil {
		t.Fatalf("NewProvider(rsa) 失败: %v", err)
	}

	// 验证 Provider 实现了三个接口
	if _, ok := provider.(Crypter); !ok {
		t.Error("Provider 应实现 Crypter 接口")
	}
	if _, ok := provider.(AsymmetricCrypter); !ok {
		t.Error("Provider 应实现 AsymmetricCrypter 接口")
	}
	if _, ok := provider.(Hasher); !ok {
		t.Error("Provider 应实现 Hasher 接口")
	}

	// ── 阶段 2：生成 RSA 密钥对 ──
	pubPEM, privPEM, err := provider.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair 失败: %v", err)
	}
	if !strings.Contains(pubPEM, "BEGIN PUBLIC KEY") {
		t.Error("公钥 PEM 格式错误")
	}
	if !strings.Contains(privPEM, "BEGIN PRIVATE KEY") {
		t.Error("私钥 PEM 格式错误（期望 PKCS#8 PRIVATE KEY）")
	}

	// ── 阶段 3：模拟前端登录密码加密 ──
	password := "contful@com"
	token := "test-token-value"
	plainPassword := password + "@@" + token // 前端拼接格式：password@@token

	cipherB64, err := provider.AsymEncrypt([]byte(pubPEM), []byte(plainPassword))
	if err != nil {
		t.Fatalf("AsymEncrypt 失败: %v", err)
	}
	if len(cipherB64) == 0 {
		t.Error("密文不应为空")
	}

	// ── 阶段 4：模拟后端解密 ──
	plaintext, err := provider.AsymDecrypt([]byte(privPEM), cipherB64)
	if err != nil {
		t.Fatalf("AsymDecrypt 失败: %v", err)
	}

	// 验证解密结果格式：password@@token
	parts := strings.SplitN(string(plaintext), "@@", 2)
	if len(parts) != 2 {
		t.Fatalf("解密结果格式错误，期望 password@@token，实际: %s", string(plaintext))
	}
	if parts[0] != password {
		t.Errorf("解密后密码不匹配: 期望 %q, 实际 %q", password, parts[0])
	}
	if parts[1] != token {
		t.Errorf("解密后 token 不匹配: 期望 %q, 实际 %q", token, parts[1])
	}

	// ── 阶段 5：RSA 签名/验签 ──
	data := []byte("test data for signing")
	sig, err := provider.Sign([]byte(privPEM), data)
	if err != nil {
		t.Fatalf("Sign 失败: %v", err)
	}
	valid, err := provider.Verify([]byte(pubPEM), data, sig)
	if err != nil {
		t.Fatalf("Verify 失败: %v", err)
	}
	if !valid {
		t.Error("RSA 签名验证失败")
	}

	// ── 阶段 6：SHA-256 哈希 ──
	testData := []byte("hello contful")
	hash := provider.Sum(testData)
	if len(hash) != 32 {
		t.Errorf("SHA-256 哈希长度错误: 期望 32, 实际 %d", len(hash))
	}

	// ── 阶段 7：HMAC-SHA256 数据签名（审计日志用） ──
	signingKey, _ := hex.DecodeString("aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899")
	hmacSig := provider.HMAC(signingKey, []byte("action=login&level=info"))
	if len(hmacSig) != 32 {
		t.Errorf("HMAC-SHA256 签名长度错误: 期望 32, 实际 %d", len(hmacSig))
	}

	// ── 阶段 8：对称加密 AES-256-GCM ──
	plaintext2 := []byte("sensitive data to encrypt")
	cipherHex, err := provider.Encrypt(plaintext2)
	if err != nil {
		t.Fatalf("AES Encrypt 失败: %v", err)
	}
	decrypted, err := provider.Decrypt(cipherHex)
	if err != nil {
		t.Fatalf("AES Decrypt 失败: %v", err)
	}
	if string(decrypted) != string(plaintext2) {
		t.Errorf("AES 加解密不匹配: 期望 %q, 实际 %q", plaintext2, decrypted)
	}

	t.Log("✅ crypto_mode=rsa 全链路验证通过：RSA密钥生成/加解密/签名 + SHA-256哈希/HMAC + AES-256-GCM对称加密")
}

// TestNewProviderModeRSA 验证 NewProvider 在 rsa 模式下返回正确的类型
func TestNewProviderModeRSA(t *testing.T) {
	provider, err := NewProvider(ModeRSA, "test-secret-32-bytes-long-key!!")
	if err != nil {
		t.Fatalf("NewProvider(rsa) 失败: %v", err)
	}

	// 类型断言验证
	if _, ok := provider.(CryptoProvider); !ok {
		t.Error("返回值应实现 CryptoProvider 接口")
	}
}

// TestNewHasher 验证 Hasher 工厂函数
func TestNewHasher(t *testing.T) {
	// rsa 模式应返回 SHA256Hasher
	h := NewHasher(ModeRSA)
	if _, ok := h.(*SHA256Hasher); !ok {
		t.Error("NewHasher(rsa) 应返回 *SHA256Hasher")
	}

	// sm 模式应返回 SM3Hasher
	h = NewHasher(ModeSM)
	if _, ok := h.(*SM3Hasher); !ok {
		t.Error("NewHasher(sm) 应返回 *SM3Hasher")
	}
}

// TestNewAsymmetricCrypter 验证非对称加密器工厂
func TestNewAsymmetricCrypter(t *testing.T) {
	// rsa 模式
	a, err := NewAsymmetricCrypter(ModeRSA)
	if err != nil {
		t.Fatalf("NewAsymmetricCrypter(rsa) 失败: %v", err)
	}
	if _, ok := a.(*RSACrypter); !ok {
		t.Error("NewAsymmetricCrypter(rsa) 应返回 *RSACrypter")
	}

	// sm 模式
	a, err = NewAsymmetricCrypter(ModeSM)
	if err != nil {
		t.Fatalf("NewAsymmetricCrypter(sm) 失败: %v", err)
	}
	if _, ok := a.(*SM2Crypter); !ok {
		t.Error("NewAsymmetricCrypter(sm) 应返回 *SM2Crypter")
	}
}

// TestSHA256Hasher 验证 SHA-256 哈希器
func TestSHA256Hasher(t *testing.T) {
	h := &SHA256Hasher{}

	// Sum
	hash := h.Sum([]byte("hello"))
	if len(hash) != 32 {
		t.Errorf("SHA-256 Sum 长度错误: 期望 32, 实际 %d", len(hash))
	}

	// HMAC
	mac := h.HMAC([]byte("key"), []byte("data"))
	if len(mac) != 32 {
		t.Errorf("HMAC-SHA256 长度错误: 期望 32, 实际 %d", len(mac))
	}

	// 确定性
	h1 := h.Sum([]byte("hello"))
	h2 := h.Sum([]byte("hello"))
	if hex.EncodeToString(h1) != hex.EncodeToString(h2) {
		t.Error("SHA-256 Sum 应具有确定性")
	}
}
