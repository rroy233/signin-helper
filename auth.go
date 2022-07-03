package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"net/url"
	"signin/Logger"
	"strconv"
	"strings"
	"time"
)

type JWTStruct struct {
	UserID  int    `json:"user_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	ClassId int    `json:"class_id"`
	IsAdmin int    `json:"is_admin"`
	jwt.RegisteredClaims
}

type WxPusherCallback struct {
	Action string `json:"action"`
	Data   struct {
		AppID       int    `json:"appId"`
		AppKey      string `json:"appKey"`
		AppName     string `json:"appName"`
		Source      string `json:"source"`
		UserName    string `json:"userName"`
		UserHeadImg string `json:"userHeadImg"`
		Time        int64  `json:"time"`
		UID         string `json:"uid"`
		Extra       string `json:"extra"`
	} `json:"data"`
}

func (j *JWTStruct) ClassIdString() string {
	return strconv.FormatInt(int64(j.ClassId), 10)
}

func (j *JWTStruct) UserIdString() string {
	return strconv.FormatInt(int64(j.UserID), 10)
}

func ssoRedirectHandler(c *gin.Context) {
	redirectUrl, _ := url.PathUnescape(c.Query("redirect"))
	if redirectUrl == "" {
		redirectUrl = "null"
	}
	state := MD5_short(strconv.FormatInt(time.Now().UnixNano(), 10))
	rdb.Set(ctx, "SIGNIN_APP:state:"+state, redirectUrl, 5*time.Minute)
	c.Redirect(302, ssoClient.RedirectUrl(state))
}

func generateJwt(user *dbUser, JwtID string, expTime time.Duration) (s string, err error) {
	playLoad := &JWTStruct{
		UserID:  user.UserID,
		Name:    user.Name,
		Email:   user.Email,
		ClassId: user.Class,
		IsAdmin: user.IsAdmin,
	}
	playLoad.ExpiresAt = jwt.NewNumericDate(time.Now().Add(expTime))
	//playLoad.ExpiresAt = jwt.NewNumericDate(time.Now().Add(5*time.Second))
	playLoad.ID = JwtID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, playLoad)
	s, err = token.SignedString([]byte(config.General.JwtKey))
	if err != nil {
		return s, err
	}

	//存redis
	rdb.Set(ctx, fmt.Sprintf("SIGNIN_APP:JWT:USER_%d:%s", user.UserID, JwtID), user.UserID, expTime)

	return s, nil
}

func verifyJWTSigning(tokenString string, checkRedis bool) (auth *JWTStruct, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.General.JwtKey), nil
	})
	if err != nil {
		return auth, err
	}
	if checkRedis == true {
		auth = new(JWTStruct)
		var t *jwt.Token
		t, err = jwt.ParseWithClaims(tokenString, &JWTStruct{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.General.JwtKey), nil
		})
		if claims, ok := t.Claims.(*JWTStruct); ok && token.Valid {
			uid := rdb.Get(ctx, fmt.Sprintf("SIGNIN_APP:JWT:USER_%d:%s", claims.UserID, claims.ID)).Val()
			if uid != strconv.FormatInt(int64(claims.UserID), 10) {
				err = errors.New("JWT已失效")
				return nil, err
			}
			auth = claims
		} else {
			fmt.Println(err)
		}
	}
	return auth, err
}

func ssoCallBackHandler(c *gin.Context) {

	accessToken := c.Query("access_token")
	state := c.Query("state")

	if accessToken == "" || state == "" {
		returnErrorView(c, "参数无效(-1)(state无效)")
		return
	}

	redirectUrl := rdb.Get(ctx, "SIGNIN_APP:state:"+state).Val()
	if redirectUrl == "" {
		returnErrorView(c, "参数无效(-2)")
		return
	}
	if redirectUrl == "null" {
		redirectUrl = "/user/"
	}

	userInfo, err := ssoClient.GetUserInfo(accessToken)
	if err != nil {
		Logger.Info.Println("登录失败:Failed To Get UserInfo!" + err.Error())
		returnErrorView(c, "登录失败:Failed To Get UserInfo!")
		return
	}

	checkDB()
	user := new(dbUser)
	err = db.Get(user, "select * from `user` where `sso_uid`=?", userInfo.UserID)
	if err != nil {
		//未初始化
		//定义管理员群组
		isAdmin := 0
		if userInfo.UserGroup == "7" {
			isAdmin = 1
		}
		//创建用户
		userForm := &dbUser{
			Email:            userInfo.Email,
			Class:            0,
			IsAdmin:          isAdmin,
			SsoUid:           userInfo.UserID,
			NotificationType: NOTIFICATION_TYPE_EMAIL,
		}
		uid, err := createUser(userForm)
		if err != nil {
			Logger.Error.Println("[创建用户]失败:", err)
			returnErrorView(c, "系统异常(-1)")
			return
		}
		//签发临时JWT
		userForm.UserID = uid
		JID := generateJwtID()
		token, err := generateJwt(userForm, JID, 10*time.Minute)
		if err != nil {
			Logger.Error.Println("[签发临时JWT]失败:", err)
			returnErrorView(c, "系统异常(-2)")
			return
		}

		//存入cookie
		storeToken(c, token)

		c.Redirect(302, "/user/reg")
		return
	}

	//正常签发jwt
	expTime := time.Hour
	if config.General.Production == false {
		expTime = 13 * time.Hour
		Logger.Debug.Println("[JWT]已颁发测试环境JWT")
	}
	jID := generateJwtID()
	token, err := generateJwt(user, jID, expTime)
	if err != nil {
		Logger.Error.Println("[正常签发JWT]失败:", err)
		returnErrorView(c, "系统异常")
		return
	}
	storeToken(c, token) //存入cookie

	_, err = csrfMake(jID, c)
	if err != nil {
		Logger.Info.Println("[用户csrf]发生错误", err)
		returnErrorView(c, "返回csrfToken失败")
		return
	}

	c.Redirect(302, redirectUrl)

}

//吊销jwt
func killJwtByJID(jID string) error {
	keys := rdb.Keys(ctx, "*:"+jID).Val()
	if len(keys) == 0 {
		return errors.New("jwt不存在")
	}
	if strings.Contains(keys[0], "SIGNIN_APP") == false {
		return errors.New("jwt不存在，前缀不匹配")
	}
	return doKillJWT(keys[0])
}

func doKillJWT(JwtKey string) error {
	r, err := rdb.Del(ctx, JwtKey).Result()
	if r == int64(0) || err != nil {
		if err != nil {
			return errors.New("吊销失败:" + err.Error())
		} else {
			return errors.New("吊销失败:未知错误")
		}
	}
	return nil
}

func killJwtByUID(UID int) {
	keys := rdb.Keys(ctx, fmt.Sprintf("SIGNIN_APP:JWT:USER_%d:*", UID)).Val()
	if len(keys) == 0 {
		return
	}
	for i := range keys {
		if err := doKillJWT(keys[i]); err != nil {
			Logger.Error.Println("[killJwtByUID]发生错误:", err, keys[i])
		}
	}
	return
}

func loginHandler(c *gin.Context) {
	tmp := c.Query("jwt")
	if tmp == "" {
		returnErrorView(c, "参数无效(-1)")
		return
	}

	tokens := strings.Split(tmp, ".")
	if len(tokens) != 2 || (len(tokens) == 2 && tokens[1] != Cipher.Sha256Hex([]byte(tokens[0]))) {
		Logger.Info.Printf("[login]query:%s,error:%s", tmp, "参数无效")
		returnErrorView(c, "参数无效")
		return
	}

	token, err := Cipher.Decrypt(tokens[0])
	if err != nil {
		Logger.Info.Printf("[login]query:%s,error:%s", tmp, "解密失败")
		returnErrorView(c, "token无效")
		return
	}

	auth, err := verifyJWTSigning(token, true)
	if err != nil {
		Logger.Info.Printf("[login]token:%s,error:%s", token, err.Error())
		returnErrorView(c, "token已过期")
		return
	}
	storeToken(c, token)

	_, err = csrfMake(auth.ID, c)
	if err != nil {
		Logger.Info.Println("[用户csrf]发生错误", err)
		returnErrorView(c, "返回csrfToken失败")
		return
	}

	c.Redirect(302, "/")

}

func logoutHandler(c *gin.Context) {

	token, err := c.Cookie("token")
	if err != nil {
		returnErrorJson(c, "参数无效(-1)")
		return
	}

	auth, err := verifyJWTSigning(token, true)
	if token == "" {
		returnErrorJson(c, "参数无效(-2)")
		return
	}

	//吊销凭证
	err = killJwtByJID(auth.ID)
	if err != nil {
		returnErrorJson(c, "吊销失败")
		return
	}

	//清除cookie
	c.SetCookie("token", "", -1, "/", "", true, true)

	res := new(ResEmpty)
	res.Status = 0
	c.JSON(200, res)
	return
}

func wxpusherCallbackHandler(c *gin.Context) {
	callBackData := new(WxPusherCallback)
	err := c.ShouldBindJSON(callBackData)
	if err != nil {
		Logger.Error.Println("[微信扫码回调]参数绑定失败:", err, c)
		c.Status(400)
		return
	}

	extra := callBackData.Data.Extra
	wxUserid := callBackData.Data.UID
	if wxUserid == "" {
		Logger.Info.Println("[微信扫码回调]wxUserid为空:", c)
		c.Status(400)
		return
	}

	userID, err := rdb.Get(ctx, "SIGNIN_APP:Wechat_Bind:"+extra).Result()
	token, err := rdb.Get(ctx, " SIGNIN_APP:Wechat_Bind:"+userID).Result()
	if err != nil {
		Logger.Info.Println("[微信扫码回调]参数无效:", err)
		c.Status(400)
		return
	}
	if userID == "" && token == "" {
		c.Status(400)
		return
	}

	//存储
	_, err = db.Exec("update `user` set `wx_pusher_uid`=? , `notification_type`=? where `user_id`=?", wxUserid, NOTIFICATION_TYPE_WECHAT, userID)
	if err != nil {
		Logger.Error.Println("[微信扫码回调]存储mysql失败:", err)
		c.Status(400)
		return
	}

	//设置redis
	err = rdb.Set(ctx, "SIGNIN_APP:Wechat_Bind:"+extra, "DONE", redis.KeepTTL).Err()
	err = rdb.Set(ctx, " SIGNIN_APP:Wechat_Bind:"+userID, "DONE", redis.KeepTTL).Err()
	if err != nil {
		Logger.Error.Println("[微信扫码回调]存储redis失败:", err)
		c.Status(400)
		return
	}

	//发送欢迎信息
	task := new(NotifyJob)
	task.NotificationType = NOTIFICATION_TYPE_WECHAT
	task.Addr = wxUserid
	task.Title = "【成功绑定】您已成功绑定微信账号。请确保您处于\"接受信息\"状态(即不要关闭公众号的消息开关)，不要\"删除订阅\"，如果出现问题导致无法推送，请联系管理员。"
	task.Body = "请确保您处于\"接受信息\"状态(即不要关闭公众号的消息开关)，不要\"删除订阅\"，如果出现问题导致无法推送，请联系管理员。"
	WechatQueue <- task

	c.Status(200)
	return

}

func getAuthFromContext(c *gin.Context) (*JWTStruct, error) {
	i, exist := c.Get("auth")
	if exist == false {
		return nil, errors.New("上下文中auth不存在")
	}
	if auth, ok := i.(*JWTStruct); ok {
		return auth, nil
	} else {
		return nil, errors.New("解析auth失败")
	}
}

func storeToken(c *gin.Context, token string) {
	c.SetCookie("token", token, 1*60*60, "/", "", true, true)
}

func generateJwtID() string {
	return MD5_short(strconv.FormatInt(time.Now().UnixNano(), 10) + config.General.JwtKey)
}
