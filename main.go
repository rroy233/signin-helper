package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"signin/Logger"
	sso "signin/SSO"
	"time"
)

var BackEndVer string
var router *gin.Engine
var ssoClient *sso.Client
var TZ = time.FixedZone("CST", 8*3600)

func main() {
	Logger.New(config.General.Production)
	initDB()
	initRedis()
	initMail()

	if BackEndVer == "" {
		BackEndVer = "开发环境"
	}

	//启动邮件发送后台进程
	go mailSender(MailQueue)
	//启动微信发送后台进程
	go wechatSender(WechatQueue)

	//启动定时任务
	initJob()

	//初始化cos
	cosClientUpdate()

	ssoClient = sso.NewClient(config.General.Production, config.SSO.ServiceName, config.SSO.ClientId, config.SSO.ClientSecret)

	log.Println("[系统]已启动")
	err := router.Run(":" + config.General.ListenPort)
	if err != nil {
		Logger.FATAL.Fatalln(err)
	}
}
