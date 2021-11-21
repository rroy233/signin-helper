package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"signin/Logger"
	"strconv"
	"time"
)

type CacheIDS struct {
	Total   int   `json:"total"`
	Easy    []int `json:"easy"`
	Careful []int `json:"careful"`
}

const (
	ACT_NOTI_TYPE_CH_NOTI = "ch_noti"
	ACT_NOTI_TYPE_TIME_WARN = "time_warn"
)

func cacheClass(classID int) (class *dbClass, err error) {
	class = new(dbClass)
	err = db.Get(class, "select * from `class` where `class_id`=?", classID)
	if err != nil {
		Logger.Error.Println("[cache][class]班级信息回源读取失败:", err)
		return
	}
	data, err := json.Marshal(class)
	if err != nil {
		Logger.Error.Println("[cache][class]信息回源读取,json失败:", err)
		return
	}
	rdb.Set(ctx, "SIGNIN_APP:Class:"+strconv.FormatInt(int64(classID), 10), data, 5*time.Minute)
	return
}

func cacheAct(ActID int) (Act *dbAct, err error) {
	Act = new(dbAct)
	err = db.Get(Act, "select * from `activity` where `act_id`=?", ActID)
	if err != nil {
		Logger.Error.Println("[cache][act]信息回源读取失败:", err)
		return
	}
	data, err := json.Marshal(Act)
	if err != nil {
		Logger.Error.Println("[cache][act]信息回源读取,json失败:", err)
		return
	}
	rdb.Set(ctx, "SIGNIN_APP:Act:"+strconv.FormatInt(int64(ActID), 10), data, 30*time.Second)
	return
}

//部分缓存机制
func cacheIDs(classID int) (ids *CacheIDS, err error) {
	ids = new(CacheIDS)
	ids.Easy = make([]int, 0)
	ids.Careful = make([]int, 0)

	acts := make([]dbAct, 0)
	err = db.Select(&acts, "select * from `activity` where `class_id`=? and `active`=?", classID, 1)
	if err != nil {
		return nil, err
	}

	for i := range acts {
		var et int64
		et, err = strconv.ParseInt(acts[i].EndTime, 10, 64)
		if err != nil {
			Logger.Error.Println("[cache][cacheIDs]解析时间失败:", err)
			return
		}

		//判断是否过期
		if time.Now().Unix() > et {
			//更新active
			_, err = db.Exec("update `activity` set `active` = ? where `act_id`=?", 0, acts[i].ActID)
			if err != nil {
				Logger.Error.Println("[cache][cacheIDs]更新active失败:", err)
				return
			}
			Logger.Info.Println("[cache][cacheIDs]", acts[i].Name, "活动过期已处理。")
			continue
		}
		if et-time.Now().Unix() > 1*60 { //<1h >5min
			ids.Easy = append(ids.Easy, acts[i].ActID)
		} else if et-time.Now().Unix() < 1*60 {
			ids.Careful = append(ids.Careful, acts[i].ActID)
		}
	}

	ids.Total = len(ids.Easy) + len(ids.Careful)

	data, err := json.Marshal(ids)
	if err != nil {
		Logger.Error.Println("[cache][cacheIDs]信息回源读取,json失败:", err)
		return
	}
	err = rdb.Set(ctx, "SIGNIN_APP:Class_Active_Act:"+strconv.FormatInt(int64(classID), 10), data, 1*time.Minute).Err()
	if err != nil {
		return nil,err
	}
	return
}

func cacheActStatistics(actID int) (res *ActStatistic, err error) {
	act, err := getAct(actID)
	if err != nil {
		return nil, err
	}

	users := make([]dbUser, 0)
	err = db.Select(&users, "select * from `user` where `class`=? ORDER BY `name` DESC;", act.ClassID)
	if err != nil {
		return nil, err
	}

	class := new(dbClass)
	err = db.Get(class, "select * from `class` where `class_id`=?", act.ClassID)
	if err != nil {
		return nil, err
	}

	logs := make([]dbLog, 0)
	err = db.Select(&logs, "select * from `signin_log` where `act_id`=? order by `log_id` desc;", actID)
	if err != nil {
		return nil, err
	}

	res = new(ActStatistic)
	res.Done = len(logs)
	res.Total = len(users)
	res.UnfinishedList = make([]*ActStatisticItem, 0)
	res.FinishedList = make([]*ActStatisticItem, 0)

	logMap := make(map[int]int, len(logs))
	for i := range logs {
		logMap[logs[i].UserID] = logs[i].LogID //不为0
	}

	fC := 0
	ufC := 0
	usernameMap := make(map[int]string, len(logs))
	for i := range users {
		if logMap[users[i].UserID] != 0 {
			fC++
			usernameMap[users[i].UserID] = users[i].Name
		} else {
			//未完成的
			ufC++
			res.UnfinishedList = append(res.UnfinishedList, &ActStatisticItem{
				Id:       ufC,
				UserID:   users[i].UserID,
				ActID:    act.ActID,
				LogID:    -1,
				Name:     users[i].Name,
				DateTime: "N/A",
			})
		}
	}

	//对已完成的用户按照提交时间倒叙排列
	for i := range logs {
		res.FinishedList = append(res.FinishedList, &ActStatisticItem{
			Id:       fC,
			UserID:   logs[i].UserID,
			ActID:    act.ActID,
			LogID:    logs[i].LogID,
			Name:     usernameMap[logs[i].UserID],
			DateTime: ts2DateString(logs[i].CreateTime),
		})
		fC--
	}

	data, err := json.Marshal(res)
	if err != nil {
		Logger.Error.Println("[cache][Act_Statistic]json格式化失败:", err)
		return nil, err
	}
	rdb.Set(ctx, "SIGNIN_APP:Act_Statistic:"+strconv.FormatInt(int64(actID), 10), string(data), 10*time.Second)

	return
}

func queryActIdByActToken(actToken string) (id int, err error) {
	r := rdb.Get(ctx, "SIGNIN_APP:actToken:"+actToken).Val()
	if r == "" {
		return 0, errors.New("actToken不存在")
	}
	id, err = strconv.Atoi(r)
	return
}

//获取用户特定活动已提醒次数
func actNotiUserTimesGet(act *dbAct, userID int) (int,error) {
	tmp,err := rdb.Get(ctx,fmt.Sprintf("SIGNIN_APP:ActNotiTimes:%d:%d",act.ActID,userID)).Result()
	if err != nil {
		return 0,err
	}
	data,_ := strconv.Atoi(tmp)
	return data,nil
}

//存储用户特定活动已提醒次数
func actNotiUserTimesIncr(act *dbAct, userID int) error {
	err := rdb.Incr(ctx,fmt.Sprintf("SIGNIN_APP:ActNotiTimes:%d:%d",act.ActID,userID)).Err()
	return err
}

//删除用户特定活动已提醒次数
func actNotiUserTimesDel(act *dbAct,  userID int) error {
	err := rdb.Del(ctx,fmt.Sprintf("SIGNIN_APP:ActNotiTimes:%d:%d",act.ActID,userID)).Err()
	return err
}