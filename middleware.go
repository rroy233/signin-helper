package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/url"
	"signin/Logger"
	"strconv"
	"strings"
	"time"
)

func UserMiddleware(c *gin.Context) {
	authorizationMiddleware(c, 0)
}

func AdminMiddleware(c *gin.Context) {
	authorizationMiddleware(c, 1)
}

func authorizationMiddleware(c *gin.Context, isAdmin int) {
	token := ""
	tokenHeader := c.GetHeader("Authorization")
	tokenCookie := getCookie(c, "token")
	if tokenHeader == "" && tokenCookie == "" {
		middleWareRedirect(c)
		c.Abort()
		return
	} else {
		if tokenCookie != "" {
			token = tokenCookie
		} else {
			token = tokenHeader
		}
	}

	auth, err := verifyJWTSigning(token, true)
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
		returnErrorJson(c, "您无权限访问")
		c.Abort()
		return
	}

	//POST验证header:csrf
	if c.Request.Method == "POST" {
		err = csrfVerify(c.GetHeader("X-CSRF-TOKEN"), auth)
		if err != nil {
			Logger.Info.Printf("[中间件]csrf验证失败 err:%s,auth=%v", err.Error(), auth)
			returnErrorJson(c, "CSRF-TOKEN Mismatch")
			c.Abort()
			return
		}
	}

	c.Set("auth", auth)
	c.Next()
}

func securityMiddleware(c *gin.Context) {
	c.Header("X-Content-Type-Options", "nosniff")

	//跨站
	c.Header("Access-Control-Allow-Origin", config.General.BaseUrl)
	c.Header("Access-Control-Allow-Headers", "Authorization,X-CSRF-TOKEN")

	//xss
	c.Header("X-XSS-Protection", "1; mode=block;")

}

func middleWareRedirect(c *gin.Context) {
	if strings.Contains(c.FullPath(), "/api/") == true {
		returnErrorJson(c, "未授权访问")
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

//生成csrfToken
func csrfMake(jwtID string, c *gin.Context) (r string, err error) {
	tmp := rdb.Get(ctx, "SIGNIN_APP:csrfToken:"+jwtID).Val()
	if tmp == "" {
		r = hex.EncodeToString(Sha256([]byte(jwtID + strconv.FormatInt(time.Now().UnixNano(), 10) + strconv.FormatInt(int64(rand.Intn(99)), 10))))
		err = rdb.Set(ctx, "SIGNIN_APP:csrfToken:"+jwtID, r, 1*time.Hour).Err()
		if err != nil {
			return "", err
		}
	} else {
		r = tmp
	}
	csrfSetCookie(c, r)
	return r, err
}

// Sha256 sha256散列原始值
func Sha256(data []byte) []byte {
	digest := sha256.New()
	digest.Write(data)
	return digest.Sum(nil)
}

//验证csrfToken
func csrfVerify(token string, auth *JWTStruct) (err error) {
	if auth.ClassId == 0 {
		//未初始化
		return nil
	}
	val := ""
	val, err = rdb.Get(ctx, "SIGNIN_APP:csrfToken:"+auth.ID).Result()
	if val != token {
		return errors.New("CSRF-TOKEN Mismatch")
	}
	return err
}

func csrfSetCookie(c *gin.Context, token string) {
	c.SetCookie("CSRF-TOKEN", token, 1*60*60, "/", "", true, false)
}

func CacheMiddleware(c *gin.Context) {
	if c.Request.Method != "GET" {
		c.Next()
		return
	}
	ext := []string{"gif", "jpg", "png", "js", "css", "js.map"}
	has := false
	for _, s := range ext {
		if strings.HasSuffix(c.Request.URL.Path, s) == true {
			c.Header("cache-control", "max-age=43201")
			has = true
			break
		}
	}
	if has == false {
		c.Header("cache-control", "no-cache")
	}
	c.Next()
}
