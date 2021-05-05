package job

import (
	"context"
	"go.etcd.io/etcd/clientv3"
	"learn-crontab/master/common"
	"learn-crontab/worker/pkg/worker"
	"net"
	"time"
)

// 注册节点到etcd： /cron/workers/IP地址
type Register struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease

	localIP string // 本机IP
}

func getLocalIP() (string, error) {
	var (
		err     error
		addrs   []net.Addr
		addr    net.Addr
		isIpNet bool
		ipNet   *net.IPNet // IP地址
		ipv4    string
	)

	//获取所有网卡
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return "", err
	}

	// 取第一个非lo的网卡IP
	for _, addr = range addrs {
		// 这个网络地址是IP地址: ipv4, ipv6
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 跳过IPV6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String() // 192.168.1.1
				return ipv4, nil
			}
		}
	}
	err = common.ERR_NO_LOCAL_IP_FOUND
	return "", err
}

var (
	G_register *Register
)

func InitRegister() error {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		err     error
		localIp string
		lease   clientv3.Lease
	)

	//初始化配置
	config = clientv3.Config{
		Endpoints:   worker.WorkerSetting.EtcdEndpoints,                                     // 集群地址
		DialTimeout: time.Duration(worker.WorkerSetting.EtcdDialTimeout) * time.Millisecond, // 连接超时
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		return err
	}

	// 本机IP
	if localIp, err = getLocalIP(); err != nil {
		return err
	}

	// 得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	G_register = &Register{
		client:  client,
		kv:      kv,
		lease:   lease,
		localIP: localIp,
	}

	//服务注册
	go G_register.keepOnline()

	return nil
}

// 注册到/cron/workers/IP, 并自动续租
func (register *Register) keepOnline() {
	var (
		err            error
		regKey         string
		leaseGrantResp *clientv3.LeaseGrantResponse
		cancelFunc     context.CancelFunc
		cancelCtx      context.Context
		leaseKeepAlivechan <- chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp *clientv3.LeaseKeepAliveResponse
	)

	for {
		// 注册路径
		regKey = common.JOB_WORKER_DIR + register.localIP

		cancelFunc = nil

		//创建租约
		if leaseGrantResp, err = register.lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}
		
		// 自动续租
		if leaseKeepAlivechan,err = register.lease.KeepAlive(context.TODO(),leaseGrantResp.ID);err != nil{
			goto RETRY
		}

		cancelCtx, cancelFunc = context.WithCancel(context.TODO())

		// 注册到etcd
	   if _,err = register.kv.Put(cancelCtx,regKey,"",clientv3.WithLease(leaseGrantResp.ID));err != nil{
		   goto RETRY
	   }

		// 处理续租应答
		for  {
			select {
			case keepAliveResp = <- leaseKeepAlivechan :
				if keepAliveResp == nil {	// 续租失败
					goto RETRY
				}
			}
		}

	RETRY:
		time.Sleep(time.Second * 1)
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}
