package common

import (
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
func ExtractWorkerIP(regKey string) (string) {
	return strings.TrimPrefix(regKey, JOB_WORKER_DIR)
}

// 从 /cron/killer/job10提取job10
func ExtractKillerName(killerKey string) (string) {
	return strings.TrimPrefix(killerKey, JOB_KILLER_DIR)
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
