package worker

import (
	"gopkg.in/ini.v1"
	"log"
)

type Worker struct {
	EtcdEndpoints         []string
	EtcdDialTimeout       int
	MongodbUri            string
	MongodbConnectTimeout int
	JobLogBatchSize       int
	JobLogCommitTimeout   int
}

var WorkerSetting = &Worker{}

func InitConfig() error {
	cfg, err := ini.Load("worker/conf/worker.ini")
	if err != nil {
		log.Fatalf("Fail to parse worker/conf/worker.ini: %v", err)
		return err
	}
	err = cfg.Section("worker").MapTo(WorkerSetting)
	if err != nil {
		log.Fatalf("Cfg.MapTo WorkerSetting err: %v", err)
		return err
	}
	return nil
}
