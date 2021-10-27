# db



## user用户表



### user_id 主键

### name 真实姓名

 varchar(10)

### email 邮箱地址

 varchar(40)

### class 班级

 int - default(0)

### notification_type 推送类型

int 

### wx_pusher_uid 微信推送uid

varchar(100)

### is_admin 是否具有管理员权限

int - default(0)

### sso_uid 账号服务系统uid

int



## class班级表

### class_id 主键

### name 班级名称

varchar(10)

### class_code 班级代码

给用户选择班级

varchar(32)

### total 人数

int

###  act_id 当前活动

int



## activity 活动表

### act_id 主键

### class_id 班级id

int

### name 活动名称

varchar(40)

### announcement 公告

varchar(50)

### cheer_text 恭喜文本

varchar(20)

### pic头图

var(100)

### begin_time 开始时间

varchar(40)

### end_time 结束时间

varchar(40)

### create_time 创建时间

varchar(40)

### update_time 更新时间

varchar(40)

### create_by 创建人uid

int





## signin_log 记录表



### log_id 主键

### class_id 班级id

int

### act_id 活动id

int

### user_id 用户id

int

### create_time 创建时间

varchar(40)



## msg_template 消息模板

### tpl_id 主键

### msg_type 消息类型

int

### level 层次

int

### title 标题

varchar(20)

### body 正文

varchar(200)

### enabled 是否启用

int





