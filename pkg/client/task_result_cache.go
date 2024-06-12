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
)

// TaskCache
// TODO: 临时实现，后续优化
type TaskCache struct {
	sync.RWMutex
	items map[int64]chan TaskResult
}

type TaskResult struct {
	PlanId  int64
	Name    string
	StartAt time.Time
	EndAt   time.Time
	Err     error
}

func NewTaskCache() *TaskCache {
	return &TaskCache{
		items: map[int64]chan TaskResult{},
	}
}

func (s *TaskCache) Get(planId int64) (chan TaskResult, bool) {
	s.RLock()
	defer s.RUnlock()

	t, ok := s.items[planId]
	return t, ok
}

func (s *TaskCache) Set(planId int64, result chan TaskResult) {
	s.RLock()
	defer s.RUnlock()

	if s.items == nil {
		s.items = map[int64]chan TaskResult{}
	}
	s.items[planId] = result
}

func (s *TaskCache) Delete(planId int64) {
	s.RLock()
	defer s.RUnlock()

	delete(s.items, planId)
}

func (s *TaskCache) Clear() {
	s.RLock()
	defer s.RUnlock()

	s.items = map[int64]chan TaskResult{}
}
