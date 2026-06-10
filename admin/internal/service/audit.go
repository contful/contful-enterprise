// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/contful/contful/admin/pkg/uid"
	"github.com/rs/zerolog/log"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
)

// AuditService 审计日志服务
type AuditService struct {
	auditRepo *repository.AuditRepository
	configSvc *ConfigService
}

func NewAuditService(auditRepo *repository.AuditRepository, configSvc *ConfigService) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
		configSvc: configSvc,
	}
}

// LogOption 审计日志记录选项
type LogOption func(*model.AuditLog)

// WithAuditSiteID 设置站点 ID
func WithAuditSiteID(siteID uid.UID) LogOption {
	return func(a *model.AuditLog) {
		a.SiteID = &siteID
	}
}

// WithResource 设置资源类型和资源 ID
func WithResource(resourceType string, resourceID uid.UID) LogOption {
	return func(a *model.AuditLog) {
		a.ResourceType = resourceType
		a.ResourceID = &resourceID
	}
}

// WithDetails 设置详细信息
func WithDetails(details string) LogOption {
	return func(a *model.AuditLog) {
		a.Details = details
	}
}

// WithIPAddress 设置 IP 地址
func WithIPAddress(ip string) LogOption {
	return func(a *model.AuditLog) {
		a.IPAddress = ip
	}
}

// WithUserAgent 设置 User-Agent
func WithUserAgent(ua string) LogOption {
	return func(a *model.AuditLog) {
		a.UserAgent = ua
	}
}

// Log 记录审计日志（高层接口）
func (s *AuditService) Log(ctx context.Context, userID uid.UID, level model.AuditLevel, category model.AuditType, action string, opts ...LogOption) error {
	auditLog := &model.AuditLog{
		UserID:   &userID,
		Action:   action,
		Level:    level,
		Category: category,
	}

	// 应用选项
	for _, opt := range opts {
		opt(auditLog)
	}

	// 获取签名密钥并记录
	signingKey, err := s.configSvc.GetAuditSigningKey()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get audit signing key, logging without signature")
		return s.auditRepo.Create(ctx, auditLog)
	}

	return s.auditRepo.CreateWithSigningKey(ctx, auditLog, signingKey)
}

// LogFromGin 从 Gin 上下文记录审计日志（自动提取用户信息）
func (s *AuditService) LogFromGin(c *gin.Context, level model.AuditLevel, category model.AuditType, action string, opts ...LogOption) error {
	// 从 Gin 上下文获取用户 ID（中间件存入 key="user"，类型 uid.UID）
	userIDVal, exists := c.Get("user")
	if !exists {
		log.Warn().Msg("user_id not found in gin context, skipping audit log")
		return nil
	}

	userID, ok := userIDVal.(uid.UID)
	if !ok {
		log.Warn().Msg("invalid user_id type in gin context, skipping audit log")
		return nil
	}

	// 自动提取 IP 和 User-Agent
	opts = append(opts, WithIPAddress(c.ClientIP()))
	opts = append(opts, WithUserAgent(c.GetHeader("User-Agent")))

	return s.Log(c.Request.Context(), userID, level, category, action, opts...)
}

// LogAuth 记录认证相关审计日志
func (s *AuditService) LogAuth(ctx context.Context, userID uid.UID, action string, ipAddress string, userAgent string, success bool) error {
	level := model.AuditLevelInfo
	details := "success"
	if !success {
		level = model.AuditLevelWarn
		details = "failed"
	}

	return s.Log(ctx, userID, level, model.AuditTypeAuth, action,
		WithIPAddress(ipAddress),
		WithUserAgent(userAgent),
		WithDetails(details),
	)
}

// LogUser 记录用户管理相关审计日志
func (s *AuditService) LogUser(ctx context.Context, operatorID uid.UID, action string, targetUserID uid.UID, opts ...LogOption) error {
	opts = append(opts, WithResource("user", targetUserID))
	return s.Log(ctx, operatorID, model.AuditLevelInfo, model.AuditTypeUser, action, opts...)
}

// LogRole 记录角色管理相关审计日志
func (s *AuditService) LogRole(ctx context.Context, operatorID uid.UID, action string, targetRoleID uid.UID, opts ...LogOption) error {
	opts = append(opts, WithResource("role", targetRoleID))
	return s.Log(ctx, operatorID, model.AuditLevelInfo, model.AuditTypeUser, action, opts...)
}

// LogSite 记录站点管理相关审计日志
func (s *AuditService) LogSite(ctx context.Context, operatorID uid.UID, action string, targetSiteID uid.UID, opts ...LogOption) error {
	opts = append(opts, WithResource("site", targetSiteID))
	return s.Log(ctx, operatorID, model.AuditLevelInfo, model.AuditTypeSetting, action, opts...)
}

// LogToken 记录 Token 管理相关审计日志
func (s *AuditService) LogToken(ctx context.Context, operatorID uid.UID, action string, targetTokenID uid.UID, opts ...LogOption) error {
	opts = append(opts, WithResource("token", targetTokenID))
	return s.Log(ctx, operatorID, model.AuditLevelInfo, model.AuditTypeSystem, action, opts...)
}

// LogContent 记录内容管理相关审计日志
func (s *AuditService) LogContent(ctx context.Context, operatorID uid.UID, action string, siteID uid.UID, schemaID uid.UID, entryID uid.UID, opts ...LogOption) error {
	opts = append(opts, WithAuditSiteID(siteID))
	opts = append(opts, WithResource("entry", entryID))
	opts = append(opts, WithDetails("schema_id="+schemaID.String()))
	return s.Log(ctx, operatorID, model.AuditLevelInfo, model.AuditTypeContent, action, opts...)
}

// LogError 记录错误审计日志
func (s *AuditService) LogError(ctx context.Context, userID uid.UID, category model.AuditType, action string, err error, opts ...LogOption) error {
	opts = append(opts, WithDetails("error="+err.Error()))
	return s.Log(ctx, userID, model.AuditLevelError, category, action, opts...)
}

// List 查询审计日志列表（支持筛选和分页）
func (s *AuditService) List(ctx context.Context, filter *model.AuditLogFilter, page, pageSize int) ([]model.AuditLog, int64, error) {
	return s.auditRepo.List(ctx, filter, page, pageSize)
}

// GetByID 根据 ID 获取审计日志详情
func (s *AuditService) GetByID(ctx context.Context, id uid.UID) (*model.AuditLog, error) {
	return s.auditRepo.GetByID(ctx, id)
}

// GetSigningKey 获取签名密钥（供其他服务使用）
func (s *AuditService) GetSigningKey(ctx context.Context) (string, error) {
	return s.configSvc.GetAuditSigningKey()
}

// ExportCSV 导出审计日志为 CSV 格式
// 返回: CSV 字节流、实际记录数、总数、error
func (s *AuditService) ExportCSV(ctx context.Context, filter *model.AuditLogFilter, maxRows int) ([]byte, int64, int64, error) {
	logs, total, err := s.auditRepo.ExportAll(ctx, filter, maxRows)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("export query: %w", err)
	}

	var buf bytes.Buffer

	// UTF-8 BOM（Excel/Numbers 兼容中文）
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(&buf)

	// 表头
	headers := []string{
		"id", "action", "category", "level", "resource_type", "resource_id",
		"user_id", "site_id", "ip_address", "user_agent", "details",
		"created_time", "data_signature",
	}
	if err := w.Write(headers); err != nil {
		return nil, 0, 0, fmt.Errorf("write csv header: %w", err)
	}

	// 数据行
	for _, log := range logs {
		row := []string{
			log.ID.String(),
			log.Action,
			string(log.Category),
			string(log.Level),
			log.ResourceType,
			uuidOrEmpty(log.ResourceID),
			uuidOrEmpty(log.UserID),
			uuidOrEmpty(log.SiteID),
			log.IPAddress,
			log.UserAgent,
			log.Details,
			log.CreatedTime.Format("2006-01-02T15:04:05Z07:00"),
			log.DataSignature,
		}
		if err := w.Write(row); err != nil {
			return nil, 0, 0, fmt.Errorf("write csv row: %w", err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, 0, 0, fmt.Errorf("csv flush: %w", err)
	}

	// 追加 HMAC 签名行
	bodyHash := sha256.Sum256(buf.Bytes())
	sig := signBody(bodyHash[:])

	sigLine := fmt.Sprintf("\n#SIGNATURE: %s\n", sig)
	buf.WriteString(sigLine)

	return buf.Bytes(), int64(len(logs)), total, nil
}

// ExportXLSX 导出审计日志为 XLSX 格式（含条件着色 + 完整性声明 sheet）
// 使用标准库 archive/zip + encoding/xml，零外部依赖
func (s *AuditService) ExportXLSX(ctx context.Context, filter *model.AuditLogFilter, maxRows int) ([]byte, int64, int64, error) {
	logs, total, err := s.auditRepo.ExportAll(ctx, filter, maxRows)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("export query: %w", err)
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// [Content_Types].xml
	_ = writeZipEntry(zw, "[Content_Types].xml", xlsxContentTypes)

	// _rels/.rels
	_ = writeZipEntry(zw, "_rels/.rels", xlsxRels)

	// xl/workbook.xml
	_ = writeZipEntry(zw, "xl/workbook.xml", xlsxWorkbook)

	// xl/_rels/workbook.xml.rels
	_ = writeZipEntry(zw, "xl/_rels/workbook.xml.rels", xlsxWorkbookRels)

	// xl/styles.xml
	_ = writeZipEntry(zw, "xl/styles.xml", xlsxStyles)

	// xl/sharedStrings.xml
	ss := newSharedStrings()
	_ = writeZipEntry(zw, "xl/sharedStrings.xml", ss.toXML())

	// xl/worksheets/sheet1.xml — 审计日志
	ws1 := newWorksheet("审计日志", logs)
	ws1.addHeaderRow(
		"ID", "操作", "类别", "级别", "资源类型", "资源ID",
		"用户ID", "站点ID", "IP", "User-Agent", "详情", "时间", "数据签名",
	)
	for _, log := range logs {
		style := 0
		switch log.Level {
		case model.AuditLevelError:
			style = 3 // red
		case model.AuditLevelWarn:
			style = 2 // yellow
		case model.AuditLevelInfo:
			style = 1 // green
		}
		ws1.addRow(style,
			log.ID.String(),
			log.Action,
			string(log.Category),
			string(log.Level),
			log.ResourceType,
			uuidOrEmpty(log.ResourceID),
			uuidOrEmpty(log.UserID),
			uuidOrEmpty(log.SiteID),
			log.IPAddress,
			log.UserAgent,
			log.Details,
			log.CreatedTime.Format("2006-01-02T15:04:05Z07:00"),
			log.DataSignature,
		)
	}
	ws1XML := ws1.toXML(ss)
	_ = writeZipEntry(zw, "xl/worksheets/sheet1.xml", ws1XML)

	// xl/worksheets/sheet2.xml — 完整性声明
	ws2 := newWorksheet("完整性声明", nil)
	ws2.addRow(0, "完整性声明")
	ws2.addRow(0, "")
	ws2.addRow(0, "导出时间", time.Now().Format("2006-01-02 15:04:05"))
	ws2.addRow(0, "筛选条件", filterDescription(filter))
	ws2.addRow(0, "总记录数", fmt.Sprintf("%d", total))
	ws2.addRow(0, "导出记录数", fmt.Sprintf("%d", len(logs)))
	ws2.addRow(0, "签名算法", "HMAC-SHA256")
	ws2XML := ws2.toXML(ss)
	_ = writeZipEntry(zw, "xl/worksheets/sheet2.xml", ws2XML)

	zw.Close()

	// Prepend UTF-8 BOM
	result := append([]byte{0xEF, 0xBB, 0xBF}, buf.Bytes()...)

	return result, int64(len(logs)), total, nil
}

// minimal XLSX writer helpers (standard library only)

func writeZipEntry(zw *zip.Writer, name, content string) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(content))
	return err
}

type sharedStrings struct {
	index map[string]int
	items []string
}

func newSharedStrings() *sharedStrings {
	return &sharedStrings{index: make(map[string]int)}
}

func (s *sharedStrings) add(v string) int {
	if i, ok := s.index[v]; ok {
		return i
	}
	i := len(s.items)
	s.items = append(s.items, v)
	s.index[v] = i
	return i
}

func (s *sharedStrings) toXML() string {
	if len(s.items) == 0 {
		return `<?xml version="1.0" encoding="UTF-8"?><sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="0" uniqueCount="0"/>`
	}
	parts := make([][2]int, 0, len(s.items))
	for _, v := range s.items {
		parts = append(parts, [2]int{s.index[v], len(v)})
	}
	// collect all items in order
	itemsXML := ""
	for _, v := range s.items {
		itemsXML += fmt.Sprintf("<si><t>%s</t></si>", xmlEscape(v))
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="%d" uniqueCount="%d">%s</sst>`,
		len(s.items), len(s.items), itemsXML)
}

type worksheet struct {
	name  string
	rows  [][]xlsxCell
	cols  int
	frozen bool
}

type xlsxCell struct {
	value string
	style int // 0=default, 1=green, 2=yellow, 3=red
	isStr bool // true = shared string, false = inline
}

func newWorksheet(name string, _ []model.AuditLog) *worksheet {
	return &worksheet{name: name}
}

func (w *worksheet) addHeaderRow(cells ...string) {
	row := make([]xlsxCell, len(cells))
	for i, c := range cells {
		row[i] = xlsxCell{value: c, style: -1, isStr: true} // -1 = header style
	}
	w.rows = append(w.rows, row)
	if len(cells) > w.cols {
		w.cols = len(cells)
	}
}

func (w *worksheet) addRow(style int, cells ...string) {
	row := make([]xlsxCell, len(cells))
	for i, c := range cells {
		row[i] = xlsxCell{value: c, style: style, isStr: true}
	}
	w.rows = append(w.rows, row)
	if len(cells) > w.cols {
		w.cols = len(cells)
	}
}

func (w *worksheet) toXML(ss *sharedStrings) string {
	rowsXML := ""
	for _, row := range w.rows {
		rowsXML += "<row>"
		for i, cell := range row {
			si := ss.add(cell.value)
			s := cell.style
			if s < 0 {
				s = 4 // header
			}
			rowsXML += fmt.Sprintf(`<c r="%s%d" s="%d" t="s"><v>%d</v></c>`, colName(i), 1, s, si)
		}
		rowsXML += "</row>"
	}

	// Simple auto-filter + frozen panes for sheet1
	frozen := ""
	if len(w.rows) > 1 {
		frozen = `<sheetViews><sheetView tabSelected="1" workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>`
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">%s<sheetData>%s</sheetData><autoFilter ref="A1:%s%d"/></worksheet>`,
		frozen, rowsXML, colName(w.cols-1), len(w.rows))
}

func colName(i int) string {
	if i < 26 {
		return string(rune('A' + i))
	}
	return string(rune('A'+i/26-1)) + string(rune('A'+i%26))
}

func xmlEscape(s string) string {
	b := new(bytes.Buffer)
	xml.EscapeText(b, []byte(s))
	return b.String()
}

// XLSX template files
const xlsxContentTypes = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/worksheets/sheet2.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
  <Override PartName="/xl/sharedStrings.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"/>
</Types>`

const xlsxRels = `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Target="xl/workbook.xml" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"/>
</Relationships>`

const xlsxWorkbook = `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="审计日志" sheetId="1" r:id="rId1"/>
    <sheet name="完整性声明" sheetId="2" r:id="rId2"/>
  </sheets>
</workbook>`

const xlsxWorkbookRels = `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Target="worksheets/sheet1.xml" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet"/>
  <Relationship Id="rId2" Target="worksheets/sheet2.xml" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet"/>
  <Relationship Id="rId3" Target="styles.xml" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles"/>
  <Relationship Id="rId4" Target="sharedStrings.xml" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings"/>
</Relationships>`

const xlsxStyles = `<?xml version="1.0" encoding="UTF-8"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="5">
    <font><sz val="11"/><name val="Microsoft YaHei"/></font>
    <font><sz val="11"/><color rgb="FF006100"/><name val="Microsoft YaHei"/></font>
    <font><sz val="11"/><color rgb="FF9C6500"/><name val="Microsoft YaHei"/></font>
    <font><sz val="11"/><color rgb="FF9C0006"/><name val="Microsoft YaHei"/></font>
    <font><sz val="12"/><b/><color rgb="FFFFFFFF"/><name val="Microsoft YaHei"/></font>
  </fonts>
  <fills count="5">
    <fill><patternFill patternType="none"/></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FFC6EFCE"/></patternFill></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FFFFEB9C"/></patternFill></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FFFFC7CE"/></patternFill></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FF4472C4"/></patternFill></fill>
  </fills>
  <borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders>
  <cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
  <cellXfs count="5">
    <xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
    <xf numFmtId="0" fontId="1" fillId="1" borderId="0" xfId="0" applyFont="1" applyFill="1"/>
    <xf numFmtId="0" fontId="2" fillId="2" borderId="0" xfId="0" applyFont="1" applyFill="1"/>
    <xf numFmtId="0" fontId="3" fillId="3" borderId="0" xfId="0" applyFont="1" applyFill="1"/>
    <xf numFmtId="0" fontId="4" fillId="4" borderId="0" xfId="0" applyFont="1" applyFill="1"/>
  </cellXfs>
</styleSheet>`

func filterDescription(filter *model.AuditLogFilter) string {
	parts := []string{}
	if filter.Category != "" {
		parts = append(parts, fmt.Sprintf("category=%s", filter.Category))
	}
	if filter.Level != "" {
		parts = append(parts, fmt.Sprintf("level=%s", filter.Level))
	}
	if filter.Action != "" {
		parts = append(parts, fmt.Sprintf("action=%s", filter.Action))
	}
	if !filter.StartTime.IsZero() {
		parts = append(parts, fmt.Sprintf("from=%s", filter.StartTime.Format("2006-01-02")))
	}
	if !filter.EndTime.IsZero() {
		parts = append(parts, fmt.Sprintf("to=%s", filter.EndTime.Format("2006-01-02")))
	}
	if len(parts) == 0 {
		return "全部"
	}
	return fmt.Sprintf("%v", parts)
}

func uuidOrEmpty(id *uid.UID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

func signBody(bodyHash []byte) string {
	// 使用固定密钥签名，与审计日志数据签名独立
	key := []byte("contful-audit-export-v1")
	mac := hmac.New(sha256.New, key)
	mac.Write(bodyHash)
	return hex.EncodeToString(mac.Sum(nil))
}

// Helper: 从 HTTP 请求中提取 IP 和 User-Agent
func getClientInfo(r *http.Request) (string, string) {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}

	ua := r.Header.Get("User-Agent")
	return ip, ua
}
