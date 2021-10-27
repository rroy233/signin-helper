package main

import "github.com/gin-gonic/gin"

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
		ActID           int    `json:"act_id"`
		ActName         string `json:"act_name"`
		ActAnnouncement string `json:"act_announcement"`
		ActPic          string `json:"act_pic"`
		BeginTime       string `json:"begin_time"`
		EndTime         string `json:"end_time"`
		Status          int    `json:"status"`
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
		Name         string `json:"name"`
		Announcement string `json:"announcement"`
		Pic          string `json:"pic"`
		CheerText    string `json:"cheer_text"`
		BeginTime    struct {
			D string `json:"d"`
			T string `json:"t"`
		} `json:"begin_time"`
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
		ActId     int    `json:"act_id"`
		ActName   string `json:"act_name"`
	} `json:"data"`
}

//自定义数据结构
type actStatisticUser struct {
	Id     int    `json:"id"`
	UserID int `json:"user_id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type resActLogItem struct {
	Id       int    `json:"id"`
	ActId    int    `json:"act_id"`
	ActName  string `json:"act_name"`
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


