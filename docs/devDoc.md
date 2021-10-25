# Back-End Dev Document

后端开发文档&备忘



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

### 2.3 /api/user/act/info 获取活动信息

* get
* resp->resUserActInfo
  * act_id
    * int
  * act_name 
    * string
  * act_announcement
    * string
  * act_pic
    * string
  * begin_time
    * string
    * 后端作格式化
  * end_time
    * string
    * 后端作格式化
  * status
    * int
    * 我的参与情况，0为“未参与”，1为“已参与”

### 2.4 /api/user/act/statistic 获取活动参与数据

* get
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
  * ts:时间戳(秒)
    * string
* resp -> resUserSignin
  * text

### 2.6 /api/user/act/log 我的参与记录

* get
* resp -> resActLog
  * total
    * int
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





## 3. 管理端API

### 3.1 /api/admin/act/info

获取当前的活动

* get
* resp -> resAdminActInfo
  * name
  * announcement
  * pic
  * cheer_text
  * begin_time
    * d
    * t
  * end_time
    * d
    * t

### 3.2 /api/admin/act/new

新建活动来覆盖掉旧活动

* post
* data -> formDataAdminAct json
  * name
  * announcement
  * pic
  * cheer_text
  * begin_time
    * d
    * t
  * end_time
    * d
    * t

### 3.3 /api/admin/act/edit

修改当前活动

* post
* data -> formDataAdminAct json
  * name
  * announcement
  * pic
  * cheer_text
  * begin_time
    * d
    * t
  * end_time
    * d
    * t

### 3.4 /api/admin/act/history

获取历史活动完成名单

* get
* resp -> ResAdminActHistory
  * id
    * 顺序id
  * user_id
  * user_name
  * act_id
  * act_name
  * act_sort_id
    * 活动内排名
  * datetime

### 3.5 /api/admin/class/info

获取班级信息

* get
* resp -> resAdminClassInfo
  * class_name
  * class_code
  * total
  * act_id
  * act_name

### 3.6 /api/admin/class/edit

* post
* data -> formDataAdminClass json
  * class_name
  * class_code





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

[key] SIGNIN_APP:JWT:{ JWT ID }

[val] user_id

[exp] 1h

### 5.2 sso 存储state

[key]  SIGNIN_APP:state:{state}

[val] 1

[exp] 5min

### 5.3 班级信息缓存

[key] SIGNIN_APP:Class:{CLASS_ID}

[val] dbClass{}

[exp] 5 min

**回源，用于读取班级的基本信息，管理员更新是手动刷新缓存**

### 5.4 活动信息缓存

[key] SIGNIN_APP:Act:{ACT_ID}

[val] dbAct{}

[exp] 30 s

**回源，用于读取活动的基本信息，管理员更新是手动刷新缓存**

### 5.5 班级统计结果缓存

[key] SIGNIN_APP:Class_Statistics:{CLASS_ID}

[val] resUserActStatistic{}

[exp] 10 s

**回源，用于读取活动的基本信息，管理员更新是手动刷新缓存**

## 6. 自定义数据结构

### 6.1 actStatisticUser

用于 #2.3`/user/act/statistic `列举用户列表

* id
  * 顺序id
* name
* avatar
  * 头像

### 6.2 resActLogItem

用于 #2.6`/user/act/log `列举参与列表

* id
  * 顺序id
* act_id
* act_name
* datetime
  * string
  * 后端格式化



## 7. 用户群组ID

普通用户:8

管理员:7
