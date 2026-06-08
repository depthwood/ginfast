-- ============================================================
-- clientapp 插件：设备自动发现 表结构 + 菜单 / API / 权限
-- 适用：MySQL
-- ============================================================

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- 设备表
-- ----------------------------
CREATE TABLE IF NOT EXISTS `plugin_clientapp_device` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `device_uuid` varchar(64) NOT NULL COMMENT '设备唯一标识(UUID)',
  `tenant_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '租户ID',
  `client_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '客户端ID',
  `device_name` varchar(100) DEFAULT NULL COMMENT '设备名称',
  `platform` varchar(32) NOT NULL DEFAULT 'android' COMMENT '平台: android/ios/web',
  `app_version` varchar(32) DEFAULT NULL COMMENT 'APP版本',
  `device_info` text COMMENT '设备详情(JSON)',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态 0停用 1正常 2待审核',
  `registered_at` datetime DEFAULT NULL COMMENT '注册时间',
  `last_seen_at` datetime DEFAULT NULL COMMENT '最后活跃时间',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_device_uuid` (`device_uuid`) USING BTREE,
  KEY `idx_tenant_client` (`tenant_id`,`client_id`) USING BTREE,
  KEY `idx_tenant_status` (`tenant_id`,`status`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='客户端设备（自动发现注册）';

-- ----------------------------
-- 清理旧数据（可重复执行）
-- ----------------------------
DELETE FROM `sys_menu_api` WHERE `menu_id` BETWEEN 140370 AND 140376;
DELETE FROM `sys_role_menu` WHERE `menu_id` BETWEEN 140370 AND 140376;
DELETE FROM `sys_casbin_rule` WHERE `v1` LIKE '/api/plugins/clientapp/admin/device/%';
DELETE FROM `sys_menu` WHERE `id` BETWEEN 140370 AND 140376;
DELETE FROM `sys_api` WHERE `id` BETWEEN 248 AND 254;

-- ----------------------------
-- sys_api（ID: 248-254）
-- ----------------------------
INSERT INTO `sys_api` (`id`, `title`, `path`, `method`, `api_group`, `created_at`, `updated_at`, `deleted_at`, `created_by`) VALUES
(248, '设备发现', '/api/plugins/clientapp/device/discover', 'POST', '客户端应用', NOW(), NOW(), NULL, 1),
(249, '设备列表', '/api/plugins/clientapp/admin/device/list', 'GET', '客户端应用', NOW(), NOW(), NULL, 1),
(250, '更新设备状态', '/api/plugins/clientapp/admin/device/status', 'PUT', '客户端应用', NOW(), NOW(), NULL, 1),
(251, '绑定设备客户端', '/api/plugins/clientapp/admin/device/bind', 'PUT', '客户端应用', NOW(), NOW(), NULL, 1),
(252, '删除设备', '/api/plugins/clientapp/admin/device/delete', 'DELETE', '客户端应用', NOW(), NOW(), NULL, 1),
(253, '模拟设备发现', '/api/plugins/clientapp/admin/device/simulate', 'POST', '客户端应用', NOW(), NOW(), NULL, 1),
(254, '设备心跳', '/api/plugins/clientapp/device/heartbeat', 'POST', '客户端应用', NOW(), NOW(), NULL, 1);

-- ----------------------------
-- sys_menu（ID: 140370-140376）
-- ----------------------------
INSERT INTO `sys_menu` VALUES
('140370', '140350', '/clientapp/devices', 'ClientappDevices', '', 'plugins/clientapp/views/device-list', 'clientapp-device', '0', '0', '0', '1', '0', '', '0', '', 'IconConnection', '5', '2', '0', '', NOW(), NOW(), NULL, '1'),
('140371', '140370', '', '', '', '', '启停设备', '0', '0', '0', '1', '0', '', '0', '', '', '1', '3', '0', 'plugins:clientapp:device:status', NOW(), NOW(), NULL, '1'),
('140372', '140370', '', '', '', '', '绑定客户端', '0', '0', '0', '1', '0', '', '0', '', '', '2', '3', '0', 'plugins:clientapp:device:bind', NOW(), NOW(), NULL, '1'),
('140373', '140370', '', '', '', '', '删除设备', '0', '0', '0', '1', '0', '', '0', '', '', '3', '3', '0', 'plugins:clientapp:device:delete', NOW(), NOW(), NULL, '1');

-- ----------------------------
-- sys_menu_api
-- ----------------------------
INSERT INTO `sys_menu_api` (`menu_id`, `api_id`) VALUES
('140370', 248),
('140370', 249),
('140370', 253),
('140370', 254),
('140371', 250),
('140372', 251),
('140373', 252);

-- ----------------------------
-- sys_casbin_rule
-- ----------------------------
INSERT INTO `sys_casbin_rule` (`ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES
('p', 'role_1', '/api/plugins/clientapp/admin/device/list', 'GET', '*', '', ''),
('p', 'role_1', '/api/plugins/clientapp/admin/device/status', 'PUT', '*', '', ''),
('p', 'role_1', '/api/plugins/clientapp/admin/device/bind', 'PUT', '*', '', ''),
('p', 'role_1', '/api/plugins/clientapp/admin/device/delete', 'DELETE', '*', '', ''),
('p', 'role_1', '/api/plugins/clientapp/admin/device/simulate', 'POST', '*', '', ''),
('p', 'role_2', '/api/plugins/clientapp/admin/device/list', 'GET', '*', '', ''),
('p', 'role_2', '/api/plugins/clientapp/admin/device/status', 'PUT', '*', '', ''),
('p', 'role_2', '/api/plugins/clientapp/admin/device/bind', 'PUT', '*', '', ''),
('p', 'role_2', '/api/plugins/clientapp/admin/device/delete', 'DELETE', '*', '', ''),
('p', 'role_2', '/api/plugins/clientapp/admin/device/simulate', 'POST', '*', '', '');

-- ----------------------------
-- sys_role_menu
-- ----------------------------
INSERT INTO `sys_role_menu` (`role_id`, `menu_id`) VALUES
('1', '140370'), ('1', '140371'), ('1', '140372'), ('1', '140373'),
('2', '140370'), ('2', '140371'), ('2', '140372'), ('2', '140373');

SET FOREIGN_KEY_CHECKS=1;
