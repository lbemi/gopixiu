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
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/pixiu/pkg/client"
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

	for _, handler := range handlers {
		taskResult := &client.TaskResult{
			PlanId:  planId,
			Name:    handler.Name(),
			StartAt: time.Now(),
			EndAt:   time.Now(),
			Status:  model.UnStartPlanStatus,
			Step:    task.Step(),
			Message: "",
		}
		//初始化缓存
		fmt.Printf("创建%s 缓存\n", handler.Name())
		taskCache.SetTaskResult(planId, taskResult)
	}
	p.taskQueue = make(chan Handler, len(handlers))
	for _, handler := range handlers {
		p.taskQueue <- handler
	}
	close(p.taskQueue)
	taskCh := make(chan client.TaskResult, len(handlers))
	taskCache.SetCh(planId, taskCh)
	go p.syncTasks(planId)
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
	taskResults, ok := taskCache.GetResult(planId)
	if !ok {
		return nil
	}
	//批量同步缓存到数据库
	for index, taskResult := range taskResults {
		fmt.Println("步骤", index, " : ", *taskResult)
		// if err := p.factory.Plan().UpdateTask(context.TODO(), planId, taskResult.Name, map[string]interface{}{
		// 	"status": taskResult.Status, "message": taskResult.Err.Error(),
		// }); err != nil {
		// 	return err
		// }
		// time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (p *plan) syncTasks(planId int64) {
	errorCh := make(chan error)
	for i := 0; i < len(p.taskQueue); i++ {
		p.mutex.Lock()
		task, ok := <-p.taskQueue
		if !ok {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		resultCh, ok := taskCache.GetCh(task.GetPlanId())
		if !ok {
			resultCh = make(chan client.TaskResult, len(p.taskQueue))
		}
		go p.handlerTask(task, errorCh, resultCh)
	}

	for err := range errorCh {
		klog.Errorf("任务运行失败: %v", err.Error())
		close(errorCh)
		taskCache.CloseCh(planId)
		err = p.syncStatus(planId)
		if err != nil {
			klog.Errorf("failed to sync plan(%d) task status: %v", planId, err)
			return
		}
	}

}

func (p *plan) handlerTask(task Handler, errorCh chan error, resultCh chan client.TaskResult) {
	defer p.mutex.Unlock()
	planId := task.GetPlanId()
	name := task.Name()

	// 获取当前任务缓存信息
	taskResult, ok := taskCache.GetTaskResult(planId, name)
	if !ok {
		// 不存在则新建
		taskResult = &client.TaskResult{
			PlanId:  planId,
			Name:    name,
			StartAt: time.Now(),
			EndAt:   time.Now(),
			Status:  model.UnStartPlanStatus,
			Message: "",
		}
		//初始化缓存
		taskCache.SetTaskResult(planId, taskResult)
	}

	// 设置启动任务时间
	taskResult.StartAt = time.Now()
	taskResult.Status = model.RunningPlanStatus
	klog.Infof("starting plan(%d) task(%s)", planId, name)

	step := task.Step()
	runErr := task.Run()
	if runErr != nil {
		step = model.FailedPlanStep
		taskResult.Message = runErr.Error()
		taskResult.EndAt = time.Now()
		taskResult.Status = model.FailedPlanStatus
		taskResult.Step = step
		resultCh <- *taskResult
		errorCh <- runErr
		return
	}

	taskResult.EndAt = time.Now()
	taskResult.Status = model.SuccessPlanStatus
	taskResult.Step = step
	resultCh <- *taskResult
	klog.Infof("completed plan(%d) task(%s),result: %v,len: %d", planId, name, taskResult, len(resultCh))

	defer func() {
		//todo 用指针了,是不是就不需要再去保存了?
		taskCache.SetCh(planId, resultCh)
		taskCache.SetTaskResult(planId, taskResult)
	}()
}
