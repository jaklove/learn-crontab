package main

import (
	"fmt"
	"learn-crontab/master/pkg/setting"
	"learn-crontab/master/router"
	"learn-crontab/master/service"
	"log"
	"net/http"
	"runtime"
	"time"
)

//设置线程数
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var err error
	initEnv()

	//初始化配置
	setting.InitConfig()

	// 初始化服务发现模块
	if err = service.InitWorkerMgr(); err != nil {
		log.Fatalf("service.InitWorkerMgr err: %v", err)
	}

	// 日志管理器
	if err = service.InitLogMgr(); err != nil {
		log.Fatalf("service.InitLogMgr err: %v", err)
	}

	//初始化任务
	err = service.InitJobMgr()
	if err != nil {
		log.Fatalf("service.InitJobMgr err: %v", err)
	}

	//初始化api
	engine := router.InitRouter()
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", setting.AppSetting.ApiPort),
		Handler:      engine,
		ReadTimeout:  time.Duration(setting.AppSetting.ApiReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(setting.AppSetting.ApiWriteTimeout) * time.Millisecond,
	}

	//启动监听服务
	server.ListenAndServe()
}
