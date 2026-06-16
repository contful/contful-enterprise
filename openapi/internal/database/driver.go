// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package database

import "github.com/contful/contful-enterprise/shared/database"

type DSNConfig = database.DSNConfig

var (
	CurrentDBType = database.CurrentDBType
	Open          = database.Open
)
