package main

import (
	"context"
	"fmt"
	"learn-crontab/master/pkg/setting"
	"learn-crontab/master/router"
	"learn-crontab/master/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

//设置线程数
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

const (
	StateHealth   = "health"
	StateUnHealth = "unhealth"
)

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
	go func() {
		if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server run err: %+v", err)
		}
	}()

	// 用于捕获退出信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 捕获到退出信号之后将健康检查状态设置为 unhealth
	state := StateUnHealth
	log.Println("Shutting down state: ", state)

	// 设置超时时间，两个心跳周期，假设一次心跳 3s
	ctx, cancelFunc := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancelFunc()

	// Shutdown 接口，如果没有新的连接了就会释放，传入超时 context
	// 调用这个接口会关闭服务，但是不会中断活动连接
	// 首先会将端口监听移除
	// 然后会关闭所有的空闲连接
	// 然后等待活动的连接变为空闲后关闭
	// 如果等待时间超过了传入的 context 的超时时间，就会强制退出
	// 调用这个接口 server 监听端口会返回 ErrServerClosed 错误
	// 注意，这个接口不会关闭和等待websocket这种被劫持的链接，如果做一些处理。可以使用 RegisterOnShutdown 注册一些清理的方法
	if err := server.Shutdown(ctx);err != nil{
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
