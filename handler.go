package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mgutz/str"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func readLogByLine(params comParams, logChan chan string) {
	logF, err := os.OpenFile(params.logFilePath, os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Warning("func readLogByLine not find log file")
		return
	}
	defer logF.Close()
	reader := bufio.NewReader(logF)
	count := 0
	for {
		line, err := reader.ReadString('\n')
		logChan <- line
		count++
		if err == io.EOF {
			time.Sleep(time.Second * 1)
			log.Infof("func readLogByLine read EOF,wait 2 second, readLine:%d\n", count)
		}
		if err != nil {
			log.Warningf("func readLogByLine err:%s\n", err)
		}
		if count%(10) == 0 {
			log.Infof("readLogFile line:%d\n", count)
			return
		}
		//log.Infof("logchan>>%v\n",<-logChan)
	}

}
func formatUrl(url, time string) urlNode {
	//一定从量大的着手，详情页》列表页》首页
	pos1 := str.IndexOf(url, HandleMovie, 0)

	if pos1 != -1 {
		//详情页
		pos1 += len(HandleMovie)
		pos2 := str.IndexOf(url, HandleHtml, 0)
		idStr := str.Substr(url, pos1, pos2-pos1)
		id, _ := strconv.Atoi(idStr)
		return urlNode{"movie", id, url, time}
	} else {
		//列表页
		pos1 = str.IndexOf(url, HandleList, 0)
		if pos1 != -1 {
			pos1 += len(HandleList)
			pos2 := str.IndexOf(url, HandleHtml, 0)
			idStr := str.Substr(url, pos1, pos2-pos1)
			id, _ := strconv.Atoi(idStr)
			return urlNode{"list", id, url, time}
		} else {
			return urlNode{"home", 1, url, time}
		}
	}
}
func logConsumer(logChan chan string, pvChan chan urlData, uvChan chan urlData) {
	for logStr := range logChan {
		//切割日志字符串 ，抠出打点上报的数据
		date := cutLogFetchDate(logStr)

		//uid
		//模拟uid， md5(refer+ua)
		hash := md5.New()
		hash.Write([]byte(date.refer + date.ua))
		uid := hex.EncodeToString(hash.Sum(nil))

		//很多解析的工作可以放这里
		uDate := urlData{date, uid, formatUrl(date.url, date.time)}

		//log.Infoln(uDate)
		pvChan <- uDate
		uvChan <- uDate
	}
	return
}
func cutLogFetchDate(logStr string) digData {
	logStr = strings.TrimSpace(logStr)
	pos1 := str.IndexOf(logStr, HandleDig, 0)
	if pos1 == -1 {
		return digData{}
	}
	pos1 += len(HandleDig)
	pos2 := str.IndexOf(logStr, "HTTP/", pos1)
	d := str.Substr(logStr, pos1, pos2-pos1)
	urlInfo, err := url.Parse("HTTP://localhost/?" + d)
	if err != nil {
		return digData{}
	}
	data := urlInfo.Query()
	return digData{
		data.Get("time"),
		data.Get("url"),
		data.Get("refer"),
		data.Get("ua"),
	}
}

func pvCounter(pvChan chan urlData, storageChan chan storageBlock) {
	//pv 有访问多少次
	for data := range pvChan {
		stoBlock := storageBlock{"pv", "ZINCRBY", data.uNode}
		storageChan <- stoBlock

	}
}

func getTime(logTime, timeType string) string {
	var format string
	switch timeType {
	case "day":
		format = "2006-01-02"
	case "hour":
		format = "2006-01-02 15"
	case "min":
		format = "2006-01-02 15:04"
	}
	t, _ := time.Parse(format, time.Now().Format(format))
	return strconv.FormatInt(t.Unix(), 10)
}
func uvCounter(uvChan chan urlData, storageChan chan storageBlock, redisPool *pool.Pool) {
	//有多少人访问(去重)
	for data := range uvChan {
		//HyperLoglog redis去重
		hyperLoglogKey := "uv_hpll_" + getTime(data.data.time, "day")
		ret, err := redisPool.Cmd("PFADD", hyperLoglogKey, data.uid, "EX", 86400).Int()
		if err != nil {
			log.Warningln("uvCounter check redis hyperLoglog failed:", err)
		}
		if ret != 1 {
			continue
		}
		sItem := storageBlock{"uv", "ZINCREBY", data.uNode}
		storageChan <- sItem
	}
}
func dataStorage(storageChan chan storageBlock, redisPool *pool.Pool) {
	for block := range storageChan {
		prefix := block.counterType + "_"
		//逐层添加，
		//维度：天-小时-分钟
		//层级：顶级-大分类-小分类-最终页面
		//存储模型：redis sortedSet
		setKeys := []string{
			prefix + "day_" + getTime(block.uNode.uNTime, "day"),
			prefix + "hour_" + getTime(block.uNode.uNTime, "hour"),
			prefix + "min_" + getTime(block.uNode.uNTime, "min"),
			prefix + block.uNode.uNType + "_day_" + getTime(block.uNode.uNTime, "day"),
			prefix + block.uNode.uNType + "_hour_" + getTime(block.uNode.uNTime, "hour"),
			prefix + block.uNode.uNType + "_min_" + getTime(block.uNode.uNTime, "min"),
		}
		rowId := block.uNode.uNRid
		for _, key := range setKeys {
			ret, err := redisPool.Cmd(block.storageModel, key, 1).Int()
			if ret <= 0 || err != nil {
				log.Errorln("dataStorage redis err:", block.storageModel, key, rowId)
			}
		}
	}
}
