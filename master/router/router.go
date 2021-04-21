package router

import (
	"github.com/gin-gonic/gin"
	v1 "learn-crontab/master/router/v1"
)

func InitRouter()*gin.Engine  {
	r := gin.New()
	apiv1 := r.Group("/api/v1")
	{
		//添加任务
		apiv1.POST("/job/save",v1.SaveJob)
		apiv1.POST("/job/delete",v1.DeleteJob)
	}

	return r
}