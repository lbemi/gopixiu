package core

import (
	"context"

	"gorm.io/gorm"

	"github.com/caoyingjunz/gopixiu/pkg/db"
	"github.com/caoyingjunz/gopixiu/pkg/db/model"
	"github.com/caoyingjunz/gopixiu/pkg/log"
)

type MachineGetter interface {
	Machine() MachineInterface
}

type machine struct {
	app     *pixiu
	factory db.ShareDaoFactory
}

func newMachine(c *pixiu) MachineInterface {
	return &machine{
		app:     c,
		factory: c.factory,
	}
}

// MachineInterface 主机操作接口
type MachineInterface interface {
	Create(ctx context.Context, machine *model.MachineReq) error
	Delete(ctx context.Context, machineId int64) error
	Update(ctx context.Context, machineId int64, machine *model.MachineReq) error
	List(ctx context.Context) (machines *[]model.Machine, err error)
	GetByMachineId(ctx context.Context, machineId int64) (machine *model.Machine, err error)
	UpdateFiledStatus(ctx context.Context, machineId int64, updateFiled string, status int8) error
	CheckMachineExist(ctx context.Context, machineId int64) bool
}

func (m *machine) Create(ctx context.Context, machine *model.MachineReq) error {
	return m.factory.Machine().Create(ctx, &model.Machine{
		Name:       machine.Name,
		Label:      machine.Label,
		Remark:     machine.Remark,
		Ip:         machine.Ip,
		Port:       machine.Port,
		Username:   machine.Username,
		AuthMethod: machine.AuthMethod,
		Password:   machine.Password,
		Status:     machine.Status,
		EnableSSH:  machine.EnableSSH,
	})
}

func (m *machine) Delete(ctx context.Context, machineId int64) error {
	err := m.factory.Machine().Delete(ctx, machineId)
	if err != nil {
		log.Logger.Error(err)
	}
	return err
}

func (m *machine) Update(ctx context.Context, machineId int64, machine *model.MachineReq) error {
	err := m.factory.Machine().Update(ctx, machineId, &model.Machine{
		Name:       machine.Name,
		Label:      machine.Label,
		Remark:     machine.Remark,
		Ip:         machine.Ip,
		Port:       machine.Port,
		Username:   machine.Username,
		AuthMethod: machine.AuthMethod,
		Password:   machine.Password,
		Status:     machine.Status,
		EnableSSH:  machine.EnableSSH,
	})
	if err != nil {
		log.Logger.Error(err)
	}
	return err
}

func (m *machine) List(ctx context.Context) (machines *[]model.Machine, err error) {
	machines, err = m.factory.Machine().List(ctx)
	if err != nil {
		log.Logger.Error(err)
	}
	return
}

func (m *machine) GetByMachineId(ctx context.Context, machineId int64) (machine *model.Machine, err error) {
	machine, err = m.factory.Machine().GetByMachineId(ctx, machineId)
	if err != nil {
		log.Logger.Error(err)
	}
	return
}

func (m *machine) UpdateFiledStatus(ctx context.Context, machineId int64, updateFiled string, status int8) error {
	err := m.factory.Machine().UpdateFiledStatus(ctx, machineId, updateFiled, status)
	if err != nil {
		log.Logger.Error(err)
	}
	return err
}

func (m *machine) CheckMachineExist(ctx context.Context, machineId int64) bool {
	_, err := m.factory.Machine().GetByMachineId(ctx, machineId)
	if err != nil || err == gorm.ErrRecordNotFound {
		return false
	}
	return true
}
