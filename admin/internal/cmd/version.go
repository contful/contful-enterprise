// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"fmt"
	"os"
)

// Version 构建时的版本号，通过 -ldflags 注入。默认 "dev"。
var Version = "dev"

// BuildDate 构建日期，通过 -ldflags 注入。默认 "unknown"。
var BuildDate = "unknown"

// SchemaVer 数据库 schema 版本。默认 "v2.0.0"。
var SchemaVer = "v2.0.0"

// PrintVersion 输出构建信息到 stderr。
func PrintVersion() {
	fmt.Fprintf(os.Stderr, "Version:    %s\n", Version)
	fmt.Fprintf(os.Stderr, "BuildDate:  %s\n", BuildDate)
	fmt.Fprintf(os.Stderr, "SchemaVer:  %s\n", SchemaVer)
}
