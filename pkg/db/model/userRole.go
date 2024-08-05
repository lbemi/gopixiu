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
	"github.com/caoyingjunz/pixiu/pkg/db/model/pixiu"

	"time"

	"gorm.io/gorm"
)

func init() {
	register(&UserRole{})
}

type UserRole struct {
	pixiu.Model

	RoleID int64 `gorm:"primaryKey;column:role_id;type:int(11)" json:"role_id"`
	UserID int64 `gorm:"primaryKey;column:user_id;type:int(11)" json:"user_id"`
}

func (*UserRole) TableName() string {
	return "user_roles"
}

func (u *UserRole) BeforeCreate(*gorm.DB) error {
	u.GmtCreate = time.Now()
	u.GmtModified = time.Now()
	return nil
}

func (u *UserRole) BeforeUpdate(*gorm.DB) error {
	u.GmtModified = time.Now()
	u.ResourceVersion++
	return nil
}
