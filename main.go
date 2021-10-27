package main

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"signin/Logger"
	sso "signin/SSO"
	"strconv"
	"time"
)

var BackEndVer string
var router *gin.Engine
var ssoClient *sso.Client
var TZ = time.FixedZone("CST", 8*3600)

func main() {
	Logger.New()
	getConfig()
	initDB()
	initRedis()
	initMail()

	if BackEndVer == ""{
		BackEndVer = "开发环境"
	}

	//启动邮件发送后台进程
	go mailSender(MailQueue)

	//启动定时任务
	initJob()

	ssoClient = sso.NewClient(config.General.Production, config.SSO.ServiceName, config.SSO.ClientId, config.SSO.ClientSecret)

	err := router.Run(":" + config.General.ListenPort)
	if err != nil {
		Logger.FATAL.Fatalln(err)
	}
}

// MD5_short 生成6位MD5
func MD5_short(v string) string {
	d := []byte(v)
	m := md5.New()
	m.Write(d)
	return hex.EncodeToString(m.Sum(nil)[0:5])
}

// MD5 生成MD5
func MD5(v string) string {
	d := []byte(v)
	m := md5.New()
	m.Write(d)
	return hex.EncodeToString(m.Sum(nil))
}

func ts2DateString(ts string) string {
	timestamp, _ := strconv.ParseInt(ts, 10, 64)
	return time.Unix(timestamp, 0).In(TZ).Format("2006-01-02 15:04:05")
}

func dateString2ts(datetime string) (int64, error) {
	tmp, err := time.ParseInLocation("2006-01-02 15:04", datetime, TZ)
	if err != nil {
		return 0, err
	}
	return tmp.Unix(), nil
}
