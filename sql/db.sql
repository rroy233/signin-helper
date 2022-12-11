SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


-- --------------------------------------------------------

--
-- 表的结构 `activity`
--

CREATE TABLE `activity` (
  `act_id` int(11) NOT NULL,
  `class_id` int(11) NOT NULL,
  `active` int(11) NOT NULL DEFAULT '0',
  `type` int(11) NOT NULL DEFAULT '0' COMMENT '签到操作类型',
  `name` varchar(40) NOT NULL,
  `announcement` varchar(50) NOT NULL,
  `cheer_text` varchar(20) NOT NULL DEFAULT '签到成功',
  `pic` varchar(500) NOT NULL DEFAULT '',
  `daily_noti_enabled` int(11) NOT NULL DEFAULT '1' COMMENT '是否每日提醒',
  `begin_time` varchar(40) NOT NULL DEFAULT '0',
  `end_time` varchar(40) NOT NULL DEFAULT '0',
  `create_time` varchar(40) NOT NULL DEFAULT '0',
  `update_time` varchar(40) NOT NULL DEFAULT '0',
  `create_by` int(11) NOT NULL,
  `file_opts` varchar(1000) NOT NULL DEFAULT '' COMMENT '文件选项'
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
-- 表的结构 `file`
--

CREATE TABLE `file` (
  `file_id` int(11) NOT NULL COMMENT '主键',
  `status` int(11) NOT NULL COMMENT '文件状态',
  `user_id` int(11) NOT NULL,
  `act_id` int(11) NOT NULL,
  `file_name` varchar(100) NOT NULL COMMENT '不含后缀文件名称',
  `content_type` varchar(100) NOT NULL,
  `local` varchar(100) NOT NULL,
  `remote` varchar(100) NOT NULL,
  `exp_time` varchar(20) NOT NULL,
  `upload_time` varchar(20) NOT NULL
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

-- --------------------------------------------------------

--
-- 表的结构 `signin_log`
--

CREATE TABLE `signin_log` (
  `log_id` int(11) NOT NULL,
  `class_id` int(11) NOT NULL,
  `act_id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `create_time` varchar(40) NOT NULL DEFAULT '0',
  `file_id` int(11) NOT NULL DEFAULT '-1'
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
-- 表的索引 `file`
--
ALTER TABLE `file`
  ADD PRIMARY KEY (`file_id`),
  ADD KEY `user_id` (`user_id`),
  ADD KEY `act_id` (`act_id`);

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
-- 使用表AUTO_INCREMENT `file`
--
ALTER TABLE `file`
  MODIFY `file_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键';

--
-- 使用表AUTO_INCREMENT `msg_template`
--
ALTER TABLE `msg_template`
  MODIFY `tpl_id` int(11) NOT NULL AUTO_INCREMENT;

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

