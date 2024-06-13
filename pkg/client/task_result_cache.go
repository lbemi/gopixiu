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

// TaskCache
// TODO: 临时实现，后续优化
type TaskCache struct {
	sync.RWMutex
	items      map[int64]chan TaskResult        // 存储每个任务的通道
	taskResult map[int64]map[string]*TaskResult // 存储每个任务的结果
}

type TaskResult struct {
	PlanId  int64            `json:"plan_id"`
	Name    string           `json:"name"`
	StartAt time.Time        `json:"start_at"`
	EndAt   time.Time        `json:"end_at"`
	Status  model.TaskStatus `json:"status"`
	Step    model.PlanStep   `json:"step"`
	Message string           `json:"message"`
}

func NewTaskCache() *TaskCache {
	return &TaskCache{
		items:      map[int64]chan TaskResult{},
		taskResult: map[int64]map[string]*TaskResult{},
	}
}

func (s *TaskCache) GetCh(planId int64) (chan TaskResult, bool) {
	s.RLock()
	defer s.RUnlock()

	t, ok := s.items[planId]
	return t, ok
}
func (s *TaskCache) GetResult(planId int64) (map[string]*TaskResult, bool) {
	s.RLock()
	defer s.RUnlock()

	t, ok := s.taskResult[planId]
	return t, ok
}
func (s *TaskCache) GetTaskResult(planId int64, taskName string) (*TaskResult, bool) {
	s.RLock()
	defer s.RUnlock()

	t, ok := s.taskResult[planId][taskName]
	return t, ok
}
func (s *TaskCache) SetCh(planId int64, result chan TaskResult) {
	s.RLock()
	defer s.RUnlock()

	if s.items == nil {
		s.items = map[int64]chan TaskResult{}
	}
	if s.taskResult[planId] == nil {
		//	初始化plan的taskResult
		s.taskResult[planId] = map[string]*TaskResult{}
	}
	s.items[planId] = result
}

func (s *TaskCache) SetTaskResult(planId int64, result *TaskResult) {
	s.RLock()
	defer s.RUnlock()
	// 如果channel没有启动，直接返回,不做操作
	if s.items == nil {
		return
	}
	// 如果没有plan的taskResult，初始化
	if s.taskResult[planId] == nil {
		//	初始化plan的taskResult
		s.taskResult[planId] = map[string]*TaskResult{}
	}
	s.taskResult[planId][result.Name] = result
}

func (s *TaskCache) DeleteCh(planId int64) {
	s.RLock()
	defer s.RUnlock()

	delete(s.items, planId)
	//同步清空plan缓存
	delete(s.taskResult, planId)
}

func (s *TaskCache) Clear() {
	s.RLock()
	defer s.RUnlock()

	s.items = map[int64]chan TaskResult{}
	s.taskResult = map[int64]map[string]*TaskResult{}
}

func (s *TaskCache) CloseCh(planId int64) {
	s.RLock()
	defer s.RUnlock()
	close(s.items[planId])
}
