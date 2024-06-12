/*
Copyright 2021 The Pixiu Authors.

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

package client

import (
	"sync"
	"time"

	"github.com/caoyingjunz/pixiu/pkg/db/model"
)

type TaskCache struct {
	sync.RWMutex
	items map[int64]*PlanInfo
}
type PlanInfo struct {
	Tasks    map[string]*Task
	Status   model.PlanStep
	TaskCh   chan *struct{}
	ErrCh    chan error
	ResultCh chan Task
}

type Task struct {
	Name    string           `json:"name"`
	PlanId  int64            `json:"plan_id"`
	Status  model.TaskStatus `json:"Status"`
	Message string           `json:"message"`
	StartAt time.Time        `json:"start_at"`
	EndAt   time.Time        `json:"end_at"`
}

func NewTaskCache() *TaskCache {
	return &TaskCache{
		items: map[int64]*PlanInfo{},
	}
}

func (t *TaskCache) GetPlainInfo(planId int64) (*PlanInfo, bool) {
	t.RLock()
	defer t.RUnlock()

	plain, ok := t.items[planId]
	return plain, ok
}

func (t *TaskCache) GetPlainTaskInfo(planId int64, taskName string) (*Task, bool) {
	t.RLock()
	defer t.RUnlock()

	status, ok := t.GetPlainInfo(planId)
	if !ok {
		return &Task{}, false
	}

	task, ok := status.Tasks[taskName]
	return task, ok
}

func (t *TaskCache) SetPlainInfo(planId int64, p *PlanInfo) {
	t.Lock()
	defer t.Unlock()
	if t.items == nil {
		t.items = map[int64]*PlanInfo{}
	}
	t.items[planId] = p
}
func (t *TaskCache) SetPlainTaskInfo(planId int64, taskInfo *Task) {
	t.Lock()
	defer t.Unlock()

	if t.items == nil {
		t.items = map[int64]*PlanInfo{}
	}
	plain, ok := t.GetPlainInfo(planId)
	if !ok {
		plain = &PlanInfo{Tasks: map[string]*Task{}}
	}
	plain.Tasks[taskInfo.Name] = taskInfo
	t.SetPlainInfo(planId, plain)
}

func (t *TaskCache) Delete(planId int64) {
	t.RLock()
	defer t.RUnlock()

	delete(t.items, planId)
}

func (t *TaskCache) Clear() {
	t.RLock()
	defer t.RUnlock()

	t.items = map[int64]*PlanInfo{}
}
