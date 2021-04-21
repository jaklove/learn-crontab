package setting

import (
	"gopkg.in/ini.v1"
	"log"
)

type App struct {
	RUN_MODE string
	ApiPort int
	ApiReadTimeout int
	ApiWriteTimeout int
	EtcdEndpoints []string
	EtcdDialTimeout int
	Webroot string
	MongodbUri string
	MongodbConnectTimeout int
}

var AppSetting = &App{}

func InitConfig()  {
	cfg, err := ini.Load("conf/conf.ini")
	if err != nil{
		log.Fatalf("Fail to parse 'conf/app.ini': %v",err)
	}
	err = cfg.Section("app").MapTo(AppSetting)
	if err != nil{
		log.Fatalf("Cfg.MapTo AppSetting err: %v",err)
	}
}