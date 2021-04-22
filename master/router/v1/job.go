package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"learn-crontab/master/pkg/app"
	"learn-crontab/master/service"
)

type AddJobForm struct {
	Name     string `form:"name"`
	Command  string `form:"command"`
	CronExpr string `form:"cronExpr"`
}

//添加任务
func SaveJob(c *gin.Context) {
	//数据绑定
	var (
		appG = app.Gin{C: c}
		form AddJobForm
	)

	err := app.BindAndValid(c, &form)
	if err != nil {
		appG.Response(400,"绑定发生错误",nil)
		return
	}

	//添加任务到etcd中
	oldJob, err := service.Global_JobMgr.AddJob(form.Name,form.Command,form.CronExpr)
	if err != nil{
		appG.Response(400,"添加任务失败",err)
		return
	}
	appG.Response(200,"success",oldJob)
	return
}

//删除任务
func DeleteJob(c *gin.Context)  {
	var (
		appG = app.Gin{C: c}
	)
	//获取name
	name := c.Query("name")
	if name == ""{
		appG.Response(400,"添加任务失败",errors.New("当前name不能为空"))
		return
	}

	oldJob, err := service.Global_JobMgr.DeleteJobByName(name)
	if err != nil{
		appG.Response(400,"任务失败失败",nil)
		return
	}
	appG.Response(200,"success",oldJob)
	return
}

//获取任务列表
func JobList(c * gin.Context)  {
	var (
		appG = app.Gin{C: c}
	)

	list, err := service.Global_JobMgr.GetJobList()
	if err != nil{
		appG.Response(500,err.Error(),nil)
		return
	}
	appG.Response(200,"success",list)
	return
}

//杀死任务
func KillJob(c *gin.Context)  {
	var (
		appG = app.Gin{C: c}
	)
	//任务名称
	name := c.Query("name")
	if name == ""{
		appG.Response(400,"添加任务失败",errors.New("当前name不能为空"))
		return
	}

	err := service.Global_JobMgr.KillJob(name)
	if err != nil{
		appG.Response(400,"kill任务失败",err)
		return
	}

	appG.Response(200,"success",nil)
	return
}