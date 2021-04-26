package Scheduler

import (
	"fmt"
	"learn-crontab/master/common"
	"time"
)

// 任务调度
type Scheduler struct {
	jobEventChan      chan *common.JobEvent              // etcd任务事件队列
	jobPlanTable      map[string]*common.JobSchedulePlan // 任务调度计划表
	jobExecutingTable map[string]*common.JobExecuteInfo  // 任务执行表
	jobResultChan     chan *common.JobExecuteResult      // 任务结果队列
}

var (
	G_scheduler *Scheduler
)



//处理任务事件
func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulePlan *common.JobSchedulePlan
		jobExisted      bool
		err             error
	)
	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE: //保存任务事件
		if jobSchedulePlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
			return
		}
		//放入执行计划表
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case common.JOB_EVENT_DELETE: //删除事件
		if jobSchedulePlan, jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		}
	}
}

// 尝试执行任务
func (scheduler *Scheduler) TryStartJob(jobPlan *common.JobSchedulePlan) {
	// 调度 和 执行 是2件事情
	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting   bool
	)
	// 执行的任务可能运行很久, 1分钟会调度60次，但是只能执行1次, 防止并发！

	// 如果任务正在执行，跳过本次调度
	if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobPlan.Job.Name]; jobExecuting {
		return
	}

	// 构建执行状态信息
	jobExecuteInfo = common.BuildJobExecuteInfo(jobPlan)

	//保存执行状态
	scheduler.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo

	//执行任务
	fmt.Println("执行任务:", jobExecuteInfo.Job.Name, jobExecuteInfo.PlanTime, jobExecuteInfo.RealTime)
	G_executor.ExecuteJob(jobExecuteInfo)
}

//重新计算任务调度状态
func (scheduler *Scheduler) TrySchedule() time.Duration {
	var (
		jobPlan        *common.JobSchedulePlan
		nowTime        time.Time
		nearTime       *time.Time
		schedulerAfter time.Duration
	)

	//执行的计划map
	if len(scheduler.jobPlanTable) == 0 {
		schedulerAfter = time.Second * 1
		return schedulerAfter
	}

	//当前时间
	nowTime = time.Now()

	//1.遍历所有任务
	for _, jobPlan = range scheduler.jobPlanTable {
		//当前任务的下次任务执行时间在当前时间之前或者等于当前时间则实现对应的任务，14.10 > 14:11 下次任务执行
		if jobPlan.NextTime.Before(nowTime) || jobPlan.NextTime.Equal(nowTime) {
			//TODO 尝试执行任务
			scheduler.TryStartJob(jobPlan)
			jobPlan.NextTime = jobPlan.Expr.Next(nowTime) //更新下次时间
		}

		//统计最近一个要过期的任务时间
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}

	//下次调度间隔 (最近要执行的任务调度时间 - 当前时间)
	schedulerAfter = (*nearTime).Sub(nowTime)
	return schedulerAfter

}

//初始化对应的调度任务
func InitScheduler() error {
	G_scheduler = &Scheduler{
		jobEventChan:      make(chan *common.JobEvent, 1000),
		jobPlanTable:      make(map[string]*common.JobSchedulePlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan:     make(chan *common.JobExecuteResult, 100),
	}

	//启动调度协程
	go G_scheduler.scheduleLoop()
	return nil
}

//调度协程
func (scheduler *Scheduler) scheduleLoop() {
	var
	(
		jobEvent       *common.JobEvent
		schedulerAfter time.Duration
		schedulerTimer *time.Timer
		jobResult      *common.JobExecuteResult
	)

	//初始化一次1秒
	schedulerAfter = scheduler.TrySchedule()  //获取任务执行后执行的最近一个时间间隔

	//调度的定时器
	schedulerTimer = time.NewTimer(schedulerAfter)

	//定时任务
	for {
		select {
		case jobEvent = <- scheduler.jobEventChan: //监听任务变化事件
			//对内存中的维护的任务列表做增删改查
			scheduler.handleJobEvent(jobEvent)
		case <-schedulerTimer.C: //最近的任务过期了

		case jobResult = <- scheduler.jobResultChan: //监听任务的执行结果进行处理
			fmt.Println("任务执行结果:",string(jobResult.Output))
			//scheduler.handleJobEvent(jobResult)
		}

		//调度一次任务,获取调度的定时器间隔时间
		schedulerAfter = scheduler.TrySchedule()
		//重置调度间隔
		schedulerTimer.Reset(schedulerAfter)
	}

}

// 推送任务变化事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

// 回传任务执行结果
func (scheduler *Scheduler) PushJobResult(jobResult *common.JobExecuteResult) {
	scheduler.jobResultChan <- jobResult
}
