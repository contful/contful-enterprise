// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"strings"
	"testing"
)

// TestRootPublicKeyMatches 验证 RootPublicKey 与 reepu/service/license 中的根密钥一致
// 根私钥 hex 后 32 字节即为根公钥
func TestRootPublicKeyMatches(t *testing.T) {
	rootPrivHex := "8b1786f00ae99d018a0bff45bbab43502da35905a67ac84e6c6ce35945e0abd8ec318b2fa41fdea387889c9ef4ebecfb79bc90ca8af9a4e3f3f704e76f234bee"
	rootPriv, err := hexDecode(rootPrivHex)
	if err != nil {
		t.Fatalf("根私钥 hex 解码失败: %v", err)
	}
	if len(rootPriv) != ed25519.PrivateKeySize {
		t.Fatalf("根私钥长度错误: 期望 %d, 实际 %d", ed25519.PrivateKeySize, len(rootPriv))
	}

	// Ed25519 私钥后 32 字节是公钥
	extractedPub := rootPriv[32:]

	if len(extractedPub) != len(RootPublicKey) {
		t.Fatalf("提取的公钥长度错误: 期望 %d, 实际 %d", len(RootPublicKey), len(extractedPub))
	}

	for i := range RootPublicKey {
		if RootPublicKey[i] != extractedPub[i] {
			t.Fatalf("RootPublicKey 不匹配，确保 license_v2.go::RootPublicKey 来源于: %x", extractedPub)
		}
	}
	t.Logf("✅ RootPublicKey 与根私钥匹配")
}

// TestVerifyLicenseFileFormat 验证 license.dat 格式正确且能通过双重验证
func TestVerifyLicenseFileFormat(t *testing.T) {
	testFile := "../../../conf/license.dat"
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Skipf("跳过: 无法读取 license.dat (%v) — 此测试仅验证正确性", err)
		return
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		t.Fatal("license.dat 为空")
	}

	info, err := verifyLicenseFile(token)
	if err != nil {
		t.Fatalf("❌ license 验证失败: %v", err)
	}

	t.Logf("✅ License 验证通过")
	t.Logf("   客户: %s", info.Customer)
	t.Logf("   产品: %s", info.ProductName)
	t.Logf("   版本: %s", info.ProductVersion)
	t.Logf("   试用: %v", info.IsTrial)
	t.Logf("   到期: %s", info.ExpiryDate.Format("2006-01-02"))
}

// TestRootSignatureVerification 单独测试根签名:用 RootPublicKey 验证子公钥签名
func TestRootSignatureVerification(t *testing.T) {
	childPubRaw := "ADO/bkmhYkwsntnbbUuoftPYixz7RkF0R5IYBXSiOe4="
	rootSigBase64 := "+hIXsJvs7W5S/9dyfZL9FWdmnxqJ7Qx+Ca6rGlUkaKwItVkhM5MSIpCtRdQzaGNGcL2ZL4QfJpCz8ZZXICOiAg=="

	childPub, err := base64.StdEncoding.DecodeString(childPubRaw)
	if err != nil {
		t.Fatalf("子公钥解码失败: %v", err)
	}

	rootSig, err := base64.StdEncoding.DecodeString(rootSigBase64)
	if err != nil {
		t.Fatalf("根签名解码失败: %v", err)
	}

	if !ed25519.Verify(RootPublicKey, childPub, rootSig) {
		t.Fatal("❌ 根签名验证失败: RootPublicKey 与签发时使用的根密钥不匹配")
	}
	t.Log("✅ 根签名验证通过: 子公钥确由根密钥签发")
}

// hexDecode helper
func hexDecode(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		s = "0" + s
	}
	result := make([]byte, len(s)/2)
	for i := 0; i < len(result); i++ {
		hi := hexVal(s[i*2])
		lo := hexVal(s[i*2+1])
		if hi < 0 || lo < 0 {
			return nil, fmtError("invalid hex character")
		}
		result[i] = byte(hi<<4 | lo)
	}
	return result, nil
}

func hexVal(c byte) int {
	switch {
	case '0' <= c && c <= '9':
		return int(c - '0')
	case 'a' <= c && c <= 'f':
		return int(c - 'a' + 10)
	case 'A' <= c && c <= 'F':
		return int(c - 'A' + 10)
	}
	return -1
}

type errString string

func (e errString) Error() string { return string(e) }

func fmtError(msg string) error { return errString(msg) }
