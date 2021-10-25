package main

import (
	"github.com/gin-gonic/gin"
	"net/url"
	"signin/Logger"
)

func UserMiddleware(c *gin.Context) {
	headerFilter(c, 0)
}

func AdminMiddleware(c *gin.Context) {
	headerFilter(c, 1)
}

func headerFilter(c *gin.Context, isAdmin int) {
	if getCookie(c, "token") == "" {
		Logger.Info.Println("[中间件]cookie无效")
		c.Redirect(302, "/sso?redirect="+url.PathEscape(c.FullPath()))
		c.Abort()
	} else {
		auth, err := verifyJWTSigning(getCookie(c, "token"), true)
		if err != nil {
			Logger.Info.Println("[中间件]", err)
			c.Redirect(302, "/sso?redirect="+url.PathEscape(c.FullPath()))
			c.Abort()
			return
		}
		if auth.ClassId == 0 && c.FullPath() != "/api/user/init" && c.FullPath() != "/user/reg"{
			//未完成初始化
			c.Redirect(302, "/user/reg")
			c.Abort()
			return
		}
		if isAdmin == 1 && auth.IsAdmin != 1 {
			returnErrorJson(c, "您无权限访问")
			c.Abort()
			return
		}
		c.Set("auth", auth)
	}
}

func getCookie(c *gin.Context, key string) string {
	tmp, _ := c.Cookie(key)
	return tmp
}

func redirectToLogin(c *gin.Context) {
	c.Redirect(302, "/sso")
}
