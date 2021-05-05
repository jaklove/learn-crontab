package job

import (
	"fmt"
	"learn-crontab/master/common"
	"learn-crontab/worker/pkg/lock"
	"math/rand"
	"os/exec"
	"time"
)

// 任务执行器
type Executor struct {

}

var (
	G_executor *Executor
)

func (executor *Executor)ExecuteJob(info *common.JobExecuteInfo)  {
	go func() {
		var (
			cmd *exec.Cmd
			err error
			output []byte
			result *common.JobExecuteResult
			jobLock *lock.JobLock
		)

		// 任务结果
		result = &common.JobExecuteResult{
			ExecuteInfo: info,
			Output: make([]byte, 0),
		}

		//初始化分布式锁
		jobLock = Worker_JobMgr.CreateJobLock(info.Job.Name)

		//任务开始时间
		result.StartTime = time.Now()

		// 随机睡眠(0~1s)
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		err = jobLock.TryLock()
		defer jobLock.Unlock()

		//上锁失败
		if err != nil{
			result.Err = err
			result.EndTime = time.Now()
		}else {
			// 上锁成功后，重置任务启动时间
			result.StartTime = time.Now()

			//执行shell命令
			fmt.Println(info.Job.Command)
			//cmd = exec.CommandContext(info.CancelCtx,"/bin/bash","-c",info.Job.Command)
			cmd = exec.CommandContext(info.CancelCtx,"c:\\cygwin64\\bin\\bash.exe","-c",info.Job.Command)

			//执行并捕获输出
			output,err = cmd.CombinedOutput()

			//把结果给Scheduler
			result.EndTime = time.Now()
			result.Output  = output
			result.Err = err
		}
		// 任务执行完成后，把执行的结果返回给Scheduler，Scheduler会从executingTable中删除掉执行记录
		fmt.Println("任务执行完成：",result.Err.Error())
		G_scheduler.PushJobResult(result)

	}()
}

//  初始化执行器
func InitExecutor() (err error) {
	G_executor = &Executor{}
	return
}