package main

import (
	"crypto/tls"
	"errors"
	"github.com/domodwyer/mailyak/v3"
	"net/smtp"
	"signin/Logger"
	"strings"
	"time"
)

var MailQueue chan *mailyak.MailYak

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

type dailyNotifyJob struct {
	NotificationType int `json:"notification_type"`
	Addr string `json:"addr"`
	Title string `json:"title"`
	Body string `json:"body"`
}

func initMail() {
	MailQueue = make(chan *mailyak.MailYak, config.Mail.QueueBufferSize)
}

//新活动通知，群发给班级的所有成员
func newActBulkSend(classID int,act *dbAct) error {
	class,err := getClass(classID)
	if err != nil {
		return err
	}

	users := make([]dbUser,0)
	err = db.Select(&users,"select * from `user` where `class` = ?",class.ClassID)
	if err != nil {
		Logger.Error.Println("[邮件][异常]群发，读取数据库失败",err)
		return errors.New("读取数据库失败，请联系管理员")
	}

	if len(users) == 0{
		return errors.New("班级还没人呢")
	}

	title := "<新打卡任务>「"+act.Name+"」开启啦！"
	body := "{{username}}您好:{{EOL}}{{space}}{{space}}您有一个新的打卡任务哦，截止日期{{act_end_time}}，快快点击下方的链接签到吧~{{EOL}}{{EOL}}{{space}}{{space}}{{login_url_withToken}}"

	for i := range users{
		if users[i].NotificationType ==NOTIFICATION_TYPE_EMAIL || users[i].NotificationType == NOTIFICATION_TYPE_NONE {
			var task *mailyak.MailYak
			task,err = newMailTask(users[i].Email,title,parseTemplate(body,&users[i],class,act))
			if err != nil {
				Logger.Error.Println("[邮件发送][异常]新建发送任务失败->",users[i].Name,err)
				continue
			}
			Logger.Info.Println("[邮件发送]已创建发生任务->",users[i].Name,task.String())
			//推入队列
			MailQueue <- task
		}else if users[i].NotificationType ==NOTIFICATION_TYPE_WECHAT {
			//微信
		}
	}
	return err
}



func newMailTask(mailAddr string,title string,body string) (*mailyak.MailYak,error) {
	var mail *mailyak.MailYak
	var err error
	if config.Mail.TLS == true {
		mail,err = mailyak.NewWithTLS(config.Mail.SmtpServer +":"+ config.Mail.Port, smtp.PlainAuth("", config.Mail.Username, config.Mail.Password, config.Mail.SmtpServer),&tls.Config{
			ServerName: config.Mail.SmtpServer,
		})
		if err != nil {
			return nil,err
		}
	}else{
		mail = mailyak.New(config.Mail.SmtpServer +":"+ config.Mail.Port, smtp.PlainAuth("", config.Mail.Username, config.Mail.Password, config.Mail.SmtpServer))
	}

	mail.To(mailAddr)
	mail.From(config.Mail.Username)
	mail.FromName("签到提醒")

	mail.Subject(title)
	mail.HTML().Set(body)

	return mail,nil
}

func mailSender(queue chan *mailyak.MailYak)  {
	Logger.Info.Println("[邮件]异步发送协程已启动")
	for  {
		sendConfig,ok := <-queue
		Logger.Debug.Println("[邮件][协程]接收到发送任务",sendConfig.String())
		if ok == false {
			Logger.Info.Println("[邮件]",sendConfig.String(),"管道关闭")
			break
		}
		err := sendConfig.Send()
		if err != nil {
			Logger.Info.Println("[邮件]->",sendConfig.String(),"发送失败:",err)
			continue
		}
		Logger.Info.Println("[邮件]",sendConfig.String(),"异步发送成功")
	}
}

//解析并替换模板中的参数
func parseTemplate(s string,user *dbUser,class *dbClass,act *dbAct) string {

	s = strings.Replace(s, "{{EOL}}", "<br>", -1)
	s = strings.Replace(s, "{{space}}", "&nbsp;&nbsp;&nbsp;&nbsp;", -1)

	if user != nil{
		s = strings.Replace(s, "{{username}}", user.Name, -1)
	}

	if act != nil{
		s = strings.Replace(s, "{{act_name}}", act.Name, -1)
		s = strings.Replace(s, "{{act_begin_time}}", ts2DateString(act.BeginTime), -1)
		s = strings.Replace(s, "{{act_end_time}}", ts2DateString(act.EndTime), -1)
		s = strings.Replace(s, "{{act_creator}}", queryUserName(act.CreateBy), -1)
	}

	if class != nil{
		s = strings.Replace(s, "{{class_name}}", class.Name, -1)
	}

	if strings.Contains(s,"{{login_url_withToken}}")==true && user!=nil{
		//签发jwt
		token := "(生成失败，请手动登录)"
		jwt,err := generateJwt(user,generateJwtID(),40*time.Minute)
		if err != nil {
			Logger.Error.Println("[解析模板]生成jwt失败",err)
		}else{
			token = config.General.BaseUrl + "/api/login?jwt=" + jwt
		}
		s = strings.Replace(s, "{{login_url_withToken}}", token+" （有效期40分钟）", -1)
	}

	return s
}