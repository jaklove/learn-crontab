package service

import (
	"context"
	"go.etcd.io/etcd/clientv3"
	"learn-crontab/master/common"
	"learn-crontab/master/pkg/setting"
	"time"
)

// /cron/workers/
type WorkerMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

var (
	G_workerMgr *WorkerMgr
)

func (workerMgr *WorkerMgr)ListWorkers()([]string,error)  {

	// 初始化数组
	workerArr := make([]string, 0)
	getResponse, err := workerMgr.kv.Get(context.TODO(), common.JOB_WORKER_DIR, clientv3.WithPrefix())
	if err != nil{
		return nil, err
	}

	// 解析每个节点的IP
	for _,kv := range getResponse.Kvs{
		// kv.Key : /cron/workers/192.168.2.1
		workerIP := common.ExtractWorkerIP(string(kv.Key))
		workerArr = append(workerArr, workerIP)
	}
	return workerArr,nil
}

//初始化任务
func InitWorkerMgr() error {
	//初始化配置
	config := clientv3.Config{
		Endpoints:   setting.AppSetting.EtcdEndpoints,
		DialTimeout: time.Duration(setting.AppSetting.EtcdDialTimeout) * time.Millisecond,
	}

	//建立连接
	client, err := clientv3.New(config)
	if err != nil{
		return err
	}

	//得到KV和Lease的API子集
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)

	G_workerMgr = &WorkerMgr{
		client: client,
		kv: kv,
		lease: lease,
	}
	return nil
}
