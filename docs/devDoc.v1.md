# 	Back-End Dev Document

v1.x版本

## 1. 通用API

### 1.1 /api/ssoCallback 单点登录回调接口

* get

* param
  * access_token
  * state

### 1.2 /api/login 登录凭证登录接口

* get

* param
  * jwt

### 1.3 /api/logout 登出

* post

### 1.4 /user/version 后端版本

* get
* resp -> resVersion
  * version





## 2. 用户端API

### 2.1 /api/user/init 初始化个人信息

用于用户第一次使用SSO进入系统，尚未选择班级和填写基本信息

初次设置后展示不予修改

* post

* data-json

  * calss_code
    * int
  * name
    * string

* resp->resUserInit

  * new_jwt 新分配一个jwt

    

### 2.2 /api/user/profile 获取个人信息

* get
* resp->resUserProfile
  * user_id
    * int
  * user_name
    * string
  * email
    * string
  * class_name
    * string
  * class_code
    * string
  * is_admin
    * int

### ~~2.3 /api/user/act/info 获取活动信息(v0)~~

* ~~get~~
* ~~resp->resUserActInfo~~
  *  ~~act_id act_token~~
    * ~~int string~~
  * ~~act_name~~ 
    * ~~string~~
  * ~~act_announcement~~
    * ~~string~~
  * ~~act_pic~~
    * ~~string~~
  * ~~begin_time~~
    * ~~string~~
    * ~~后端作格式化~~
  * ~~end_time~~
    * ~~string~~
    * ~~后端作格式化~~
  * ~~status~~
    * ~~int~~
    * ~~我的参与情况，0为“未参与”，1为“已参与”~~

### 2.3 /api/user/act/info 获取活动信息(v1)

* get
* resp->resUserActInfo
  *  total
  *  list
     *  []userActInfo



### 2.4 /api/user/act/statistic 获取活动参与数据

* get
* params:
  * act_token
* resp -> resUserActStatistic
  * done
    * int
    * 参与人数
  * total
    * int
    * 总人数
  * unfinished_list
    * []actStatisticUser
  * finished_list
    * []actStatisticUser



### 2.5 /api/user/act/signin 签到操作

* post
* data - json
  * act_token
    * string
  * upload_token
    * string
* resp -> resUserSignin
  * text

### 2.5.1 /api/user/act/upload 签到操作

* post
* data - json
  * act_token
    * string
* resp -> resUserUpload
  * upload_token

### 2.6 /api/user/act/log 我的参与记录

* get
* params -> formUserActLog
  * page

* resp -> resActLog
  * total
    * 总数
  * pages_num
    * 页数总数
  * list -> []resActLogItem

### 2.7 /api/user/act/query 搜索活动详情

* get
* param
  * act_id
    * int
* resp -> resUserActQuery
  * name
  * announcement
  * cheer_text
  * pic
  * begin_time
  * end_time
  * create_time
  * update_time
  * create_by
    * string
    * 姓名

### 2.8 /api/user/noti/get 我的通知方式

* get
* resp -> resUserNotiGet
  * noti_type
    * string

### 2.9 /api/user/noti/edit 修改我的通知方式

* post
* data -> formDataUserNotiEdit
  * noti_type
    * string

### 2.10 /api/user/wechat/qrcode 获取微信绑定二维码

* get
* resp -> resUserWechatQrcode
  * token
  * qrcode_url

**（请求此接口可以刷新Extra）**

### 2.11 /api/user/wechat/qrcode/bind 微信绑定轮询接口

* get

* data

  * token

* resp -> empty

  业务码(status)为0，即正在等待回调数据。

  业务码为-1，即绑定失败（超时）或参数无效。

  业务码为1，即绑定成功。



### 2.12 /api/user/csrfToken 获取csrfToken

* GET

存cookie



### 2.13 /api/user/noti/check 信息已读

* post
* data -> formDataNotiCheck
  * token
* resp -> empty

### 2.14 /api/user/noti/fetch 拉取顶置信息

* get
* resp -> resUserNotiFetch
  * status
  * msg
  * data
    * []UserNotiFetchItem

### 2.15 /api/user/act/cancel 取消签到

* post
* data - json
  * act_token
    * string



## 3. 管理端API

### 3.1 /api/admin/act/info 返回单个活动信息

获取当前的活动

* get
* params
  * act_id
* resp -> resAdminActInfo
  * act_id
  * name
  * active
    * bool
  * announcement
  * pic
  * cheer_text
  * end_time
    * d
    * t
  * upload
    * enabled
      * bool
    * type
      * []string
    * max_size
      * int
    * rename
      * bool

### 3.2 /api/admin/act/new 新建活动

* post
* data -> FormDataAdminAct json
  * name
  * announcement
  * pic
  * daily_notify
  * upload
    * enabled
      * bool
    * type
      * []string
    * max_size
      * int
    * rename
      * bool
  * cheer_text
  * end_time
    * d
    * t

### 3.3 /api/admin/act/edit 编辑活动

修改指定活动

* post
* data
  * act_id
* data -> FormDataAdminAct json
  * name
  * active
    * bool
  * announcement
  * pic
  * daily_notify
  * upload
    * enabled
      * bool
    * type
      * []string
    * max_size
      * int
    * rename
      * bool
  * cheer_text
  * end_time
    * d
    * t

### 3.4 /api/admin/act/statistic 某次活动的统计数据

获取历史活动完成名单

* get
* params
  * act_id
* resp -> ResAdminActStatistic
  * done
    * int
    * 参与人数
  * total
    * int
    * 总人数
  * unfinished_list
    * []AdminActStatisticItem
  * finished_list
    * []AdminActStatisticItem

### 3.5 /api/admin/class/info 获取班级信息

获取班级信息

* get
* resp -> resAdminClassInfo
  * class_name
  * class_code
  * total
  * act_id
  * act_name

### 3.6 /api/admin/class/edit 编辑班级信息

* post
* data -> formDataAdminClass json
  * class_name
  * class_code

### 3.7 /api/admin/act/list 活动列表

* get
* resp -> resAdminActList
  * active_num
  * active_list
    * []adminActListItem
  * history_list
    * []adminActListItem

### 3.8 /api/admin/csrfToken 获取csrfToken

* GET

存cookie

### 3.9 /api/admin/user/list 用户列表

* get
* resp -> resAdminUserList
  * count
  * data []adminUserListItem

### 3.10 /api/admin/user/del 删除用户

* post
* data: formDataAdminUserDel
  * user_id
  * sign

### 3.10.1 /api/admin/user/setAdmin 设置admin权限

* post
* data: formDataAdminUserSetAdmin
  * user_id
  * set_to
  * sign

### 3.11 /api/admin/act/export 导出所有文件

* post
* data -> formDataAdminActExport
  * act_id
* resp -> resAdminActExport
  * download_url

### 3.11 /api/admin/act/viewFile 查看用户文件

* post
* data -> formDataAdminViewFile
  * user_id
* resp -> resAdminActViewFile
  * type
    * string
  * img_url
  * download_url

### 3.12 /api/admin/act/getRandomPic 管理员获取随机图片

* pic
* resp -> adminActGetRandomPic
  * url



## 4. JWT 结构

### 4.1 header

```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

### 4.2 playload

```json
{
  "user_id":1,
  "name":"张三",
  "email":"abc@example.com",
  "class_id":1,//班级id，未分配时为-1
  "is_admin":0,//是否为管理员
  "exp":1516239022,//过期时间，1小时有效期
  "jti":"",//JWT ID，随机生成，存储在redis中，便于吊销
}
```

### 4.3 secret

保存在config.yaml



## 5. Redis-key-list

### 5.1 JWT存根

[key] SIGNIN_APP:JWT:USER_{UID}:{ JWT ID }

[val] user_id

[exp] 1h

### 5.2 sso state 存储state

[key]  SIGNIN_APP:state:{state}

[val] 1

[exp] 5min

### 5.3 Class班级信息缓存

[key] SIGNIN_APP:Class:{CLASS_ID}

[val] dbClass{}

[exp] 5 min

**回源，用于读取班级的基本信息，管理员更新是手动刷新缓存**

### 5.4 Act活动信息缓存

[key] SIGNIN_APP:Act:{ACT_ID}

[val] dbAct{}

[exp] 30 s

**回源，用于读取活动的基本信息，管理员更新是手动刷新缓存**

### ~~5.5 Class_Statistics班级统计结果缓存~~

~~[key] SIGNIN_APP:Class_Statistics:{CLASS_ID}~~

~~[val] resUserActStatistic{}~~

~~[exp] 10 s~~

~~回源，用于读取活动的基本信息，管理员更新是手动刷新缓存~~

### 5.6 NOTI_LIST每日提醒消息列表

[key] SIGNIN_APP:NOTI_LIST

[val] dailyNotifyJob{}

[exp] -1

### 5.7 Wechat_Bind微信绑定Extra=>user_id

[key] SIGNIN_APP:Wechat_Bind:{{Extra}}

[val] user_id (或为DONE，即已完成绑定)

[exp] 30min

### 5.8 Wechat_Bind微信绑定user_id=>Extra

[key] SIGNIN_APP:Wechat_Bind:{{user_id}}

[val] Extra (或为DONE，即已完成绑定)

[exp] 30min

### 5.9 actToken 用户活动查询凭证

[key] SIGNIN_APP:actToken:{{token}}

[val] act_id

[exp] 10min

### 5.10 Class_Active_Act当前班级正在生效的活动

[key] SIGNIN_APP:Class_Active_Act:{{CLASS_ID}}

[val]  CacheIDS

[exp] 1min

* 距离结束时间>1min
  * easy
* 距离结束时间<1min
  * careful

### 5.11 Act_Statistic活动统计结果缓存

[key] SIGNIN_APP:Act_Statistic:{ACT_ID}

[val] resUserActStatistic{}

[exp] 10 s

**回源，用于读取活动的基本信息，管理员更新是手动刷新缓存**

### 5.12 scrfToken活动统计结果缓存

[key] SIGNIN_APP:csrfToken:{{JWT_ID}}

[val] csrfToken

[exp] 1h

### 5.13 UserNoti 用户消息列表

[key] SIGNIN_APP:UserNoti:USER_{{USER_ID}}:{{noti_token}}

[val] UserNotiFetchItem.json

[exp] -1

### 5.14 ActNotiTimes 此活动提醒次数

[key] SIGNIN_APP:ActNotiTimes:{{act_id}}:{{UserID}}

[val] (int)

[exp] -1

**在活动(手动/自动)结束时结算**

### 5.15 UserSignUpload 上传文件记录

[key] SIGNIN_APP:UserSignUpload:{{upload_token}}

[val] file_id

[exp] 2min

### 5.16 TempFile 临时文件

[key] SIGNIN_APP:TempFile:{{tempFile_token}}

[val] base64

[exp] -

### 5.17 UrlToken 短网址

[key] SIGNIN_APP:UrlToken:{{url_token}}

[val] md5_short

[exp] va

## 6. 自定义数据结构

### 6.1 actStatisticUser

用于 #2.4`/user/act/statistic `列举用户列表

* id
  * 顺序id
* name

### 6.2 resActLogItem

用于 #2.6`/user/act/log `列举参与列表

* id
  * 顺序id
* act_id
* act_name
* datetime
  * string
  * 后端格式化

### 6.3 微信扫码关注回调

```go
type WxPusherCallback struct {
	Action string `json:"action"`
	Data   struct {
		AppID       int    `json:"appId"`
		AppKey      string `json:"appKey"`
		AppName     string `json:"appName"`
		Source      string `json:"source"`
		UserName    string `json:"userName"`
		UserHeadImg string `json:"userHeadImg"`
		Time        int64  `json:"time"`
		UID         string `json:"uid"`
		Extra       string `json:"extra"`
	} `json:"data"`
}
```

### 6.4 userActInfo

#2.3.1 用户首页获取活动列表

*  act_token
   * string
*  act_name 
   * string
*  act_announcement
   * string
*  act_type
   *  int
*  noti_enabled
   *  int

*  act_pic
   * string
*  begin_time
   * string
   * 后端作格式化
*  end_time
   * string
   * 后端作格式化
*  status
   * int
   * 我的参与情况，0为“未参与”，1为“已参与”
*  statistic
   * done
   * total
   * info
*  file_options
   *  allow_ext
      *  string

   *  max_size
      *  string

   *  note
      *  string
*  upload
   * enabled
     * bool
   * type
     * string
   * img_url
   * download_url


### 6.11 CacheIDS 缓存活动信息

* total
  * int
* easy
  * 直接采用redis的值
  * []int
* careful
  * get时还需要查询一下mysql
  * []int

### 6.12 adminActListItem管理员获取活动列表

* id
  * 顺序id
* act_id
* type
* name
* begin_time
  * string
  * 后端作格式化
* end_time
  * string
  * 后端作格式化
* create_by
  * string
  * 姓名

### 6.13 AdminActStatisticItem 管理员-活动数据

* id
  * 顺序id(按照完成时间先后排列)
* user_id
* user_name
* datetime
  * 完成时间

### 6.14 UserNotiFetchItem 用户顶置信息

* type
* token
* text

**token生成算法:**

活动相关：md5_short(userid+actid+act_noti_type)

其他：md5_short(userid+unixnano+rand)

### 6.15 adminUserListItem 管理员用户列表

* id
* user_id
* name
* email
* noti_type
* admin
  * int
* sign
  * 签名

### 6.16 file_options 文件上传选项

* allow_content_type
  * string
* max_size
  * int
* rename
  * bool



## 7. 用户群组ID

普通用户:其他

管理员:7
