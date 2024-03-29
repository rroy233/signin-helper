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
	Data struct{} `json:"data"`
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
		Total    int             `json:"total"`
		PagesNum int             `json:"pages_num"`
		List     []resActLogItem `json:"list"`
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
		ActID        int    `json:"act_id"`
		Name         string `json:"name"`
		Active       bool   `json:"active"`
		Announcement string `json:"announcement"`
		Pic          string `json:"pic"`
		DailyNotify  bool   `json:"daily_notify"`
		CheerText    string `json:"cheer_text"`
		NeedWait     bool   `json:"need_wait"`
		BeginTime    struct {
			D string `json:"d"`
			T string `json:"t"`
		} `json:"begin_time"`
		EndTime struct {
			D string `json:"d"`
			T string `json:"t"`
		} `json:"end_time"`
		Upload struct {
			Enabled bool     `json:"enabled"`
			Type    []string `json:"type"`
			MaxSize int      `json:"max_size"`
			Rename  bool     `json:"rename"`
		} `json:"upload"`
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
		WaitingNum  int                 `json:"waiting_num"`
		ActiveList  []*adminActListItem `json:"active_list"`
		WaitingList []*adminActListItem `json:"waiting_list"`
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
	Data struct {
		Version string `json:"version"`
	} `json:"data"`
}

type ResCsrfToken struct {
	Res
	Data struct {
		Token string `json:"X-CSRF-TOKEN"`
	} `json:"data"`
}

type ResUserNotiFetch struct {
	Res
	Data []*UserNotiFetchItem `json:"data"`
}

type ResAdminUserList struct {
	Res
	Data struct {
		Count int                 `json:"count"`
		Data  []AdminUserListItem `json:"data"`
	} `json:"data"`
}

type ResUserUpload struct {
	Res
	Data struct {
		UploadToken string `json:"upload_token"`
	} `json:"data"`
}

type ResAdminActExport struct {
	Res
	Data struct {
		DownloadUrl string `json:"download_url"`
	} `json:"data"`
}

type ResAdminActViewFile struct {
	Res
	Data struct {
		Type        string `json:"type"`
		ImgUrl      string `json:"img_url"`
		DownloadUrl string `json:"download_url"`
	} `json:"data"`
}

type ResAdminActGetRandomPic struct {
	Res
	Data struct {
		Url string `json:"url"`
	} `json:"data"`
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
	ActType         int    `json:"act_type"`
	NotiEnabled     int    `json:"noti_enabled"`
	BeginTime       string `json:"begin_time"`
	EndTime         string `json:"end_time"`
	Status          int    `json:"status"`
	Statistic       struct {
		Done  int    `json:"done"`
		Total int    `json:"total"`
		Info  string `json:"info"`
	} `json:"statistic"`
	FileOptions struct {
		AllowExt string `json:"allow_ext"`
		MaxSize  string `json:"max_size"`
		Note     string `json:"note"`
	} `json:"file_options"`
	Upload struct {
		Enabled     bool   `json:"enabled"`
		Type        string `json:"type"`
		ImgUrl      string `json:"img_url"`
		DownloadUrl string `json:"download_url"`
	} `json:"upload"`
}

type adminActListItem struct {
	Id        int    `json:"id"`
	ActID     int    `json:"act_id"`
	Type      int    `json:"type"`
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

type UserNotiFetchItem struct {
	Type     string   `json:"type"`
	NotiType NotiType `json:"noti_type"`
	Token    string   `json:"token"`
	Text     string   `json:"text"`
}

type AdminUserListItem struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	NotiType string `json:"noti_type"`
	Admin    int    `json:"admin"`
	Sign     string `json:"sign"`
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

func returnErrorView(c *gin.Context, text string) {
	c.Data(200, ContentTypeHTML, views("error1", map[string]string{"text": text}))
}

func returnErrorText(c *gin.Context, code int, text string) {
	c.Data(code, "text/plain; charset=UTF-8", []byte(text))
}
