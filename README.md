# 基于go开发的任务调度系统

######  当前项目基于ectd的注册中心中间件开发的任务调度系统

 1.发布部署，用git clone https://github.com/jaklove/learn-crontab.git 下载对应的软件目录

当前目录中已创建好对应的DockerfileMaster、DockerfileWorker文件.

#### 构建管理界面的镜像

```bash
docker build -t go-corntab-master -f DockerfileMaster. 
```

#### 构建后台worker的镜像

```bash
docker build -t go-corntab-worker -f DockerfileWorker. 
```

#### 运行对应的镜像文件

```bash
docker run -it -p 8070:8070 -d go-corntab
docker run -it -d  go-corntab-worker
```

