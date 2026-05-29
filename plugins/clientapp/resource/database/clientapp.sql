SET FOREIGN_KEY_CHECKS=0;

DROP TABLE IF EXISTS `plugin_clientapp_client`;
CREATE TABLE `plugin_clientapp_client` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '租户ID',
  `client_key` varchar(64) NOT NULL COMMENT '客户端Key',
  `client_name` varchar(100) NOT NULL COMMENT '客户端名称',
  `client_type` varchar(32) NOT NULL DEFAULT 'mini_program' COMMENT '客户端类型: mini_program/h5/native',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态 0停用 1启用',
  `wallet_evm_enabled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否启用EVM钱包登录',
  `allowed_chain_ids` varchar(500) DEFAULT NULL COMMENT '允许的EVM链ID(JSON数组)',
  `wallet_sign_message` varchar(500) DEFAULT NULL COMMENT '钱包签名消息模板',
  `logo` varchar(500) DEFAULT NULL COMMENT 'Logo',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_by` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建人',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_tenant_client_key` (`tenant_id`,`client_key`) USING BTREE,
  KEY `idx_tenant_status` (`tenant_id`,`status`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='客户端应用';

DROP TABLE IF EXISTS `plugin_clientapp_platform`;
CREATE TABLE `plugin_clientapp_platform` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '租户ID',
  `client_id` int(11) unsigned NOT NULL COMMENT '客户端ID',
  `platform` varchar(32) NOT NULL COMMENT '平台: wechat/alipay/douyin/baidu/qq/kuaishou/feishu/jd',
  `platform_app_id` varchar(128) NOT NULL COMMENT '平台AppID',
  `platform_app_name` varchar(100) DEFAULT NULL COMMENT '渠道展示名',
  `credentials` text COMMENT '平台凭证(JSON)',
  `features` varchar(1000) DEFAULT NULL COMMENT '平台特性(JSON)',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态 0停用 1启用',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_client_platform` (`client_id`,`platform`,`platform_app_id`) USING BTREE,
  KEY `idx_tenant_client` (`tenant_id`,`client_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='客户端平台渠道';

DROP TABLE IF EXISTS `plugin_clientapp_user`;
CREATE TABLE `plugin_clientapp_user` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '租户ID',
  `nickname` varchar(100) DEFAULT NULL COMMENT '昵称',
  `avatar` varchar(500) DEFAULT NULL COMMENT '头像',
  `gender` tinyint(1) NOT NULL DEFAULT '0' COMMENT '性别 0未知 1男 2女',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态 0禁用 1正常',
  `register_source` varchar(32) NOT NULL DEFAULT 'admin' COMMENT '注册来源',
  `register_client_id` int(11) unsigned DEFAULT NULL COMMENT '注册来源客户端ID',
  `register_platform_id` int(11) unsigned DEFAULT NULL COMMENT '注册来源平台渠道ID',
  `last_login_at` datetime DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` varchar(64) DEFAULT NULL COMMENT '最后登录IP',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_by` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建人',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_tenant_status` (`tenant_id`,`status`) USING BTREE,
  KEY `idx_tenant_created` (`tenant_id`,`created_at`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='客户端用户';

DROP TABLE IF EXISTS `plugin_clientapp_user_identity`;
CREATE TABLE `plugin_clientapp_user_identity` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '租户ID',
  `user_id` int(11) unsigned NOT NULL COMMENT '用户ID',
  `client_id` int(11) unsigned DEFAULT NULL COMMENT '客户端ID',
  `platform_id` int(11) unsigned DEFAULT NULL COMMENT '平台渠道ID',
  `platform` varchar(32) DEFAULT NULL COMMENT '平台码',
  `identity_type` varchar(32) NOT NULL COMMENT '身份类型',
  `identity_key` varchar(128) NOT NULL COMMENT '身份标识',
  `provider_id` varchar(128) DEFAULT NULL COMMENT '辅助ID',
  `union_key` varchar(128) DEFAULT NULL COMMENT 'UnionID',
  `extra_data` text COMMENT '扩展数据(JSON)',
  `verified_at` datetime DEFAULT NULL COMMENT '验证时间',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态 0解绑 1有效',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_identity` (`tenant_id`,`identity_type`,`platform`,`identity_key`,`provider_id`,`platform_id`) USING BTREE,
  KEY `idx_user_id` (`user_id`) USING BTREE,
  KEY `idx_platform_id` (`platform_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='客户端用户身份';

DROP TABLE IF EXISTS `plugin_clientapp_login_log`;
CREATE TABLE `plugin_clientapp_login_log` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '租户ID',
  `client_id` int(11) unsigned DEFAULT NULL COMMENT '客户端ID',
  `platform_id` int(11) unsigned DEFAULT NULL COMMENT '平台渠道ID',
  `platform` varchar(32) DEFAULT NULL COMMENT '平台码',
  `user_id` int(11) unsigned DEFAULT NULL COMMENT '用户ID',
  `identity_type` varchar(32) DEFAULT NULL COMMENT '身份类型',
  `identity_key` varchar(128) DEFAULT NULL COMMENT '身份标识(脱敏)',
  `login_channel` varchar(32) DEFAULT NULL COMMENT '登录渠道',
  `ip` varchar(64) DEFAULT NULL COMMENT 'IP',
  `user_agent` varchar(500) DEFAULT NULL COMMENT 'UserAgent',
  `device_info` text COMMENT '设备信息(JSON)',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态 0失败 1成功',
  `fail_reason` varchar(255) DEFAULT NULL COMMENT '失败原因',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_tenant_created` (`tenant_id`,`created_at`) USING BTREE,
  KEY `idx_client_id` (`client_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='客户端登录日志';

SET FOREIGN_KEY_CHECKS=1;
