// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RootPublicKey 根公钥（编译进二进制，永远不变）
var RootPublicKey = []byte{
	0xec, 0x31, 0x8b, 0x2f, 0xa4, 0x1f, 0xde, 0xa3,
	0x87, 0x88, 0x9c, 0x9e, 0xf4, 0xeb, 0xec, 0xfb,
	0x79, 0xbc, 0x90, 0xca, 0x8a, 0xf9, 0xa4, 0xe3,
	0xf3, 0xf7, 0x04, 0xe7, 0x6f, 0x23, 0x4b, 0xee,
}

// Info 许可证信息结构体
type Info struct {
	Customer       string    `json:"customer"`                  // 授权客户名
	ExpiryDate     time.Time `json:"expiry_date,omitempty"`     // 有效期
	IsTrial        bool      `json:"is_trial,omitempty"`        // 免费试用
	IssuedDate     time.Time `json:"issued_date,omitempty"`     // 签发日期
	ProductCode    string    `json:"product_code,omitempty"`    // 产品编号
	ProductName    string    `json:"product_name,omitempty"`    // 产品名称
	ProductVersion string    `json:"product_version,omitempty"` // 产品版本
}

// IsExpired 判断 license 是否已过期
func (i *Info) IsExpired() bool {
	return time.Now().After(i.ExpiryDate)
}

// Status 返回 license 状态
func (i *Info) Status() string {
	if i.IsExpired() {
		return "expired"
	}
	return "active"
}

// Load 从 conf/license.dat 加载并验证 license
func Load() (*Info, error) {
	searchPaths := []string{"./conf", "../conf", "."}
	var data []byte
	var foundPath string

	for _, dir := range searchPaths {
		candidate := filepath.Join(dir, "license.dat")
		if _, err := os.Stat(candidate); err == nil {
			var err error
			data, err = os.ReadFile(candidate)
			if err != nil {
				continue
			}
			foundPath = candidate
			break
		}
	}

	if data == nil {
		return nil, fmt.Errorf("license.dat not found in search paths: %v", searchPaths)
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return nil, fmt.Errorf("license.dat is empty (%s)", foundPath)
	}

	info, err := verifyLicenseFile(token)
	if err != nil {
		return nil, fmt.Errorf("license verification failed (%s): %v", foundPath, err)
	}

	return info, nil
}

// verifyLicenseFile 客户端验证授权文件（仅依赖 RootPublicKey）
// licenseFile: license.dat 文件内容（4 段 base64 以 . 分隔）
// 验证通过返回 License 信息，失败返回 error
func verifyLicenseFile(licenseFile string) (*Info, error) {
	parts := strings.Split(licenseFile, ".")
	if len(parts) != 4 {
		return nil, errors.New("授权文件格式无效: 需要 4 段数据")
	}

	// 1. 解码子公钥
	childPub, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("子公钥解码失败: %v", err)
	}
	if len(childPub) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("子公钥长度无效: %d", len(childPub))
	}

	// 2. 解码根签名
	rootSig, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("根签名解码失败: %v", err)
	}

	// 3. 解码 License JSON
	licenseJSON, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("授权数据解码失败: %v", err)
	}

	// 4. 解码子签名
	childSig, err := base64.StdEncoding.DecodeString(parts[3])
	if err != nil {
		return nil, fmt.Errorf("授权签名解码失败: %v", err)
	}

	// ① 用根公钥验证子公钥签名
	if !ed25519.Verify(RootPublicKey, childPub, rootSig) {
		return nil, errors.New("授权文件已损坏: 根签名验证失败")
	}

	// ② 用子公钥验证 License 签名
	if !ed25519.Verify(childPub, licenseJSON, childSig) {
		return nil, errors.New("授权文件已损坏: License 签名验证失败，可能被篡改")
	}

	// ③ 解析 License
	var info Info
	if err := json.Unmarshal(licenseJSON, &info); err != nil {
		return nil, fmt.Errorf("授权数据解析失败: %v", err)
	}

	// ④ 检查是否过期
	if !info.ExpiryDate.IsZero() && time.Now().After(info.ExpiryDate) {
		return nil, fmt.Errorf("授权已过期: %s (过期时间: %s)",
			info.Customer, info.ExpiryDate.Format("2006-01-02"))
	}

	return &info, nil
}
