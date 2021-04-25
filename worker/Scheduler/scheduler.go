package Scheduler

import (
	"fmt"
	"learn-crontab/master/common"
	"time"
)

// 任务调度
type Scheduler struct {
	jobEventChan chan *common.JobEvent              //  etcd任务事件队列
	jobPlanTable map[string]*common.JobSchedulePlan // 任务调度计划表
	//jobExecutingTable map[string]*common.JobExecuteInfo // 任务执行表
	//jobResultChan chan *common.JobExecuteResult	// 任务结果队列
}

var (
	G_scheduler *Scheduler
)

func InitScheduler() error {
	G_scheduler = &Scheduler{
		jobEventChan: make(chan *common.JobEvent, 1000),
		jobPlanTable: make(map[string]*common.JobSchedulePlan),
	}

	//启动调度协程
	go G_scheduler.scheduleLoop()
	return nil
}

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

//重新计算任务调度状态
func (scheduler *Scheduler) TrySchedule() time.Duration {
	var (
		jobPlan        *common.JobSchedulePlan
		nowTime        time.Time
		nearTime       *time.Time
		schedulerAfter time.Duration
	)

	if len(scheduler.jobPlanTable) == 0 {
		schedulerAfter = time.Second * 1;
		return schedulerAfter
	}

	//当前时间
	nowTime = time.Now()

	//1.遍历所有任务
	for _, jobPlan = range scheduler.jobPlanTable {
		if jobPlan.NextTime.Before(nowTime) || jobPlan.NextTime.Equal(nowTime) {
			//TODO 尝试执行任务
			fmt.Println("执行任务:",jobPlan.Job.Name)
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

//调度协程
func (scheduler *Scheduler) scheduleLoop() {
	var
	(
		jobEvent *common.JobEvent
		schedulerAfter time.Duration
		schedulerTimer *time.Timer
	)

	//初始化一次1秒
	schedulerAfter = scheduler.TrySchedule()

	//调度的定时器
	schedulerTimer = time.NewTimer(schedulerAfter)

	//定时任务
	for {
		select {
		case jobEvent = <-scheduler.jobEventChan: //监听任务变化事件
			//对内存中的维护的任务列表做增删改查
			scheduler.handleJobEvent(jobEvent)

		case <- schedulerTimer.C: //最近的任务过期了
		}

		//调度一次任务
		schedulerAfter = scheduler.TrySchedule()
		//重置调度间隔
		schedulerTimer.Reset(schedulerAfter)
	}

}

// 推送任务变化事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}
