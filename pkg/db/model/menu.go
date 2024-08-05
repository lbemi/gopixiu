/*
Copyright 2024 The Pixiu Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package model

import (
	"time"

	"github.com/caoyingjunz/pixiu/pkg/db/model/pixiu"
	"gorm.io/gorm"
)

// Menu 菜单
type Menu struct {
	pixiu.Model

	Status      CommonStatus `gorm:"column:status;type:tinyint(1);not null;default:1;comment:状态(1:启用 2:不启用)" json:"status" `                       // 状态(1:启用 2:不启用)
	Description string       `gorm:"type:text" json:"description"`                                                                                 // 备注
	ParentID    uint64       `gorm:"column:parentID;not null;comment:父级ID" json:"parentID" form:"parent_id"`                                       // 父级ID
	Path        string       `gorm:"column:path;size:128;comment:菜单URL" json:"path,omitempty" form:"path"`                                         // 菜单URL
	Component   string       `gorm:"column:component;" json:"component"`                                                                           // 前端组件
	Name        string       `gorm:"column:name;size:128;not null;comment:菜单名称" json:"name" form:"name"`                                           // 菜单名称
	Sequence    int          `gorm:"column:sequence;not null;comment:排序值" json:"sequence" form:"sequence"`                                         // 排序值
	MenuType    int8         `gorm:"column:menuType;type:tinyint(1);not null;comment:菜单类型(1 左侧菜单,2 按钮, 3 非展示权限)" json:"menuType" form:"menu_type"` // 菜单类型 1 左侧菜单,2 按钮, 3 非展示权限
	Redirect    string       `json:"redirect" gorm:"column:redirect;size:128;"`
	Method      string       `gorm:"column:method;size:32;not null;comment:操作类型 none/GET/POST/PUT/DELETE" json:"method,omitempty" form:"method"` // 操作类型 none/GET/POST/PUT/DELETE
	Code        string       `gorm:"column:code;size:128;not null;" json:"code"`                                                                 // 前端鉴权code 例： user:role:add, user:role:delete
	Group       string       `gorm:"column:group;size:256;not null;" json:"group"`
	Meta        `json:"meta"`
	Children    []Menu `gorm:"-" json:"children"`
}

type Meta struct {
	Title       string `json:"title" gorm:"column:title;comment:标题"`
	IsLink      string `json:"is_link" gorm:"column:is_link;type:varchar(256)"`
	IsHide      bool   `json:"is_hide" gorm:"column:is_hide"`
	IsKeepAlive bool   `json:"is_keepalive" gorm:"column:is_keepalive"`
	IsAffix     bool   `json:"is_affix" gorm:"column:is_affix"`
	IsIframe    bool   `json:"is_iframe" gorm:"column:is_iframe"`
	Icon        string `gorm:"column:icon;size:128;comment:icon图标" json:"icon" form:"icon"`
}

// TableName 表名
func (m *Menu) TableName() string {
	return "menus"
}

// BeforeCreate 添加前
func (m *Menu) BeforeCreate(*gorm.DB) error {
	m.GmtCreate = time.Now()
	m.GmtModified = time.Now()
	return nil
}

// BeforeUpdate 更新前
func (m *Menu) BeforeUpdate(tx *gorm.DB) error {
	m.GmtModified = time.Now()
	m.ResourceVersion = m.ResourceVersion + 1
	return nil
}
