package main

import (
	"github.com/gin-gonic/gin"
	wx "github.com/wxpusher/wxpusher-sdk-go"
	wxModel "github.com/wxpusher/wxpusher-sdk-go/model"
	"signin/Logger"
	"strconv"
	"time"
)

type FormDataUserInit struct {
	ClassCode string `json:"class_code" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type FormDataSignIn struct {
	TS string `json:"ts" binding:"required"`
}

type FormDataUserNotiEdit struct {
	NotiType string `json:"noti_type" binding:"required"`
}

func initHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}
	if auth.ClassId != 0 {
		returnErrorJson(c, "您无需再初始化")
		return
	}

	form := new(FormDataUserInit)
	err = c.ShouldBindJSON(form)
	if err != nil {
		Logger.Info.Println("[初始化]数据绑定错误", err)
		returnErrorJson(c, "数据无效")
		return
	}

	checkDB()

	//查询班级是否存在
	classId := 0
	err = db.Get(&classId, "select `class_id` from `class` where `class_code`=?", form.ClassCode)
	if auth.IsAdmin == 0 && (err != nil || classId == 0) {
		returnErrorJson(c, "班级码无效")
		return
	}

	//判断是否新建班级
	if auth.IsAdmin == 1 && form.ClassCode == "new" {
		//新建班级
		dbretClass, err := db.Exec("INSERT INTO `class` (`class_id`, `name`, `class_code`, `total`, `act_id`) VALUES (NULL, ?, ?, ?, ?);",
			"新建班级",
			MD5_short(strconv.FormatInt(time.Now().UnixNano(), 10)),
			0,
			0,
		)
		if err != nil {
			Logger.Error.Println("[初始化][管理员]创建班级失败", err, auth)
			returnErrorJson(c, "创建班级失败")
			return
		}
		tmpClass, err := dbretClass.LastInsertId()
		classId = int(tmpClass)

		//创建活动
		dbretAct, err := db.Exec("INSERT INTO `activity` (`act_id`, `class_id`, `name`, `announcement`, `cheer_text`, `pic`, `begin_time`, `end_time`, `create_time`, `update_time`, `create_by`) VALUES (NULL, ?, ?, ?, '恭喜', '', ?, ?, ?,?, ?);",
			classId,
			"新建活动",
			"默认公告",
			strconv.FormatInt(time.Now().Unix(), 10),
			strconv.FormatInt(time.Now().Unix(), 10),
			strconv.FormatInt(time.Now().Unix(), 10),
			strconv.FormatInt(time.Now().Unix(), 10),
			auth.UserID,
		)
		if err != nil {
			Logger.Error.Println("[初始化][管理员]创建活动失败", err, auth)
			returnErrorJson(c, "创建活动失败")
			return
		}
		tmpAct, err := dbretAct.LastInsertId()
		actId := int(tmpAct)

		//更新班级actID
		_, err = db.Exec("update `class` set `act_id`=? where `class_id`=?", actId, classId)
		if err != nil {
			Logger.Error.Println("[初始化][管理员]更新班级actID", err, auth)
			returnErrorJson(c, "更新班级actID失败")
			return
		}

		//更新缓存
		_, err = cacheClass(classId)
		if err != nil {
			Logger.Error.Println("[初始化][管理员]更新class缓存", err, auth)
		}
		_, err = cacheAct(actId)
		if err != nil {
			Logger.Error.Println("[初始化][管理员]更新act缓存", err, auth)
		}
	}

	//更新数据库
	_, err = db.Exec("UPDATE `user` SET `name` = ?,`class` = ? WHERE `user`.`user_id` = ?", form.Name, classId, auth.UserID)
	if err != nil {
		Logger.Error.Println("[初始化]更新用户信息失败:", err)
		returnErrorJson(c, "系统异常")
		return
	}
	//更新班级人数
	_, err = db.Exec("UPDATE `class` SET `total` = `total`+1 WHERE `class`.`class_id` = ?", classId)
	if err != nil {
		Logger.Error.Println("[初始化]更新班级人数失败:", err)
	}

	//重新拉取用户信息
	user := new(dbUser)
	err = db.Get(user, "select * from `user` where `user_id` = ?", auth.UserID)
	if err != nil {
		Logger.Error.Println("[初始化]重新拉取用户信息失败:", err)
		returnErrorJson(c, "系统异常")
		return
	}

	//生成新的JWT
	newJwt, err := generateJwt(user, generateJwtID(), 1*time.Hour)
	if err != nil {
		Logger.Error.Println("[初始化]生成新的JWT失败:", err)
		returnErrorJson(c, "系统异常")
		return
	}

	//返回新的JWT,setcookie
	c.SetCookie("token", newJwt, 60*60, "/", "", true, true)
	res := new(ResUserInit)
	res.Status = 0
	res.Data.NewJWT = newJwt

	//吊销旧jwt
	err = killJwt(auth.ID)
	if err != nil {
		Logger.Error.Println("[初始化]吊销旧jwt失败:", err)
		res.Status = -1
		res.Msg = "成功了但并未完全成功，请重新登录"
	}

	c.JSON(200, res)
	return
}

func profileHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	//获取班级
	class, err := getClass(auth.ClassId)
	if err != nil {
		Logger.Error.Println("[个人信息查询]班级信息查询失败:", err)
		returnErrorJson(c, "查询失败")
		return
	}

	res := new(ResUserProfile)
	res.Status = 0
	res.Data.UserId = auth.UserID
	res.Data.UserName = auth.Name
	res.Data.Email = auth.Email
	res.Data.ClassName = class.Name
	res.Data.ClassCode = class.ClassCode
	res.Data.IsAdmin = auth.IsAdmin

	c.JSON(200, res)

}

func UserActInfoHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	//获取班级
	class, err := getClass(auth.ClassId)
	if err != nil {
		Logger.Error.Println("[个人信息查询]班级信息查询失败:", err)
		returnErrorJson(c, "查询失败")
		return
	}

	//获取班级
	act, err := getAct(class.ActID)
	if err != nil {
		Logger.Error.Println("[个人信息查询]活动信息查询失败:", err)
		returnErrorJson(c, "查询失败(系统异常或是班级负责人配置错误)")
		return
	}

	res := new(ResUserActInfo)
	res.Status = 0
	res.Data.ActID = act.ActID
	res.Data.ActName = act.Name
	res.Data.ActAnnouncement = act.Announcement
	//判断是否需要使用默认图片
	if act.Pic == "" {
		res.Data.ActPic = "/static/image/default.jpg"
	} else {
		res.Data.ActPic = act.Pic
	}
	res.Data.BeginTime = ts2DateString(act.BeginTime)
	res.Data.EndTime = ts2DateString(act.EndTime)

	//查询是否已参与
	logId := 0
	_ = db.Get(&logId, "select `log_id` from `signin_log` where `user_id`=? and `act_id`=?",
		auth.UserID,
		act.ActID)
	if logId == 0 {
		res.Data.Status = 0 //未参与
	} else {
		res.Data.Status = 1 //已参与
	}

	c.JSON(200, res)
	return

}

func UserActStatisticHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	res, err := getClassStatistics(auth.ClassId)
	if err != nil {
		returnErrorJson(c, err.Error())
	}

	c.JSON(200, res)
}

func UserActSigninHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	form := new(FormDataSignIn)
	err = c.ShouldBindJSON(form)
	if err != nil {
		Logger.Info.Println("[签到]json绑定失败", err, auth)
		returnErrorJson(c, "参数无效(-1)")
		return
	}

	//解析时间戳
	//ts, err := strconv.ParseInt(form.TS, 10, 64)
	//if err != nil {
	//	Logger.Info.Println("[签到]解析时间戳失败", err, auth)
	//	returnErrorJson(c, "参数无效(-2)")
	//	return
	//}
	//nowTs := time.Now().Unix()
	//if !(ts <= nowTs && ts > nowTs-2) && config.General.Production == true {
	//	Logger.Info.Println("[签到]时间戳不合法->", ts, nowTs, auth)
	//	returnErrorJson(c, "参数无效(-3)")
	//	return
	//}

	//TODO 性能优化
	//获取班级
	class, err := getClass(auth.ClassId)
	if err != nil {
		Logger.Info.Println("[签到]班级查找失败", err, auth)
		returnErrorJson(c, "系统异常")
		return
	}

	//查找活动
	act, err := getAct(class.ActID)
	if err != nil {
		Logger.Info.Println("[签到]活动查找失败", err, auth)
		returnErrorJson(c, "参数无效(-4)")
		return
	}

	//查询是否已参与
	logId := 0
	_ = db.Get(&logId, "select `log_id` from `signin_log` where `user_id`=? and `act_id`=?",
		auth.UserID,
		act.ActID)
	if logId != 0 {
		Logger.Info.Println("[签到]重复参与", err, auth)
		returnErrorJson(c, "请勿重复签到")
		return
	}

	//写入log表
	_, err = db.Exec("INSERT INTO `signin_log` (`log_id`, `class_id`, `act_id`, `user_id`, `create_time`) VALUES (NULL, ?, ?, ?, ?);",
		auth.ClassId,
		act.ActID,
		auth.UserID,
		strconv.FormatInt(time.Now().Unix(), 10))
	if err != nil {
		Logger.Info.Println("[签到]写入log表失败", err, auth)
		returnErrorJson(c, "系统异常，请联系管理员")
		return
	}

	res := new(ResUserSignIn)
	res.Status = 0
	res.Data.Text = act.CheerText

	c.JSON(200, res)
	return
}

func UserNotiGetHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	notiType := 0
	err = db.Get(&notiType, "select `notification_type` from `user` where `user_id` = ?", auth.UserID)
	if err != nil {
		Logger.Error.Println("[查询通知方式]", err, auth)
		returnErrorJson(c, "查询失败")
		return
	}

	res := new(ResUserNotiGet)
	res.Status = 0
	if notiType == NOTIFICATION_TYPE_NONE {
		res.Data.NotiType = "none"
	} else if notiType == NOTIFICATION_TYPE_EMAIL {
		res.Data.NotiType = "email"
	}else if notiType == NOTIFICATION_TYPE_WECHAT {
		res.Data.NotiType = "wechat"
	}

	c.JSON(200, res)
	return
}

func UserNotiEditHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	form := new(FormDataUserNotiEdit)
	err = c.ShouldBindJSON(form)
	if err != nil {
		returnErrorJson(c, "参数无效(-1)")
		return
	}

	notiType := 0
	if form.NotiType == "none" {
		notiType = 0
	} else if form.NotiType == "email" {
		notiType = 1
	} else if form.NotiType == "wechat" {
		notiType = 2
	} else{
		returnErrorJson(c, "参数无效(-2)")
		return
	}

	//检查是否绑定微信
	wxID:=""
	err = db.Get(&wxID,"select `wx_pusher_uid` from `user` where `user_id`=?",auth.UserID)
	if err != nil {
		Logger.Error.Println("[更改通知方式]查询mysql异常",err)
		returnErrorJson(c, "系统异常")
		return
	}

	if notiType == 2 && wxID == ""{
		returnErrorJson(c, "您还未绑定微信")
		return
	}

	_, err = db.Exec("UPDATE `user` SET `notification_type` = ? WHERE `user`.`user_id` = ?", notiType, auth.UserID)
	if err != nil {
		Logger.Error.Println("[更改通知方式]", err, auth)
		returnErrorJson(c, "更新失败")
		return
	}

	res := new(ResEmpty)
	res.Status = 0
	c.JSON(200, res)
	return
}

func UserActLogHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	res := new(ResActLog)
	res.Status = 0

	logs := make([]dbLog, 0)
	err = db.Select(&logs, "select * from `signin_log` where `user_id`=? order by `log_id` DESC", auth.UserID)
	if err != nil {
		Logger.Info.Println("[用户查询参与记录]查询log表失败:", err)
		returnErrorJson(c, "系统异常")
		return
	}
	//无记录
	if len(logs) == 0 {
		res.Data.List = nil
		c.JSON(200, res)
		return
	}

	//查询活动信息
	res.Data.List = make([]resActLogItem, 0)
	id := 1
	for i := range logs {
		act, err := getAct(logs[i].ActID)
		if err != nil {
			Logger.Info.Println("[用户查询参与记录]查询活动信息失败", logs[i], err)
			res.Data.List = append(res.Data.List, resActLogItem{
				Id:       id,
				ActId:    0,
				ActName:  "(活动不存在)",
				DateTime: "null",
			})
		} else {
			res.Data.List = append(res.Data.List, resActLogItem{
				Id:       id,
				ActId:    act.ActID,
				ActName:  act.Name,
				DateTime: ts2DateString(logs[i].CreateTime),
			})
		}
		id++
	}

	c.JSON(200, res)
}

func UserActQueryHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	actId, err := strconv.Atoi(c.Query("act_id"))
	if err != nil || actId == 0 {
		Logger.Info.Println("[查询活动详情]过滤参数", err, auth)
		returnErrorJson(c, "参数无效(-1)")
		return
	}

	act, err := getAct(actId)
	if err != nil {
		Logger.Info.Println("[查询活动详情]查询活动", err, auth)
		returnErrorJson(c, "参数无效(-2)")
		return
	}

	//判断是不是同班
	if act.ClassID != auth.ClassId {
		Logger.Info.Println("[查询活动详情]判断是不是同班", auth)
		returnErrorJson(c, "您无权访问此数据")
		return
	}

	res := new(ResUserActQuery)
	res.Status = 0
	res.Data.Name = act.Name
	res.Data.Announcement = act.Announcement
	//判断是否需要使用默认图片
	if act.Pic == "" {
		res.Data.Pic = "/static/image/default.jpg"
	} else {
		res.Data.Pic = act.Pic
	}
	res.Data.CheerText = act.CheerText
	res.Data.BeginTime = ts2DateString(act.BeginTime)
	res.Data.EndTime = ts2DateString(act.EndTime)
	res.Data.UpdateTime = ts2DateString(act.UpdateTime)
	res.Data.CreateTime = ts2DateString(act.CreateTime)

	//查询创建人
	res.Data.CreateBy = queryUserName(act.CreateBy)

	c.JSON(200, res)
	return

}

func UserWechatQrcodeHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	res := new(ResUserWechatQrcode)
	res.Status = 0

	//判断是否已经绑定wx
	dbWxId := ""
	err = db.Get(&dbWxId, "select `wx_pusher_uid` from `user` where `user_id`=?", auth.UserID)
	if err != nil {
		Logger.Error.Println("[微信绑定]查询数据库失败", err)
		returnErrorJson(c, "系统异常(-1)")
		return
	}
	if rdb.Get(ctx, "SIGNIN_APP:Wechat_Bind:"+auth.UserIdString()).Val() == "DONE" || dbWxId != "" {
		res.Status = -1
		res.Msg = "您已完成绑定"
		c.JSON(200, res)
		return
	}

	//生成二维码地址
	Token := MD5_short(strconv.FormatInt(time.Now().Unix(), 10))
	err = rdb.Set(ctx, "SIGNIN_APP:Wechat_Bind:"+Token, auth.UserID, 30*time.Minute).Err()
	err = rdb.Set(ctx, " SIGNIN_APP:Wechat_Bind:"+auth.UserIdString(), Token, 30*time.Minute).Err()
	if err != nil {
		Logger.Error.Println("[微信绑定]查询redis失败", err)
		returnErrorJson(c, "系统异常(-2)")
		return
	}
	qrcode := wxModel.Qrcode{AppToken: config.WxPusher.AppToken, Extra: Token}
	qrcodeResp, err := wx.CreateQrcode(&qrcode)

	res.Data.QrcodeUrl = qrcodeResp.Url
	res.Data.Token = Token
	c.JSON(200, res)
}

func UserWechatBindHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	token := c.Query("token")
	if token == "" {
		returnErrorJson(c, "参数无效(-1)")
		return
	}

	//获取userid对应的extra
	rByToken, err := rdb.Get(ctx, "SIGNIN_APP:Wechat_Bind:"+token).Result()
	if rByToken == "" {
		returnErrorJson(c, "参数无效(-2)")
		return
	}
	rByUID, err := rdb.Get(ctx, " SIGNIN_APP:Wechat_Bind:"+auth.UserIdString()).Result()
	if err != nil {
		Logger.Error.Println("[微信绑定轮询]查询redis失败", err)
		returnErrorJson(c, "系统异常(-1)")
		return
	}

	res := new(ResEmpty)
	res.Status = 0

	//检查SIGNIN_APP:Wechat_Bind:{{user_id}}和 SIGNIN_APP:Wechat_Bind:{{Extra}}是否为DONE
	if rByUID == "DONE" && rByToken == "DONE" {
		res.Status = 1
		res.Msg = "绑定成功"
		c.JSON(200, res)
		return
	}

	c.JSON(200, res)
	return
}
