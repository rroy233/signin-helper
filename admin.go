package main

import (
	"github.com/gin-gonic/gin"
	"net/url"
	"signin/Logger"
	"strconv"
	"strings"
	"time"
)

type FormDataAdminAct struct {
	Name         string `json:"name"`
	Announcement string `json:"announcement"`
	Pic          string `json:"pic"`
	CheerText    string `json:"cheer_text"`
	BeginTime    struct {
		D string `json:"d"`
		T string `json:"t"`
	} `json:"begin_time"`
	EndTime struct {
		D string `json:"d"`
		T string `json:"t"`
	} `json:"end_time"`
}

type FormDataAdminClassEdit struct {
	ClassName string `json:"class_name"`
	ClassCode string `json:"class_code"`
}

func adminActInfoHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	//获取班级
	class, err := getClass(auth.ClassId)
	if err != nil {
		Logger.Error.Println("[管理员][获取活动信息]获取班级失败", err, auth)
		returnErrorJson(c, "系统异常(-1)")
		return
	}

	//获取活动
	act, err := getAct(class.ActID)
	if err != nil {
		Logger.Error.Println("[管理员][获取活动信息]获取活动失败", err, auth)
		returnErrorJson(c, "系统异常(-2)")
		return
	}

	//数据处理
	res := new(ResAdminActInfo)
	res.Status = 0
	res.Data.Name = act.Name
	res.Data.Announcement = act.Announcement
	if act.Pic == "" {
		res.Data.Pic = "/static/image/default.jpg"
	} else {
		res.Data.Pic = act.Pic
	}
	res.Data.CheerText = act.CheerText

	//处理时间
	bt := strings.Split(ts2DateString(act.BeginTime), " ")
	et := strings.Split(ts2DateString(act.EndTime), " ")
	res.Data.BeginTime.D = bt[0]
	res.Data.BeginTime.T = bt[1][:5]
	res.Data.EndTime.D = et[0]
	res.Data.EndTime.T = et[1][:5]

	//返回
	c.JSON(200, res)
}

func adminActNewHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	//与adminActEditHandler代码类似
	form := new(FormDataAdminAct)
	err = c.ShouldBindJSON(form)
	if err != nil {
		Logger.Info.Println("[管理员][创建活动]参数绑定失败", err, auth)
		returnErrorJson(c, "参数无效(-1)")
		return
	}

	//校验时间格式是否有效
	bt, err := dateString2ts(form.BeginTime.D + " " + form.BeginTime.T)
	if err != nil {
		Logger.Info.Println("[管理员][创建活动]时间校验失败", err, auth)
		returnErrorJson(c, "开始日期或时间无效")
		return
	}
	et, err := dateString2ts(form.EndTime.D + " " + form.EndTime.T)
	if err != nil {
		Logger.Info.Println("[管理员][创建活动]时间校验失败", err, auth)
		returnErrorJson(c, "结束日期或时间无效")
		return
	}

	//校验图片地址
	picUrl := ""
	if form.Pic != "" {
		purl, err := url.Parse(form.Pic)
		if err != nil {
			Logger.Error.Println("[管理员][创建活动]图片地址校验", err, auth)
			returnErrorJson(c, "图片地址无效(-1)")
			return
		}
		if purl.Host != "i.loli.net" {
			Logger.Error.Println("[管理员][创建活动]图片地址校验", err, auth)
			returnErrorJson(c, "图片地址无效(-2)")
			return
		}
		picUrl = purl.Scheme + "://" + purl.Host + "/" + purl.Path
		Logger.Debug.Println("[图片地址]", picUrl)
	}

	//字段长度校验name40 ann50 ct20 url100
	if len(picUrl) > 100 {
		Logger.Info.Println("[管理员][创建活动]字段长度校验", err, auth)
		returnErrorJson(c, "图片地址无效(-3)")
		return
	}

	//更新数据库activity
	dbrt, err := db.Exec("INSERT INTO `activity` (`act_id`, `class_id`, `name`, `announcement`, `cheer_text`, `pic`, `begin_time`, `end_time`, `create_time`, `update_time`, `create_by`) VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);",
		auth.ClassId,
		form.Name,
		form.Announcement,
		form.CheerText,
		picUrl,
		strconv.FormatInt(bt, 10),
		strconv.FormatInt(et, 10),
		strconv.FormatInt(time.Now().Unix(), 10),
		strconv.FormatInt(time.Now().Unix(), 10),
		auth.UserID,
	)
	if err != nil {
		Logger.Error.Println("[管理员][创建活动]更新act数据库失败", err, auth)
		returnErrorJson(c, "更新失败，请联系管理员。(-1)")
		return
	}
	actID, _ := dbrt.LastInsertId()

	//更新班级表
	_, err = db.Exec("update `class` set `act_id`=? where `class_id`=?", actID, auth.ClassId)
	if err != nil {
		Logger.Error.Println("[管理员][创建活动]更新class数据库失败", err, auth)
		returnErrorJson(c, "更新失败，请联系管理员。(-2)")
		return
	}

	//更新缓存act+class
	act, err := cacheAct(int(actID))
	if err != nil {
		Logger.Error.Println("[管理员][创建活动]缓存活动更新失败", err, auth)
	}
	_, err = cacheClass(auth.ClassId)
	if err != nil {
		Logger.Error.Println("[管理员][创建活动]缓存班级更新失败", err, auth)
	}

	//群发消息(待修改)
	err = newActBulkSend(auth.ClassId, act)
	if err != nil {
		Logger.Debug.Println("[管理员][创建活动]群发时发生错误:", err, auth)
	}

	res := new(ResEmpty)
	res.Status = 0
	c.JSON(200, res)
}

//TODO 根据begin_time and end_time 判断active
func adminActEditHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	form := new(FormDataAdminAct)
	err = c.ShouldBindJSON(form)
	if err != nil {
		Logger.Info.Println("[管理员][编辑活动信息]参数绑定失败", err, auth)
		returnErrorJson(c, "参数无效(-1)")
		return
	}

	//校验时间格式是否有效
	bt, err := dateString2ts(form.BeginTime.D + " " + form.BeginTime.T)
	if err != nil {
		Logger.Info.Println("[管理员][编辑活动信息]时间校验失败", form.BeginTime.D+" "+form.BeginTime.T, err, auth)
		returnErrorJson(c, "开始日期或时间无效")
		return
	}
	et, err := dateString2ts(form.EndTime.D + " " + form.EndTime.T)
	if err != nil {
		Logger.Info.Println("[管理员][编辑活动信息]时间校验失败", form.EndTime.D+" "+form.EndTime.T, err, auth)
		returnErrorJson(c, "结束日期或时间无效")
		return
	}

	//校验图片地址
	picUrl := ""
	if form.Pic != "" {
		purl, err := url.Parse(form.Pic)
		if err != nil {
			Logger.Info.Println("[管理员][编辑活动信息]图片地址校验", err, auth)
			returnErrorJson(c, "图片地址无效(-1)")
			return
		}
		if purl.Host != "i.loli.net" {
			Logger.Info.Println("[管理员][编辑活动信息]图片地址校验", err, auth)
			returnErrorJson(c, "图片地址无效(-2)")
			return
		}
		picUrl = purl.Scheme + "://" + purl.Host + "/" + purl.Path
		Logger.Debug.Println("[图片地址]", picUrl)
	}

	//字段长度校验name40 ann50 ct20 url100
	if len(picUrl) > 100 {
		Logger.Info.Println("[管理员][编辑活动信息]字段长度校验", err, auth)
		returnErrorJson(c, "图片地址无效(-3)")
		return
	}

	//获取班级当前活动
	class, err := getClass(auth.ClassId)
	if err != nil {
		Logger.Info.Println("[管理员][编辑活动信息]获取班级失败", err, auth)
		returnErrorJson(c, "班级信息查询失败，请联系管理员。")
		return
	}
	act, err := getAct(class.ActID)
	if err != nil {
		Logger.Info.Println("[管理员][编辑活动信息]获取活动失败", err, auth)
		returnErrorJson(c, "活动信息查询失败，请联系管理员。")
		return
	}

	//更新数据库
	_, err = db.Exec("UPDATE `activity` SET `name` = ?,`announcement`=?,`pic`=?,`cheer_text`=?,`begin_time`=?,`end_time`=?,`update_time`=? WHERE `activity`.`act_id` = ?;",
		form.Name,
		form.Announcement,
		picUrl,
		form.CheerText,
		strconv.FormatInt(bt, 10),
		strconv.FormatInt(et, 10),
		strconv.FormatInt(time.Now().Unix(), 10),
		act.ActID,
	)
	if err != nil {
		Logger.Error.Println("[管理员][编辑活动信息]更新数据库失败", err, auth)
		returnErrorJson(c, "更新失败，请联系管理员。")
		return
	}

	//更新缓存
	_, err = cacheAct(act.ActID)
	if err != nil {
		Logger.Error.Println("[管理员][编辑活动信息]缓存更新失败", err, auth)
	}

	res := new(ResEmpty)
	res.Status = 0
	c.JSON(200, res)
}

func adminClassInfoHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	class, err := getClass(auth.ClassId)
	if err != nil {
		Logger.Error.Println("[管理员][获取班级信息]err", err)
		returnErrorJson(c, "查询失败(-1)")
		return
	}
	act, err := getAct(class.ActID)
	if err != nil {
		Logger.Error.Println("[管理员][获取活动信息]err", err)
		returnErrorJson(c, "查询失败(-2)")
		return
	}

	res := new(ResAdminClassInfo)
	res.Status = 0
	res.Data.ClassName = class.Name
	res.Data.ClassCode = class.ClassCode
	res.Data.Total = class.Total
	res.Data.ActId = class.ActID
	res.Data.ActName = act.Name

	c.JSON(200, res)

}

func adminClassEditHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	form := new(FormDataAdminClassEdit)
	err = c.ShouldBindJSON(form)
	if err != nil {
		Logger.Error.Println("[管理员][编辑班级信息]err", err)
		returnErrorJson(c, "参数无效(-1)")
		return
	}

	//校验班级代码
	if form.ClassCode == "new" {
		returnErrorJson(c, "班级代码不能为\"new\"")
		return
	}

	_, err = db.Exec("update `class` set `name`=?,`class_code`=? where `class_id`=?",
		form.ClassName, form.ClassCode, auth.ClassId)
	if err != nil {
		Logger.Error.Println("[管理员][编辑班级信息]更新数据库失败", err)
		returnErrorJson(c, "更新失败，可能是因为班级代码撞车了")
		return
	}

	//更新缓存
	_, err = cacheClass(auth.ClassId)
	if err != nil {
		Logger.Error.Println("[管理员][编辑班级信息]更新缓存失败", err)
	}
	res := new(ResEmpty)
	res.Status = 0
	c.JSON(200, res)
}
