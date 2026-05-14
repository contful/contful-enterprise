// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

declare namespace API {
  interface Response<T = unknown> {
    code: number
    message: string
    msg?: string
    data: T
  }
}
