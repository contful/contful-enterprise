// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package uid

import "github.com/contful/contful-enterprise/shared/uid"

// 从 shared/uid 透传，一元维护
type UID = uid.UID

var (
	New       = uid.New
	Parse     = uid.Parse
	Nil       = uid.Nil
	SetDBType = uid.SetDBType
	GenUUID   = uid.GenUUID
)
