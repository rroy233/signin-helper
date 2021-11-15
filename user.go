package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	wx "github.com/wxpusher/wxpusher-sdk-go"
	wxModel "github.com/wxpusher/wxpusher-sdk-go/model"
	"signin/Logger"
	"strconv"
	"strings"
	"time"
)

type FormDataUserInit struct {
	ClassCode string `json:"class_code" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type FormDataSignIn struct {
	ActToken string `json:"act_token" binding:"required"`
}

type FormDataUserNotiEdit struct {
	NotiType string `json:"noti_type" binding:"required"`
}

type FormDataNotiCheck struct {
	Token string `json:"token" binding:"required"`
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
		dbretClass, err := db.Exec("INSERT INTO `class` (`class_id`, `name`, `class_code`, `total`) VALUES (NULL, ?, ?, ?);",
			"新建班级",
			MD5_short(strconv.FormatInt(time.Now().UnixNano(), 10)),
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

	actList, err := getActIDs(auth.ClassId)
	if err != nil {
		Logger.Error.Println("[个人信息查询]获取班级活动列表失败:", err)
		returnErrorJson(c, "系统异常")
		return
	}

	res := new(ResUserActInfo)
	if len(actList) == 0 {
		res.Msg = "当前无生效中的活动"
		c.JSON(200, res)
		return
	}
	res.Data.List = make([]*userActInfo, 0)

	for i := range actList {
		actItem := new(userActInfo)
		act, err := getAct(actList[i])
		if err != nil {
			Logger.Error.Println("[个人信息查询]活动信息查询失败:", err)
			returnErrorJson(c, "查询失败(系统异常或是班级负责人配置错误)")
			return
		}
		//获取统计数据
		sts, err := getActStatistics(actList[i])
		if err != nil {
			returnErrorJson(c, "查询统计数据失败")
			return
		}
		actItem.Statistic.Done = sts.Done
		actItem.Statistic.Total = sts.Total
		//完成情况概述
		stsInfo := ""
		if sts.Done == sts.Total {
			stsInfo = "🎉所有同学都完成啦🎉"
		} else {
			stsInfo = "还有"
			for j := 0; j < 3 && j < len(sts.UnfinishedList); j++ {
				if j != 0 {
					stsInfo += "、"
				}
				stsInfo += sts.UnfinishedList[j].Name
			}
			if sts.Total-sts.Done > 3 {
				stsInfo += "等" + strconv.FormatInt(int64(sts.Total-sts.Done), 10) + "名同学未完成👀"
			} else {
				stsInfo += "这" + strconv.FormatInt(int64(sts.Total-sts.Done), 10) + "名同学未完成👀"
			}
		}
		actItem.Statistic.Info = stsInfo

		//存储actToken
		actToken := MD5_short(auth.ID+fmt.Sprintf("%d",act.ActID))
		rdb.Set(ctx, "SIGNIN_APP:actToken:"+actToken, strconv.FormatInt(int64(act.ActID), 10), 10*time.Minute)

		actItem.ActToken = actToken
		actItem.ActName = act.Name
		actItem.ActAnnouncement = act.Announcement
		//判断是否需要使用默认图片
		if act.Pic == "" {
			actItem.ActPic = "/static/image/default.jpg"
		} else {
			actItem.ActPic = act.Pic
		}
		actItem.BeginTime = ts2DateString(act.BeginTime)

		//结束时间描述
		et,_ := strconv.ParseInt(act.EndTime,10,64)
		//tm次日凌晨时间
		tm := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, TZ).AddDate(0, 0, 1)
		if et < tm.Unix() {
			//今天
			actItem.EndTime = "今天" + time.Unix(et, 0).In(TZ).Format("15:04")
		}else if et < tm.AddDate(0,0,1).Unix(){
			//明天
			actItem.EndTime = "明天" + time.Unix(et, 0).In(TZ).Format("15:04")
		}else if et < tm.AddDate(0,0,2).Unix(){
			//后天
			actItem.EndTime = "后天" + time.Unix(et, 0).In(TZ).Format("15:04")
		}else{
			actItem.EndTime = ts2DateString(act.EndTime)
		}

		//查询是否已参与
		logId := 0
		_ = db.Get(&logId, "select `log_id` from `signin_log` where `user_id`=? and `act_id`=?",
			auth.UserID,
			act.ActID)
		if logId == 0 {
			actItem.Status = 0 //未参与
		} else {
			actItem.Status = 1 //已参与
		}
		res.Data.List = append(res.Data.List, actItem)
		res.Data.Total++
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

	actToken := c.Query("act_token")
	if actToken == "" {
		returnErrorJson(c, "参数无效")
		return
	}

	actID, err := queryActIdByActToken(actToken)
	if err != nil {
		Logger.Info.Println("[签到]从redis查找活动id失败", err, auth)
		returnErrorJson(c, "参数无效(-2)")
		return
	}

	sts, err := getActStatistics(actID)
	if err != nil {
		returnErrorJson(c, err.Error())
	}

	res := new(ResUserActStatistic)
	res.Status = 0
	res.Data.Done = sts.Done
	res.Data.Total = sts.Total
	res.Data.FinishedList = make([]actStatisticUser, 0)
	res.Data.UnfinishedList = make([]actStatisticUser, 0)

	for i := range sts.FinishedList {
		res.Data.FinishedList = append(res.Data.FinishedList, actStatisticUser{
			Id:   sts.FinishedList[i].Id,
			Name: sts.FinishedList[i].Name,
		})
	}
	for i := range sts.UnfinishedList {
		res.Data.UnfinishedList = append(res.Data.UnfinishedList, actStatisticUser{
			Id:   sts.UnfinishedList[i].Id,
			Name: sts.UnfinishedList[i].Name,
		})
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

	//查询正在生效的活动id
	ActiveActIDs, err := getActIDs(auth.ClassId)
	if err != nil {
		Logger.Error.Println("[签到]活动id查找失败", err, auth)
		returnErrorJson(c, "系统异常(-1)")
		return
	}

	//从redis查找活动id
	actID, err := queryActIdByActToken(form.ActToken)
	if err != nil {
		Logger.Info.Println("[签到]从redis查找活动id失败", err, auth)
		returnErrorJson(c, "参数无效(-2)")
		return
	}

	//判断是否正在生效
	if existIn(ActiveActIDs, actID) == false {
		Logger.Info.Println("[签到]从redis查找活动，活动已失效", auth)
		returnErrorJson(c, "当前活动已过期")
		return
	}

	//查询是否已参与
	logId := 0
	_ = db.Get(&logId, "select `log_id` from `signin_log` where `user_id`=? and `act_id`=?",
		auth.UserID,
		actID)
	if logId != 0 {
		Logger.Info.Println("[签到]重复参与", err, auth)
		returnErrorJson(c, "请勿重复参与")
		return
	}

	//活动活动信息
	act, err := getAct(actID)
	if err != nil {
		Logger.Info.Println("[签到]获取活动信息失败", err, auth)
		returnErrorJson(c, "系统异常(-2)")
		return
	}

	//写入log表
	_, err = db.Exec("INSERT INTO `signin_log` (`log_id`, `class_id`, `act_id`, `user_id`, `create_time`) VALUES (NULL, ?, ?, ?, ?);",
		auth.ClassId,
		actID,
		auth.UserID,
		strconv.FormatInt(time.Now().Unix(), 10))
	if err != nil {
		Logger.Info.Println("[签到]写入log表失败", err, auth)
		returnErrorJson(c, "系统异常，请联系管理员")
		return
	}

	//检查提醒次数，判断是否需要推送提醒
	notiTimes,err := actNotiUserTimesGet(act,auth.UserID)
	if err == nil {
		//成功获取
		if notiTimes > 6{
			noti,err := makeActInnerNoti(actID,auth.UserID,ACT_NOTI_TYPE_CH_NOTI)
			err = pushInnerNoti(auth.UserID,noti)
			if err != nil {
				Logger.Error.Println("[签到][检查提醒次数]推送消息失败:",err)
			}else{
				err=actNotiUserTimesDel(act,auth.UserID)
				if err != nil {
					Logger.Error.Println("[签到][检查提醒次数]删除提醒次数失败:",err)
				}
			}
		}
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
	} else if notiType == NOTIFICATION_TYPE_WECHAT {
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
	} else {
		returnErrorJson(c, "参数无效(-2)")
		return
	}

	//检查是否绑定微信
	wxID := ""
	err = db.Get(&wxID, "select `wx_pusher_uid` from `user` where `user_id`=?", auth.UserID)
	if err != nil {
		Logger.Error.Println("[更改通知方式]查询mysql异常", err)
		returnErrorJson(c, "系统异常")
		return
	}

	if notiType == 2 && wxID == "" {
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
				ActToken: "",
				ActName:  "(活动不存在)",
				DateTime: "null",
			})
		} else {
			//存储actToken
			actToken := MD5_short(auth.ID+fmt.Sprintf("%d",act.ActID))
			rdb.Set(ctx, "SIGNIN_APP:actToken:"+actToken, strconv.FormatInt(int64(logs[i].ActID), 10), 10*time.Minute)
			res.Data.List = append(res.Data.List, resActLogItem{
				Id:       id,
				ActToken: actToken,
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

	actToken := c.Query("act_token")
	if actToken == "" {
		returnErrorJson(c, "参数无效(-1)")
		return
	}

	actID, err := queryActIdByActToken(actToken)
	if err != nil {
		Logger.Info.Println("[签到]从redis查找活动id失败", err, auth)
		returnErrorJson(c, "参数无效(-2)")
		return
	}

	act, err := getAct(actID)
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


func UserNotiCheckHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	form := new(FormDataNotiCheck)
	err = c.ShouldBindJSON(form)
	if err != nil {
		Logger.Info.Println("[用户信息已读]解析参数错误:",err,auth)
		returnErrorJson(c,"参数无效(-1)")
		return
	}

	noti,err := rdb.Exists(ctx,"SIGNIN_APP:UserNoti:USER_"+auth.UserIdString()+":"+form.Token).Result()
	if noti != int64(1) || err != nil{
		Logger.Info.Println("[用户信息已读]参数无效:",err,auth)
		returnErrorJson(c,"参数无效(-2)")
		return
	}

	rdb.Del(ctx,"SIGNIN_APP:UserNoti:USER_"+auth.UserIdString()+":"+form.Token)

	res := new(ResEmpty)
	res.Status = 0
	c.JSON(200,res)
	return
}

func UserNotiFetchHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	//SIGNIN_APP:UserNoti:USER_{{USER_ID}}:{{noti_token}}
	keys := rdb.Keys(ctx,"SIGNIN_APP:UserNoti:USER_"+auth.UserIdString()+":*").Val()
	if len(keys) == 0{
		res := new(ResEmpty)
		res.Status = 0
		c.JSON(200,res)
		return
	}

	res := new(ResUserNotiFetch)
	res.Status = 0
	res.Data = make([]*UserNotiFetchItem,0)
	for i:= range keys {
		key := strings.Split(keys[i],":")
		if len(key) != 4{
			Logger.Error.Println("[拉取用户消息]keys异常:",key)
			continue
		}
		item := new(UserNotiFetchItem)
		err = json.Unmarshal([]byte(rdb.Get(ctx,"SIGNIN_APP:UserNoti:USER_"+auth.UserIdString()+":"+key[3]).Val()),item)
		if err != nil {
			Logger.Error.Println("[拉取用户消息]json反序列化失败:",err,key)
			continue
		}
		res.Data = append(res.Data,item)
	}

	c.JSON(200,res)
}
