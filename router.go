package main

import (
	"github.com/gin-gonic/gin"
)

func init() {
	router = gin.Default()

	router.GET("/", func(context *gin.Context) {
		context.Redirect(302, "/user")
	})
	router.Static("/static", "./static")

	router.GET("/sso", ssoRedirectHandler)

	// user view
	userViewGroup := router.Group("/user")
	userViewGroup.Use(UserMiddleware)
	{
		userViewGroup.GET("/reg", viewReg)
		userViewGroup.GET("/", viewUser)
	}

	//admin view
	adminViewGroup := router.Group("/admin")
	adminViewGroup.Use(AdminMiddleware)
	{
		adminViewGroup.GET("/", viewAdmin)
	}

	//通用api
	router.GET("/api/ssoCallback", ssoCallBackHandler)
	router.GET("/api/login", loginHandler)
	router.POST("/api/logout", logoutHandler)

	//用户api
	userApiGroup := router.Group("/api/user")
	userApiGroup.Use(UserMiddleware)
	{
		userApiGroup.POST("/init", initHandler)
		userApiGroup.GET("/profile", profileHandler)
		userApiGroup.GET("/act/info", UserActInfoHandler)
		userApiGroup.GET("/act/statistic", UserActStatisticHandler)
		userApiGroup.POST("/act/signin", UserActSigninHandler)
		userApiGroup.GET("/act/log", UserActLogHandler)
		userApiGroup.GET("/act/query", UserActQueryHandler)
		userApiGroup.GET("/noti/get", UserNotiGetHandler)
		userApiGroup.POST("/noti/edit", UserNotiEditHandler)
	}

	//管理员api
	adminApiGroup := router.Group("/api/admin")
	adminApiGroup.Use(AdminMiddleware)
	{
		adminApiGroup.GET("/act/info", adminActInfoHandler)
		adminApiGroup.POST("/act/edit", adminActEditHandler)
		adminApiGroup.POST("/act/new", adminActNewHandler)
		adminApiGroup.GET("/class/info", adminClassInfoHandler)
		adminApiGroup.POST("/class/edit", adminClassEditHandler)
	}
}
