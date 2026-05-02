// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

// Response API 统一响应结构（Open API）
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) Response {
	return Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, msg string) Response {
	return Response{
		Code: code,
		Msg:  msg,
	}
}

// Error codes（Open API 专用）
const (
	CodeSuccess              = 0
	CodeBadRequest           = 400
	CodeUnauthorized         = 401
	CodeForbidden            = 403
	CodeNotFound             = 404
	CodeTooManyRequests      = 429
	CodeInternalError        = 500
	CodeInvalidToken         = 40101 // Token 无效
	CodeTokenExpired         = 40102 // Token 过期
	CodeTokenRevoked         = 40103 // Token 已撤销
	CodeInsufficientScope    = 40301 // 权限不足
	CodeRateLimitExceeded    = 42901 // 速率超限
)
