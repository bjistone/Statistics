package main

import (
	"flag"
	"github.com/mediocregopher/radix.v2/redis"

	//"github.com/mediocregopher/radix.v2/redis"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)


var redisCli redis.Client

func init() {
	log.Out = os.Stdout
	log.SetLevel(logrus.DebugLevel)
	/*
		redisCli, err := redis.Dial("tcp", "127.0.0.1/6379")
		if err != nil {
			log.Fatal("redis conn err")
		}
		defer redisCli.Close()
	*/
}
func main() {

	//获取参数
	logFilePath := flag.String("logFilePath", "dig.log", "log file path")
	routineNum := flag.Int("routineNum", 10, "consumer num by goroutine")
	l := flag.String("l", "logDir/log.log", "program log here")
	flag.Parse()
	params := comParams{logFilePath: *logFilePath, routineNum: *routineNum}

	//打日志
	WLogFd, err := os.OpenFile(*l, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Error("open writeLogFile err")
		return
	}
	log.Out = WLogFd
	defer WLogFd.Close()
	log.Infof("params: logFilePath:%s  routineNum:%d", params.logFilePath, params.routineNum)

	//初始化channel,用于数据传输
	var logChan = make(chan string, params.routineNum*2)
	var pvChan = make(chan urlData, params.routineNum)
	var uvChan = make(chan urlData, params.routineNum)
	var storageChan = make(chan storageBlock, params.routineNum)

	//redis pool
	redisPool, err := pool.New("tcp", "192.168.203.128:6379", 2*params.routineNum)
	if err != nil {
		log.Fatal("redis pool create failed,err:", err)
		return
	}
	go func() {
		for {
			redisPool.Cmd("PING")
			time.Sleep(3 * time.Second)
		}
	}()

	//日志消费者
	go readLogByLine(params, logChan)

	//创建一组日志处理
	for i := 0; i < *routineNum; i++ {
		go logConsumer(logChan, pvChan, uvChan)
	}

	//创建PV UV 统计器    可以扩展
	go pvCounter(pvChan, storageChan)
	go uvCounter(uvChan, storageChan, redisPool)

	//创建存储器
	go dataStorage(storageChan, redisPool)

	time.Sleep(time.Second * 3)
}

