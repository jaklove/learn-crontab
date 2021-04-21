package v1

import (
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

func DeleteJob(c *gin.Context)  {

}