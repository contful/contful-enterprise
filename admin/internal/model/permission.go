// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"time"

	"github.com/contful/contful/admin/pkg/uid"
)

// PermissionGroup 权限分组
type PermissionGroup struct {
	ID          uid.UID `json:"id" gorm:"primaryKey;default:gen_random_uuid()"`
	GroupKey    string    `json:"group_key" gorm:"type:varchar(50);unique;not null"`
	Label       string    `json:"label" gorm:"type:varchar(100);not null"`
	LabelEn     string    `json:"label_en" gorm:"type:varchar(100)"`
	SortOrder   int       `json:"sort_order" gorm:"not null;default:0"`
	CreatedTime time.Time `json:"created_time" gorm:"autoCreateTime"`
}

func (PermissionGroup) TableName() string {
	return "contful_system_permission_groups"
}

// Permission 权限项
type Permission struct {
	ID          uid.UID `json:"id" gorm:"primaryKey;default:gen_random_uuid()"`
	GroupID     uid.UID `json:"group_id" gorm:"not null;index"`
	Action      string    `json:"action" gorm:"type:varchar(50);not null"`
	Label       string    `json:"label" gorm:"type:varchar(100);not null"`
	LabelEn     string    `json:"label_en" gorm:"type:varchar(100)"`
	SortOrder   int       `json:"sort_order" gorm:"not null;default:0"`
	CreatedTime time.Time `json:"created_time" gorm:"autoCreateTime"`

	// 关联
	Group *PermissionGroup `json:"group,omitempty" gorm:"foreignKey:GroupID;references:ID"`
}

func (Permission) TableName() string {
	return "contful_system_permissions"
}
