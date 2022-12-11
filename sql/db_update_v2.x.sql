# 密码
ALTER TABLE `user` ADD `password` VARCHAR(200) NOT NULL AFTER `email`;