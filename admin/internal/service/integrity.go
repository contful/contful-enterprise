// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/contful/contful/admin/pkg/uid"
	"github.com/contful/contful/admin/internal/model"
)

// contextKey IntegrityService 在 context 中的 key
type contextKey string

const integrityKey contextKey = "integrity_service"

// SignaturePayload 签名载荷结构（存储在 data_signature 列）
type SignaturePayload struct {
	Alg         string `json:"alg"`          // 算法: "HMAC-SHA256"
	CreatedAt   string `json:"created_at"`  // 签名时间
	SignedBy    string `json:"signed_by"`   // 签名者
	PayloadHash string `json:"payload_hash"` // SHA-256(Canonical Payload)
	Signature   string `json:"signature"`   // HMAC-SHA256 签名
}

// VerifyResult 验签结果
type VerifyResult struct {
	Valid       *bool  `json:"valid,omitempty"`         // nil = 未签名
	Alg         string `json:"alg,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	PayloadHash string `json:"payload_hash,omitempty"`
	Reason      string `json:"reason,omitempty"` // not_signed / payload_hash_mismatch / signature_invalid / malformed_signature
}

// IntegrityService 数据完整性签名服务（站点级别）
type IntegrityService struct {
	siteID     uid.UID
	signingKey []byte // 解密后的 HMAC 密钥
	alg        string // "HMAC-SHA256" 或 "SM3withSM2"
}

// NewIntegrityService 创建 IntegrityService（从配置中心读取签名密钥）
func NewIntegrityService(siteID uid.UID, signingKeyHex string, alg string) (*IntegrityService, error) {
	if signingKeyHex == "" {
		// 密钥不存在，签名功能关闭
		return &IntegrityService{
			siteID:     siteID,
			signingKey: nil,
			alg:        alg,
		}, nil
	}

	key, err := hex.DecodeString(signingKeyHex)
	if err != nil {
		return nil, fmt.Errorf("签名密钥格式错误（应为 hex）: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("签名密钥长度错误，需要 32 字节，当前 %d 字节", len(key))
	}

	return &IntegrityService{
		siteID:     siteID,
		signingKey: key,
		alg:        alg,
	}, nil
}

// IsEnabled 是否启用签名
func (s *IntegrityService) IsEnabled() bool {
	return s != nil && len(s.signingKey) == 32
}

// siteIDKey 用于在 context 中传递 siteID
type siteIDKey struct{}

// WithSiteID 将 siteID 存入 context（供 AuditLog callback 使用）
func WithSiteID(ctx context.Context, siteID uid.UID) context.Context {
	return context.WithValue(ctx, siteIDKey{}, siteID)
}

// GetSiteID 从 context 取出 siteID
func GetSiteID(ctx context.Context) (uid.UID, bool) {
	if v := ctx.Value(siteIDKey{}); v != nil {
		if id, ok := v.(uid.UID); ok {
			return id, true
		}
	}
	return uid.UID{}, false
}

// WithIntegrityService 将 IntegrityService 存入 context
func WithIntegrityService(ctx context.Context, svc *IntegrityService) context.Context {
	return context.WithValue(ctx, integrityKey, svc)
}

// GetIntegrityService 从 context 取出 IntegrityService
func GetIntegrityService(ctx context.Context) *IntegrityService {
	if v := ctx.Value(integrityKey); v != nil {
		return v.(*IntegrityService)
	}
	return nil
}

// SignEntry 为 Entry 生成签名（联动签名字段值）
func (s *IntegrityService) SignEntry(entry *model.Entry, values []model.EntryValue) error {
	if !s.IsEnabled() {
		return nil
	}

	payload := s.buildEntryCanonicalPayload(entry, values)
	return s.sign(entry.ID.String(), payload, func(sig model.JSONB) {
		entry.DataSignature = sig
	})
}

// SignAsset 为 Asset 生成签名
func (s *IntegrityService) SignAsset(asset *model.Asset) error {
	if !s.IsEnabled() {
		return nil
	}
	payload := s.buildAssetCanonicalPayload(asset)
	return s.sign(asset.ID.String(), payload, func(sig model.JSONB) {
		asset.DataSignature = sig
	})
}

// SignAuditLog 为 AuditLog 生成签名（强制，存储 hex HMAC-SHA256）
func (s *IntegrityService) SignAuditLog(log *model.AuditLog) error {
	if !s.IsEnabled() {
		log.DataSignature = ""
		return nil
	}
	payload := s.buildAuditCanonicalPayload(log)

	mac := hmac.New(sha256.New, s.signingKey)
	mac.Write([]byte(payload))
	log.DataSignature = hex.EncodeToString(mac.Sum(nil))
	return nil
}

// VerifyEntry 验签 Entry
func (s *IntegrityService) VerifyEntry(entry *model.Entry, values []model.EntryValue) (*VerifyResult, error) {
	if !s.IsEnabled() {
		return &VerifyResult{Valid: nil, Reason: "not_enabled"}, nil
	}
	payload := s.buildEntryCanonicalPayload(entry, values)
	return s.verify(entry.ID.String(), payload, entry.DataSignature)
}

// VerifyAsset 验签 Asset
func (s *IntegrityService) VerifyAsset(asset *model.Asset) (*VerifyResult, error) {
	if !s.IsEnabled() {
		return &VerifyResult{Valid: nil, Reason: "not_enabled"}, nil
	}
	payload := s.buildAssetCanonicalPayload(asset)
	return s.verify(asset.ID.String(), payload, asset.DataSignature)
}

// VerifyAuditLog 验签 AuditLog（对比 hex HMAC-SHA256）
func (s *IntegrityService) VerifyAuditLog(log *model.AuditLog) (*VerifyResult, error) {
	if !s.IsEnabled() {
		return &VerifyResult{Valid: nil, Reason: "not_enabled"}, nil
	}
	if log.DataSignature == "" {
		return &VerifyResult{Valid: nil, Reason: "not_signed"}, nil
	}

	payload := s.buildAuditCanonicalPayload(log)
	mac := hmac.New(sha256.New, s.signingKey)
	mac.Write([]byte(payload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(log.DataSignature), []byte(expectedSig)) != 1 {
		return &VerifyResult{
			Valid:  boolPtr(false),
			Alg:    s.alg,
			Reason: "signature_invalid",
		}, nil
	}

	return &VerifyResult{
		Valid: boolPtr(true),
		Alg:   s.alg,
	}, nil
}

// VerifyRaw 验签原始 data_signature JSON
func (s *IntegrityService) VerifyRaw(entityID, payload string, storedSig model.JSONB) (*VerifyResult, error) {
	if !s.IsEnabled() {
		return &VerifyResult{Valid: nil, Reason: "not_enabled"}, nil
	}
	return s.verify(entityID, payload, storedSig)
}

// ============ Canonical Payload 构建 ============

// buildEntryCanonicalPayload 构建 Entry 规范载荷
func (s *IntegrityService) buildEntryCanonicalPayload(entry *model.Entry, values []model.EntryValue) string {
	valuesHash := s.computeValuesHash(values)

	// 按固定字段顺序构建（Go map 迭代顺序不确定，用显式 slice 保证顺序）
	type kv struct{ k, v any }
	pairs := []kv{
		{"id", entry.ID.String()},
		{"schema_id", entry.ContentSchemaID.String()},
		{"site_id", entry.SiteID.String()},
		{"locale", entry.Locale},
		{"status", string(entry.Status)},
		{"version", entry.Version},
		{"values_hash", valuesHash},
		{"seo_title", entry.SEOTitle},
		{"seo_description", entry.SEODescription},
		{"sort_weight", entry.SortWeight},
	}
	if entry.PublishedTime != nil {
		pairs = append(pairs, kv{"published_time", entry.PublishedTime.Format(time.RFC3339)})
	}
	if entry.PublishedBy != nil {
		pairs = append(pairs, kv{"published_by", entry.PublishedBy.String()})
	}
	if entry.CreatedBy != nil {
		pairs = append(pairs, kv{"created_by", entry.CreatedBy.String()})
	}
	if len(entry.SEOKeywords) > 0 {
		pairs = append(pairs, kv{"seo_keywords", entry.SEOKeywords})
	}

	canonical := make(map[string]any, len(pairs))
	for _, p := range pairs {
		if s.isZero(p.v) {
			continue
		}
		canonical[p.k.(string)] = p.v
	}

	payload, _ := json.Marshal(canonical)
	return string(payload)
}

// computeValuesHash 计算字段值哈希（前 16 字符）
func (s *IntegrityService) computeValuesHash(values []model.EntryValue) string {
	if len(values) == 0 {
		return sha256Hex("")[:16]
	}

	sorted := make([]model.EntryValue, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FieldID.String() < sorted[j].FieldID.String()
	})

	var parts []string
	for _, v := range sorted {
		val := "null"
		if v.Value != nil {
			if b, err := json.Marshal(v.Value); err == nil {
				val = string(b)
			}
		}
		parts = append(parts, fmt.Sprintf(`{"field_id":"%s","value":%s}`, v.FieldID.String(), val))
	}
	combined := strings.Join(parts, "|")
	return sha256Hex(combined)[:16]
}

// buildAssetCanonicalPayload 构建 Asset 规范载荷
func (s *IntegrityService) buildAssetCanonicalPayload(asset *model.Asset) string {
	// 签名覆盖元信息，不签名动态字段（download_count/used_count）
	canonical := map[string]any{
		"id":         asset.ID.String(),
		"site_id":    asset.SiteID.String(),
		"name":       asset.Name,
		"type":       string(asset.Type),
		"mime_type":  asset.MimeType,
		"extension":  asset.Extension,
		"size":       asset.Size,
		"path":       asset.Path,
		"visibility": string(asset.Visibility),
		"file_hash":  asset.FileHash,
		"disk":       asset.Disk,
	}
	if asset.Width != nil {
		canonical["width"] = *asset.Width
	}
	if asset.Height != nil {
		canonical["height"] = *asset.Height
	}
	if asset.Alt != "" {
		canonical["alt"] = asset.Alt
	}
	if asset.Title != "" {
		canonical["title"] = asset.Title
	}

	payload, _ := json.Marshal(canonical)
	return string(payload)
}

// buildAuditCanonicalPayload 构建 AuditLog 规范载荷
func (s *IntegrityService) buildAuditCanonicalPayload(log *model.AuditLog) string {
	canonical := map[string]any{
		"id":             log.ID.String(),
		"action":        log.Action,
		"resource_type": log.ResourceType,
		"level":         string(log.Level),
		"category":      string(log.Category),
		"details":       log.Details,
		"ip_address":    log.IPAddress,
		"created_time":  log.CreatedTime,
	}
	if log.SiteID != nil {
		canonical["site_id"] = log.SiteID.String()
	}
	if log.UserID != nil {
		canonical["user_id"] = log.UserID.String()
	}
	if log.ResourceID != nil {
		canonical["resource_id"] = log.ResourceID.String()
	}

	payload, _ := json.Marshal(canonical)
	return string(payload)
}

// ============ 签名/验签核心 ============

// sign 通用签名方法
func (s *IntegrityService) sign(entityID, payload string, setSig func(model.JSONB)) error {
	payloadHash := sha256Hex(payload)
	now := time.Now().UTC().Format(time.RFC3339)

	var sigHex string
	switch s.alg {
	case "HMAC-SHA256":
		mac := hmac.New(sha256.New, s.signingKey)
		mac.Write([]byte(entityID + ":" + payloadHash))
		sigHex = hex.EncodeToString(mac.Sum(nil))
	default:
		// 默认用 HMAC-SHA256
		mac := hmac.New(sha256.New, s.signingKey)
		mac.Write([]byte(entityID + ":" + payloadHash))
		sigHex = hex.EncodeToString(mac.Sum(nil))
	}

	setSig(model.JSONB{
		"alg":          s.alg,
		"created_at":    now,
		"signed_by":    s.alg,
		"payload_hash": payloadHash,
		"signature":    sigHex,
	})
	return nil
}

// verify 通用验签方法
func (s *IntegrityService) verify(entityID, payload string, storedSig model.JSONB) (*VerifyResult, error) {
	if storedSig == nil || len(storedSig) == 0 {
		return &VerifyResult{Valid: nil, Reason: "not_signed"}, nil
	}

	alg, _ := storedSig["alg"].(string)
	payloadHashStored, _ := storedSig["payload_hash"].(string)
	sigStored, _ := storedSig["signature"].(string)
	createdAt, _ := storedSig["created_at"].(string)

	if alg == "" || payloadHashStored == "" || sigStored == "" {
		return &VerifyResult{Valid: boolPtr(false), Reason: "malformed_signature"}, nil
	}

	currentPayloadHash := sha256Hex(payload)

	// 第一层：比对 payload_hash
	if subtle.ConstantTimeCompare([]byte(payloadHashStored), []byte(currentPayloadHash)) != 1 {
		return &VerifyResult{
			Valid:       boolPtr(false),
			Alg:         alg,
			CreatedAt:   createdAt,
			PayloadHash: payloadHashStored,
			Reason:      "payload_hash_mismatch",
		}, nil
	}

	// 第二层：重新计算签名比对
	var expectedSig string
	switch alg {
	case "HMAC-SHA256":
		mac := hmac.New(sha256.New, s.signingKey)
		mac.Write([]byte(entityID + ":" + currentPayloadHash))
		expectedSig = hex.EncodeToString(mac.Sum(nil))
	default:
		mac := hmac.New(sha256.New, s.signingKey)
		mac.Write([]byte(entityID + ":" + currentPayloadHash))
		expectedSig = hex.EncodeToString(mac.Sum(nil))
	}

	if subtle.ConstantTimeCompare([]byte(sigStored), []byte(expectedSig)) != 1 {
		return &VerifyResult{
			Valid:       boolPtr(false),
			Alg:         alg,
			CreatedAt:   createdAt,
			PayloadHash: currentPayloadHash,
			Reason:      "signature_invalid",
		}, nil
	}

	return &VerifyResult{
		Valid:       boolPtr(true),
		Alg:         alg,
		CreatedAt:   createdAt,
		PayloadHash: currentPayloadHash,
	}, nil
}

// ============ 辅助函数 ============

func sha256Hex(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

func boolPtr(b bool) *bool { return &b }

func (s *IntegrityService) isZero(v any) bool {
	switch x := v.(type) {
	case string:
		return x == ""
	case int:
		return x == 0
	case int64:
		return x == 0
	case float64:
		return x == 0
	case bool:
		return !x
	case []string:
		return len(x) == 0
	case *string:
		return x == nil
	case *int:
		return x == nil
	case *int64:
		return x == nil
	case *float64:
		return x == nil
	case *uid.UID:
		return x == nil
	case *time.Time:
		return x == nil
	default:
		return false
	}
}


