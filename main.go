package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"signin/Logger"
	sso "signin/SSO"
	"signin/cipher"
	"time"
)

var BackEndVer string
var router *gin.Engine
var ssoClient *sso.Client
var TZ = time.FixedZone("CST", 8*3600)
var Cipher *cipher.Cipher

func main() {
	Logger.New(config.General.Production)
	initDB()
	initRedis()
	initMail()

	if BackEndVer == "" {
		BackEndVer = "开发环境"
	}

	//初始化加密模块
	var err error
	Cipher, err = cipher.Init([]byte(config.General.AESKey), []byte(config.General.AESIv))
	if err != nil {
		Logger.FATAL.Fatalln("初始化加密模块初始化失败")
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
	err = router.Run(":" + config.General.ListenPort)
	if err != nil {
		Logger.FATAL.Fatalln(err)
	}
}
