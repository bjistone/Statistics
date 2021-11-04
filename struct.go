package main

import "github.com/sirupsen/logrus"

type comParams struct {
	logFilePath string
	routineNum  int
}
type digData struct {
	time  string
	url   string
	refer string
	ua    string
}

type urlData struct {
	data  digData
	uid   string
	uNode urlNode
}

type urlNode struct {
	//uN==urlNode
	uNType string ///move || /list/  ||首页
	uNRid  int    //resource id
	uNUrl  string //current url
	uNTime string //当前访问这个页面的时间
}

type storageBlock struct {
	counterType  string //区分统计类型
	storageModel string
	uNode        urlNode
}

var log = logrus.New()

const HandleDig = "/dig?"
const HandleMovie = "/movie/"
const HandleList = "/list/"
const HandleHtml = "/.html/"
