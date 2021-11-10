package main

import (
	"github.com/gin-gonic/gin"
	"net/url"
	"signin/Logger"
	"strconv"
	"strings"
	"time"
)

type FormDataAdminActNew struct {
	Name         string `json:"name"`
	Announcement string `json:"announcement"`
	Pic          string `json:"pic"`
	CheerText    string `json:"cheer_text"`
	EndTime      struct {
		D string `json:"d"`
		T string `json:"t"`
	} `json:"end_time"`
}
type FormDataAdminActEdit struct {
	ActID        int    `json:"act_id"`
	Name         string `json:"name"`
	Active       bool    `json:"active"`
	Announcement string `json:"announcement"`
	Pic          string `json:"pic"`
	CheerText    string `json:"cheer_text"`
	EndTime      struct {
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

	actID, _ := strconv.Atoi(c.Query("act_id"))
	if actID == 0 {
		returnErrorJson(c, "参数无效")
		return
	}

	//获取活动
	act, err := getAct(actID)
	if err != nil {
		Logger.Error.Println("[管理员][获取活动信息]获取活动失败", err, auth)
		returnErrorJson(c, "系统异常(-2)")
		return
	}

	//数据处理
	res := new(ResAdminActInfo)
	res.Status = 0
	res.Data.ActID = act.ActID
	res.Data.Name = act.Name
	res.Data.Announcement = act.Announcement
	if act.Pic == "/static/image/default.jpg" {
		res.Data.Pic = ""
	}else{
		res.Data.Pic = act.Pic
	}
	res.Data.CheerText = act.CheerText
	if act.Active == 1{
		res.Data.Active = true
	}else{
		res.Data.Active = false
	}

	//处理时间
	et := strings.Split(ts2DateString(act.EndTime), " ")
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
	form := new(FormDataAdminActNew)
	err = c.ShouldBindJSON(form)
	if err != nil {
		Logger.Info.Println("[管理员][创建活动]参数绑定失败", err, auth)
		returnErrorJson(c, "参数无效(-1)")
		return
	}

	//校验时间格式是否有效
	et, err := dateString2ts(form.EndTime.D + " " + form.EndTime.T)
	if err != nil || et <= time.Now().Unix() {
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
		Logger.Info.Println("[管理员][创建活动][图片地址]", picUrl)
	}

	//字段长度校验name40 ann50 ct20 url100
	if len(picUrl) > 100 {
		Logger.Info.Println("[管理员][创建活动]字段长度校验", err, auth)
		returnErrorJson(c, "图片地址无效(-3)")
		return
	}

	//更新数据库activity
	dbrt, err := db.Exec("INSERT INTO `activity` (`act_id`, `class_id`,`active`, `name`, `announcement`, `cheer_text`, `pic`, `begin_time`, `end_time`, `create_time`, `update_time`, `create_by`) VALUES (NULL, ?, ?,?, ?, ?, ?, ?, ?, ?, ?, ?);",
		auth.ClassId,
		1, //active
		form.Name,
		form.Announcement,
		form.CheerText,
		picUrl,
		strconv.FormatInt(time.Now().Unix(), 10),
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

	//更新缓存act+class+actIDs
	act, err := cacheAct(int(actID))
	if err != nil {
		Logger.Error.Println("[管理员][创建活动]缓存活动失败", err, auth)
	}
	_, err = cacheClass(auth.ClassId)
	if err != nil {
		Logger.Error.Println("[管理员][创建活动]缓存班级失败", err, auth)
	}
	_, err = cacheIDs(auth.ClassId)
	if err != nil {
		Logger.Error.Println("[管理员][创建活动]缓存活动id失败", err, auth)
	}

	//群发消息
	err = newActBulkSend(auth.ClassId, act)
	if err != nil {
		Logger.Info.Println("[管理员][创建活动]群发时发生错误:", err, auth)
	}

	res := new(ResEmpty)
	res.Status = 0
	c.JSON(200, res)
}

func adminActEditHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	form := new(FormDataAdminActEdit)
	err = c.ShouldBindJSON(form)
	if err != nil {
		Logger.Info.Println("[管理员][编辑活动信息]参数绑定失败", err, auth)
		returnErrorJson(c, "参数无效(-1)")
		return
	}

	//判断活动是否是自己班的
	act, err := getAct(form.ActID)
	if err != nil {
		returnErrorJson(c, "系统异常")
		return
	}
	if act.ClassID != auth.ClassId {
		Logger.Info.Println("[管理员][编辑活动信息]无权限修改", auth, form)
		returnErrorJson(c, "您没有权限修改此活动")
		return
	}

	var et int64
	//校验时间格式是否有效
	if form.Active == false{
		et = time.Now().Unix()
	}else{
		et, err = dateString2ts(form.EndTime.D + " " + form.EndTime.T)
		if err != nil || et < time.Now().Unix() {
			Logger.Info.Println("[管理员][编辑活动信息]时间校验失败", form.EndTime.D+" "+form.EndTime.T, err, auth)
			returnErrorJson(c, "结束日期或时间无效")
			return
		}
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

	active := 0
	if form.Active == true{
		active = 1
	}

	//更新数据库
	_, err = db.Exec("UPDATE `activity` SET `name` = ?,`active`=?,`announcement`=?,`pic`=?,`cheer_text`=?,`end_time`=?,`update_time`=? WHERE `activity`.`act_id` = ?;",
		form.Name,
		active,
		form.Announcement,
		picUrl,
		form.CheerText,
		strconv.FormatInt(et, 10),
		strconv.FormatInt(time.Now().Unix(), 10),
		form.ActID,
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
	_, err = cacheIDs(auth.ClassId)
	if err != nil {
		Logger.Error.Println("[管理员][编辑活动信息]cacheIDs失败", err, auth)
	}

	res := new(ResEmpty)
	res.Status = 0
	c.JSON(200, res)
}

func adminActStatisticHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	actId, err := strconv.Atoi(c.Query("act_id"))
	if actId == 0 || err != nil {
		Logger.Info.Println("[管理员活动数据]解析参数失败", err, auth, c.Query("act_id"))
		returnErrorJson(c, "参数无效")
		return
	}

	//判断是不是自己班的
	act, err := getAct(actId)
	if err != nil || act.ClassID != auth.ClassId {
		returnErrorJson(c, "您没有权限查询此数据")
		return
	}

	//获取统计数据
	sts, err := getActStatistics(actId)
	if err != nil {
		returnErrorJson(c, "系统异常(-2)")
		return
	}

	res := new(ResAdminActStatistic)
	res.Status = 0
	res.Data.Done = sts.Done
	res.Data.Total = sts.Total
	res.Data.FinishedList = make([]*AdminActStatisticItem, 0)
	res.Data.UnfinishedList = make([]*AdminActStatisticItem, 0)

	//装载数据
	for i := range sts.FinishedList {
		res.Data.FinishedList = append(res.Data.FinishedList, &AdminActStatisticItem{
			ID:       sts.FinishedList[i].Id,
			UserId:   sts.FinishedList[i].UserID,
			UserName: sts.FinishedList[i].Name,
			DateTime: sts.FinishedList[i].DateTime,
		})
	}
	for i := range sts.UnfinishedList {
		res.Data.UnfinishedList = append(res.Data.UnfinishedList, &AdminActStatisticItem{
			ID:       sts.UnfinishedList[i].Id,
			UserId:   sts.UnfinishedList[i].UserID,
			UserName: sts.UnfinishedList[i].Name,
			DateTime: sts.UnfinishedList[i].DateTime,
		})
	}

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

	res := new(ResAdminClassInfo)
	res.Status = 0
	res.Data.ClassName = class.Name
	res.Data.ClassCode = class.ClassCode
	res.Data.Total = class.Total

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

func adminActListHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	res := new(ResAdminActList)
	res.Status = 0
	res.Data.ActiveList = make([]*adminActListItem, 0)
	res.Data.HistoryList = make([]*adminActListItem, 0)

	//检查活动状态
	_, err = getActIDs(auth.ClassId)
	if err != nil {
		returnErrorJson(c, "系统异常")
		return
	}

	acts := make([]dbAct, 0)
	err = db.Select(&acts, "select * from `activity` where `class_id`=?", auth.ClassId)
	if err != nil {
		Logger.Error.Println("[管理员活动列表]查询数据库异常", err)
		returnErrorJson(c, "查询失败")
		return
	}
	if len(acts) == 0 {
		returnErrorJson(c, "当前无活动")
		return
	}

	activeCnt := 0
	HistoryCnt := 0
	for i := range acts {
		item := new(adminActListItem)
		item.Name = acts[i].Name
		item.ActID = acts[i].ActID
		item.BeginTime = ts2DateString(acts[i].BeginTime)
		item.EndTime = ts2DateString(acts[i].EndTime)
		item.CreateBy = queryUserName(acts[i].CreateBy)
		if acts[i].Active == 1 {
			activeCnt++
			item.Id = activeCnt
			res.Data.ActiveList = append(res.Data.ActiveList, item)
		} else {
			HistoryCnt++
			item.Id = HistoryCnt
			res.Data.HistoryList = append(res.Data.HistoryList, item)
		}
	}

	res.Data.ActiveNum = activeCnt
	c.JSON(200, res)
	return
}

func AdminCsrfTokenHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	_,err = csrfMake(auth,c)
	if err != nil {
		Logger.Info.Println("[管理员csrf]发生错误",err)
		returnErrorJson(c,"返回csrfToken失败")
		return
	}

	c.Status(200)
}
