SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";

--
-- activity追加字段
--
ALTER TABLE `activity` ADD `type` INT NOT NULL DEFAULT '0' COMMENT '签到操作类型' AFTER `active`;
ALTER TABLE `activity` ADD `daily_noti_enabled` INT NOT NULL DEFAULT '1' COMMENT '是否每日提醒' AFTER `pic`;
ALTER TABLE `activity` ADD `file_opts` VARCHAR(400) NOT NULL DEFAULT '' COMMENT '文件选项' AFTER `create_by`;

--
-- signin_log追加字段
--
ALTER TABLE `signin_log` ADD `file_id` INT NOT NULL DEFAULT '-1' AFTER `create_time`;

--
-- 新建数据表file
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
ALTER TABLE `file`
    ADD PRIMARY KEY (`file_id`),
    ADD KEY `user_id` (`user_id`),
    ADD KEY `act_id` (`act_id`);
ALTER TABLE `file`
    MODIFY `file_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键';
COMMIT;