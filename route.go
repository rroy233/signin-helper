package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"signin/Logger"
)

func init() {
	getConfig()

	//gin日志
	logFile, err := os.OpenFile("./log/gin.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Logger.FATAL.Fatalln(err)
	}
	gin.DefaultWriter = io.MultiWriter(logFile)
	gin.SetMode(gin.ReleaseMode)

	router = gin.Default()
	router.Use(securityMiddleware)
	
	router.NoRoute(func(c *gin.Context) {
		c.Data(404,ContentTypeHTML,views("error1",map[string]string{"text":"该页面不存在"}))
	})

	router.GET("/", viewIndex)
	router.Static("/js", "./views/dist/js")
	router.Static("/css", "./views/dist/css")
	router.Static("/img", "./views/dist/img")

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
		userApiGroup.GET("/act/log", UserActLogHandler)
		userApiGroup.GET("/act/query", UserActQueryHandler)
		userApiGroup.GET("/noti/get", UserNotiGetHandler)
		userApiGroup.POST("/noti/edit", UserNotiEditHandler)
		userApiGroup.POST("/noti/check", UserNotiCheckHandler)
		userApiGroup.GET("/noti/fetch", UserNotiFetchHandler)

		userApiGroup.GET("/wechat/qrcode", UserWechatQrcodeHandler)
		userApiGroup.GET("/wechat/bind", UserWechatBindHandler)

		userApiGroup.GET("/csrfToken",UserCsrfTokenHandler)
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
		adminApiGroup.GET("/class/info", adminClassInfoHandler)
		adminApiGroup.POST("/class/edit", adminClassEditHandler)

		adminApiGroup.GET("/csrfToken",AdminCsrfTokenHandler)
	}

	testGroup := router.Group("/test")
	testGroup.Use(UserMiddleware)
	{
		testGroup.GET("tpl",testTplHandler)

		//消息推送测试
		if config.General.Production == false{
			testGroup.GET("noti/act",testActNotiHandler)
			testGroup.GET("noti/normal",testNotiHandler)
		}
	}
}
