# 消息模板

实例

标题：`<新打卡任务>「{{act_name}}」开启啦！`

模板文本：`{{username}}您好:{{EOL}}{{space}}{{space}}您有一个新的打卡任务哦，截止日期{{act_end_time}}，快快点击下方的链接签到吧~{{EOL}}{{space}}{{space}}{{login_url_withToken}}`

渲染结果：

```
<新打卡任务>「默认活动」开启啦！
```

```
小明您好:
   您有一个新的打卡任务哦，截止日期2021-10-26 10:10:00，快快点击下方的链接签到吧~
   https://qd.roy233.com/api/loginByToken?jwt=xxxx
```



## 格式

换行符：`{{EOL}}`

空格：`{{space}}`

当前时间：`{{time_now}}`

## 用户信息

用户名：`{{username}}`



## 活动信息

活动名称：`{{act_name}}`

活动开始时间：`{{act_begin_time}}`

活动结束时间：`{{act_end_time}}`

创建人：`{{act_creator}}`



## 班级信息

班级名称：`{{class_name}}`



## 免登录入口

用户点击邮箱中的链接即可直接完成登录，进入签到主页

入口：`{{login_url_withToken}}`
