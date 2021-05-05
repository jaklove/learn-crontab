package router

import (
	"github.com/gin-gonic/gin"
	v1 "learn-crontab/master/router/v1"
	"net/http"
)

func InitRouter()*gin.Engine  {
	r := gin.New()

    //显示对于的html文件
	r.StaticFS("/webroot/index", http.Dir("./master/view"))
	apiv1 := r.Group("/api/v1")
	{
		//添加任务
		apiv1.POST("/job/save",v1.SaveJob)
		//删除任务
		apiv1.GET("/job/delete",v1.DeleteJob)
		//获取任务列表
		apiv1.GET("/job/list",v1.JobList)
		//杀死任务
		apiv1.GET("/job/kill",v1.KillJob)

		apiv1.GET("/job/log",v1.LogList)

		apiv1.GET("/worker/list",v1.HandleWorkerList)

	}

	return r
}