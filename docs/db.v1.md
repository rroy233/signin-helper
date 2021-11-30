# db.v1

## v0与v1对比

* 删除class.act_id

  ```sql
  ALTER TABLE `class` DROP `act_id`;
  ```

* 新增activity.active

  ```sql
  ALTER TABLE `activity` ADD `active` INT NOT NULL DEFAULT '0' AFTER `class_id`;
  ```

  



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



## activity 活动表

### act_id 主键

### class_id 班级id

int

### active是否有效

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



## file 文件记录

### file_id 主键
### status 文件状态

* -1:文件(远端)已删除
* 0:暂存在本地
* 1:保存在远端

### user_id

### act_id

### file_name

文件名称(不含后缀)

### content_type

文件类型

### local

本地地址

`/storage/temp/xxx`

### remote

cos对象存储的key

`/upload/xxx`

### exp_time

### upload_time



