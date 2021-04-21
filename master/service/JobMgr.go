package service

import (
	"context"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"learn-crontab/master/common"
	"learn-crontab/master/pkg/setting"
	"time"
)

type JobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

var Global_JobMgr *JobMgr

func InitJobMgr() error {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
		err    error
	)

	//初始化配置
	config = clientv3.Config{
		Endpoints:   setting.AppSetting.EtcdEndpoints,
		DialTimeout: time.Duration(setting.AppSetting.EtcdDialTimeout) * time.Millisecond,
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		return err
	}

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	Global_JobMgr = &JobMgr{
		client: client,
		kv: kv,
		lease: lease,
	}

	return nil
}


type AddJob struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

func (j *JobMgr)AddJob(name,command,cornexpr string)(*AddJob, error)  {
	// Jobkey
	jobKey := common.JOB_SAVE_DIR + name

	//组装数据
	var jobvalue AddJob
	jobvalue = AddJob{
		Name: name,
		Command: command,
		CronExpr: cornexpr,
	}

	//json任务对象
	jobValue, err := json.Marshal(jobvalue)
	if err != nil{
		return nil,err
	}

	// 保存到etcd
	putResponse, err := j.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV())
	if err != nil{
		return nil,err
	}

	var oldJob AddJob
	// 如果是更新, 那么返回旧值
	if putResponse.PrevKv != nil{
		// 对旧值做一个反序列化
		err := json.Unmarshal(putResponse.PrevKv.Value, &oldJob)
		if err != nil{
			err = nil
			return nil,err
		}
		return &oldJob,nil
	}
	return &oldJob,nil
}
