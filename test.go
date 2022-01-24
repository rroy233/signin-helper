package main

import (
	"fmt"
	"github.com/domodwyer/mailyak/v3"
	"github.com/gin-gonic/gin"
)

type formDataTestTplSend struct {
	ID   int    `form:"id" binding:"required"`
	Type string `form:"type" binding:"required"`
	Key  string `form:"key" binding:"required"`
}

func testTplHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorView(c, "登录状态无效")
		return
	}

	tpls := make([]dbTplItem, 0)
	err = db.Select(&tpls, "select * from `msg_template`")
	if err != nil {
		returnErrorJson(c, "出现异常")
		return
	}

	//测试用登录凭证
	loginToken, err := Cipher.Encrypt([]byte(getCookie(c, "token")))
	if err != nil {
		returnErrorJson(c, "出现异常"+err.Error())
		return
	}
	loginUrl := fmt.Sprintf("%s/api/login?jwt=%s.%s", config.General.BaseUrl, loginToken, Cipher.Sha256Hex([]byte(loginToken)))

	act := &dbAct{
		Name:    "测试活动",
		EndTime: "0",
	}
	html := ""
	html += `<h1>通知模板</h1><p>以下为当前数据库下所有的通知模板，解析规则请看<a href="https://github.com/rroy233/signin-helper/blob/main/docs/msgTemplate.md" target="_blank">github</a>，投稿请联系管理员。</p>`

	html += "<h3>当前登录凭证</h3><p>" + loginUrl + "</p>"

	html += `<h3>邮件推送解析</h3><table border="1"><tr><th>ID</th><th>title</th><th>body</th><th>操作(管理员)</th></tr>`
	for i := range tpls {
		html += "<tr>"
		html += fmt.Sprintf("<td>%d</td>", i+1)
		html += fmt.Sprintf("<td>%s</td>", parseEmailTemplate(tpls[i].Title, nil, nil, act))
		html += fmt.Sprintf("<td>%s</td>", parseEmailTemplate(tpls[i].Body, nil, nil, act))
		html += fmt.Sprintf("<td><a href='/debug/noti/send?id=%d&type=email&key=%s' target='_blank'>发送推送</td>", tpls[i].TplID, MD5_short(fmt.Sprintf("%d%d%s", auth.UserID, tpls[i].TplID, "email"+config.General.JwtKey)))
		html += "</tr>"
	}
	html += `</table>`

	html += `<h3>微信推送解析</h3><table border="1"><tr><th>ID</th><th>title</th><th>body</th><th>操作(管理员)</th></tr>`
	for i := range tpls {
		task := &NotifyJob{NotificationType: NOTIFICATION_TYPE_EMAIL, Addr: ""}
		html += "<tr>"
		task.Title = parseEmailTemplate(tpls[i].Title, nil, nil, act)
		task.Body = parseWechatBodyTitle(tpls[i].Body, nil, nil, nil, task)
		html += fmt.Sprintf("<td>%d</td>", i+1)
		html += fmt.Sprintf("<td>%s</td>", task.Title)
		html += fmt.Sprintf("<td>%s</td>", task.Body)
		html += fmt.Sprintf("<td><a href='/debug/noti/send?id=%d&type=wechat&key=%s' target='_blank'>发送推送</td>", tpls[i].TplID, MD5_short(fmt.Sprintf("%d%d%s", auth.UserID, tpls[i].TplID, "wechat"+config.General.JwtKey)))
		html += "</tr>"
	}
	html += `</table>`
	c.Data(200, ContentTypeHTML, []byte(html))
}

func testTplSendHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorView(c, "登录状态无效")
		return
	}

	form := new(formDataTestTplSend)
	err = c.ShouldBindQuery(form)
	if err != nil {
		returnErrorView(c, "参数无效(-1)"+err.Error())
		return
	}

	if form.Key != MD5_short(fmt.Sprintf("%d%d%s", auth.UserID, form.ID, form.Type+config.General.JwtKey)) {
		returnErrorView(c, "参数无效(-2)")
		return
	}

	tpl := new(dbTplItem)
	err = db.Get(tpl, "select * from `msg_template` where `tpl_id`=?", form.ID)
	if err != nil {
		returnErrorView(c, "检索失败:"+err.Error())
		return
	}

	act := &dbAct{
		Name:    "测试活动",
		EndTime: "0",
	}

	if form.Type == "email" {
		var task *mailyak.MailYak
		title := "<DEBUG>" + parseEmailTemplate(tpl.Title, nil, nil, act)
		body := parseEmailTemplate(tpl.Body, nil, nil, act)
		task, err = newMailTask(auth.Email, title, body)
		if err != nil {
			returnErrorView(c, "新建发送任务失败:"+err.Error())
			return
		}
		MailQueue <- task
	} else if form.Type == "wechat" {
		//获取管理员wxid
		wxId := ""
		err = db.Get(&wxId, "select `wx_pusher_uid`from `user` where `user_id`=?", auth.UserID)
		if err != nil {
			returnErrorView(c, "查询wx_pusher_uid失败:"+err.Error())
			return
		}
		if wxId == "" {
			returnErrorView(c, "您尚未绑定微信")
			return
		}
		task := new(NotifyJob)
		task.NotificationType = NOTIFICATION_TYPE_WECHAT
		task.Addr = wxId
		task.Title = "<DEBUG>" + parseEmailTemplate(tpl.Title, nil, nil, act)
		task.Body = parseWechatBodyTitle(tpl.Body, nil, nil, act, task)
		WechatQueue <- task
	} else {
		returnErrorView(c, "type无效")
		return
	}
	c.Data(200, ContentTypeHTML, []byte("发送成功"))
}

func testActNotiHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	noti, err := makeActInnerNoti(1, auth.UserID, ACT_NOTI_TYPE_CH_NOTI)
	if err != nil {
		returnErrorJson(c, err.Error())
		return
	}

	err = pushInnerNoti(auth.UserID, noti)
	if err != nil {
		returnErrorJson(c, err.Error())
		return
	}

	res := new(ResEmpty)
	res.Status = 0
	res.Msg = "已推送测试消息"
	c.JSON(200, res)
}

func testNotiHandler(c *gin.Context) {
	auth, err := getAuthFromContext(c)
	if err != nil {
		returnErrorJson(c, "登录状态无效")
		return
	}

	noti, err := makeInnerNoti(auth.UserID)
	if err != nil {
		returnErrorJson(c, err.Error())
		return
	}

	err = pushInnerNoti(auth.UserID, noti)
	if err != nil {
		returnErrorJson(c, err.Error())
		return
	}

	res := new(ResEmpty)
	res.Status = 0
	res.Msg = "已推送测试消息"
	c.JSON(200, res)
}
