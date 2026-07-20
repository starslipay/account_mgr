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
  `balance` BIGINT NOT NULL COMMENT '余额',
  `amount` BIGINT NOT NULL COMMENT '金额',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uid_inout_type_transaction_id_biz_type` (`uid`,`inout_type`,`transaction_id`),
  INDEX `idx_transaction_id` (`transaction_id`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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

CREATE TABLE `t_c2c_bill` (
  `transaction_id` VARCHAR(64) NOT NULL COMMENT '订单ID',
  `buyer_uid` BIGINT NOT NULL COMMENT '买家用户UID',
  `seller_uid` BIGINT NOT NULL COMMENT '卖家用户UID',
  `buyer_user_id` VARCHAR(64) NOT NULL COMMENT '买家用户ID',
  `seller_user_id` VARCHAR(64) NOT NULL COMMENT '卖家用户ID',
  `amount` BIGINT NOT NULL COMMENT '金额',
  `state` TINYINT NOT NULL COMMENT '单状态',
  `biz_type` INTEGER NOT NULL COMMENT '业务类型',
  `desc` VARCHAR(256) NOT NULL COMMENT '转账描述',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`transaction_id`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `t_save_bill` (
  `transaction_id` VARCHAR(64) NOT NULL COMMENT '订单ID',
  `uid` BIGINT NOT NULL COMMENT '用户UID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
  `bank_type` VARCHAR(64) NOT NULL COMMENT '银行类型',
  `amount` BIGINT NOT NULL COMMENT '金额',
  `state` TINYINT NOT NULL COMMENT '单状态',
  `desc` VARCHAR(256) NOT NULL COMMENT '充值描述',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`transaction_id`),
  INDEX `idx_create_time` (`create_time`),
  INDEX `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
 

-- goctl model mysql ddl -src account.sql -dir .
-- -c：开启缓存（redis，可选，不加则无缓存）
