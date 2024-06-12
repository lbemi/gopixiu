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
	"github.com/caoyingjunz/pixiu/pkg/db/model"
	"sync"
)

type TaskCache struct {
	sync.RWMutex
	items map[int64]PlainInfo
}
type PlainInfo struct {
	tasks map[string]model.Task
}

func NewTaskCache() *TaskCache {
	return &TaskCache{
		items: map[int64]PlainInfo{},
	}
}

func (t *TaskCache) GetPlainInfo(plainId int64) (PlainInfo, bool) {
	t.RLock()
	defer t.RUnlock()

	plain, ok := t.items[plainId]
	return plain, ok
}

func (t *TaskCache) GetPlainTaskInfo(plainId int64, taskName string) (model.Task, bool) {
	t.RLock()
	defer t.RUnlock()

	status, ok := t.GetPlainInfo(plainId)
	if !ok {
		return model.Task{}, false
	}

	task, ok := status.tasks[taskName]
	return task, ok
}

func (t *TaskCache) SetPlainInfo(plainId int64, p PlainInfo) {
	t.Lock()
	defer t.Unlock()
	if t.items == nil {
		t.items = map[int64]PlainInfo{}
	}
	t.items[plainId] = p
}
func (t *TaskCache) SetPlainTaskInfo(plainId int64, taskInfo model.Task) {
	t.Lock()
	defer t.Unlock()

	if t.items == nil {
		t.items = map[int64]PlainInfo{}
	}
	plain, ok := t.GetPlainInfo(plainId)
	if !ok {
		plain = PlainInfo{tasks: map[string]model.Task{}}
	}
	plain.tasks[taskInfo.Name] = taskInfo
	t.SetPlainInfo(plainId, plain)
}

func (t *TaskCache) Delete(plainId int64) {
	t.RLock()
	defer t.RUnlock()

	delete(t.items, plainId)
}

func (t *TaskCache) Clear() {
	t.RLock()
	defer t.RUnlock()

	t.items = map[int64]PlainInfo{}
}
