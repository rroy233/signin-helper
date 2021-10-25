SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

--
-- Database: `sign_in`
--

-- --------------------------------------------------------

--
-- Table structure for table `activity`
--

CREATE TABLE `activity` (
  `act_id` int(11) NOT NULL,
  `class_id` int(11) NOT NULL,
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
-- Table structure for table `class`
--

CREATE TABLE `class` (
  `class_id` int(11) NOT NULL,
  `name` varchar(10) NOT NULL,
  `class_code` varchar(32) NOT NULL,
  `total` int(11) NOT NULL,
  `act_id` int(11) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Table structure for table `msg_template`
--

CREATE TABLE `msg_template` (
  `tpl_id` int(11) NOT NULL,
  `msg_type` int(11) NOT NULL,
  `title` varchar(20) NOT NULL,
  `body` varchar(200) NOT NULL,
  `enabled` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Table structure for table `signin_log`
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
-- Table structure for table `user`
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
-- Indexes for dumped tables
--

--
-- Indexes for table `activity`
--
ALTER TABLE `activity`
  ADD PRIMARY KEY (`act_id`);

--
-- Indexes for table `class`
--
ALTER TABLE `class`
  ADD PRIMARY KEY (`class_id`),
  ADD UNIQUE KEY `class_code` (`class_code`);

--
-- Indexes for table `msg_template`
--
ALTER TABLE `msg_template`
  ADD PRIMARY KEY (`tpl_id`);

--
-- Indexes for table `signin_log`
--
ALTER TABLE `signin_log`
  ADD PRIMARY KEY (`log_id`),
  ADD KEY `user_id` (`user_id`);

--
-- Indexes for table `user`
--
ALTER TABLE `user`
  ADD PRIMARY KEY (`user_id`),
  ADD UNIQUE KEY `sso_uid` (`sso_uid`),
  ADD KEY `uder_id` (`user_id`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `activity`
--
ALTER TABLE `activity`
  MODIFY `act_id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `class`
--
ALTER TABLE `class`
  MODIFY `class_id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `msg_template`
--
ALTER TABLE `msg_template`
  MODIFY `tpl_id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `signin_log`
--
ALTER TABLE `signin_log`
  MODIFY `log_id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `user`
--
ALTER TABLE `user`
  MODIFY `user_id` int(11) NOT NULL AUTO_INCREMENT;
