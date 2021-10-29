package main

import (
	"encoding/json"
	"signin/Logger"
	"strconv"
	"time"
)

type CacheIDS struct {
	Total int `json:"total"`
	Easy []int `json:"easy"`
	Careful []int `json:"careful"`
}

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
func cacheIDs(classID int) (ids *CacheIDS,err error) {
	ids = new(CacheIDS)
	ids.Easy = make([]int,0)
	ids.Careful = make([]int,0)

	acts := make([]dbAct,0)
	err = db.Select(&acts,"select * from `activity` where `class_id`=? and `active`=?",classID,1)
	if err != nil {
		return nil,err
	}
	if len(acts)==0{
		return ids,nil
	}

	for i:= range acts{
		//不判断begin_time了
		var et int64
		et,err = strconv.ParseInt(ts2DateString(acts[i].EndTime),10,64)
		if err != nil {
			Logger.Error.Println("[cache][cacheIDs]解析时间失败:", err)
			return
		}

		//判断是否过期
		if time.Now().Unix()>et {
			//更新active
			_,err = db.Exec("update `activity` set `active` = ? where `act_id`=?",0,acts[i].ActID)
			if err != nil {
				Logger.Error.Println("[cache][cacheIDs]更新active失败:", err)
				return
			}
			continue
		}
		if et - time.Now().Unix() > 1*60{//<1h >5min
			ids.Easy = append(ids.Easy,acts[i].ActID)
		}else if et - time.Now().Unix() < 1*60{
			ids.Careful = append(ids.Careful,acts[i].ActID)
		}
	}

	ids.Total = len(ids.Easy)+len(ids.Careful)

	data, err := json.Marshal(ids)
	if err != nil {
		Logger.Error.Println("[cache][cacheIDs]信息回源读取,json失败:", err)
		return
	}
	rdb.Set(ctx, "SIGNIN_APP:Class_Active_Act:"+strconv.FormatInt(int64(classID), 10), data, 1*time.Minute)
	return
}

func cacheClassStatistics(classID int) (res *ResUserActStatistic, err error) {
	users := make([]dbUser, 0)
	err = db.Select(&users, "select * from `user` where `class`=?;", classID)
	if err != nil {
		return nil, err
	}

	class := new(dbClass)
	err = db.Get(class, "select * from `class` where `class_id`=?", classID)
	if err != nil {
		return nil, err
	}

	logItem := make([]dbLog, 0)
	err = db.Select(&logItem, "select * from `signin_log` where `act_id`=? order by `log_id` desc;", class.ActID)
	if err != nil {
		return nil, err
	}

	res = new(ResUserActStatistic)
	res.Status = 0
	res.Msg = strconv.FormatInt(time.Now().Unix(), 10)
	res.Data.Done = len(logItem)
	res.Data.Total = len(users)
	res.Data.UnfinishedList = make([]actStatisticUser, 0)
	res.Data.FinishedList = make([]actStatisticUser, 0)

	logMap := make(map[int]int, len(logItem))
	for i := range logItem {
		logMap[logItem[i].UserID] = 1
	}

	fC := 0
	ufC := 0
	usernameMap := make(map[int]string, len(logItem))
	for i := range users {
		if logMap[users[i].UserID] == 1 {
			fC++
			usernameMap[users[i].UserID] = users[i].Name
		} else {
			ufC++
			res.Data.UnfinishedList = append(res.Data.UnfinishedList, actStatisticUser{
				Id:     ufC,
				UserID: users[i].UserID,
				Name:   users[i].Name,
				Avatar: "null",
			})
		}
	}

	//对已完成的用户按照提交时间倒叙排列
	for i := range logItem {
		res.Data.FinishedList = append(res.Data.FinishedList, actStatisticUser{
			Id:     fC,
			UserID: 0,
			Name:   usernameMap[logItem[i].UserID],
			Avatar: "null",
		})
		fC--
	}

	data, err := json.Marshal(res)
	if err != nil {
		Logger.Error.Println("[cache][ClassStatistics]json格式化失败:", err)
		return nil, err
	}
	rdb.Set(ctx, "SIGNIN_APP:Class_Statistics:"+strconv.FormatInt(int64(classID), 10), string(data), 10*time.Second)

	return
}
