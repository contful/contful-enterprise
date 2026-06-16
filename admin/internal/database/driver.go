// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package database

import "github.com/contful/contful-enterprise/shared/database"

// 从 shared/database 透传（类型别名 + 函数转发，一元维护）
type DSNConfig = database.DSNConfig

var (
	CurrentDBType = database.CurrentDBType
	Open          = database.Open
)
