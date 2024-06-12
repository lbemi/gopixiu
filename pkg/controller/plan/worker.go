/*
Copyright 2021 The Pixiu Authors.

Licensed under the Apache License, Version 2.0 (phe "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plan

import (
	"context"
	"github.com/caoyingjunz/pixiu/pkg/client"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/pixiu/pkg/db/model"
	"github.com/caoyingjunz/pixiu/pkg/util/errors"
)

type Handler interface {
	GetPlanId() int64

	Name() string         // 检查项名称
	Step() model.PlanStep // 未开始，运行中，异常和完成
	Run() error           // 执行
}

type handlerTask struct {
	data TaskData
}

func (t handlerTask) GetPlanId() int64     { return t.data.PlanId }
func (t handlerTask) Step() model.PlanStep { return model.RunningPlanStep }

func newHandlerTask(data TaskData) handlerTask {
	return handlerTask{data: data}
}

func (p *plan) Run(ctx context.Context, workers int) error {
	klog.Infof("Starting Plan Manager")
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, p.worker, time.Second)
	}
	return nil
}

func (p *plan) worker(ctx context.Context) {
	for p.process(ctx) {
	}
}

func (p *plan) process(ctx context.Context) bool {
	key, quit := taskQueue.Get()
	if quit {
		return false
	}
	defer taskQueue.Done(key)

	p.syncHandler(ctx, key.(int64))
	return true
}

type TaskData struct {
	PlanId int64
	Config *model.Config
	Nodes  []model.Node
}

func (t TaskData) validate() error {
	return nil
}

func (p *plan) getTaskData(ctx context.Context, planId int64) (TaskData, error) {
	nodes, err := p.factory.Plan().ListNodes(ctx, planId)
	if err != nil {
		return TaskData{}, err
	}
	cfg, err := p.factory.Plan().GetConfigByPlan(ctx, planId)
	if err != nil {
		return TaskData{}, err
	}

	return TaskData{
		PlanId: planId,
		Config: cfg,
		Nodes:  nodes,
	}, nil
}

// 实际处理函数
// 处理步骤:
// 1. 检查部署参数是否符合要求
// 2. 渲染环境
// 3. 执行部署
// 4. 部署后环境清理
func (p *plan) syncHandler(ctx context.Context, planId int64) {
	klog.Infof("starting plan(%d) task", planId)

	taskData, err := p.getTaskData(ctx, planId)
	if err != nil {
		klog.Errorf("failed to get task data: %v", err)
		return
	}

	task := newHandlerTask(taskData)
	handlers := []Handler{
		Check{handlerTask: task},
		Render{handlerTask: task},
		BootStrap{handlerTask: task},
		Deploy{handlerTask: task},
		DeployNode{handlerTask: task},
		DeployChart{handlerTask: task},
	}
	err = p.createPlanTasksIfNotExist(handlers...)
	if err != nil {
		klog.Errorf("failed to create plan(%d) tasks: %v", planId, err)
		return
	}
	p.taskQueue = make(chan Handler, len(handlers))
	for _, handler := range handlers {
		p.taskQueue <- handler
	}
	close(p.taskQueue)
	taskCh := make(chan client.TaskResult, len(handlers))
	taskCache.Set(planId, taskCh)
	go p.syncTasks()
}

func (p *plan) createPlanTasksIfNotExist(tasks ...Handler) error {
	for _, task := range tasks {
		planId := task.GetPlanId()
		name := task.Name()
		step := task.Step()

		_, err := p.factory.Plan().GetTaskByName(context.TODO(), planId, name)
		// 存在则直接返回
		if err == nil {
			return nil
		}
		if err != nil {
			// 非不存在报错则报异常
			if !errors.IsRecordNotFound(err) {
				klog.Infof("failed to get plan(%d) tasks(%s) for first created: %v", planId, name, err)
				return err
			}
		}

		// 不存在记录则新建
		if _, err = p.factory.Plan().CreatTask(context.TODO(), &model.Task{
			Name:   name,
			PlanId: planId,
			Step:   step,
			Status: model.UnStartPlanStatus,
		}); err != nil {
			klog.Errorf("failed to init plan(%d) task(%s): %v", planId, name, err)
			return err
		}
	}

	return nil
}

// 同步任务状态
// 任务启动时设置为运行中，结束时同步为结束状态(成功或者失败)
// TODO: 后续优化
func (p *plan) syncStatus(planId int64) error {
	return nil
}

func (p *plan) syncTasks() {
	errorCh := make(chan error)
	for i := 0; i < len(p.taskQueue); i++ {
		p.mutex.Lock()
		task, ok := <-p.taskQueue
		if !ok {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		resultCh, ok := taskCache.Get(task.GetPlanId())
		if !ok {
			//  不存在则创建缓存
			resultCh = make(chan client.TaskResult, len(p.taskQueue))
		}

		go func() {
			taskResult := newTaskResult()
			planId := task.GetPlanId()
			name := task.Name()
			taskResult.PlanId = planId
			taskResult.Name = task.Name()
			klog.Infof("starting plan(%d) task(%s)", planId, name)
			// TODO: 通过闭包方式优化
			if err := p.factory.Plan().UpdateTask(context.TODO(), planId, name, map[string]interface{}{
				"status": model.RunningPlanStatus, "message": "",
			}); err != nil {
				taskResult.Err = err
				taskResult.EndAt = time.Now()
				resultCh <- *taskResult
				errorCh <- err
				return
			}

			status := model.SuccessPlanStatus
			step := task.Step()
			message := ""

			// 执行检查
			runErr := task.Run()
			if runErr != nil {
				status = model.FailedPlanStatus
				step = model.FailedPlanStep
				message = runErr.Error()
				taskResult.Err = runErr
				taskResult.EndAt = time.Now()
				resultCh <- *taskResult
				if err := p.factory.Plan().UpdateTask(context.TODO(), planId, name, map[string]interface{}{
					"status": status, "message": message, "step": step,
				}); err != nil {
					return
				}
				errorCh <- runErr
				return
			}

			// 执行完成之后更新状态
			if err := p.factory.Plan().UpdateTask(context.TODO(), planId, name, map[string]interface{}{
				"status": status, "message": message, "step": step,
			}); err != nil {
				taskResult.Err = err
				taskResult.EndAt = time.Now().UTC()
				resultCh <- *taskResult
				errorCh <- err
				return
			}

			klog.Infof("completed plan(%d) task(%s)", planId, name)
			taskResult.EndAt = time.Now().UTC()
			resultCh <- *taskResult
			klog.Infof("completed plan(%d) task(%s),result: %v,len: %d", planId, name, taskResult, len(resultCh))
			defer p.mutex.Unlock()
			defer taskCache.Set(planId, resultCh)
		}()
	}

	select {
	case err := <-errorCh:
		klog.Errorf("-------task error: %v", err.Error())
		close(errorCh)
	default:
	}

	//// 初始化记录
	//if err := p.createPlanTasksIfNotExist(p.taskQueue...); err != nil {
	//	p.resultCh <- TaskResult{Err: err}
	//	close(p.resultCh)
	//}
	//
	//var wg sync.WaitGroup

	// 确保task的执行顺序

	//var runTask func(Handler) = func(task Handler) {
	//	planId := task.GetPlanId()
	//	name := task.Name()
	//	klog.Infof("starting plan(%d) task(%s)", planId, name)
	//
	//	// TODO: 通过闭包方式优化
	//	if err := p.factory.Plan().UpdateTask(context.TODO(), planId, name, map[string]interface{}{
	//		"status": model.RunningPlanStatus, "message": "",
	//	}); err != nil {
	//		p.mutex.Lock()
	//		resultCh <- TaskResult{PlanId: planId, TaskId: 0, Err: err}
	//		p.mutex.Unlock()
	//		return
	//	}
	//
	//	status := model.SuccessPlanStatus
	//	step := task.Step()
	//	message := ""
	//
	//	// 执行检查
	//	runErr := task.Run()
	//	if runErr != nil {
	//		status = model.FailedPlanStatus
	//		step = model.FailedPlanStep
	//		message = runErr.Error()
	//		// 如果某个task执行失败，则整个task结束
	//	}
	//
	//	// 执行完成之后更新状态
	//	if err := p.factory.Plan().UpdateTask(context.TODO(), planId, name, map[string]interface{}{
	//		"status": status, "message": message, "step": step,
	//	}); err != nil {
	//		p.mutex.Lock()
	//		resultCh <- TaskResult{PlanId: planId, TaskId: 0, Err: err}
	//		p.mutex.Unlock()
	//		return
	//	}
	//
	//	klog.Infof("completed plan(%d) task(%s)", planId, name)
	//	p.mutex.Lock()
	//	resultCh <- TaskResult{PlanId: planId, TaskId: 0, Err: runErr}
	//	p.mutex.Unlock()
	//}

	//for _, task := range tasks {
	//	wg.Add(1)
	//	go runTask(task)
	//}
	//
	//go func() {
	//	wg.Wait()
	//	close(resultCh)
	//}()
}
