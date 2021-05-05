package common

import (
	"context"
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

// 定时任务
type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

// 提取worker的IP
func ExtractWorkerIP(regKey string) string {
	return strings.TrimPrefix(regKey, JOB_WORKER_DIR)
}

// 从 /cron/killer/job10提取job10
func ExtractKillerName(killerKey string) string {
	return strings.TrimPrefix(killerKey, JOB_KILLER_DIR)
}

// 从etcd的key中提取任务名
// /cron/jobs/job10抹掉/cron/jobs/
func ExtractJobName(jobKey string) (string) {
	return strings.TrimPrefix(jobKey, JOB_SAVE_DIR)
}

// 变化事件
type JobEvent struct {
	EventType int //  SAVE, DELETE
	Job       *Job
}

// 任务变化事件有2种：1）更新任务 2）删除任务
func BuildJobEvent(eventType int, job *Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

//反序列化操作
func UnpackJob(value []byte) (*Job, error) {
	var (
		job *Job
		err error
	)
	job = &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return nil, err
	}
	return job, nil
}

// 任务调度计划
type JobSchedulePlan struct {
	Job      *Job                 // 要调度的任务信息
	Expr     *cronexpr.Expression // 解析好的cronexpr表达式
	NextTime time.Time            // 下次调度时间
}

//构建执行计划
func BuildJobSchedulePlan(job *Job) (*JobSchedulePlan, error) {
	var (
		expression *cronexpr.Expression
		err        error
	)

	if expression, err = cronexpr.Parse(job.CronExpr); err != nil {
		return nil, err
	}

	//生成调度计划
	jobSchedulePlan := &JobSchedulePlan{
		Job:      job,
		Expr:     expression,
		NextTime: expression.Next(time.Now()),
	}
	return jobSchedulePlan, nil
}

//任务执行状态
type JobExecuteInfo struct {
	Job *Job // 任务信息
	PlanTime time.Time // 理论上的调度时间
	RealTime time.Time // 实际的调度时间
	CancelCtx context.Context // 任务command的context
	CancelFunc context.CancelFunc//  用于取消command执行的cancel函数
}

//构造执行状态信息
func BuildJobExecuteInfo(jobSchedulePlan *JobSchedulePlan)*JobExecuteInfo  {
	jobExecuteInfo := &JobExecuteInfo{
		Job:jobSchedulePlan.Job,
		PlanTime:jobSchedulePlan.NextTime, //  计算调度时间
		RealTime:time.Now(),               //  真实调度时间
	}

	jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return jobExecuteInfo
}

// 任务执行结果
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo	// 执行状态
	Output []byte // 脚本输出
	Err error // 脚本错误原因
	StartTime time.Time // 启动时间
	EndTime time.Time // 结束时间
}

// 任务执行日志
type JobLog struct {
	JobName string `json:"jobName" bson:"jobName"` // 任务名字
	Command string `json:"command" bson:"command"` // 脚本命令
	Err string `json:"err" bson:"err"` // 错误原因
	Output string `json:"output" bson:"output"`	// 脚本输出
	PlanTime int64 `json:"planTime" bson:"planTime"` // 计划开始时间
	ScheduleTime int64 `json:"scheduleTime" bson:"scheduleTime"` // 实际调度时间
	StartTime int64 `json:"startTime" bson:"startTime"` // 任务执行开始时间
	EndTime int64 `json:"endTime" bson:"endTime"` // 任务执行结束时间
}
// 日志批次
type LogBatch struct {
	Logs []interface{}	// 多条日志
}

// 任务日志过滤条件
type JobLogFilter struct {
	JobName string `bson:"jobName"`
}

// 任务日志排序规则
type SortLogByStartTime struct {
	SortOrder int `bson:"startTime"`	// {startTime: -1}
}