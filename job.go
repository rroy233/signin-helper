package main

import (
	"encoding/json"
	"github.com/robfig/cron/v3"
	"math/rand"
	"signin/Logger"
	"strconv"
	"time"
)

var crontab *cron.Cron

func initJob() {
	crontab = cron.New(cron.WithLocation(TZ))
	var err error


	//数据准备
	//每天12:29+18:29
	_, err = crontab.AddFunc("29 12 * * ?", PrepareDailyNotification)
	_, err = crontab.AddFunc("29 18 * * ?", PrepareDailyNotification)
	if err != nil {
		Logger.FATAL.Fatalln("[定时任务][异常]添加PrepareDailyNotification失败:", err)
	} else {
		Logger.Info.Println("[定时任务][成功]添加PrepareDailyNotification成功")
	}

	//发送任务
	//每天12:30+19:30
	_, err = crontab.AddFunc("30 12 * * ?", SendDailyNotification)
	_, err = crontab.AddFunc("30 18 * * ?", SendDailyNotification)
	if err != nil {
		Logger.FATAL.Fatalln("[定时任务][异常]添加SendDailyNotification失败:", err)
	} else {
		Logger.Info.Println("[定时任务][成功]添加SendDailyNotification成功")
	}

	crontab.Start()

}

func PrepareDailyNotification() {

	classList := make([]int, 0)
	err := db.Select(&classList, "SELECT  `class_id` FROM `class`")
	if err != nil {
		Logger.Error.Println("[定时任务][准备数据]获取班级列表失败", err)
	}

	//遍历每个班级
	for i, classID := range classList {
		Logger.Info.Println("[定时任务][准备数据]开始处理班级ID:", classID, "(", i+1, "/", len(classList), ")")

		//获取班级内所有用户
		users := make([]dbUser, 0)
		err = db.Select(&users, "select * from `user` where `class`=?", classID)
		if err != nil {
			Logger.Error.Println("[定时任务][准备数据]读取user表失败", err)
		}
		if len(users) == 0 {
			//太惨了，都关闭通知
			Logger.Info.Println("[定时任务][准备数据]开始处理班级ID:", classID, "无可发送用户")
			continue
		}

		//获取班级信息
		class, err := getClass(classID)
		if err != nil {
			Logger.Info.Println("[定时任务][准备数据]获取班级信息失败:", classID, err)
			continue
		}

		//获取班级有效活动
		actIDs, err := getActIDs(classID)
		if err != nil {
			Logger.Info.Println("[定时任务][准备数据]获取班级有效活动失败:", classID, err)
			continue
		}

		//遍历班级的每一个有效活动
		for _, actID := range actIDs {
			//获取参与数据
			sts, err := cacheActStatistics(actID)
			if err != nil {
				Logger.Info.Println("[定时任务][准备数据]获取参与数据失败")
				continue
			}

			act, err := getAct(actID)
			if err != nil {
				Logger.Info.Println("[定时任务][准备数据]获取活动信息失败:", actID, err)
				continue
			}

			//若全员完成，则不发送
			if sts.Done == sts.Total {
				Logger.Info.Println("[定时任务][准备数据]班级ID:", classID, "全员完成(", i+1, "/", len(classList), ")")
				continue
			}

			//id->user映射
			userMap := make(map[int]*dbUser, 0)
			for i := range users {
				userMap[users[i].UserID] = &users[i]
			}

			//遍历班级内所有未完成的同学
			for i := range sts.UnfinishedList {
				thisUser := userMap[sts.UnfinishedList[i].UserID]
				job, err := getTemplate(TPL_MSGTYPE_daily, TPL_LEVEL_LOW)
				if err != nil {
					Logger.Error.Println("[定时任务][准备数据]获取模板失败", err, thisUser)
					continue
				}
				var msgJson []byte
				//判断通知方式
				if thisUser.NotificationType == NOTIFICATION_TYPE_EMAIL {
					job.NotificationType = NOTIFICATION_TYPE_EMAIL
					job.Addr = thisUser.Email
					job.Title = parseEmailTemplate(job.Title, thisUser, class, act)
					job.Body = parseEmailTemplate(job.Body, thisUser, class, act)
				} else if thisUser.NotificationType == NOTIFICATION_TYPE_WECHAT {
					//微信
					job.NotificationType = NOTIFICATION_TYPE_WECHAT
					job.Title = parseEmailTemplate(job.Title, thisUser, class, act)
					job.Body = parseWechatBodyTitle(job.Body, thisUser, class, act, job)
					job.Addr = thisUser.WxPusherUid
				}else if thisUser.NotificationType == NOTIFICATION_TYPE_NONE {
					//已关闭通知
					continue
				}
				msgJson, err = json.Marshal(job)
				if err != nil {
					Logger.Error.Println("[定时任务][准备数据]json格式化失败", err, thisUser)
					continue
				}
				Logger.Info.Println("[定时任务][准备数据]已添加任务:", string(msgJson))
				//存入redis
				rdb.LPush(ctx, "SIGNIN_APP:NOTI_LIST", string(msgJson))
			}

			//活动参与率未达标，群发给管理员
			if sts.Done<sts.Total {
				timeHour := time.Now().Format("15")
				et,err := strconv.ParseInt(act.EndTime,10,64)
				if err != nil {
					Logger.Error.Println("[定时任务][准备数据][群发给管理员]时间转换失败")
				}else{
					if (timeHour == "12" && et - time.Now().Unix() < 6*60*60) || (timeHour == "18" && et - time.Now().Unix() < 18*60*60){
						err = ActEndingBulkSend(classID,act)
					}
					if err != nil {
						Logger.Error.Println("[定时任务][准备数据][群发给管理员]发送报错：",err)
					}
				}
			}
		}

	}

}

func SendDailyNotification() {
	jobJsons, err := rdb.LRange(ctx, "SIGNIN_APP:NOTI_LIST", int64(0), int64(-1)).Result()
	if err != nil {
		Logger.Error.Println("[定时任务][发送]读取队列内容失败", err)
		return
	}
	if len(jobJsons) == 0 {
		//无发送任务
		return
	}

	total := len(jobJsons)
	for i := range jobJsons {
		Logger.Info.Println("[定时任务][发送][", i+1, "/", total, "]已读取", jobJsons[i])
		job := new(NotifyJob)
		err = json.Unmarshal([]byte(jobJsons[i]), job)
		if err != nil {
			Logger.Error.Println("[定时任务][发送][", i+1, "/", total, "]json解析失败", err)
			continue
		}
		if job.NotificationType == NOTIFICATION_TYPE_EMAIL {
			mailTask, err := newMailTask(job.Addr, job.Title, job.Body)
			if err != nil {
				Logger.Error.Println("[定时任务][发送][", i+1, "/", total, "]创建邮件任务失败", err)
				continue
			}
			MailQueue <- mailTask
		} else if job.NotificationType == NOTIFICATION_TYPE_WECHAT {
			//微信
			WechatQueue <- job
		}
		Logger.Info.Println("[定时任务][发送][", i+1, "/", total, "]发送成功")
	}
	rdb.Del(ctx, "SIGNIN_APP:NOTI_LIST")
}

//随机选取发送模板
func getTemplate(msgType int, level int) (*NotifyJob, error) {
	job := new(NotifyJob)
	tpls := make([]dbTplItem, 0)
	err := db.Select(&tpls, "select `title`,`body` from `msg_template` where `msg_type`=? and `level`=?", msgType, level)
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	id := rand.Intn(len(tpls))
	job.Title = tpls[id].Title
	job.Body = tpls[id].Body
	return job, err
}
