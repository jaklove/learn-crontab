package main

import (
	"learn-crontab/worker/Scheduler"
	"learn-crontab/worker/job"
	"learn-crontab/worker/pkg/worker"
	"log"
	"runtime"
	"time"
)

// 初始化线程数量
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main()  {
	var err error
	initEnv()

	//初始化配置
	if err = worker.InitConfig();err != nil{
		log.Fatalf("worker.InitConfig err: %v",err)
	}

	//调度任务
	if err = Scheduler.InitScheduler();err != nil{
		log.Fatalf("Scheduler.InitScheduler err: %v",err)
	}

	//监听任务
	if err = job.InitJobMgr();err != nil{
		log.Fatalf("job.InitJobMgr err: %v",err)
	}

	//hold on 进程
	for {
		time.Sleep(time.Millisecond)
	}
}