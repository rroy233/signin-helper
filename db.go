package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"signin/Logger"
	"strconv"
	"time"
)

type dbUser struct {
	UserID           int    `db:"user_id"`
	Name             string `db:"name"`
	Email            string `db:"email"`
	Class            int    `db:"class"`
	NotificationType int    `db:"notification_type"`
	WxPusherUid      string `db:"wx_pusher_uid"`
	IsAdmin          int    `db:"is_admin"`
	SsoUid           int    `db:"sso_uid"`
}

type dbClass struct {
	ClassID   int    `db:"class_id" json:"class_id"`
	Name      string `db:"name" json:"name"`
	ClassCode string `db:"class_code" json:"class_code"`
	Total     int    `db:"total" json:"total"`
}

type dbAct struct {
	ActID            int    `db:"act_id" json:"act_id"`
	ClassID          int    `db:"class_id" json:"class_id"`
	Active           int    `json:"active" db:"active"`
	Type             int    `json:"type" db:"type"`
	Name             string `db:"name" json:"name"`
	Announcement     string `db:"announcement" json:"announcement"`
	CheerText        string `db:"cheer_text" json:"cheer_text"`
	Pic              string `json:"pic" db:"pic"`
	DailyNotiEnabled int    `json:"daily_noti_enabled" db:"daily_noti_enabled"`
	BeginTime        string `db:"begin_time" json:"begin_time"`
	EndTime          string `db:"end_time" json:"end_time"`
	CreateTime       string `db:"create_time" json:"create_time"`
	UpdateTime       string `db:"update_time" json:"update_time"`
	CreateBy         int    `db:"create_by" json:"create_by"`
	FileOpts         string `json:"file_opts" db:"file_opts"`
}

type dbLog struct {
	LogID      int    `json:"log_id" db:"log_id"`
	ClassID    int    `json:"class_id" db:"class_id"`
	ActID      int    `json:"act_id" db:"act_id"`
	UserID     int    `json:"user_id" db:"user_id"`
	CreateTime string `json:"create_time" db:"create_time"`
	FileID     int    `json:"file_id" db:"file_id"`
}

type dbTplItem struct {
	TplID   int    `db:"tpl_id" json:"tpl_id"`
	MsgType int    `json:"msg_type" db:"msg_type"`
	Level   int    `json:"level" db:"level"`
	Title   string `db:"title" json:"title"`
	Body    string `db:"body" json:"body"`
	Enabled int    `json:"enabled" db:"enabled"`
}

type dbFile struct {
	FileID      int    `json:"file_id" db:"file_id"`
	Status      int    `json:"status" db:"status"`
	UserID      int    `json:"user_id" db:"user_id"`
	ActID       int    `json:"act_id" db:"act_id"`
	FileName    string `json:"file_name" db:"file_name"`
	ContentType string `json:"content_type" db:"content_type"`
	Local       string `json:"local" db:"local"`
	Remote      string `json:"remote" db:"remote"`
	ExpTime     string `json:"exp_time" db:"exp_time"`
	UploadTime  string `json:"upload_time" db:"upload_time"`
}

const (
	FILE_STATUS_DELETED = -1
	FILE_STATUS_LOCAL   = 0
	FILE_STATUS_REMOTE  = 1
)

const (
	NOTIFICATION_TYPE_NONE = iota
	NOTIFICATION_TYPE_EMAIL
	NOTIFICATION_TYPE_WECHAT
)

var db *sqlx.DB       //mysql client
var rdb *redis.Client //redis client
var ctx = context.Background()

// initDB mysql?????????
func initDB() {
	var err error
	dsn := config.Db.Username + ":" + config.Db.Password + "@tcp(" + config.Db.Server + ":" + config.Db.Port + ")/" + config.Db.Db + "?charset=utf8mb4&parseTime=True"
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		Logger.FATAL.Fatalln("[????????????][??????]Mysql????????????" + err.Error())
		return
	}
	Logger.Info.Println("[????????????][??????]Mysql?????????")

	//??????????????????
	db.SetConnMaxIdleTime(5 * time.Second)
	//??????????????????????????????
	db.SetMaxOpenConns(1000)
	//????????????????????????????????????
	db.SetMaxIdleConns(20)
	return
}

func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	err := rdb.Ping(ctx).Err()
	if err != nil {
		Logger.FATAL.Fatalln("[????????????][??????]Redis????????????")
		return
	}
	Logger.Info.Println("[????????????][??????]Redis?????????")
	return
}

// checkDB ??????mysql???????????????
func checkDB() {
	if db == nil {
		Logger.FATAL.Fatalln("Mysql ???????????????")
	}
	err := db.Ping()
	if err != nil {
		Logger.Info.Println("[????????????]Mysql??????????????????")
	}
}

// checkRedis ??????redis???????????????
func checkRedis() {
	if rdb == nil {
		Logger.FATAL.Fatalln("Redis ???????????????")
	}
	err := rdb.Ping(ctx).Err()
	if err != nil {
		Logger.Info.Println("[????????????]Redis??????????????????")
	}
}

func getAffectedRows(r sql.Result) int64 {
	tmp, _ := r.RowsAffected()
	return tmp
}

func createUser(user *dbUser) (int, error) {
	if user == nil {
		return -1, errors.New("user is nil")
	}
	res, err := db.Exec("INSERT INTO `user` (`user_id`, `name`, `email`, `class`, `notification_type`, `wx_pusher_uid`, `is_admin`, `sso_uid`) VALUES (NULL, ?, ?,?, ?, ?, ?, ?);",
		user.Name,
		user.Email,
		user.Class,
		user.NotificationType,
		user.WxPusherUid,
		user.IsAdmin,
		user.SsoUid,
	)
	if err != nil {
		return -1, err
	}
	uid, _ := res.LastInsertId()
	return int(uid), err
}

func getClass(classID int) (class *dbClass, err error) {
	class = new(dbClass)
	classCache := rdb.Get(ctx, "SIGNIN_APP:Class:"+strconv.FormatInt(int64(classID), 10)).Val()
	if classCache == "" {
		//????????????????????????????????????
		class, err = cacheClass(classID)
		if err != nil {
			Logger.Error.Println("[DB]????????????????????????:", err)
			return
		}
	} else {
		err = json.Unmarshal([]byte(classCache), class)
		if err != nil {
			Logger.Error.Println("[DB]??????????????????????????????:", err)
			return
		}
	}
	return class, err
}

func getAct(actID int) (act *dbAct, err error) {
	act = new(dbAct)
	actCache := rdb.Get(ctx, "SIGNIN_APP:Act:"+strconv.FormatInt(int64(actID), 10)).Val()
	if actCache == "" {
		//????????????????????????????????????
		act, err = cacheAct(actID)
		if err != nil {
			Logger.Error.Println("[DB][act]??????????????????:", err)
			return
		}
	} else {
		err = json.Unmarshal([]byte(actCache), act)
		if err != nil {
			Logger.Error.Println("[DB][act]????????????????????????:", err)
			return
		}
	}
	return act, err
}

func getActIDs(classID int) (res []int, err error) {
	ids := new(CacheIDS)
	res = make([]int, 0)

	idsCache := rdb.Get(ctx, "SIGNIN_APP:Class_Active_Act:"+strconv.FormatInt(int64(classID), 10)).Val()
	if idsCache == "" {
		//????????????????????????????????????
		ids, err = cacheIDs(classID)
		if err != nil {
			Logger.Error.Println("[DB][getActIDs]??????????????????:", err)
			return
		}
	} else {
		err = json.Unmarshal([]byte(idsCache), &ids)
		if err != nil {
			Logger.Error.Println("[DB][getActIDs]????????????????????????:", err)
			return
		}
	}

	//???easy???????????????res
	for i := range ids.Easy {
		res = append(res, ids.Easy[i])
	}

	//careful???????????????mysql
	for i := range ids.Careful {
		act := new(dbAct)
		err = db.Get(act, "select * from `activity` where `act_id`=?", ids.Careful[i])
		if err != nil {
			Logger.Error.Println("[DB][getActIDs]careful???????????????mysql??????:", err)
			return
		}
		var et int64
		et, err = strconv.ParseInt(act.EndTime, 10, 64)
		if err != nil {
			Logger.Error.Println("[cache][cacheIDs]??????????????????:", err)
			return
		}
		if time.Now().Unix() < et {
			//?????????
			res = append(res, ids.Careful[i])
		}
	}

	return res, err
}

func getActStatistics(actID int) (sts *ActStatistic, err error) {
	sts = new(ActStatistic)
	stsCache := rdb.Get(ctx, "SIGNIN_APP:Act_Statistic::"+strconv.FormatInt(int64(actID), 10)).Val()
	if stsCache == "" {
		//????????????????????????????????????
		sts, err = cacheActStatistics(actID)
		if err != nil {
			Logger.Error.Println("[DB][Act_Statistic]??????????????????:", err)
			return
		}
	} else {
		err = json.Unmarshal([]byte(stsCache), sts)
		if err != nil {
			Logger.Error.Println("[DB][Act_Statistic]????????????????????????:", err)
			return
		}
	}
	return sts, err
}

func queryUserName(userid int) string {
	name := ""
	err := db.Get(&name, "select `name` from `user` where `user_id`=?", userid)
	if err != nil {
		Logger.Info.Println("[DB]??????username??????", err)
		return "??????"
	}
	return name
}

func existIn(src []int, val int) bool {
	if len(src) == 0 {
		return false
	}
	for i := range src {
		if src[i] == val {
			return true
		}
	}
	return false
}
