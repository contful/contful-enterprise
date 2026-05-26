// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"

	"github.com/contful/contful/admin/internal/crypto"
)

const (
	defaultKeyDir   = "/app/conf/keys"
	pubKeyFileName  = "public.pem"
	privKeyFileName = "private.pem"
)

// GenKey 生成非对称密钥对（RSA 或 SM2，由 CRYPTO_MODE 环境变量决定）。
//
// 行为:
//   - 默认检查 /app/conf/keys/public.pem 是否存在；已存在则跳过（SKIPPED）。
//   - 传递 --force 标志可忽略已有文件，强制重新生成。
//   - 生成后写入 /app/conf/keys/public.pem (0644) 和 /app/conf/keys/private.pem (0600)。
//   - 输出 SHA-256 指纹到 stderr。
func GenKey() {
	mode := os.Getenv("CRYPTO_MODE")
	if mode == "" {
		mode = "rsa"
	}

	force := false
	for _, arg := range os.Args[2:] {
		if arg == "--force" {
			force = true
		}
	}

	pubPath := defaultKeyDir + "/" + pubKeyFileName
	privPath := defaultKeyDir + "/" + privKeyFileName

	if !force {
		if _, err := os.Stat(pubPath); err == nil {
			fmt.Fprintf(os.Stderr, "GEN_KEY: key pair already exists at %s — SKIPPED (use --force to regenerate)\n", pubPath)
			os.Exit(0)
		}
	}

	asym, err := crypto.NewAsymmetricCrypter(mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GEN_KEY: failed to create asymmetric crypter for mode %q: %v\n", mode, err)
		os.Exit(1)
	}

	pubPEM, privPEM, err := asym.GenerateKeyPair()
	if err != nil {
		fmt.Fprintf(os.Stderr, "GEN_KEY: failed to generate key pair: %v\n", err)
		os.Exit(1)
	}

	// 确保目录存在
	if err := os.MkdirAll(defaultKeyDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "GEN_KEY: failed to create directory %s: %v\n", defaultKeyDir, err)
		os.Exit(1)
	}

	if err := os.WriteFile(pubPath, []byte(pubPEM), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "GEN_KEY: failed to write public key: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(privPath, []byte(privPEM), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "GEN_KEY: failed to write private key: %v\n", err)
		os.Exit(1)
	}

	fingerprint := sha256Hex(pubPEM)
	fmt.Fprintf(os.Stderr,
		"GEN_KEY: key pair generated (mode=%s)\n  public:  %s\n  private: %s\n  SHA256:  %s\n",
		mode, pubPath, privPath, fingerprint,
	)
}

// sha256Hex 返回 data 的 SHA-256 指纹的前 16 个字符。
func sha256Hex(data string) string {
	h := sha256.Sum256([]byte(data))
	return strings.ToUpper(fmt.Sprintf("%x", h[:8]))
}
