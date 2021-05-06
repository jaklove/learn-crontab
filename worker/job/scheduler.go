package job

import (
	"fmt"
	"learn-crontab/worker/common"
	"time"
)

// 任务调度
type Scheduler struct {
	jobEventChan      chan *common.JobEvent              //  etcd任务事件队列
	jobPlanTable      map[string]*common.JobSchedulePlan // 任务调度计划表
	jobExecutingTable map[string]*common.JobExecuteInfo  // 任务执行表
	jobResultChan     chan *common.JobExecuteResult      // 任务结果队列
}

var (
	G_scheduler *Scheduler
)

func InitScheduler() error {
	G_scheduler = &Scheduler{
		jobEventChan:      make(chan *common.JobEvent, 1000),
		jobPlanTable:      make(map[string]*common.JobSchedulePlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan:     make(chan *common.JobExecuteResult, 1000),
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
		jobExecuteInfo  *common.JobExecuteInfo
		jobExecuting    bool
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
	case common.JOB_EVENT_KILL: //强杀任务事件
		fmt.Println("强杀任务")
		if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobEvent.Job.Name]; jobExecuting {
			fmt.Println("任务名字:",jobEvent.Job.Name)
			jobExecuteInfo.CancelFunc() // 触发command杀死shell子进程, 任务得到退出
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
			scheduler.TryStartJob(jobPlan)
			//fmt.Println("执行任务:",jobPlan.Job.Name)
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

func (scheduler *Scheduler) TryStartJob(jobPlan *common.JobSchedulePlan) {
	//调度 和 执行 是2件事情
	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting   bool
	)
	if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobPlan.Job.Name]; jobExecuting {
		fmt.Println("尚未退出,跳过执行:", jobPlan.Job.Name)
		return
	}

	// 构建执行状态信息
	jobExecuteInfo = common.BuildJobExecuteInfo(jobPlan)

	//保存信息状态
	scheduler.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo
	// 执行任务
	fmt.Println("执行任务:", jobExecuteInfo.Job.Name, jobExecuteInfo.PlanTime, jobExecuteInfo.RealTime)
	G_executor.ExecuteJob(jobExecuteInfo)

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
	schedulerAfter = scheduler.TrySchedule()

	//调度的定时器
	schedulerTimer = time.NewTimer(schedulerAfter)

	//定时任务
	for {
		select {
		case jobEvent = <-scheduler.jobEventChan: //监听任务变化事件
			//对内存中的维护的任务列表做增删改查
			scheduler.handleJobEvent(jobEvent)

		case <-schedulerTimer.C: //最近的任务过期了

		case jobResult = <-scheduler.jobResultChan:
			scheduler.handleJobResult(jobResult)
		}

		//调度一次任务
		schedulerAfter = scheduler.TrySchedule()
		//重置调度间隔
		schedulerTimer.Reset(schedulerAfter)
	}

}

//执行结果
func (scheduler *Scheduler) handleJobResult(jobResult *common.JobExecuteResult) {
	var (
		jobLog *common.JobLog
	)
	//删除执行状态
	delete(scheduler.jobExecutingTable, jobResult.ExecuteInfo.Job.Name)

	//生成日志
	if jobResult.Err != common.ERR_LOCK_ALREADY_REQUIRED{
		jobLog = &common.JobLog{
			JobName:jobResult.ExecuteInfo.Job.Name,
			Command:jobResult.ExecuteInfo.Job.Command,
			Output:string(jobResult.Output),
			PlanTime:jobResult.ExecuteInfo.PlanTime.UnixNano() /1000 / 1000,
			ScheduleTime:jobResult.ExecuteInfo.RealTime.UnixNano() / 1000 /1000,
			StartTime:jobResult.StartTime.UnixNano() / 1000 / 1000,
			EndTime:jobResult.EndTime.UnixNano() / 1000 / 1000,
		}

		if jobResult.Err != nil{
			jobLog.Err = jobResult.Err.Error()
		}else {
			jobLog.Err = ""
		}
		G_logSink.Append(jobLog)
	}

	fmt.Println("任务执行完成：",jobResult.ExecuteInfo.Job.Name,string(jobResult.Output),jobResult.Err)
}

// 推送任务变化事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

// 回传任务执行结果
func (scheduler *Scheduler) PushJobResult(jobResult *common.JobExecuteResult) {
	scheduler.jobResultChan <- jobResult
}
