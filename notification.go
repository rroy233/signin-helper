package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/domodwyer/mailyak/v3"
	"math/rand"
	"net/http"
	"net/smtp"
	"signin/Logger"
	"strings"
	"time"
)

var MailQueue chan *mailyak.MailYak
var WechatQueue chan *NotifyJob

const (
	TPL_MSGTYPE_newAct = iota
	TPL_MSGTYPE_daily
)
const (
	TPL_LEVEL_LOW = iota
	TPL_LEVEL_MID
	TPL_LEVEL_HIGH
	TPL_LEVEL_URGE
)

type NotifyJob struct {
	NotificationType int    `json:"notification_type"`
	Addr             string `json:"addr"`
	Title            string `json:"title"`
	Body             string `json:"body"`
	Url              string `json:"url"`
}

type wxMsgData struct {
	AppToken    string   `json:"appToken"`
	Content     string   `json:"content"`
	Summary     string   `json:"summary"`
	ContentType int      `json:"contentType"`
	TopicIds    []string `json:"topicIds"`
	Uids        []string `json:"uids"`
	URL         string   `json:"url"`
}

func initMail() {
	MailQueue = make(chan *mailyak.MailYak, config.Mail.QueueBufferSize)
	WechatQueue = make(chan *NotifyJob, 50)
}

//新活动通知，群发给班级的所有成员
func newActBulkSend(classID int, act *dbAct) error {
	class, err := getClass(classID)
	if err != nil {
		return err
	}

	users := make([]dbUser, 0)
	err = db.Select(&users, "select * from `user` where `class` = ?", class.ClassID)
	if err != nil {
		Logger.Error.Println("[邮件][异常]群发，读取数据库失败", err)
		return errors.New("读取数据库失败，请联系管理员")
	}

	if len(users) == 0 {
		return errors.New("班级还没人呢")
	}

	title := "<新打卡任务>「" + act.Name + "」开启啦！"
	body := "{{username}}您好:{{EOL}}{{space}}{{space}}您有一个新的打卡任务哦，截止日期{{act_end_time}}，快快点击下方的链接签到吧~{{EOL}}{{EOL}}{{space}}{{space}}{{login_url_withToken}}"

	for i := range users {
		if users[i].NotificationType == NOTIFICATION_TYPE_EMAIL || users[i].NotificationType == NOTIFICATION_TYPE_NONE {
			var task *mailyak.MailYak
			task, err = newMailTask(users[i].Email, title, parseEmailTemplate(body, &users[i], class, act))
			if err != nil {
				Logger.Error.Println("[邮件发送][异常]新建发送任务失败->", users[i].Name, err)
				continue
			}
			Logger.Info.Println("[邮件发送]已创建发生任务->", users[i].Name)
			//推入队列
			MailQueue <- task
		} else if users[i].NotificationType == NOTIFICATION_TYPE_WECHAT {
			//微信
			task := new(NotifyJob)
			task.NotificationType = NOTIFICATION_TYPE_WECHAT
			task.Addr = users[i].WxPusherUid
			task.Title = title
			task.Body = parseWechatBodyTitle(body, &users[i], class, act, task)
			Logger.Info.Println("[微信推送]已创建发生任务->", users[i].Name, task)
			WechatQueue <- task
		}
	}
	return err
}

func ActEndingBulkSend(classID int, act *dbAct) error {
	class, err := getClass(classID)
	if err != nil {
		return err
	}

	admins := make([]dbUser, 0)
	err = db.Select(&admins, "select * from `user` where `class` = ? and `is_admin`=1;", class.ClassID)
	if err != nil {
		Logger.Error.Println("[群发][异常]群发，读取数据库失败", err)
		return errors.New("读取数据库失败，请联系管理员")
	}

	if len(admins) == 0 {
		return errors.New("没人是管理员")
	}

	title := "<提醒>「" + act.Name + "」参与率未达标"
	timeHour := time.Now().Format("15")
	body := ""
	if timeHour == "12"{
		body = "管理员{{username}}您好:{{EOL}}{{space}}{{space}}感谢您使用本系统。{{EOL}}{{space}}{{space}}活动「{{act_name}}」将于今天下午6:30之前结束，此时活动参与率并未达到100%，详细情况请点击链接进入系统查看。{{EOL}}{{EOL}}快捷入口：{{login_url_withToken}}"
	}else if timeHour == "18"{
		body = "管理员{{username}}您好:{{EOL}}{{space}}{{space}}感谢您使用本系统。{{EOL}}{{space}}{{space}}活动「{{act_name}}」将于明天中午12:30之前结束，此时活动参与率并未达到100%，详细情况请点击链接进入系统查看。{{EOL}}{{EOL}}快捷入口：{{login_url_withToken}}"
	}else{
		return errors.New("时间不是规定值")
	}

	for i := range admins {
		if admins[i].NotificationType == NOTIFICATION_TYPE_EMAIL || admins[i].NotificationType == NOTIFICATION_TYPE_NONE {
			var task *mailyak.MailYak
			task, err = newMailTask(admins[i].Email, title, parseEmailTemplate(body, &admins[i], class, act))
			if err != nil {
				Logger.Error.Println("[邮件发送][异常]新建发送任务失败->", admins[i].Name, err)
				continue
			}
			Logger.Info.Println("[邮件发送]已创建发生任务->", admins[i].Name, task.String())
			//推入队列
			MailQueue <- task
		} else if admins[i].NotificationType == NOTIFICATION_TYPE_WECHAT {
			//微信
			task := new(NotifyJob)
			task.NotificationType = NOTIFICATION_TYPE_WECHAT
			task.Addr = admins[i].WxPusherUid
			task.Title = title
			task.Body = parseWechatBodyTitle(body, &admins[i], class, act, task)
			Logger.Info.Println("[微信推送]已创建发生任务->", admins[i].Name, task)
			WechatQueue <- task
		}
	}
	return err
}

func newMailTask(mailAddr string, title string, body string) (*mailyak.MailYak, error) {
	var mail *mailyak.MailYak
	var err error
	if config.Mail.TLS == true {
		mail, err = mailyak.NewWithTLS(config.Mail.SmtpServer+":"+config.Mail.Port, smtp.PlainAuth("", config.Mail.Username, config.Mail.Password, config.Mail.SmtpServer), &tls.Config{
			ServerName: config.Mail.SmtpServer,
		})
		if err != nil {
			return nil, err
		}
	} else {
		mail = mailyak.New(config.Mail.SmtpServer+":"+config.Mail.Port, smtp.PlainAuth("", config.Mail.Username, config.Mail.Password, config.Mail.SmtpServer))
	}

	mail.To(mailAddr)
	mail.From(config.Mail.Username)
	mail.FromName("签到提醒")

	mail.Subject(title)
	mail.HTML().Set(body)

	return mail, nil
}

func mailSender(queue chan *mailyak.MailYak) {
	Logger.Info.Println("[邮件]异步发送协程已启动")
	for {
		sendConfig, ok := <-queue
		if ok == false {
			Logger.Info.Println("[邮件]", sendConfig.String(), "管道关闭")
			break
		}
		err := sendConfig.Send()
		if err != nil {
			Logger.Info.Println("[邮件]->", sendConfig.String(), "发送失败:", err)
			continue
		}
		Logger.Info.Println("[邮件]", sendConfig, "异步发送成功")
	}
}

func wechatSender(queue chan *NotifyJob) {
	Logger.Info.Println("[微信推送]异步发送协程已启动")
	for {
		sendConfig, ok := <-queue
		if ok == false {
			Logger.Info.Println("[微信推送]", sendConfig, "管道关闭")
			break
		}
		//构造发送
		data := new(wxMsgData)
		data.AppToken = config.WxPusher.AppToken
		data.Content = sendConfig.Body
		data.Summary = sendConfig.Title
		data.ContentType = 2 //html
		data.Uids = make([]string, 1)
		data.Uids[0] = sendConfig.Addr
		data.URL = sendConfig.Url

		dataJson, err := json.Marshal(data)
		if err != nil {
			Logger.Error.Println("[微信推送]", sendConfig, "json格式化失败", err)
			break
		}
		resp, err := http.Post("http://wxpusher.zjiecode.com/api/send/message", "application/json", bytes.NewReader(dataJson))
		if err != nil {
			Logger.Error.Println("[微信推送]", sendConfig, "请求api失败", err)
			break
		}
		Logger.Info.Println("[微信推送]", sendConfig, "异步发送成功",resp)
	}
}

//解析并替换模板中的参数
func parseEmailTemplate(s string, user *dbUser, class *dbClass, act *dbAct) string {

	s = strings.Replace(s, "{{EOL}}", "<br>", -1)
	s = strings.Replace(s, "{{space}}", "&nbsp;&nbsp;&nbsp;&nbsp;", -1)
	s = strings.Replace(s, "{{time_now}}", time.Now().Format("2006年01月02日 15时04分05秒"), -1)

	if user != nil {
		s = strings.Replace(s, "{{username}}", user.Name, -1)
	}

	if act != nil {
		s = strings.Replace(s, "{{act_name}}", act.Name, -1)
		s = strings.Replace(s, "{{act_begin_time}}", ts2DateString(act.BeginTime), -1)
		s = strings.Replace(s, "{{act_end_time}}", ts2DateString(act.EndTime), -1)
		if strings.Contains(s,"{{act_creator}}") == true{
			s = strings.Replace(s, "{{act_creator}}", queryUserName(act.CreateBy), -1)
		}
	}

	if class != nil {
		s = strings.Replace(s, "{{class_name}}", class.Name, -1)
	}

	if strings.Contains(s, "{{login_url_withToken}}") == true && user != nil {
		//签发jwt
		token := "(生成失败，请手动登录)"
		jwt, err := generateJwt(user, generateJwtID(), 40*time.Minute)
		if err != nil {
			Logger.Error.Println("[解析模板]生成jwt失败", err)
		} else {
			token = config.General.BaseUrl + "/api/login?jwt=" + jwt
		}
		s = strings.Replace(s, "{{login_url_withToken}}", token+" （有效期40分钟,若出现错误请复制到浏览器打开）", -1)
	}

	return s
}

func parseWechatBodyTitle(s string, user *dbUser, class *dbClass, act *dbAct, task *NotifyJob) string {

	s = strings.Replace(s, "{{EOL}}", "<br>", -1)
	s = strings.Replace(s, "{{space}}", "&nbsp;&nbsp;", -1)
	s = strings.Replace(s, "{{time_now}}", time.Now().Format("2006年01月02日 15时04分05秒"), -1)

	if task.Title != "" {
		if strings.Contains(task.Title, "<") == true {
			task.Title = strings.Replace(task.Title, "<", "【", -1)
			task.Title = strings.Replace(task.Title, ">", "】", -1)
		}
	}

	if user != nil {
		s = strings.Replace(s, "{{username}}", user.Name, -1)
	}

	if act != nil {
		s = strings.Replace(s, "{{act_name}}", act.Name, -1)
		s = strings.Replace(s, "{{act_begin_time}}", ts2DateString(act.BeginTime), -1)
		s = strings.Replace(s, "{{act_end_time}}", ts2DateString(act.EndTime), -1)
		s = strings.Replace(s, "{{act_creator}}", queryUserName(act.CreateBy), -1)
	}

	if class != nil {
		s = strings.Replace(s, "{{class_name}}", class.Name, -1)
	}

	if strings.Contains(s, "{{login_url_withToken}}") == true && user != nil {
		//签发jwt
		token := "(生成失败，请手动登录)"
		jwt, err := generateJwt(user, generateJwtID(), 40*time.Minute)
		if err != nil {
			Logger.Error.Println("[解析模板]生成jwt失败", err)
			s = strings.Replace(s, "{{login_url_withToken}}", token, -1)
		} else {
			token = config.General.BaseUrl + "/api/login?jwt=" + jwt
			task.Url = token
			s = strings.Replace(s, "{{login_url_withToken}}", "(点击\"阅读原文\"快速签到，入口有效期40分钟)", -1)
		}
	}
	s = task.Title + "<br>" + s
	return s
}

//推送站内信息
func pushInnerNoti(userID int,notiItem *UserNotiFetchItem) error {
	if notiItem == nil {
		return errors.New("notiItem为空")
	}
	data,err := json.Marshal(notiItem)
	if err != nil {
		return err
	}
	err = rdb.SetNX(ctx,fmt.Sprintf("SIGNIN_APP:UserNoti:USER_%d:%s",userID,notiItem.Token),data,-1).Err()
	return err
}

func makeActInnerNoti(actID int,userID int,actNotiType string) (*UserNotiFetchItem,error) {
	token := MD5_short(fmt.Sprintf("%d%d%s",userID,actID,actNotiType))
	item := new(UserNotiFetchItem)
	item.Token = token

	switch actNotiType {
	case ACT_NOTI_TYPE_CH_NOTI:
		item.Type = "info"
		item.Text = "请选择有效的\"通知方式\"以确保您能及时收到提醒推送"
	case ACT_NOTI_TYPE_TIME_WARN:
		item.Type = "warning"
		item.Text = "请您及时完成任务并进行签到"
	default:
		return nil,errors.New("actNotiType无效")
	}
	data,err := json.Marshal(item)
	if err != nil {
		return nil,err
	}
	err = rdb.SetNX(ctx,"SIGNIN_APP:UserNoti:USER_"+fmt.Sprintf("%d",userID)+":"+token,string(data),-1).Err()
	if err != nil {
		return nil,err
	}
	return item,nil
}

func makeInnerNoti(userID int) (*UserNotiFetchItem,error) {
	rand.Seed(time.Now().UnixNano())
	token := MD5_short(fmt.Sprintf("%d%d%d",userID,time.Now().UnixNano(),rand.Intn(999)))
	item := new(UserNotiFetchItem)
	item.Token = token
	item.Type = "info"
	item.Text = "测试信息"+time.Now().Format("2006-01-02 15:04:05")
	data,err := json.Marshal(item)
	if err != nil {
		return nil,err
	}
	err = rdb.SetNX(ctx,"SIGNIN_APP:UserNoti:USER_"+fmt.Sprintf("%d",userID)+":"+token,string(data),-1).Err()
	if err != nil {
		return nil,err
	}
	return item,nil
}