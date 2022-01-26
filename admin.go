package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"signin/Logger"
	"strconv"
	"strings"
	"time"
)

type FormDataAdminAct struct {
	ActID        int    `json:"act_id"`
	Name         string `json:"name"`
	Active       bool   `json:"active"`
	Announcement string `json:"announcement"`
	Pic          string `json:"pic"`
	DailyNotify  bool   `json:"daily_notify"`
	CheerText    string `json:"cheer_text"`
	EndTime      struct {
		D string `json:"d"`
		T string `json:"t"`
	} `json:"end_time"`
	Upload struct {
		Enabled bool     `json:"enabled"`
		Type    []string `json:"type"`
		MaxSize int      `json:"max_size"`
		Rename  bool     `json:"rename"`
	} `json:"upload"`
}

type FormDataAdminClassEdit struct {
	ClassName string `json:"class_name"`
	ClassCode string `json:"class_code"`
}

type FormDataAdminUserDel struct {
	UserID int    `json:"user_id"`
	Sign   string `json:"sign"`
}

type FormDataAdminUserSetAdmin struct {
	UserID int    `json:"user_id"`
	SetTo  int    `json:"set_to"`
	Sign   string `json:"sign"`
}

type FormDataAdminActExport struct {
	ActID int `json:"act_id" binding:"required"`
}

type FormDataAdminActViewFile struct {
	UserID int `json:"user_id" binding:"required"`
	ActID  int `json:"act_id" binding:"required"`
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
	if act.DailyNotiEnabled == 1 {
		res.Data.DailyNotify = true
	} else {
		res.Data.DailyNotify = false
	}
	if act.Pic == "/static/image/default.jpg" {
		res.Data.Pic = ""
	} else {
		res.Data.Pic = act.Pic
	}
	res.Data.CheerText = act.CheerText
	if act.Active == 1 {
		res.Data.Active = true
	} else {
		res.Data.Active = false
	}

	//处理时间
	et := strings.Split(ts2DateString(act.EndTime), " ")
	res.Data.EndTime.D = et[0]
	res.Data.EndTime.T = et[1][:5]

	//处理上传规则
	if act.Type == ACT_TYPE_UPLOAD {
		res.Data.Upload.Enabled = true
		opts := new(FileOptions)
		res.Data.Upload.Type = make([]string, 0)
		err = json.Unmarshal([]byte(act.FileOpts), opts)
		if err != nil {
			Logger.Error.Println("[管理员][获取活动信息]解析文件上传要求失败", err, auth)
			returnErrorJson(c, "系统异常(-3)")
			return
		}

		//解析文件类型
		for i := range opts.AllowContentType {
			if strings.Contains(opts.AllowContentType[i], "png") {
				res.Data.Upload.Type = append(res.Data.Upload.Type, "图片")
				continue
			}
			if strings.Contains(opts.AllowContentType[i], "zip") {
				res.Data.Upload.Type = append(res.Data.Upload.Type, "压缩包")
				continue
			}
			if strings.Contains(opts.AllowContentType[i], "msword") {
				res.Data.Upload.Type = append(res.Data.Upload.Type, "word")
				continue
			}
			if strings.Contains(opts.AllowContentType[i], "powerpoint") {
				res.Data.Upload.Type = append(res.Data.Upload.Type, "ppt")
				continue
			}
			if strings.Contains(opts.AllowContentType[i], "excel") {
				res.Data.Upload.Type = append(res.Data.Upload.Type, "excel")
				continue
			}
			if strings.Contains(opts.AllowContentType[i], "pdf") {
				res.Data.Upload.Type = append(res.Data.Upload.Type, "pdf")
				continue
			}
		}
		res.Data.Upload.Rename = opts.Rename
		res.Data.Upload.MaxSize = opts.MaxSize
	} else {
		res.Data.Upload.Enabled = false
	}

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

		tmpPicToken, err := Cipher.Encrypt([]byte("redis_tempImage/" + auth.ID))
		if err != nil {
			Logger.Error.Println("[管理员][创建活动]获取加密字符串失败", err, auth)
			returnErrorJson(c, "系统异常")
			return
		}
		if strings.Contains(form.Pic, config.General.BaseUrl) == true && strings.Contains(form.Pic, tmpPicToken) { //随机图片
			if rdb.Exists(ctx, fmt.Sprintf("SIGNIN_APP:TempFile:file_%s", auth.ID)).Val() != int64(1) {
				returnErrorJson(c, "随机图片临时地址失效")
				return
			}
			fileData, err := base64.StdEncoding.DecodeString(rdb.Get(ctx, fmt.Sprintf("SIGNIN_APP:TempFile:file_%s", auth.ID)).Val())
			if err != nil {
				returnErrorJson(c, "图片解码失败，请联系管理员")
				Logger.Error.Println("[管理员][创建活动]随机图片解码失败", err, auth)
				return
			}
			//暂存图片
			fileName := fmt.Sprintf("RandomPic_USER%d_%s_%d.jpg", auth.UserID, auth.ID, time.Now().UnixNano())
			err = ioutil.WriteFile("./storage/upload/"+fileName, fileData, 0666)
			if err != nil {
				returnErrorJson(c, "暂存图片文件失败，请联系管理员")
				Logger.Error.Println("[管理员][创建活动]暂存图片文件失败", err, auth)
				return
			}
			//写入file表
			fileDB := new(dbFile)
			fileDB.Status = FILE_STATUS_LOCAL
			fileDB.FileName = fileName
			fileDB.UserID = auth.UserID
			fileDB.ActID = 0
			fileDB.ContentType = "image/jpeg"
			fileDB.Local = "./storage/upload/" + fileName
			fileDB.ExpTime = strconv.FormatInt(time.Now().Add(30*24*time.Hour).Unix(), 10)
			fileDB.UploadTime = strconv.FormatInt(time.Now().Unix(), 10)
			_, err = db.Exec("INSERT INTO `file` (`file_id`, `status`, `user_id`, `act_id`, `file_name`, `content_type`, `local`, `remote`, `exp_time`, `upload_time`) VALUES (NULL,?, ?, ?, ?, ?, ?, ?, ?, ?);",
				fileDB.Status,
				fileDB.UserID,
				fileDB.ActID,
				fileDB.FileName,
				fileDB.ContentType,
				fileDB.Local,
				fileDB.Remote,
				fileDB.ExpTime,
				fileDB.UploadTime,
			)
			if err != nil {
				Logger.Error.Println("[管理员][创建活动]文件登记失败:", err)
				returnErrorJson(c, "文件登记失败，请联系管理员:"+fileName)
				return
			}
			//生成图片访问地址
			fileToken, err := Cipher.Encrypt([]byte(fmt.Sprintf("local_file#%s#%s#%d", fileDB.ContentType, fileDB.Local, time.Now().UnixNano())))
			if err != nil {
				Logger.Error.Println("[管理员][创建活动]生成图片访问地址加密失败:", err)
				returnErrorJson(c, "文件地址生成失败，请联系管理员:"+fileName)
				return
			}
			picUrl = config.General.BaseUrl + "/file/" + fileToken + "." + MD5(fileToken+config.General.AESKey)
		} else if purl.Host == "i.loli.net" { //图床
			picUrl = purl.Scheme + "://" + purl.Host + "/" + purl.Path
		} else {
			Logger.Error.Println("[管理员][创建活动]图片地址校验", err, auth)
			returnErrorJson(c, "图片地址无效(-2)")
			return
		}

		Logger.Info.Println("[管理员][创建活动][图片地址]", picUrl)
	}

	//字段长度校验name40 ann50 ct20 url500
	if len(picUrl) > 500 {
		Logger.Info.Println("[管理员][编辑活动信息]字段长度校验", err, auth)
		returnErrorJson(c, "图片地址无效(-3)")
		return
	}

	//文件上传
	actType := ACT_TYPE_NORMAL
	fileOpts := ""
	if form.Upload.Enabled == true {
		actType = ACT_TYPE_UPLOAD
		opts := new(FileOptions)
		opts.Rename = form.Upload.Rename
		if form.Upload.MaxSize < 1 || form.Upload.MaxSize > 100 {
			returnErrorJson(c, "文件大小无效")
			return
		}
		if len(form.Upload.Type) == 0 {
			returnErrorJson(c, "请至少选择一种文件类型")
			return
		}
		opts.MaxSize = form.Upload.MaxSize
		opts.AllowContentType = make([]string, 0)
		for i := range form.Upload.Type {
			if form.Upload.Type[i] == "image" {
				opts.AllowContentType = append(opts.AllowContentType, "image/png")
				opts.AllowContentType = append(opts.AllowContentType, "image/jpeg")
				continue
			}
			if form.Upload.Type[i] == "archive" {
				opts.AllowContentType = append(opts.AllowContentType, "application/zip")
				opts.AllowContentType = append(opts.AllowContentType, "application/x-rar-compressed")
				opts.AllowContentType = append(opts.AllowContentType, "application/x-zip-compressed")
				continue
			}
			if form.Upload.Type[i] == "word" {
				opts.AllowContentType = append(opts.AllowContentType, "application/msword")
				opts.AllowContentType = append(opts.AllowContentType, "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
				continue
			}
			if form.Upload.Type[i] == "ppt" {
				opts.AllowContentType = append(opts.AllowContentType, "application/vnd.ms-powerpoint")
				opts.AllowContentType = append(opts.AllowContentType, "application/vnd.openxmlformats-officedocument.presentationml.presentation")
				continue
			}
			if form.Upload.Type[i] == "excel" {
				opts.AllowContentType = append(opts.AllowContentType, "application/vnd.ms-excel")
				opts.AllowContentType = append(opts.AllowContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
				continue
			}
			if form.Upload.Type[i] == "pdf" {
				opts.AllowContentType = append(opts.AllowContentType, "application/pdf")
				continue
			}
			returnErrorJson(c, "文件类型无效")
			return
		}

		tmp, err := json.Marshal(opts)
		if err != nil {
			Logger.Error.Println("[管理员][创建活动]FileOptions格式化失败:", err)
			returnErrorJson(c, "FileOptions格式化失败")
			return
		}
		fileOpts = string(tmp)
	}

	//每日提醒开关
	dailyNoti := 1
	if form.DailyNotify == false {
		dailyNoti = 0
	}
	//更新数据库activity
	dbrt, err := db.Exec("INSERT INTO `activity` (`act_id`, `class_id`,`active`,`type`, `name`, `announcement`, `cheer_text`, `pic`,`daily_noti_enabled`, `begin_time`, `end_time`, `create_time`, `update_time`, `create_by`,`file_opts`) VALUES (NULL, ?, ?,?, ?,?, ?, ?, ?,?, ?, ?, ?, ?,?);",
		auth.ClassId,
		1, //active
		actType,
		form.Name,
		form.Announcement,
		form.CheerText,
		picUrl,
		dailyNoti, //每日提醒
		strconv.FormatInt(time.Now().Unix(), 10),
		strconv.FormatInt(et, 10),
		strconv.FormatInt(time.Now().Unix(), 10),
		strconv.FormatInt(time.Now().Unix(), 10),
		auth.UserID,
		fileOpts,
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

	form := new(FormDataAdminAct)
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

	var formEtTS int64
	//校验时间格式是否有效
	endTime := ""
	formEtTS, err = dateString2ts(form.EndTime.D + " " + form.EndTime.T)
	dbEtString := ts2DateString(act.EndTime)
	dbEtString = dbEtString[:len(dbEtString)-3]
	if err != nil {
		Logger.Info.Println("[管理员][编辑活动信息]时间校验失败", form.EndTime.D+" "+form.EndTime.T, err, auth)
		returnErrorJson(c, "结束日期或时间无效")
		return
	}
	if formEtTS < time.Now().Unix() {
		if form.Active == true {
			Logger.Info.Printf("[管理员][编辑活动信息]时间校验失败:%s != %s,%v\n", dbEtString, form.EndTime.D+" "+form.EndTime.T, auth)
			returnErrorJson(c, "结束时间不能早于当前时间")
			return
		} else {
			if dbEtString != form.EndTime.D+" "+form.EndTime.T {
				Logger.Info.Printf("[管理员][编辑活动信息]时间校验失败:%s != %s,%v\n", dbEtString, form.EndTime.D+" "+form.EndTime.T, auth)
				returnErrorJson(c, "活动已结束，无法更改结束时间")
				return
			}
		}
	}

	if form.Active == true {
		endTime = strconv.FormatInt(formEtTS, 10)
	} else {
		if act.Active == 0 {
			//不作更改
			if dbEtString != form.EndTime.D+" "+form.EndTime.T {
				returnErrorJson(c, "活动已结束，无法修改结束时间。")
				return
			} else {
				endTime = act.EndTime
			}
		} else {
			endTime = strconv.FormatInt(time.Now().Unix(), 10)
		}
	}

	//校验图片地址
	picUrl := ""
	if form.Pic != "" && form.Pic != act.Pic {
		purl, err := url.Parse(form.Pic)
		if err != nil {
			Logger.Error.Println("[管理员][创建活动]图片地址校验", err, auth)
			returnErrorJson(c, "图片地址无效(-1)")
			return
		}
		tmpPicToken, err := Cipher.Encrypt([]byte("redis_tempImage/" + auth.ID))
		if err != nil {
			Logger.Error.Println("[管理员][创建活动]获取加密字符串失败", err, auth)
			returnErrorJson(c, "系统异常")
			return
		}

		if strings.Contains(form.Pic, config.General.BaseUrl) == true && strings.Contains(form.Pic, tmpPicToken) { //随机图片预览地址
			if rdb.Exists(ctx, fmt.Sprintf("SIGNIN_APP:TempFile:file_%s", auth.ID)).Val() != int64(1) {
				returnErrorJson(c, "随机图片临时地址失效")
				return
			}
			fileData, err := base64.StdEncoding.DecodeString(rdb.Get(ctx, fmt.Sprintf("SIGNIN_APP:TempFile:file_%s", auth.ID)).Val())
			if err != nil {
				returnErrorJson(c, "图片解码失败，请联系管理员")
				Logger.Error.Println("[管理员][创建活动]随机图片解码失败", err, auth)
				return
			}
			//暂存图片
			fileName := fmt.Sprintf("RandomPic_USER%d_%s_%d.jpg", auth.UserID, auth.ID, time.Now().UnixNano())
			err = ioutil.WriteFile("./storage/upload/"+fileName, fileData, 0666)
			if err != nil {
				returnErrorJson(c, "暂存图片文件失败，请联系管理员")
				Logger.Error.Println("[管理员][创建活动]暂存图片文件失败", err, auth)
				return
			}
			//写入file表
			fileDB := new(dbFile)
			fileDB.Status = FILE_STATUS_LOCAL
			fileDB.FileName = fileName
			fileDB.UserID = auth.UserID
			fileDB.ActID = 0
			fileDB.ContentType = "image/jpeg"
			fileDB.Local = "./storage/upload/" + fileName
			fileDB.ExpTime = strconv.FormatInt(time.Now().Add(90*24*time.Hour).Unix(), 10)
			fileDB.UploadTime = strconv.FormatInt(time.Now().Unix(), 10)
			_, err = db.Exec("INSERT INTO `file` (`file_id`, `status`, `user_id`, `act_id`, `file_name`, `content_type`, `local`, `remote`, `exp_time`, `upload_time`) VALUES (NULL,?, ?, ?, ?, ?, ?, ?, ?, ?);",
				fileDB.Status,
				fileDB.UserID,
				fileDB.ActID,
				fileDB.FileName,
				fileDB.ContentType,
				fileDB.Local,
				fileDB.Remote,
				fileDB.ExpTime,
				fileDB.UploadTime,
			)
			if err != nil {
				Logger.Error.Println("[管理员][创建活动]文件登记失败:", err)
				returnErrorJson(c, "文件登记失败，请联系管理员:"+fileName)
				return
			}
			//生成图片访问地址
			fileToken, err := Cipher.Encrypt([]byte(fmt.Sprintf("local_file#%s#%s#%d", fileDB.ContentType, fileDB.Local, time.Now().UnixNano())))
			if err != nil {
				Logger.Error.Println("[管理员][创建活动]生成图片访问地址加密失败:", err)
				returnErrorJson(c, "文件地址生成失败，请联系管理员:"+fileName)
				return
			}
			picUrl = config.General.BaseUrl + "/file/" + fileToken + "." + MD5(fileToken+config.General.AESKey)
			rdb.Del(ctx, fmt.Sprintf("SIGNIN_APP:TempFile:file_%s", auth.ID))
		} else if purl.Host == "i.loli.net" { //图床
			picUrl = purl.Scheme + "://" + purl.Host + "/" + purl.Path
		} else {
			Logger.Error.Println("[管理员][创建活动]图片地址校验", err, auth)
			returnErrorJson(c, "图片地址无效(-2)")
			return
		}
		Logger.Info.Println("[管理员][创建活动][图片地址]", picUrl)
	}

	//字段长度校验name40 ann50 ct20 url500
	if len(picUrl) > 500 {
		Logger.Info.Println("[管理员][编辑活动信息]字段长度校验", err, auth)
		returnErrorJson(c, "图片地址无效(-3)")
		return
	}

	active := 0
	if form.Active == true {
		active = 1
	}

	//每日提醒开关
	dailyNoti := 1
	if form.DailyNotify == false {
		dailyNoti = 0
	}

	//更新数据库
	_, err = db.Exec("UPDATE `activity` SET `name` = ?,`active`=?,`announcement`=?,`pic`=?,`cheer_text`=?,`end_time`=?,`update_time`=?,`daily_noti_enabled`=? WHERE `activity`.`act_id` = ?;",
		form.Name,
		active,
		form.Announcement,
		picUrl,
		form.CheerText,
		endTime,
		strconv.FormatInt(time.Now().Unix(), 10),
		dailyNoti,
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
	err = db.Select(&acts, "select * from `activity` where `class_id`=? order by `act_id` desc;", auth.ClassId)
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
		item.Type = acts[i].Type
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

func adminUserListHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	users := make([]dbUser, 0)
	err = db.Select(&users, "select * from `user` where `class`=? order by `is_admin` desc;", auth.ClassId)
	if err != nil {
		Logger.Error.Println("[管理员用户列表]查询db失败:", err)
		returnErrorJson(c, "系统异常")
		return
	}

	res := new(ResAdminUserList)
	res.Status = 0
	res.Data.Count = len(users)
	res.Data.Data = make([]AdminUserListItem, 0)

	for i := range users {
		res.Data.Data = append(res.Data.Data, AdminUserListItem{
			ID:     i + 1,
			UserID: users[i].UserID,
			Name:   users[i].Name,
			Admin:  users[i].IsAdmin,
			Sign:   Cipher.Sha256Hex([]byte(fmt.Sprintf("%d%d%s", users[i].UserID, auth.ClassId, auth.ID))),
		})
	}

	c.JSON(200, res)

}

func adminUserDelHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	form := new(FormDataAdminUserDel)
	err = c.ShouldBindJSON(form)
	if err != nil {
		Logger.Info.Println("[管理员删除用户]参数绑定失败:", err)
		returnErrorJson(c, "参数无效")
		return
	}

	if form.Sign != Cipher.Sha256Hex([]byte(fmt.Sprintf("%d%d%s", form.UserID, auth.ClassId, auth.ID))) {
		Logger.Info.Println("[管理员删除用户]数据签名无效:", form)
		returnErrorJson(c, "数据签名无效")
		return
	}

	res := new(ResEmpty)
	res.Status = 0

	//查询用户信息
	user := new(dbUser)
	err = db.Get(user, "select * from `user` where `user_id`=?", form.UserID)
	if err != nil {
		Logger.Info.Println("[管理员删除用户]查询用户信息失败:", err)
		returnErrorJson(c, "查询用户信息失败")
		return
	}

	//获取记录，删除文件
	files := make([]dbFile, 0)
	err = db.Select(&files, "select * from `file` where `user_id`=? and `status`=1", auth.UserID)
	if err != nil {
		Logger.Info.Println("[管理员删除用户]获取文件上传记录失败:", err)
		returnErrorJson(c, "获取文件上传记录失败")
		return
	}
	for i := range files {
		err := cosFileDel(files[i].Remote)
		if err != nil {
			Logger.Error.Println("[管理员删除用户]远端文件删除失败:", files[i].Remote, err)
		}
	}
	_, err = db.Exec("update `file` set `status`=-1 where `user_id`=?", auth.UserID)
	if err != nil {
		Logger.Error.Println("[管理员删除用户]更新db.file失败:", err)
	}

	_, err = db.Exec("update `user` set `class`=0 where `user_id`=? and `class` = ?", form.UserID, auth.ClassId)
	if err != nil {
		Logger.Info.Println("[管理员删除用户]更新数据库失败:", err)
		returnErrorJson(c, "更新数据库失败")
		return
	}

	result, err := db.Exec("delete from `signin_log`  where `user_id`=?", form.UserID)
	if err != nil {
		Logger.Info.Println("[管理员删除用户]更新数据库失败:", err)
		returnErrorJson(c, "系统异常")
		return
	}
	num, _ := result.RowsAffected()
	res.Msg = fmt.Sprintf("已删除%d条签到记录", num)
	Logger.Info.Println("[管理员删除用户]:", user, "已被踢出班级，操作者：", auth, "删除记录条数:", num)

	killJwtByUID(form.UserID)

	c.JSON(200, res)
}

func adminUserSetAdminHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	form := new(FormDataAdminUserSetAdmin)
	err = c.ShouldBindJSON(form)
	if err != nil || (form.SetTo != 0 && form.SetTo != 1) {
		Logger.Info.Println("[管理员设置管理员]参数绑定失败:", err)
		returnErrorJson(c, "参数无效")
		return
	}

	if form.Sign != Cipher.Sha256Hex([]byte(fmt.Sprintf("%d%d%s", form.UserID, auth.ClassId, auth.ID))) {
		Logger.Info.Println("[管理员设置管理员]数据签名无效:", form)
		returnErrorJson(c, "数据签名无效")
		return
	}

	res := new(ResEmpty)
	res.Status = 0

	//查询用户信息
	user := new(dbUser)
	err = db.Get(user, "select * from `user` where `user_id`=?", form.UserID)
	if err != nil {
		Logger.Info.Println("[管理员设置管理员]查询用户信息失败:", err)
		returnErrorJson(c, "查询用户信息失败")
		return
	}
	if user.IsAdmin == form.SetTo {
		returnErrorJson(c, "您似乎未作任何修改")
		return
	}

	_, err = db.Exec("update `user` set `is_admin`=? where `user_id`=? and `class` = ?", form.SetTo, form.UserID, auth.ClassId)
	if err != nil {
		Logger.Info.Println("[管理员设置管理员]更新数据库失败:", err)
		returnErrorJson(c, "更新数据库失败")
		return
	}

	killJwtByUID(form.UserID)

	c.JSON(200, res)
}

func AdminActExportHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	form := new(FormDataAdminActExport)
	err = c.ShouldBindJSON(form)
	if err != nil {
		Logger.Info.Println("[管理员导出数据]参数绑定失败：", err)
		returnErrorJson(c, "参数无效(-1)")
		return
	}

	act, err := getAct(form.ActID)
	if err != nil {
		Logger.Error.Println("[管理员导出数据]获取活动失败：", err)
		returnErrorJson(c, "获取活动失败")
		return
	}

	if act.ClassID != auth.ClassId {
		Logger.Error.Println("[管理员导出数据]班级不匹配：", auth, form)
		returnErrorJson(c, "参数无效(-2)")
		return
	}

	if act.Type != ACT_TYPE_UPLOAD {
		returnErrorJson(c, "当前活动未开启文件上传")
		return
	}

	files := make([]dbFile, 0)
	err = db.Select(&files, "select * from `file` where `act_id`=? and `status` = 1", act.ActID)
	if err != nil {
		Logger.Error.Println("[管理员导出数据]查询数据库失败：", err)
		returnErrorJson(c, "查询数据库失败")
		return
	}

	if len(files) == 0 {
		returnErrorJson(c, "文件数为0")
		return
	}

	//创建目录
	dirName := act.Name + "_批量导出_" + MD5_short(fmt.Sprintf("%d%d", act.ActID, time.Now().UnixNano()))
	err = os.Mkdir("./storage/temp/"+dirName, os.ModePerm)
	if err != nil {
		Logger.Error.Println("[管理员导出数据]创建目录失败：", err)
		returnErrorJson(c, "系统异常(-1)")
		return
	}

	errCnt := 0
	for i := range files {
		fileName := files[i].FileName + fileExt[files[i].ContentType]
		if e, _ := FsExists(fmt.Sprintf("./storage/temp/%s/%s", dirName, fileName)); e != false {
			fileName = fmt.Sprintf("%s_%s%s", files[i].FileName, MD5_short(fmt.Sprintf("%d", time.Now().UnixNano())), fileExt[files[i].ContentType])
		}
		err = cosDownload(files[i].Remote, fmt.Sprintf("./storage/temp/%s/%s", dirName, fileName))
		if err != nil {
			Logger.Error.Printf("[管理员导出数据]下载文件失败：%s --> %s : err:%s\n", files[i].Remote, fileName, err)
			errCnt++
		}
	}

	f, err := os.Open("./storage/temp/" + dirName)
	if err != nil {
		Logger.Error.Println("[管理员导出数据]打开目录失败：", err)
		returnErrorJson(c, "系统异常(-2)")
		return
	}
	var fs = []*os.File{f}
	err = Compress(fs, "./storage/export/"+dirName+".zip")
	if err != nil {
		Logger.Error.Println("[管理员导出数据]压缩失败：", err)
		returnErrorJson(c, "系统异常(-3)")
		return
	}

	cleanTime := 5
	if config.General.Production == false {
		cleanTime = 1
	}
	go trashCleaner("./storage/temp/"+dirName, 0)
	go trashCleaner("./storage/export/"+dirName+".zip", cleanTime)

	res := new(ResAdminActExport)
	res.Status = 0
	res.Data.DownloadUrl = config.General.BaseUrl + "/export/" + dirName + ".zip"
	Logger.Info.Println("[管理员导出数据]操作者：", auth)
	c.JSON(200, res)
}

func AdminActViewFileHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	form := new(FormDataAdminActViewFile)
	err = c.ShouldBindJSON(form)
	if err != nil {
		returnErrorJson(c, "参数无效(-1)")
		return
	}

	//获取用户
	user := new(dbUser)
	err = db.Get(user, "select * from `user` where `user_id`=?", form.UserID)
	if err != nil {
		Logger.Error.Println("[管理员查看用户文件]查询用户信息失败：", err)
		returnErrorJson(c, "系统异常(-1)")
		return
	}
	if user.UserID == 0 {
		returnErrorJson(c, "参数无效(-2)")
		return
	}

	//获取活动
	act, err := getAct(form.ActID)
	if err != nil {
		Logger.Error.Println("[管理员查看用户文件]查询活动信息失败：", err)
		returnErrorJson(c, "参数无效(-3)")
		return
	}

	//判断是否有权限
	if user.Class != auth.ClassId || act.ClassID != auth.ClassId {
		returnErrorJson(c, "无权限查看")
		return
	}

	//判断是否满足条件
	if act.Type != ACT_TYPE_UPLOAD {
		returnErrorJson(c, "无权限查看")
		return
	}
	//查询记录
	logItem := new(dbLog)
	err = db.Get(logItem, "select * from `signin_log` where `user_id`=? and `act_id`=?", form.UserID, form.ActID)
	if err != nil || logItem.FileID == 0 {
		Logger.Error.Println("[管理员查看用户文件]查询签到记录失败：", err)
		returnErrorJson(c, "未查询到此用户的参与记录")
		return
	}
	//查询文件
	fileItem := new(dbFile)
	err = db.Get(fileItem, "select * from `file` where `file_id`=?", logItem.FileID)
	if err != nil {
		Logger.Error.Println("[管理员查看用户文件]查询文件信息失败：", err)
		returnErrorJson(c, "查询文件信息失败")
		return
	}

	res := new(ResAdminActViewFile)
	res.Status = 0
	if fileItem.Status == FILE_STATUS_DELETED {
		returnErrorJson(c, "该文件已过期")
		return
	}
	res.Data.Type = "other"
	if strings.Contains(fileItem.ContentType, "image") == true {
		res.Data.Type = "image"
		res.Data.ImgUrl, err = cosGetUrl(fileItem.Remote, 5*time.Minute)
	} else {
		res.Data.DownloadUrl, err = cosGetUrl(fileItem.Remote, 5*time.Minute)
	}
	if err != nil {
		Logger.Error.Println("[管理员查看用户文件]签发url失败：", err)
		returnErrorJson(c, "文件签名失败")
		return
	}

	c.JSON(200, res)
}

func AdminActRandomPicHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	req, err := http.NewRequest("GET", "https://pximg.rainchan.win/img", nil)
	if err != nil {
		returnErrorJson(c, "接口异常")
		Logger.Error.Println("[管理员获取随机图片] 请求接口 - 新建请求异常：", err)
		return
	}
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Ch-Ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"96\", \"Google Chrome\";v=\"96\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"macOS\"")
	req.Header.Set("Dnt", "1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,ja;q=0.7,zh-TW;q=0.6")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
		returnErrorJson(c, "接口异常")
		Logger.Error.Println("[管理员获取随机图片] 请求接口 - 异常：", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		returnErrorJson(c, "接口异常")
		Logger.Error.Println("[管理员获取随机图片] 请求接口 - 异常：", err, resp.Header)
		return
	}

	imageData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		returnErrorJson(c, "接口异常")
		Logger.Error.Println("[管理员获取随机图片] 请求接口 - 读取body异常：", err)
		return
	}

	rdb.Set(ctx, fmt.Sprintf("SIGNIN_APP:TempFile:file_%s", auth.ID), base64.StdEncoding.EncodeToString(imageData), 10*time.Minute)

	originUrl := "redis_tempImage/" + auth.ID

	//获取代理地址
	fileToken, err := Cipher.Encrypt([]byte(originUrl))
	if err != nil {
		Logger.Error.Println("[COS]生成代理url失败:", err.Error())
		return
	}
	imgUrl := config.General.BaseUrl + "/file/" + fileToken + "." + MD5(fileToken+config.General.AESKey) + fmt.Sprintf("?ts=%d", time.Now().Unix())

	res := new(ResAdminActGetRandomPic)
	res.Status = 0
	res.Data.Url = imgUrl

	c.JSON(200, res)
}
