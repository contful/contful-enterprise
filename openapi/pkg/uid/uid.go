// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package uid

import "github.com/contful/contful-enterprise/shared/uid"

type UID = uid.UID

var (
	New       = uid.New
	Parse     = uid.Parse
	Nil       = uid.Nil
	SetDBType = uid.SetDBType
	GenUUID   = uid.GenUUID
)
