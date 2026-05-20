// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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

// GetKey 获取 AES-256 密钥（32 字节）
// 从 LICENSE_KEY 环境变量读取，去除连字符后转为字节切片
// 开发/测试环境使用默认 UUID
func GetKey() []byte {
	rawKey := os.Getenv("LICENSE_KEY")
	if rawKey == "" {
		rawKey = "7f16aef8-4587-5b94-8bd2-fa156882da6b"
	}
	cleanKey := strings.ReplaceAll(rawKey, "-", "")
	return []byte(cleanKey)
}

// Load 从 conf/license.dat 加载并解密 license
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

	info, err := Decrypt(token, GetKey())
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt license (%s): %v", foundPath, err)
	}

	return info, nil
}

// Decrypt AES-256-CBC 解密 license
func Decrypt(encrypted string, key []byte) (*Info, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("invalid key: %v", err)
	}

	blockSize := block.BlockSize()
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, fmt.Errorf("base64 decode error: %v", err)
	}

	if len(data) < blockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// 分离 IV 和密文
	iv := data[:blockSize]
	ciphertext := data[blockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// 去除 PKCS7 填充
	unpadded, err := pkcs7Unpad(plaintext, blockSize)
	if err != nil {
		return nil, fmt.Errorf("padding error: %v", err)
	}

	var info Info
	if err := json.Unmarshal(unpadded, &info); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %v", err)
	}

	return &info, nil
}

// pkcs7Unpad 去除 PKCS7 填充
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("empty data")
	}
	if length%blockSize != 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	padding := int(data[length-1])
	if padding > length || padding == 0 {
		return nil, fmt.Errorf("invalid padding size")
	}
	return data[:length-padding], nil
}

// Encrypt AES-256-CBC 加密 license（用于测试/验证）
func Encrypt(info Info, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("invalid key: %v", err)
	}

	blockSize := block.BlockSize()
	plaintext, err := json.Marshal(info)
	if err != nil {
		return "", fmt.Errorf("json marshal error: %v", err)
	}

	// PKCS7 填充
	padded := pkcs7Pad(plaintext, blockSize)

	// 生成随机 IV
	iv := make([]byte, blockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("iv generation failed: %v", err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(padded))
	mode.CryptBlocks(ciphertext, padded)

	encrypted := append(iv, ciphertext...)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// pkcs7Pad PKCS7 填充
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}
