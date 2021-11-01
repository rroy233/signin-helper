-- phpMyAdmin SQL Dump
-- version 5.0.4
-- https://www.phpmyadmin.net/
--
-- 主机： localhost
-- 生成日期： 2021-11-01 20:18:50
-- 服务器版本： 5.6.50-log
-- PHP 版本： 7.4.16

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- 数据库： `sign_in_app`
--

-- --------------------------------------------------------

--
-- 表的结构 `activity`
--

CREATE TABLE `activity` (
  `act_id` int(11) NOT NULL,
  `class_id` int(11) NOT NULL,
  `active` int(11) NOT NULL DEFAULT '0',
  `name` varchar(40) NOT NULL,
  `announcement` varchar(50) NOT NULL,
  `cheer_text` varchar(20) NOT NULL DEFAULT '签到成功',
  `pic` varchar(100) NOT NULL DEFAULT '',
  `begin_time` varchar(40) NOT NULL DEFAULT '0',
  `end_time` varchar(40) NOT NULL DEFAULT '0',
  `create_time` varchar(40) NOT NULL DEFAULT '0',
  `update_time` varchar(40) NOT NULL DEFAULT '0',
  `create_by` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 表的结构 `class`
--

CREATE TABLE `class` (
  `class_id` int(11) NOT NULL,
  `name` varchar(10) NOT NULL,
  `class_code` varchar(32) NOT NULL,
  `total` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 表的结构 `msg_template`
--

CREATE TABLE `msg_template` (
  `tpl_id` int(11) NOT NULL,
  `msg_type` int(11) NOT NULL,
  `level` int(11) NOT NULL DEFAULT '0',
  `title` varchar(200) CHARACTER SET utf8 NOT NULL,
  `body` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL,
  `enabled` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- 转存表中的数据 `msg_template`
--

INSERT INTO `msg_template` (`tpl_id`, `msg_type`, `level`, `title`, `body`, `enabled`) VALUES
(1, 0, 0, '<新打卡任务>「{{act_name}}」开启啦！', '{{username}}您好:{{EOL}}{{space}}{{space}}您有一个新的打卡任务哦，截止日期{{act_end_time}}，快快点击下方的链接签到吧~{{EOL}}{{EOL}}{{space}}{{space}}{{login_url_withToken}}', 1),
(2, 1, 0, '<打卡提醒>「{{act_name}}」', '阿伟你又在玩电动吼，休息一下吧，签个到好不好？{{EOL}}{{space}}{{space}}{{login_url_withToken}}', 1),
(3, 1, 0, '<打卡提醒>「{{act_name}}」', '阿伟你又在用功读书吼，休息一下吧，签个到好不好？{{EOL}}{{space}}{{space}}{{login_url_withToken}}', 1),
(4, 1, 0, '<打卡提醒>「{{act_name}}」', 'Olah muhe {{username}}，yo aba unta 签到，mosi mita！{{EOL}}{{space}}{{space}}{{login_url_withToken}}', 1),
(5, 1, 0, '<打卡提醒>「{{act_name}}」', '封印黑暗力量的超维之杖啊{{EOL}}在此显现你真正的力量吧{{EOL}}奇迹之力化为星尘 光之魔法汇聚吾心{{EOL}}Dimension{{space}}Sonata{{EOL}}{{EOL}}尊敬的{{username}}，您再不签到，我就把你TeRiTeRi掉哟～{{EOL}}{{space}}{{space}}{{login_url_withToken}}', 1),
(6, 1, 0, '<打卡提醒>「{{act_name}}」', '呐呐呐，{{username}}欧尼酱(桑)~~{{EOL}}诶多捏诶多捏，瓦塔西就是那个，签到得斯~~{{EOL}}阿里嘎多欧尼酱！！！呆siki了！！{{EOL}}{{space}}{{space}}{{login_url_withToken}}', 1),
(7, 1, 0, '<打卡提醒>「{{act_name}}」', '( ﹁ ﹁ ) ~→{{login_url_withToken}}', 1),
(8, 1, 0, '<打卡提醒>「{{act_name}}」', '✧(≖ ◡ ≖✿) ~→{{login_url_withToken}}', 1),
(9, 1, 0, '<打卡提醒>「{{act_name}}」', '球球你快点签到吧{{EOL}}(/◕ヮ◕)/ ~→{{login_url_withToken}}', 1),
(10, 1, 0, '<打卡提醒>「{{act_name}}」', '嘿，我亲爱的{{username}}，瞧瞧我发现了什么，您今天还有签到任务没完成喔，真是见鬼，隔壁的汤姆叔叔已经等不及了，哦，{{username}}，愿上帝保佑你，我的老伙计。{{EOL}}{{EOL}}{{space}}{{space}}{{login_url_withToken}}', 1),
(11, 1, 0, '<打卡提醒>「{{act_name}}」', '无内鬼。{{EOL}}{{login_url_withToken}}', 1),
(12, 1, 0, '<打卡提醒>「{{act_name}}」', 'S{{space}}T{{space}}O{{space}}P{{space}}✋🏻{{EOL}}很抱歉打扰你{{EOL}}但是我想提醒您一句{{EOL}}{{EOL}}你签到了吗{{EOL}}{{login_url_withToken}}', 1);

-- --------------------------------------------------------

--
-- 表的结构 `signin_log`
--

CREATE TABLE `signin_log` (
  `log_id` int(11) NOT NULL,
  `class_id` int(11) NOT NULL,
  `act_id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `create_time` varchar(40) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 表的结构 `user`
--

CREATE TABLE `user` (
  `user_id` int(11) NOT NULL,
  `name` varchar(10) NOT NULL,
  `email` varchar(40) NOT NULL,
  `class` int(11) NOT NULL DEFAULT '0',
  `notification_type` int(11) NOT NULL DEFAULT '0',
  `wx_pusher_uid` varchar(100) DEFAULT NULL,
  `is_admin` int(11) NOT NULL DEFAULT '0',
  `sso_uid` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- 转储表的索引
--

--
-- 表的索引 `activity`
--
ALTER TABLE `activity`
  ADD PRIMARY KEY (`act_id`);

--
-- 表的索引 `class`
--
ALTER TABLE `class`
  ADD PRIMARY KEY (`class_id`),
  ADD UNIQUE KEY `class_code` (`class_code`);

--
-- 表的索引 `msg_template`
--
ALTER TABLE `msg_template`
  ADD PRIMARY KEY (`tpl_id`);

--
-- 表的索引 `signin_log`
--
ALTER TABLE `signin_log`
  ADD PRIMARY KEY (`log_id`),
  ADD KEY `user_id` (`user_id`);

--
-- 表的索引 `user`
--
ALTER TABLE `user`
  ADD PRIMARY KEY (`user_id`),
  ADD UNIQUE KEY `sso_uid` (`sso_uid`),
  ADD KEY `uder_id` (`user_id`);

--
-- 在导出的表使用AUTO_INCREMENT
--

--
-- 使用表AUTO_INCREMENT `activity`
--
ALTER TABLE `activity`
  MODIFY `act_id` int(11) NOT NULL AUTO_INCREMENT;

--
-- 使用表AUTO_INCREMENT `class`
--
ALTER TABLE `class`
  MODIFY `class_id` int(11) NOT NULL AUTO_INCREMENT;

--
-- 使用表AUTO_INCREMENT `msg_template`
--
ALTER TABLE `msg_template`
  MODIFY `tpl_id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=13;

--
-- 使用表AUTO_INCREMENT `signin_log`
--
ALTER TABLE `signin_log`
  MODIFY `log_id` int(11) NOT NULL AUTO_INCREMENT;

--
-- 使用表AUTO_INCREMENT `user`
--
ALTER TABLE `user`
  MODIFY `user_id` int(11) NOT NULL AUTO_INCREMENT;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
