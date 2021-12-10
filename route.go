package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"signin/Logger"
)

func init() {
	getConfig()

	logFile, err := os.OpenFile("./log/gin.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Logger.FATAL.Fatalln(err)
	}
	gin.DefaultWriter = io.MultiWriter(logFile)
	gin.SetMode(gin.ReleaseMode)

	router = gin.Default()
	router.Use(securityMiddleware)
	router.MaxMultipartMemory = 100 << 20 // 100 MiB

	router.NoRoute(func(c *gin.Context) {
		c.Data(404, ContentTypeHTML, views("error1", map[string]string{"text": "该页面不存在"}))
	})

	router.GET("/", viewIndex)
	router.Static("/js", "./views/dist/js")
	router.Static("/css", "./views/dist/css")
	router.Static("/img", "./views/dist/img")
	router.Static("/export", "./storage/export")

	router.Static("/static", "./static")

	router.GET("/sso", ssoRedirectHandler)

	// user view
	userViewGroup := router.Group("/user")
	userViewGroup.Use(UserMiddleware)
	{
		userViewGroup.GET("/reg", viewReg)
	}

	//通用api
	router.GET("/api/ssoCallback", ssoCallBackHandler)
	router.GET("/api/login", loginHandler)
	router.POST("/api/logout", logoutHandler)
	router.POST("/api/wxpusherCallback", wxpusherCallbackHandler)

	//用户api
	userApiGroup := router.Group("/api/user")
	userApiGroup.Use(UserMiddleware)
	{
		userApiGroup.GET("/version", versionHandler)
		userApiGroup.POST("/init", initHandler)
		userApiGroup.GET("/profile", profileHandler)
		userApiGroup.GET("/act/info", UserActInfoHandler)
		userApiGroup.GET("/act/statistic", UserActStatisticHandler)
		userApiGroup.POST("/act/signin", UserActSigninHandler)
		userApiGroup.POST("/act/cancel", UserActCancelHandler)
		userApiGroup.POST("/act/upload", UserActUploadHandler)
		userApiGroup.GET("/act/log", UserActLogHandler)
		userApiGroup.GET("/act/query", UserActQueryHandler)
		userApiGroup.GET("/noti/get", UserNotiGetHandler)
		userApiGroup.POST("/noti/edit", UserNotiEditHandler)
		userApiGroup.POST("/noti/check", UserNotiCheckHandler)
		userApiGroup.GET("/noti/fetch", UserNotiFetchHandler)

		userApiGroup.GET("/wechat/qrcode", UserWechatQrcodeHandler)
		userApiGroup.GET("/wechat/bind", UserWechatBindHandler)
	}

	//管理员api
	adminApiGroup := router.Group("/api/admin")
	adminApiGroup.Use(AdminMiddleware)
	{
		adminApiGroup.GET("/act/info", adminActInfoHandler)
		adminApiGroup.POST("/act/edit", adminActEditHandler)
		adminApiGroup.POST("/act/new", adminActNewHandler)
		adminApiGroup.GET("/act/list", adminActListHandler)
		adminApiGroup.GET("/act/statistic", adminActStatisticHandler)
		adminApiGroup.POST("/act/export", AdminActExportHandler)
		adminApiGroup.POST("/act/viewFile", AdminActViewFileHandler)
		adminApiGroup.GET("/class/info", adminClassInfoHandler)
		adminApiGroup.POST("/class/edit", adminClassEditHandler)

		adminApiGroup.GET("/user/list", adminUserListHandler)
		adminApiGroup.POST("/user/del", adminUserDelHandler)
	}

	testGroup := router.Group("/test")
	testGroup.Use(AdminMiddleware)
	{
		testGroup.GET("tpl", testTplHandler)
	}

	debugGroup := router.Group("/debug")
	debugGroup.Use(AdminMiddleware)
	{
		//消息推送测试
		debugGroup.GET("noti/act", testActNotiHandler)
		debugGroup.GET("noti/normal", testNotiHandler)

		debugGroup.GET("noti/send", testTplSendHandler)
	}
}
