package job

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"learn-crontab/master/common"
	"time"
)

// mongodb存储日志
type LogSink struct {
	client         *mongo.Client
	logCollection  *mongo.Collection
	logChan        chan *common.JobLog
	autoCommitChan chan *common.LogBatch
}

var (
	// 单例
	G_logSink *LogSink
)

// 批量写入日志
func (logSink *LogSink) saveLogs(batch *common.LogBatch) {
	logSink.logCollection.InsertMany(context.TODO(), batch.Logs)
}

// 日志存储协程
func (logSink *LogSink) writeLoop() {
	var (
		log          *common.JobLog
		logBatch     *common.LogBatch // 当前的批次
		commitTimer  *time.Timer
		timeoutBatch *common.LogBatch // 超时批次
	)

	for {
		select {
		case log = <-logSink.logChan:
			if logBatch == nil {
				logBatch = &common.LogBatch{}
				// 让这个批次超时自动提交(给1秒的时间）
				commitTimer = time.AfterFunc(
					time.Duration(time.Second*1),
					func(batch *common.LogBatch) func() {
						return func() {
							logSink.autoCommitChan <- batch
						}
					}(logBatch), )
			}

			// 把新日志追加到批次中
			logBatch.Logs = append(logBatch.Logs, log)

			//如果批次满了
			if len(logBatch.Logs) > 50 {
				// 发送日志
				logSink.saveLogs(logBatch)
				// 清空logBatch
				logBatch = nil
				// 取消定时器
				commitTimer.Stop()
			}
		case timeoutBatch = <-logSink.autoCommitChan:
			// 判断过期批次是否仍旧是当前的批次
			if timeoutBatch != logBatch {
				continue // 跳过已经被提交的批次
			}

			// 把批次写入到mongo中
			logSink.saveLogs(timeoutBatch)
			// 清空logBatch
			logBatch = nil
		}
	}

}

const ConnectTimeout  =  5000

//初始化操作
func InitLogSink() error {
	var (
		err    error
		client *mongo.Client
	)

	//建立mongodb连接
	if client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://192.168.111.100:27017"),
		options.Client().SetConnectTimeout(5000 * time.Millisecond) ); err != nil {
		return err
	}

	//   选择db和collection
	G_logSink = &LogSink{
		client:         client,
		logCollection:  client.Database("cron").Collection("log"),
		logChan:        make(chan *common.JobLog, 1000),
		autoCommitChan: make(chan *common.LogBatch),
	}

	//启动一个mongodb处理协程
	go G_logSink.writeLoop()

	return nil
}

// 发送日志
func (logSink *LogSink) Append(jobLog *common.JobLog) {
	select {
	case logSink.logChan <- jobLog:
	default:
		// 队列满了就丢弃
	}
}