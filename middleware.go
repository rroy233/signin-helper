package main

import (
	"github.com/gin-gonic/gin"
	"net/url"
	"strings"
)

func UserMiddleware(c *gin.Context) {
	headerFilter(c, 0)
}

func AdminMiddleware(c *gin.Context) {
	headerFilter(c, 1)
}

func headerFilter(c *gin.Context, isAdmin int) {
	if getCookie(c, "token") == "" {
		middleWareRedirect(c)
		c.Abort()
	} else {
		auth, err := verifyJWTSigning(getCookie(c, "token"), true)
		if err != nil {
			middleWareRedirect(c)
			c.Abort()
			return
		}
		if auth.ClassId == 0 && c.FullPath() != "/api/user/init" && c.FullPath() != "/user/reg" {
			//未完成初始化
			c.Redirect(302, "/user/reg")
			c.Abort()
			return
		}
		if isAdmin == 1 && auth.IsAdmin != 1 {
			returnErrorView(c,"您无权限访问")
			c.Abort()
			return
		}
		c.Set("auth", auth)
	}
}

func middleWareRedirect(c *gin.Context) {
	if strings.Contains(c.FullPath(), "/api/") == true {
		returnErrorView(c,"未授权访问")
	} else {
		c.Redirect(302, "/sso?redirect="+url.PathEscape(c.FullPath()))
	}
}

func getCookie(c *gin.Context, key string) string {
	tmp, _ := c.Cookie(key)
	return tmp
}

func redirectToLogin(c *gin.Context) {
	c.Redirect(302, "/sso")
}
