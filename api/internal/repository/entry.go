package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EntryRepository 条目数据访问层（只读）
type EntryRepository struct {
	db *gorm.DB
}

// NewEntryRepository 创建 Entry Repository
func NewEntryRepository(db *gorm.DB) *EntryRepository {
	return &EntryRepository{db: db}
}

// Field 字段定义（最小化，供 EntryValue Preload 用）
type Field struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"size:200;not null"`  // 字段标识名（即 slug）
	Label     string    `gorm:"size:200;not null"`  // 显示名称
	FieldType string    `gorm:"column:field_type;size:50;not null"`
}

func (Field) TableName() string {
	return "fields"
}

// JSONBValue JSONB 值类型（可序列化任意 JSON）
type JSONBValue map[string]interface{}

func (j *JSONBValue) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	}
	return nil
}

func (j JSONBValue) Value() (interface{}, error) {
	if j == nil {
		return "{}", nil
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// EntryValue 字段值
type EntryValue struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	EntryID   uuid.UUID  `gorm:"type:uuid;not null;index"`
	FieldID   uuid.UUID  `gorm:"type:uuid;not null;index"`
	Value     JSONBValue `gorm:"type:jsonb;not null"`
	Field     Field      `gorm:"foreignKey:FieldID"`
}

func (EntryValue) TableName() string {
	return "entry_values"
}

// Entry 内容条目
type Entry struct {
	ID            uuid.UUID    `gorm:"type:uuid;primaryKey"`
	ContentTypeID uuid.UUID    `gorm:"type:uuid;not null;index"`
	SiteID        uuid.UUID    `gorm:"type:uuid;not null;index"`
	Locale        string       `gorm:"size:20;not null;default:'zh-CN'"`
	Status        string       `gorm:"type:entry_status;not null;default:'draft'"`
	Version       int          `gorm:"not null;default:1"`
	SortWeight    int          `gorm:"not null;default:0"`
	SeoTitle      *string      `gorm:"column:seo_title;size:255"`
	SeoDescription *string     `gorm:"column:seo_description;type:text"`
	PublishedTime *time.Time   `gorm:"column:published_time;type:timestamptz"`
	CreatedTime   time.Time    `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime   time.Time    `gorm:"column:updated_time;autoUpdateTime"`
	DeletedTime   *time.Time   `gorm:"column:deleted_time;type:timestamptz"`
	Values        []EntryValue `gorm:"foreignKey:EntryID"`
}

func (Entry) TableName() string {
	return "entries"
}

// EntryListFilter 查询条件
type EntryListFilter struct {
	Locale    string // 语言，可选
	SortField string // 排序字段
	SortOrder string // 排序方向 asc/desc
}

// ListPublished 查询指定内容类型的已发布条目（携带字段值）
func (r *EntryRepository) ListPublished(ctx context.Context, siteID, contentTypeID uuid.UUID, filter EntryListFilter, page, pageSize int) ([]Entry, int64, error) {
	var entries []Entry
	var total int64

	query := r.db.WithContext(ctx).Model(&Entry{}).
		Where("site_id = ?", siteID).
		Where("content_type_id = ?", contentTypeID).
		Where("status = ?", "published").
		Where("deleted_time IS NULL")

	if filter.Locale != "" {
		query = query.Where("locale = ?", filter.Locale)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderClause := "sort_weight DESC, published_time DESC"
	if filter.SortField != "" {
		validFields := map[string]bool{
			"updated_time":   true,
			"created_time":   true,
			"published_time": true,
			"sort_weight":    true,
		}
		if validFields[filter.SortField] {
			order := "DESC"
			if filter.SortOrder == "asc" {
				order = "ASC"
			}
			orderClause = filter.SortField + " " + order
		}
	}

	offset := (page - 1) * pageSize
	err := query.
		Preload("Values", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Field")
		}).
		Order(orderClause).
		Offset(offset).
		Limit(pageSize).
		Find(&entries).Error
	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

// GetPublishedByID 获取单个已发布条目（携带字段值）
func (r *EntryRepository) GetPublishedByID(ctx context.Context, siteID uuid.UUID, entryID uuid.UUID) (*Entry, error) {
	var entry Entry
	err := r.db.WithContext(ctx).
		Where("id = ?", entryID).
		Where("site_id = ?", siteID).
		Where("status = ?", "published").
		Where("deleted_time IS NULL").
		Preload("Values", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Field")
		}).
		First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}
