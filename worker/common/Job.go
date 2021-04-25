package common

import (
	"encoding/json"
)

// 定时任务
type Job struct {
	Name string `json:"name"`
	Command string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

