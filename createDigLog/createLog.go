package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var uaList = []string{
	"Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_0) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
	"Opera/9.80 (Windows NT 6.1; U; en) Presto/2.8.131 Version/11.11",
	"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; en) Presto/2.8.131 Version/11.11",
	"Mozilla/5.0 (Windows NT 6.1; rv:2.0.1) Gecko/20100101 Firefox/4.0.1",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.6; rv:2.0.1) Gecko/20100101 Firefox/4.0.1",
	"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)",
	"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36",
}

type resource struct {
	url    string
	target string
	start  int
	end    int
}

func ruleResource() []resource {
	var res []resource
	r1 := resource{
		url:    "http://localhost:8080",
		target: "",
		start:  0,
		end:    0,
	}
	r2 := resource{
		url:    "http://localhost/list/{$id}:8080",
		target: "{$id}",
		start:  1,
		end:    21,
	}
	r3 := resource{
		url:    "http://localhost:8080/movie{$id}.html",
		target: "{$id}",
		start:  1,
		end:    12924,
	}
	res = append(append(append(res, r1), r2), r3)
	return res
}
func buildUrl(res []resource) []string {
	var list []string
	for _, r := range res {
		if len(r.target) == 0 {
			list = append(list, r.url)
		} else {
			for i := r.start; i <= r.end; i++ {
				//fmt.Println(i)
				urlStr := strings.Replace(r.url, r.target, strconv.Itoa(i), -1)
				list = append(list, urlStr)
			}
		}
	}
	//fmt.Println(list)
	return list
}
func makeLog(current, refer, ua string) string {
	u := url.Values{}
	u.Set("time", "1")
	u.Set("url", current)
	u.Set("refer", refer)
	u.Set("ua", ua)
	paramsStr := u.Encode()
	logTemplate := "127.0.0.1 - - [10/Nov/2020:21:27:52 +0800] \"OPTIONS /dig? {$paramsStr}HTTP/1.1\" 200 43 \"-\" \"{$ua}\" \"-\""
	log := strings.Replace(logTemplate, "{$paramsStr}", paramsStr, -1)
	log = strings.Replace(log, "{$ua}", ua, -1)
	return log
}
func randInt(min, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if min > max {
		return min
	}
	return r.Intn(max-min) + min
}
func main() {
	total := flag.Int("total", 1000, "how many log")
	writeLogFile := flag.String("logfile", "dig.log", "log file path")
	flag.Parse()
	fmt.Println(*writeLogFile)
	writeF, err := os.OpenFile(*writeLogFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	//需要构造出真实的网站url集合
	res := ruleResource()
	list := buildUrl(res)
	//fmt.Println(list)
	//按要求生成total行日志
	logStr := ""
	for i := 0; i < *total; i++ {
		currentUrl := list[randInt(0, len(list)-1)]
		referUrl := list[randInt(0, len(list)-1)]
		ua := uaList[randInt(0, len(uaList)-1)]
		logStr += strconv.Itoa(i) + ">>" + makeLog(currentUrl, referUrl, ua) + "\n"
		if err != nil {
			panic("write file err")
		}
	}
	_, err = writeF.Write([]byte(logStr))
	fmt.Println("done!")
}
