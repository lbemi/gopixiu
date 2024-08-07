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
	register(&UserResource{})
}

type AuthorityType string

const (
	Read    AuthorityType = "GET"
	Write   AuthorityType = "POST"
	Modifiy AuthorityType = "PUT"
	Delete  AuthorityType = "DELETE"
)

type UserResource struct {
	pixiu.Model
	UserID     int64         `gorm:"primaryKey;column:user_id;type:int(11)" json:"user_id"`
	ResourceId string        `gorm:"column:resource;type:varchar(255)" json:"resource"`
	Authority  AuthorityType `gorm:"type:varchar(10);not null;default:GET" json:"authority"`
}

func (u *UserResource) TableName() string {
	return "user_resources"
}

func (u *UserResource) BeforeCreate(*gorm.DB) error {
	u.GmtCreate = time.Now()
	u.GmtModified = time.Now()
	return nil
}

func (u *UserResource) BeforeUpdate(*gorm.DB) error {
	u.GmtModified = time.Now()
	u.ResourceVersion++
	return nil
}
