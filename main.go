package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	layOut           = "2006-01-02"
	ConditionRelease = "release"
	ConditionTest    = "test"
	picDir			 = "./pic/"

	Robotdb       = "robotdb"
)

var (
	robotdb       *sqlx.DB
	qipaidb       *sqlx.DB
	robotRedis    *redis.Pool //	平台机器人
	coinRedis     *redis.Pool //  金币场机器人
	matchRedis    *redis.Pool //  比赛场机器人
	userInfoRedis *redis.Pool
	myLog         *log.Logger
	logFile       = "./downloadrobotheadscript-%v.log"
	goLimit 	  = make(chan int, 5)
	finish sync.WaitGroup
	downCount = DownCount{}
)

type DownCount struct {
	count int64
	mutex    sync.Mutex
}

type RobotUrl struct {
	ID int64 `json:"id"`
	State int `json:"state"`
	ImgUrl string `json:"img_url"`
	LogDate string `json:"log_date"`
	LogTime string `json:"log_time"`
}

func main() {
	// 配置log
	fileName := fmt.Sprintf(logFile, time.Now().Format(layOut))
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("not find the log file")
	}
	myLog = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile|log.Lmicroseconds)
	// 命令行参数
	var fileStr string
	flag.StringVar(&fileStr, "file", "resources/avatar1.csv", "图片信息文件地址")
	flag.StringVar(&logFile, "logfile", logFile, "日志文件地址")
	limit := flag.Int("limit", 5, "concurrency limit")
	flag.Parse()
	defer func() {
		file.Close()
	}()
	//1.得到机器人头像url
	//从文件中读取机器人url
	csvFile, _ := os.Open(fileStr)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	var robotUrl []RobotUrl
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		id, err := strconv.ParseInt(line[0], 10, 64)
		if err != nil {
			log.Printf("机器人头像ID ParseInt err:%v", err)
		}
		state, err := strconv.Atoi(line[1])
		if err != nil {
			log.Printf("机器人头像状态Atoi err:%v", err)
		}
		robotUrl = append(robotUrl, RobotUrl{
			ID: id,
			State: state,
			ImgUrl: line[2],
			LogDate: line[3],
			LogTime: line[4],
		})
	}

	//2.遍历机器人头像下载
	myLog.Printf("开始下载机器人头\n")
	myLog.Printf("当前的并发限制量: %v\n", *limit)
	goLimit = make(chan int, *limit)
	finish.Add(len(robotUrl))
	for _, u := range robotUrl {
		go downloadPic(u)
	}
	finish.Wait()
	myLog.Printf("下载机器人头完成\n")
	myLog.Printf("总数量:%v  下载成功数量:%v\n", len(robotUrl), downCount.count)
}

func downloadPic(robotAvatar RobotUrl) {
	goLimit <- 1
	defer func() {
		<-goLimit
		finish.Done()
	}()
	res, err := http.Get(robotAvatar.ImgUrl)
	if err != nil {
		myLog.Printf("downloadPic 文件下载失败:%v\n", err)
		return
	}
	data ,err := ioutil.ReadAll(res.Body)
	if err != nil {
		myLog.Printf("downloadPic 读取数据失败:%v\n", err)
		return
	}
	newFile, err := os.Create(fmt.Sprintf("%vrobot-%d.png", picDir, robotAvatar.ID))
	if err != nil {
		myLog.Printf("downloadPic 创建文件失败:%v\n", err)
		return
	}
	defer newFile.Close()
	err = ioutil.WriteFile(fmt.Sprintf("%vrobot-%d.png", picDir, robotAvatar.ID), data, 0644)
	if err != nil {
		myLog.Printf("downloadPic err:%v\n", err)
		return
	}
	downCount.mutex.Lock()
	downCount.count++
	downCount.mutex.Unlock()
	myLog.Printf("下载成功%v\n", robotAvatar.ImgUrl)
}
