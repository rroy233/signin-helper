--
-- activity追加字段
--
ALTER TABLE `activity` ADD `type` INT NOT NULL DEFAULT '0' COMMENT '签到操作类型' AFTER `active`;
ALTER TABLE `activity` ADD `daily_noti_enabled` INT NOT NULL DEFAULT '1' COMMENT '是否每日提醒' AFTER `pic`;


--
-- signin_log追加字段
--
ALTER TABLE `signin_log` ADD `file_path` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '文件地址' AFTER `create_time`;