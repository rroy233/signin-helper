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
	ActID     int    `db:"act_id" json:"act_id"`
}

type dbAct struct {
	ActID        int    `db:"act_id" json:"act_id"`
	ClassID      int    `db:"class_id" json:"class_id"`
	Name         string `db:"name" json:"name"`
	Announcement string `db:"announcement" json:"announcement"`
	CheerText    string `db:"cheer_text" json:"cheer_text"`
	Pic          string `json:"pic" db:"pic"`
	BeginTime    string `db:"begin_time" json:"begin_time"`
	EndTime      string `db:"end_time" json:"end_time"`
	CreateTime   string `db:"create_time" json:"create_time"`
	UpdateTime   string `db:"update_time" json:"update_time"`
	CreateBy     int    `db:"create_by" json:"create_by"`
}

type dbLog struct {
	LogID      int    `json:"log_id" db:"log_id"`
	ClassID    int    `json:"class_id" db:"class_id"`
	ActID      int    `json:"act_id" db:"act_id"`
	UserID     int    `json:"user_id" db:"user_id"`
	CreateTime string `json:"create_time" db:"create_time"`
}

const (
	NOTIFICATION_TYPE_NONE = iota
	NOTIFICATION_TYPE_EMAIL
	NOTIFICATION_TYPE_WECHAT
)

var db *sqlx.DB       //mysql client
var rdb *redis.Client //redis client
var ctx = context.Background()

// initDB mysql初始化
func initDB() {
	var err error
	dsn := config.Db.Username + ":" + config.Db.Password + "@tcp(" + config.Db.Server + ":" + config.Db.Port + ")/" + config.Db.Db + "?charset=utf8mb4&parseTime=True"
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		Logger.FATAL.Fatalln("[系统服务][异常]Mysql启动失败" + err.Error())
		return
	}
	Logger.Info.Println("[系统服务][成功]Mysql已连接")

	//最大闲置时间
	db.SetConnMaxIdleTime(5 * time.Second)
	//设置连接池最大连接数
	db.SetMaxOpenConns(1000)
	//设置连接池最大空闲连接数
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
		Logger.FATAL.Fatalln("[系统服务][异常]Redis启动失败")
		return
	}
	Logger.Info.Println("[系统服务][成功]Redis已连接")
	return
}

// checkDB 检查mysql是否有连接
func checkDB() {
	if db == nil {
		Logger.FATAL.Fatalln("Mysql 初始化失败")
	}
	err := db.Ping()
	if err != nil {
		Logger.Info.Println("[系统服务]Mysql重新建立连接")
	}
}

// checkRedis 检查redis是否有连接
func checkRedis() {
	if rdb == nil {
		Logger.FATAL.Fatalln("Redis 初始化失败")
	}
	err := rdb.Ping(ctx).Err()
	if err != nil {
		Logger.Info.Println("[系统服务]Redis重新建立连接")
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
		Logger.Info.Println("[DB]回源读取班级信息:", err)
		//回源请求数据库，然后缓存
		class, err = cacheClass(classID)
		if err != nil {
			Logger.Error.Println("[DB]班级信息回源失败:", err)
			return
		}
	} else {
		err = json.Unmarshal([]byte(classCache), class)
		if err != nil {
			Logger.Error.Println("[DB]解析班级信息缓存失败:", err)
			return
		}
	}
	return class, err
}

func getAct(actID int) (act *dbAct, err error) {
	act = new(dbAct)
	actCache := rdb.Get(ctx, "SIGNIN_APP:Act:"+strconv.FormatInt(int64(actID), 10)).Val()
	if actCache == "" {
		Logger.Info.Println("[DB][act]回源读取信息:", err)
		//回源请求数据库，然后缓存
		act, err = cacheAct(actID)
		if err != nil {
			Logger.Error.Println("[DB][act]信息回源失败:", err)
			return
		}
	} else {
		err = json.Unmarshal([]byte(actCache), act)
		if err != nil {
			Logger.Error.Println("[DB][act]解析信息缓存失败:", err)
			return
		}
	}
	return act, err
}

func getClassStatistics(classID int) (sts *ResUserActStatistic, err error) {
	sts = new(ResUserActStatistic)
	stsCache := rdb.Get(ctx, "SIGNIN_APP:Class_Statistics::"+strconv.FormatInt(int64(classID), 10)).Val()
	if stsCache == "" {
		Logger.Info.Println("[DB][ClassStatistics]回源读取信息:", err)
		//回源请求数据库，然后缓存
		sts, err = cacheClassStatistics(classID)
		if err != nil {
			Logger.Error.Println("[DB][ClassStatistics]信息回源失败:", err)
			return
		}
	} else {
		err = json.Unmarshal([]byte(stsCache), sts)
		if err != nil {
			Logger.Error.Println("[DB][ClassStatistics]解析信息缓存失败:", err)
			return
		}
	}
	return sts, err
}
