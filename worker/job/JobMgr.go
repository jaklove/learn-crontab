package job

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"learn-crontab/master/common"
	"learn-crontab/worker/Scheduler"
	"learn-crontab/worker/pkg/worker"
	"time"
)

type JobMgr struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

var Worker_JobMgr *JobMgr

func InitJobMgr() error {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
		err     error
	)

	//初始化配置
	config = clientv3.Config{
		Endpoints:   worker.WorkerSetting.EtcdEndpoints,
		DialTimeout: time.Duration(worker.WorkerSetting.EtcdDialTimeout) * time.Millisecond,
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		return err
	}

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)
	Worker_JobMgr = &JobMgr{
		client:  client,
		kv:      kv,
		lease:   lease,
		watcher: watcher,
	}

	if err = Worker_JobMgr.watchJobs();err != nil{
		return err
	}
	return nil
}

//监听任务的变化
func (jobMgr *JobMgr) watchJobs() error {
	var (
		err           error
		getResponse   *clientv3.GetResponse
		job           *common.Job
		watchStartRev int64
		watchChan     clientv3.WatchChan
		watchResponse clientv3.WatchResponse
		watchEvents   *clientv3.Event
		jobEvent       *common.JobEvent
	)

	//1.get一下/cron/jobs/目录下的所有任务，并且获知当前集群的revision
	if getResponse, err = jobMgr.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return err
	}

	//当前有哪些任务
	for _, kv := range getResponse.Kvs {
		//反序劣化
		if job, err = common.UnpackJob(kv.Value); err != nil {
			continue;
		}
		//TODO
		jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE,job)
		Scheduler.G_scheduler.PushJobEvent(jobEvent)
	}

	//2, 从该revision向后监听变化事件
	go func() {
		//监听下一个版本
		watchStartRev = getResponse.Header.Revision + 1
		//启动监听
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRev),clientv3.WithPrefix())

		for watchResponse = range watchChan {
			for _,watchEvents = range watchResponse.Events {
				switch watchEvents.Type {
				case mvccpb.PUT: //任务保存
					if job, err = common.UnpackJob(watchEvents.Kv.Value); err != nil {
						continue
					}
					//构建一个event事件
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE,job)
					//TODO：反序列化job,推送给scheduler
				case mvccpb.DELETE: //删除任务
					jobName := common.ExtractKillerName(string(watchEvents.Kv.Key))

					job = &common.Job{
						Name:jobName,
					}
					//构建一个删除evenet
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE,job)
				}
				fmt.Println("监听任务:",jobEvent.Job.Name)

				//投递任务
				Scheduler.G_scheduler.PushJobEvent(jobEvent)
			}

		}
	}()

	return nil
}
