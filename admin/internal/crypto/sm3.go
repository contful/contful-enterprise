// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package crypto

import (
	"encoding/hex"
	"hash"

	"github.com/tjfoc/gmsm/sm3"
)

// SM3Hasher SM3 密码杂凑算法（国密）
type SM3Hasher struct{}

// Sum 计算 SM3 哈希，返回 32 字节
func (h *SM3Hasher) Sum(data []byte) []byte {
	hash := sm3.Sm3Sum(data)
	return hash
}

// HMAC 计算 HMAC-SM3
func (h *SM3Hasher) HMAC(key, data []byte) []byte {
	return hmacSM3(key, data)
}

// SumHex 计算 SM3 哈希，返回 64 字符 hex 字符串
func (h *SM3Hasher) SumHex(data []byte) string {
	return hex.EncodeToString(h.Sum(data))
}

// HMACHex 计算 HMAC-SM3，返回 64 字符 hex 字符串
func (h *SM3Hasher) HMACHex(key, data []byte) string {
	return hex.EncodeToString(h.HMAC(key, data))
}

// hmacSM3 实现 HMAC-SM3（遵循 RFC 2104）
func hmacSM3(key, data []byte) []byte {
	const blockSize = 64 // SM3 块大小

	// 零填充或哈希缩减 key
	if len(key) > blockSize {
		hashed := sm3.Sm3Sum(key)
		key = hashed[:]
	}
	if len(key) < blockSize {
		padded := make([]byte, blockSize)
		copy(padded, key)
		key = padded
	}

	// ipad / opad
	ipad := make([]byte, blockSize)
	opad := make([]byte, blockSize)
	for i := 0; i < blockSize; i++ {
		ipad[i] = key[i] ^ 0x36
		opad[i] = key[i] ^ 0x5C
	}

	// HMAC = H(opad || H(ipad || data))
	inner := sm3.New()
	inner.Write(ipad)
	inner.Write(data)
	innerHash := inner.Sum(nil)

	outer := sm3.New()
	outer.Write(opad)
	outer.Write(innerHash)
	return outer.Sum(nil)
}

// newSM3Hash 创建新的 SM3 hash.Hash 实例（用于需要 io.Writer 的场景）
func newSM3Hash() hash.Hash {
	return sm3.New()
}
