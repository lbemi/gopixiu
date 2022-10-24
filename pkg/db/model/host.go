package model

import (
	"github.com/caoyingjunz/gopixiu/pkg/db/gopixiu"
	"gorm.io/gorm"
	"time"
)

// Machine 机器信息
type Machine struct {
	gopixiu.Model

	Name       string `json:"name"`
	Label      string `json:"label"`       // 标签
	Remark     string `json:"remark"`      // 备注
	Ip         string `json:"ip"`          // IP地址
	Port       int    `json:"port"`        // 端口号
	Username   string `json:"username"`    // 用户名
	AuthMethod int8   `json:"auth_method"` // 授权认证方式
	Password   string `json:"-"`
	Status     int8   `json:"status"`     // 状态 1:启用；2:停用
	EnableSSH  int8   `json:"enable_ssh"` // 是否允许SSH 1:启用；2:停用
}

// BeforeCreate 添加前
func (m *Machine) BeforeCreate(*gorm.DB) error {
	m.GmtCreate = time.Now()
	m.GmtModified = time.Now()
	return nil
}

// BeforeUpdate 更新前
func (m *Machine) BeforeUpdate(*gorm.DB) error {
	m.GmtModified = time.Now()
	return nil
}

type MachineReq struct {
	Name       string `json:"name"`
	Label      string `json:"label"`       // 标签
	Remark     string `json:"remark"`      // 备注
	Ip         string `json:"ip"`          // IP地址
	Port       int    `json:"port"`        // 端口号
	Username   string `json:"username"`    // 用户名
	AuthMethod int8   `json:"auth_method"` // 授权认证方式
	Password   string `json:"password"`
	Status     int8   `json:"status"`     // 状态 1:启用；2:停用
	EnableSSH  int8   `json:"enable_ssh"` // 是否允许SSH 1:启用；2:停用
}
