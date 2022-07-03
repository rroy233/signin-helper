package sso

import "encoding/json"

type UserInfo struct {
	UserName     string `json:"username,omitempty"`
	UserID       int    `json:"userid,omitempty"`
	Email        string `json:"email,omitempty"`
	UserGroup    string `json:"user_group,omitempty"`
	Avatar       string `json:"avatar,omitempty"`
	WechatOpenID string `json:"wechat_openID,omitempty"`
}

type httpResUserInfo struct {
	Status int      `json:"status"`
	Msg    string   `json:"msg"`
	Data   UserInfo `json:"data"`
}

func (ui UserInfo) String() string {
	res := new(httpResUserInfo)
	res.Status = 0
	res.Data = ui
	out, _ := json.Marshal(res)
	return string(out)
}

func DecodeHttpRes(data string) (UserInfo, error) {
	res := new(httpResUserInfo)
	err := json.Unmarshal([]byte(data), res)
	if err != nil {
		return UserInfo{}, err
	}
	return res.Data, nil
}
