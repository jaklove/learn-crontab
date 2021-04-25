package service

import (
	"context"
	"encoding/json"
	"fmt"
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
		kv:     kv,
		lease:  lease,
	}

	return nil
}

type AddJob struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

func (j *JobMgr) AddJob(name, command, cornexpr string) (*AddJob, error) {
	// Jobkey
	jobKey := common.JOB_SAVE_DIR + name

	//组装数据
	var jobvalue AddJob
	jobvalue = AddJob{
		Name:     name,
		Command:  command,
		CronExpr: cornexpr,
	}

	//json任务对象
	jobValue, err := json.Marshal(jobvalue)
	if err != nil {
		return nil, err
	}

	// 保存到etcd
	putResponse, err := j.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}

	var oldJob AddJob
	// 如果是更新, 那么返回旧值
	if putResponse.PrevKv != nil {
		// 对旧值做一个反序列化
		err := json.Unmarshal(putResponse.PrevKv.Value, &oldJob)
		if err != nil {
			err = nil
			return nil, err
		}
		return &oldJob, nil
	}
	return &oldJob, nil
}

//删除任务
func (j *JobMgr) DeleteJobByName(name string) (*AddJob, error) {
	var (
		delResp   *clientv3.DeleteResponse
		err       error
		deleteJob *AddJob
	)
	// etcd中保存任务的key
	jobKey := common.JOB_SAVE_DIR + name

	// 从etcd中删除它
	if delResp, err = j.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return nil, err
	}

	//判断删除返回的数据
	if len(delResp.PrevKvs) != 0 {
		//解析值
		err = json.Unmarshal(delResp.PrevKvs[0].Value, &deleteJob)
		if err != nil {
			err = nil
			return nil, err
		}
		return deleteJob, err
	}

	return deleteJob, nil
}

//获取任务列表
func (j *JobMgr) GetJobList() ([]*AddJob, error) {
	var (
		getResponse *clientv3.GetResponse
		err         error
	)

	// 任务保存的目录
	dirKey := common.JOB_SAVE_DIR
	if getResponse, err = j.kv.Get(context.TODO(), dirKey, clientv3.WithPrefix()); err != nil {
		return nil, err
	}

	// 初始化数组空间
	jobList := make([]*AddJob, 0)

	// 遍历所有任务, 进行反序列化
	for _, kvPair := range getResponse.Kvs {
		var job *AddJob
		if err = json.Unmarshal(kvPair.Value, &job); err != nil {
			continue
		}
		jobList = append(jobList, job)
	}
	return jobList, nil
}


func (j *JobMgr) KillJob(name string)error {
	fmt.Println(name)
	var (
		err error
		leaseGrantResp *clientv3.LeaseGrantResponse
	)
	// 通知worker杀死对应任务
	killerKey := common.JOB_KILLER_DIR + name

	// 让worker监听到一次put操作, 创建一个租约让其稍后自动过期即可
	if leaseGrantResp, err = j.lease.Grant(context.TODO(), 1); err != nil {
		return err
	}

	// 租约ID
	leaseId := leaseGrantResp.ID

	// 设置killer标记
	if _, err = j.kv.Put(context.TODO(), killerKey, "", clientv3.WithLease(leaseId));err != nil{
		return err
	}

	return nil
}
