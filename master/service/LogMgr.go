package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"learn-crontab/master/common"
	"learn-crontab/master/pkg/setting"
	"time"
)

// mongodb日志管理
type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var (
	G_logMgr *LogMgr
)

//日志处理初始化
func InitLogMgr() error {
	var (
		client *mongo.Client
		err    error
	)

	// 建立mongodb连接
	if client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(setting.AppSetting.MongodbUri), options.Client().SetConnectTimeout(5000*time.Millisecond)); err != nil {
		return err
	}

	G_logMgr = &LogMgr{
		client:client,
		logCollection: client.Database("cron").Collection("log"),
	}
	return nil
}

//查看日志列表
func (j *JobMgr) LogList(name string,skip,limit int)([]*common.JobLog,error) {
	fmt.Println("日志列表")
	var (
		logArr []*common.JobLog
		filter  *common.JobLogFilter
		logSort *common.SortLogByStartTime
		jobLog  *common.JobLog
	)

	logArr = make([]*common.JobLog,0)

	// 过滤条件
	filter = &common.JobLogFilter{JobName:name}

	// 按照任务开始时间倒排
	logSort = &common.SortLogByStartTime{SortOrder: -1}

	//查询
	// 5, 查询（过滤 +翻页参数）
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(logSort)

	cursor, err := G_logMgr.logCollection.Find(context.TODO(), filter, findOptions);
	if err != nil {
		return nil,err
	}

	// 延迟释放游标
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()){
		jobLog = &common.JobLog{}

		// 反序列化BSON
		if err = cursor.Decode(jobLog);err != nil{
			continue //过滤掉不合法日志
		}

		logArr = append(logArr,jobLog)
	}
	return logArr,nil
}
