package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func testTplHandler( c *gin.Context){
	tpls := make([]dbTplItem,0)
	err := db.Select(&tpls,"select `title`,`body` from `msg_template`")
	if err != nil {
		returnErrorJson(c,"出现异常")
		return
	}

	act := &dbAct{
		Name: "测试活动",
		EndTime: "0",
	}
	html := ""
	html += `<h1>通知模板</h1><p>以下为当前数据库下所有的通知模板，解析规则请看<a href="https://github.com/rroy233/signin-helper/blob/main/docs/msgTemplate.md" target="_blank">github</a>，投稿请联系管理员。</p>`

	html += `<h3>邮件推送解析</h3><table border="1"><tr><th>ID</th><th>title</th><th>body</th></tr>`
	for i:= range tpls{
		html +="<tr>"
		html +=fmt.Sprintf("<td>%d</td>",i+1)
		html +=fmt.Sprintf("<td>%s</td>",parseEmailTemplate(tpls[i].Title,nil,nil,act))
		html +=fmt.Sprintf("<td>%s</td>",parseEmailTemplate(tpls[i].Body,nil,nil,act))
		html +="</tr>"
	}
	html += `</table>`

	html += `<h3>微信推送解析</h3><table border="1"><tr><th>ID</th><th>title</th><th>body</th></tr>`
	for i:= range tpls{
		task := &NotifyJob{NotificationType: NOTIFICATION_TYPE_EMAIL,Addr: ""}
		html +="<tr>"
		task.Title = parseEmailTemplate(tpls[i].Title,nil,nil,act)
		task.Body = parseWechatBodyTitle(tpls[i].Body,nil,nil,nil,task)
		html +=fmt.Sprintf("<td>%d</td>",i+1)
		html +=fmt.Sprintf("<td>%s</td>",task.Title)
		html +=fmt.Sprintf("<td>%s</td>",task.Body)
		html +="</tr>"
	}
	html += `</table>`
	c.Data(200,ContentTypeHTML,[]byte(html))
}
