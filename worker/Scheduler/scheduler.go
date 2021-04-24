package Scheduler

import "learn-crontab/master/common"

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
	}

	//启动调度协程
	go G_scheduler.scheduleLoop()
	return nil
}

//处理任务事件
func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulePlan *common.JobSchedulePlan
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

	}

}

//调度协程
func (scheduler *Scheduler) scheduleLoop() {
	var jobEvent *common.JobEvent
	//定时任务
	for {
		select {
		case jobEvent = <-scheduler.jobEventChan: //监听任务变化事件
			//对任务




		}
	}

}

// 推送任务变化事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}
