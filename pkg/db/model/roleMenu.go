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

func init() {
	register(&RoleMenu{})
}

type RoleMenu struct {
	pixiu.Model

	RoleID int `gorm:"primaryKey;column:role_id;type:int(11)" json:"role_id"`
	MenuID int `gorm:"primaryKey;column:menu_id;type:int(11)" json:"menu_id"`
}

func (*RoleMenu) TableName() string {
	return "role_menu"
}

func (r *RoleMenu) BeforeCreate(*gorm.DB) error {
	r.GmtCreate = time.Now()
	r.GmtModified = time.Now()
	return nil
}

func (r *RoleMenu) BeforeUpdate(*gorm.DB) error {
	r.GmtModified = time.Now()
	r.ResourceVersion++
	return nil
}
