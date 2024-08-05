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

	"gorm.io/gorm"

	"github.com/caoyingjunz/pixiu/pkg/db/model/pixiu"
)

func init() {
	register(&Role{})
}

type CommonStatus uint8

const (
	EnableStatus  CommonStatus = 1 //启用
	DisableStatus CommonStatus = 2 //禁用
)

type Role struct {
	pixiu.Model

	Name        string       `gorm:"index:idx_name" json:"name"`
	Description string       `gorm:"type:text" json:"description"`
	Sequence    int          `gorm:"type:int" json:"sequence"`
	ParentId    int64        `gorm:"type:int" json:"parent_id"`
	Status      CommonStatus `gorm:"type:tinyint;not null;default:1;comment:状态(1:启用 2:不启用)" json:"status"`
	Children    []*Role      `gorm:"-" json:"children"`
	Extension   string       `gorm:"type:text" json:"extension,omitempty"`
}

func (*Role) TableName() string {
	return "roles"
}

func (r *Role) BeforeCreate(*gorm.DB) error {
	r.GmtCreate = time.Now()
	r.GmtModified = time.Now()
	return nil
}

func (r *Role) BeforeUpdate(*gorm.DB) error {
	r.GmtModified = time.Now()
	r.ResourceVersion++
	return nil
}
