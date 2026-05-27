//go:build ignore
// +build ignore

// gen_license — 使用 Ed25519 双层签名方案生成 conf/license.dat
// 用法: go run scripts/gen_license.go > ../conf/license.dat

package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// LicenseV2 授权信息
type LicenseV2 struct {
	Customer       string    `json:"customer"`
	ExpiryDate     time.Time `json:"expiry_date,omitempty"`
	IsTrial        bool      `json:"is_trial,omitempty"`
	IssuedDate     time.Time `json:"issued_date,omitempty"`
	ProductCode    string    `json:"product_code,omitempty"`
	ProductName    string    `json:"product_name,omitempty"`
	ProductVersion string    `json:"product_version,omitempty"`
}

func main() {
	// 密钥（与 reepu/service/license 共用同一套根密钥体系）
	rootPrivHex := getEnv("LICENSE_ROOT_KEY", "8b1786f00ae99d018a0bff45bbab43502da35905a67ac84e6c6ce35945e0abd8ec318b2fa41fdea387889c9ef4ebecfb79bc90ca8af9a4e3f3f704e76f234bee")
	childPrivHex := getEnv("LICENSE_CHILD_KEY", "9bd159d38e0390d9da75d404411fdf295e7b9e988a6fcf95d3ddeb2a2f494e510033bf6e49a1624c2c9ed9db6d4ba87ed3d88b1cfb4641744792180574a239ee")

	license := LicenseV2{
		Customer:       "演示版",
		ProductName:    "Contful",
		ProductVersion: "企业版 1.3.0",
		ProductCode:    "contful-ent-001",
		IsTrial:        true,
		IssuedDate:     time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		ExpiryDate:     time.Date(2027, 12, 30, 0, 0, 0, 0, time.UTC),
	}

	authFile, err := signLicense(license, rootPrivHex, childPrivHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "签名失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(authFile)
}

func signLicense(info LicenseV2, rootPrivHex, childPrivHex string) (string, error) {
	rootPriv, err := hex.DecodeString(rootPrivHex)
	if err != nil || len(rootPriv) != ed25519.PrivateKeySize {
		return "", fmt.Errorf("根私钥无效: 需要 %d 字节 hex 编码", ed25519.PrivateKeySize)
	}

	childPriv, err := hex.DecodeString(childPrivHex)
	if err != nil || len(childPriv) != ed25519.PrivateKeySize {
		return "", fmt.Errorf("子私钥无效: 需要 %d 字节 hex 编码", ed25519.PrivateKeySize)
	}

	// 提取子公钥（Ed25519 私钥后 32 字节即为公钥）
	childPub := childPriv[32:]

	// ① 用根私钥签子公钥
	rootSig := ed25519.Sign(rootPriv, childPub)

	// ② 序列化 License
	licenseJSON, err := json.Marshal(info)
	if err != nil {
		return "", fmt.Errorf("json marshal error: %v", err)
	}

	// ③ 用子私钥签 License
	childSig := ed25519.Sign(childPriv, licenseJSON)

	// ④ 打包: 子公钥.根签名.License.子签名
	return strings.Join([]string{
		base64.StdEncoding.EncodeToString(childPub),
		base64.StdEncoding.EncodeToString(rootSig),
		base64.StdEncoding.EncodeToString(licenseJSON),
		base64.StdEncoding.EncodeToString(childSig),
	}, "."), nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
