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
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`uid`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `t_c_account_log`;
-- 用户流水日志表
CREATE TABLE `t_c_account_log` (
  `id` BIGINT AUTO_INCREMENT COMMENT '主键',
  `uid` BIGINT NOT NULL COMMENT '用户UID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
  `counterparty_id` VARCHAR(64) NOT NULL COMMENT '对方ID',
  `counterparty_uid` BIGINT NOT NULL COMMENT '对方UID',
  `transaction_id` VARCHAR(64) NOT NULL COMMENT '交易ID',
  `inout_type` TINYINT NOT NULL COMMENT '出入金类型',
  `biz_type` INTEGER NOT NULL COMMENT '业务类型',
  `balance` BIGINT NOT NULL COMMENT '余额',
  `amount` BIGINT NOT NULL COMMENT '金额',
  `desc` VARCHAR(256) NOT NULL COMMENT '描述',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uid_inout_type_transaction_id_biz_type` (`uid`,`inout_type`,`transaction_id`),
  INDEX `idx_transaction_id` (`transaction_id`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `t_b_account`;
CREATE TABLE `t_b_account` (
  `merchant_uid` BIGINT NOT NULL COMMENT '商户UID',
  `merchant_id` VARCHAR(64) NOT NULL COMMENT '商户ID',
  `balance` BIGINT NOT NULL COMMENT '余额',
  `cur_type` INTEGER NOT NULL COMMENT '货币类型',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`merchant_uid`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `t_b_account` (`merchant_uid`, `merchant_id`, `balance`, `cur_type`, `create_time`, `update_time`)
VALUES (2000000000, '2000000000', 0, 1, NOW(), NOW());


DROP TABLE IF EXISTS `t_b_account_log`;
-- 商户账户流水日志表
CREATE TABLE `t_b_account_log` (
  `id` BIGINT AUTO_INCREMENT COMMENT '主键',
  `merchant_uid` BIGINT NOT NULL COMMENT '商户UID',
  `merchant_id` VARCHAR(64) NOT NULL COMMENT '商户ID',
  `counterparty_id` VARCHAR(64) NOT NULL COMMENT '对方ID',
  `counterparty_uid` BIGINT NOT NULL COMMENT '对方UID',
  `transaction_id` VARCHAR(64) NOT NULL COMMENT '交易ID',
  `inout_type` TINYINT NOT NULL COMMENT '出入金类型',
  `biz_type` INTEGER NOT NULL COMMENT '业务类型',
  `balance` BIGINT NOT NULL COMMENT '余额',
  `amount` BIGINT NOT NULL COMMENT '金额',
  `desc` VARCHAR(256) NOT NULL COMMENT '描述',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_merchant_uid_inout_type_transaction_id_biz_type` (`merchant_uid`,`inout_type`,`transaction_id`),
  INDEX `idx_transaction_id` (`transaction_id`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `t_c2c_bill`;
CREATE TABLE `t_c2c_bill` (
  `transaction_id` VARCHAR(64) NOT NULL COMMENT '交易ID',
  `buyer_uid` BIGINT NOT NULL COMMENT '买家用户UID',
  `seller_uid` BIGINT NOT NULL COMMENT '卖家用户UID',
  `buyer_user_id` VARCHAR(64) NOT NULL COMMENT '买家用户ID',
  `seller_user_id` VARCHAR(64) NOT NULL COMMENT '卖家用户ID',
  `amount` BIGINT NOT NULL COMMENT '金额',
  `state` TINYINT NOT NULL COMMENT '单状态',
  `biz_type` INTEGER NOT NULL COMMENT '业务类型',
  `desc` VARCHAR(256) NOT NULL COMMENT '转账描述',
  `pay_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '支付时间',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',   
  PRIMARY KEY (`transaction_id`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- c2b待转账表
DROP TABLE IF EXISTS `t_c2b_bill`;
CREATE TABLE `t_c2b_bill` (
  `transaction_id` VARCHAR(64) NOT NULL COMMENT '交易ID',
  `uid` BIGINT NOT NULL COMMENT '用户UID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
  `merchant_uid` BIGINT NOT NULL COMMENT '商户UID',
  `merchant_id` VARCHAR(64) NOT NULL COMMENT '商户ID',
  `amount` VARCHAR(64) NOT NULL COMMENT '金额',
  `state` TINYINT NOT NULL COMMENT '单状态',
  `biz_type` INTEGER NOT NULL COMMENT '业务类型',
  `desc` VARCHAR(256) NOT NULL COMMENT '转账描述',
  `pay_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '支付时间',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',   
  PRIMARY KEY (`transaction_id`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `t_save_bill`;
CREATE TABLE `t_save_bill` (
  `transaction_id` VARCHAR(64) NOT NULL COMMENT '交易ID',
  `uid` BIGINT NOT NULL COMMENT '用户UID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
  `bank_type` VARCHAR(64) NOT NULL COMMENT '银行类型',
  `amount` BIGINT NOT NULL COMMENT '金额',
  `state` TINYINT NOT NULL COMMENT '单状态',
  `desc` VARCHAR(256) NOT NULL COMMENT '充值描述',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`transaction_id`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- c2c待转账表
DROP TABLE IF EXISTS `t_c2c_pending_transfer`;
CREATE TABLE `t_c2c_pending_transfer` (
  `transaction_id` VARCHAR(64) NOT NULL COMMENT '交易ID',
  `buyer_uid` BIGINT NOT NULL COMMENT '买家用户UID',
  `seller_uid` BIGINT NOT NULL COMMENT '卖家用户UID',
  `buyer_user_id` VARCHAR(64) NOT NULL COMMENT '买家用户ID',
  `seller_user_id` VARCHAR(64) NOT NULL COMMENT '卖家用户ID',
  `amount` BIGINT NOT NULL COMMENT '金额',
  `state` TINYINT NOT NULL COMMENT '状态',
  `desc` VARCHAR(256) NOT NULL COMMENT '转账描述',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',  
  PRIMARY KEY (`transaction_id`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- c2b待转账表
DROP TABLE IF EXISTS `t_c2b_pending_transfer`;
CREATE TABLE `t_c2b_pending_transfer` (
  `transaction_id` VARCHAR(64) NOT NULL COMMENT '交易ID',
  `uid` BIGINT NOT NULL COMMENT '用户UID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
  `merchant_uid` BIGINT NOT NULL COMMENT '商户UID',
  `merchant_id` VARCHAR(64) NOT NULL COMMENT '商户ID',
  `amount` BIGINT NOT NULL COMMENT '金额',
  `state` TINYINT NOT NULL COMMENT '状态',
  `desc` VARCHAR(256) NOT NULL COMMENT '转账描述',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',  
  PRIMARY KEY (`transaction_id`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- linux:  mysql -h 127.0.0.1 -P 3306 -u root -proot123456 < account_init.sql
-- windows: Get-Content -Encoding UTF8 account_init.sql | mysql -h 127.0.0.1 -P 3306 -u root -proot123456
-- 只读权限 multipass exec master1 -- sudo kubectl exec -it -n pay-ns mysql-0 -- mysql -ustarslipay -ppayClipayA2026
-- root权限 multipass exec master1 -- sudo kubectl exec -it -n pay-ns mysql-0 -- mysql -uroot -proot123456
