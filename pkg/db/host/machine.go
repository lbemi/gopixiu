package host

import (
	"context"
	"github.com/caoyingjunz/gopixiu/pkg/db/model"
	"gorm.io/gorm"
)

type machine struct {
	db *gorm.DB
}

func NewMachine(DB *gorm.DB) MachineInterface {
	return &machine{db: DB}
}

type MachineInterface interface {
	Create(ctx context.Context, machine *model.Machine) error
	Delete(ctx context.Context, machineId int64) error
	Update(ctx context.Context, machineId int64, machine *model.Machine) error
	List(ctx context.Context) (machines *[]model.Machine, err error)
	GetByMachineId(ctx context.Context, machineId int64) (machine *model.Machine, err error)
	UpdateFiledStatus(ctx context.Context, machineId int64, updateFiled string, status int8) error
}

func (m *machine) Create(ctx context.Context, machine *model.Machine) error {
	return m.db.Create(machine).Error
}

func (m *machine) Delete(ctx context.Context, machineId int64) error {
	return m.db.Where("id = ?", machineId).Delete(&model.Machine{}).Error
}

func (m *machine) Update(ctx context.Context, machineId int64, machine *model.Machine) error {
	return m.db.Where("id = ?", machineId).Updates(machine).Error
}

func (m *machine) List(ctx context.Context) (machines *[]model.Machine, err error) {
	err = m.db.Model(&model.Machine{}).Find(&machines).Error
	return
}

func (m *machine) GetByMachineId(ctx context.Context, machineId int64) (machine *model.Machine, err error) {
	err = m.db.Where("id = ?", machineId).Find(&machine).Error
	return
}

func (m *machine) UpdateFiledStatus(ctx context.Context, machineId int64, updateFiled string, status int8) error {
	return m.db.Where("id = ?", machineId).Update(updateFiled, status).Error
}
