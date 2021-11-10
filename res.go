package main

import (
	"github.com/gin-gonic/gin"
)

const (
	ContentTypeJSON = "application/json; charset=UTF-8"
	ContentTypeHTML = "text/html; charset=UTF-8"
)

type Res struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}

type ResEmpty struct {
	Res
	Data interface{} `json:"data"`
}

type ResUserInit struct {
	Res
	Data struct {
		NewJWT string `json:"new_jwt"`
	} `json:"data"`
}

type ResUserProfile struct {
	Res
	Data struct {
		UserId    int    `json:"user_id"`
		UserName  string `json:"user_name"`
		Email     string `json:"email"`
		ClassName string `json:"class_name"`
		ClassCode string `json:"class_code"`
		IsAdmin   int    `json:"is_admin"`
	} `json:"data"`
}

type ResUserActInfo struct {
	Res
	Data struct {
		Total int            `json:"total"`
		List  []*userActInfo `json:"list"`
	} `json:"data"`
}

type ResUserSignIn struct {
	Res
	Data struct {
		Text string `json:"text"`
	} `json:"data"`
}

type ResUserActStatistic struct {
	Res
	Data struct {
		Done           int                `json:"done"`
		Total          int                `json:"total"`
		UnfinishedList []actStatisticUser `json:"unfinished_list"`
		FinishedList   []actStatisticUser `json:"finished_list"`
	} `json:"data"`
}

type ActStatistic struct {
	Done           int                 `json:"done"`
	Total          int                 `json:"total"`
	UnfinishedList []*ActStatisticItem `json:"unfinished_list"`
	FinishedList   []*ActStatisticItem `json:"finished_list"`
}

type ResActLog struct {
	Res
	Data struct {
		Total int             `json:"total"`
		List  []resActLogItem `json:"list"`
	} `json:"data"`
}

type ResUserNotiGet struct {
	Res
	Data struct {
		NotiType string `json:"noti_type"`
	} `json:"data"`
}

type ResUserActQuery struct {
	Res
	Data struct {
		Name         string `json:"name"`
		Announcement string `json:"announcement"`
		Pic          string `json:"pic"`
		CheerText    string `json:"cheer_text"`
		BeginTime    string `json:"begin_time"`
		EndTime      string `json:"end_time"`
		CreateTime   string `json:"create_time"`
		UpdateTime   string `json:"update_time"`
		CreateBy     string `json:"create_by"`
	} `json:"data"`
}

type ResAdminActInfo struct {
	Res
	Data struct {
		ActID int `json:"act_id"`
		Name         string `json:"name"`
		Active bool `json:"active"`
		Announcement string `json:"announcement"`
		Pic          string `json:"pic"`
		CheerText    string `json:"cheer_text"`
		EndTime struct {
			D string `json:"d"`
			T string `json:"t"`
		} `json:"end_time"`
	} `json:"data"`
}

type ResAdminClassInfo struct {
	Res
	Data struct {
		ClassName string `json:"class_name"`
		ClassCode string `json:"class_code"`
		Total     int    `json:"total"`
	} `json:"data"`
}

type ResUserWechatQrcode struct {
	Res
	Data struct {
		Token     string `json:"token"`
		QrcodeUrl string `json:"qrcode_url"`
	} `json:"data"`
}

type ResAdminActList struct {
	Res
	Data struct {
		ActiveNum   int                 `json:"active_num"`
		ActiveList  []*adminActListItem `json:"active_list"`
		HistoryList []*adminActListItem `json:"history_list"`
	} `json:"data"`
}

type ResAdminActStatistic struct {
	Res
	Data struct {
		Done           int                      `json:"done"`
		Total          int                      `json:"total"`
		UnfinishedList []*AdminActStatisticItem `json:"unfinished_list"`
		FinishedList   []*AdminActStatisticItem `json:"finished_list"`
	} `json:"data"`
}

type resVersion struct {
	Res
	Data struct{
		Version string `json:"version"`
	}`json:"data"`
}

type ResCsrfToken struct {
	Res
	Data struct{
		Token string `json:"X-CSRF-TOKEN"`
	}`json:"data"`
}

//自定义数据结构
type actStatisticUser struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type ActStatisticItem struct {
	Id       int    `json:"id"`
	UserID   int    `json:"user_id"`
	ActID    int    `json:"act_id"`
	LogID    int    `json:"log_id"`
	Name     string `json:"name"`
	DateTime string `json:"date_time"`
}

type resActLogItem struct {
	Id       int    `json:"id"`
	ActToken string `json:"act_token"`
	ActName  string `json:"act_name"`
	DateTime string `json:"date_time"`
}

type userActInfo struct {
	ActToken        string `json:"act_token"`
	ActName         string `json:"act_name"`
	ActAnnouncement string `json:"act_announcement"`
	ActPic          string `json:"act_pic"`
	BeginTime       string `json:"begin_time"`
	EndTime         string `json:"end_time"`
	Status          int    `json:"status"`
	Statistic       struct {
		Done  int    `json:"done"`
		Total int    `json:"total"`
		Info  string `json:"info"`
	} `json:"statistic"`
}

type adminActListItem struct {
	Id        int    `json:"id"`
	ActID     int    `json:"act_id"`
	Name      string `json:"name"`
	BeginTime string `json:"begin_time"`
	EndTime   string `json:"end_time"`
	CreateBy  string `json:"create_by"`
}

type AdminActStatisticItem struct {
	ID       int    `json:"id"`
	UserId   int    `json:"user_id"`
	UserName string `json:"user_name"`
	DateTime string `json:"date_time"`
}

//函数
func errorRes(msg string) (r *ResEmpty) {
	r = new(ResEmpty)
	r.Status = -1
	r.Msg = msg
	return
}

func returnErrorJson(c *gin.Context, text string) {
	c.JSON(200, errorRes(text))
}

func returnErrorView(c *gin.Context,text string)  {
	c.Data(200,ContentTypeHTML,views("error1",map[string]string{"text":text}))
}