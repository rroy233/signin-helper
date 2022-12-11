package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	Logger "github.com/rroy233/logger"
	"github.com/steambap/captcha"
	"net/url"
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

type LoginData struct {
	Email    string `json:"email" binding:"required"`
	PassWord string `json:"password" binding:"required"`
	Captcha  string `json:"captcha"`
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

func loginHandler(c *gin.Context) {
	ip := getIP(c)

	Logger.Info.Println("[登录]IP:", ip, "尝试登录，c.Request:", c.Request)

	loginInfo := new(LoginData)
	err := c.ShouldBindJSON(loginInfo)
	if err != nil {
		returnErrorJson(c, "请输入邮箱地址和密码")
		return
	}

	checkDB()

	//检查黑名单
	num, err := blackListCheck(c)
	if err != nil {
		returnErrorJson(c, err.Error())
		return
	}
	//验证码
	if num > 0 {
		if loginInfo.Captcha == "" {
			returnErrorJson(c, "ResErrorNeedCaptcha")
			return
		}
		//校验验证码
		captchaAnswer := rdb.Get(ctx, "SIGNIN_APP:Captcha:"+MD5(getCookie(c, "_uuid"))).Val()
		if captchaAnswer == "" {
			returnErrorJson(c, "请重新获取验证码")
			return
		}
		rdb.Del(ctx, "SIGNIN_APP:Captcha:"+MD5(getCookie(c, "_uuid")))
		if strings.ToLower(loginInfo.Captcha) != captchaAnswer {
			blackListStore(c)
			returnErrorJson(c, "验证码错误")
			return
		}
		//清除黑名单
		blackListDel(c)
	}

	if govalidator.IsEmail(loginInfo.Email) == false {
		blackListStore(c)
		returnErrorJson(c, "邮箱地址无效")
		return
	}

	user := new(dbUser)
	err = db.Get(user, "SELECT * FROM `user` WHERE `email` = ?", loginInfo.Email)
	if err != nil {
		Logger.Info.Println("登录失败，用户不存在.IP:", getIP(c))
		returnErrorJson(c, "邮箱或密码错误")
		blackListStore(c)
		return
	}

	if user.Password != MD5(loginInfo.PassWord+config.General.MD5Salt) {
		//登录失败
		Logger.Info.Println("登录失败，邮箱或密码错误.IP:", getIP(c), ":", c.Request)
		returnErrorJson(c, "邮箱或密码错误")
		blackListStore(c)
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

	ress := new(ResLoginSuccess)
	ress.Status = 0
	ress.Msg = "success"
	ress.Data.Token = token
	c.JSON(200, ress)

	return
}

func registerHandler(c *gin.Context) {

}

func forgetHandler(c *gin.Context) {

}

func captchaHandler(c *gin.Context) {
	loggerPrefix := "[captchaHandler]"
	res := new(ResCaptcha)
	res.Status = 0

	data, err := captcha.New(210, 70)
	if err != nil {
		Logger.Error.Println(loggerPrefix+"生成验证码失败：", err)
		returnErrorJson(c, "生成验证码失败")
		return
	}

	img := bytes.NewBuffer(nil)
	err = data.WriteImage(img)
	if err != nil {
		Logger.Error.Println(loggerPrefix+"验证码图片输出失败：", err)
		returnErrorJson(c, "生成验证码失败")
		return
	}

	//存储
	uid := uuid.New().String()
	setCookie(c, "_uuid", uid, 3600)
	rdb.Set(ctx, "SIGNIN_APP:Captcha:"+MD5(uid), strings.ToLower(data.Text), 5*time.Minute)
	//验证时使用:
	//captchaAnswer := rdb.Get(ctx, "SIGNIN_APP:Captcha:"+MD5(getCookie(c, "_uuid"))).Val()

	res.Data.Image = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(img.Bytes())

	c.JSON(200, res)
	return
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

		c.Redirect(302, "/user/init")
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

// 吊销jwt
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

func loginByTokenHandler(c *gin.Context) {
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

func blackListDel(c *gin.Context) {
	ip := getIP(c)
	//查看黑名单
	if rdb.Get(ctx, "BL:Account:"+ip).Val() != "" {
		rdb.Del(ctx, "BL:Account:"+ip)
	}
}

func blackListCheck(c *gin.Context) (int, error) {
	ip := getIP(c)
	//查看黑名单
	if rdb.Get(ctx, "SIGNIN_APP:BL:"+ip).Val() != "" {
		return 0, errors.New("您已被封禁30分钟")
	}

	loginNum := rdb.Get(ctx, "SIGNIN_APP:BL:Account:"+ip).Val()
	if loginNum == "" {
		rdb.Set(ctx, "SIGNIN_APP:BL:Account:"+ip, 0, 3*time.Minute)
	}
	loginLimitInt, _ := strconv.Atoi(loginNum)
	if loginLimitInt > 30 {
		//放入黑名单，30分钟解除
		rdb.Set(ctx, "SIGNIN_APP:BL:"+ip, "BL", 30*time.Minute)
		Logger.Info.Println("[黑名单]新狱友IP:", ip)
		return loginLimitInt, errors.New("您已被封禁30分钟")
	}
	return loginLimitInt, nil
}

func blackListStore(c *gin.Context) {
	ip := getIP(c)
	if rdb.Get(ctx, "SIGNIN_APP:BL:Account:"+ip).Val() == "" {
		rdb.Set(ctx, "SIGNIN_APP:BL:Account:"+ip, 1, 24*time.Hour)
	} else {
		rdb.Incr(ctx, "SIGNIN_APP:BL:Account:"+ip)
	}
}
