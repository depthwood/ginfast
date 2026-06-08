-- clientapp App界面配置表迁移

CREATE TABLE IF NOT EXISTS `plugin_clientapp_app_config` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '租户ID',
  `client_id` int(11) unsigned NOT NULL COMMENT '客户端ID',
  `config_key` varchar(64) NOT NULL DEFAULT 'default' COMMENT '配置Key',
  `name` varchar(100) NOT NULL COMMENT '配置名称',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态 0停用 1启用',
  `theme` text COMMENT '主题配置(JSON)',
  `pages` text COMMENT '页面开放配置(JSON)',
  `feature_flags` text COMMENT '功能开关(JSON)',
  `navigation` text COMMENT '导航配置(JSON)',
  `extra` text COMMENT '扩展配置(JSON)',
  `cache_version` varchar(64) DEFAULT NULL COMMENT '缓存版本',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_client_config_key` (`tenant_id`,`client_id`,`config_key`) USING BTREE,
  KEY `idx_tenant_client_status` (`tenant_id`,`client_id`,`status`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='客户端App界面配置';

