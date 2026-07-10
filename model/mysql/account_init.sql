-- Drop database if exists
DROP DATABASE IF EXISTS `account_db`;

-- Create database
CREATE DATABASE IF NOT EXISTS `account_db` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

-- Use database
USE `account_db`;

DROP TABLE IF EXISTS `t_c_account`;
-- c账户表(用户)
CREATE TABLE `t_c_account` (
  `uid` BIGINT NOT NULL COMMENT '主键',
  `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
  `balance` BIGINT NOT NULL COMMENT '余额',
  `cur_type` SMALLINT NOT NULL COMMENT '货币类型',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`uid`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `t_c2c_transfer_order`;
-- c2c转账订单表
CREATE TABLE `t_c2c_transfer_order` (
  `transaction_id` VARCHAR(64) NOT NULL COMMENT '订单ID',
  `buyer_uid` BIGINT NOT NULL COMMENT '买家用户UID',
  `seller_uid` BIGINT NOT NULL COMMENT '卖家用户UID',
  `buyer_user_id` VARCHAR(64) NOT NULL COMMENT '买家用户ID',
  `seller_user_id` VARCHAR(64) NOT NULL COMMENT '卖家用户ID',
  `amount` BIGINT NOT NULL COMMENT '金额',
  `biz_type` INTEGER NOT NULL COMMENT '业务类型',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`transaction_id`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
DROP TABLE IF EXISTS `t_c_account_log`;
-- 用户流水日志表
CREATE TABLE `t_c_account_log` (
  `id` BIGINT AUTO_INCREMENT COMMENT '主键',
  `uid` BIGINT NOT NULL COMMENT '用户UID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
  `counterparty_user_id` VARCHAR(64) NOT NULL COMMENT '对方用户ID',
  `counterparty_uid` BIGINT NOT NULL COMMENT '对方用户UID',
  `transaction_id` VARCHAR(64) NOT NULL COMMENT '交易ID',
  `inout_type` TINYINT NOT NULL COMMENT '出入金类型',
  `biz_type` INTEGER NOT NULL COMMENT '业务类型',
  `amount` BIGINT NOT NULL COMMENT '金额',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uid_inout_type_transaction_id_biz_type` (`uid`,`inout_type`,`transaction_id`),
  INDEX `idx_transaction_id` (`transaction_id`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `t_b_account`;
-- b账户表(商户)
CREATE TABLE `t_b_account` (
  `uid` BIGINT NOT NULL COMMENT '主键',
  `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
  `balance` BIGINT NOT NULL COMMENT '余额',
  `cur_type` INTEGER NOT NULL COMMENT '货币类型',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`uid`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
-- linux:  mysql -h 127.0.0.1 -P 3307 -u root -p123456 < account_init.sql
-- windows: Get-Content -Encoding UTF8 account_init.sql | mysql -h 127.0.0.1 -P 3307 -u root -p123456
-- 只读权限 multipass exec master1 -- sudo kubectl exec -it -n pay-ns mysql-0 -- mysql -ustarslipay -ppayClipayA2026
-- root权限 multipass exec master1 -- sudo kubectl exec -it -n pay-ns mysql-0 -- mysql -uroot -proot123456
