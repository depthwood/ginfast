-- ============================================================
-- clientapp 插件：菜单 / API / 权限 SQL
-- 适用：MySQL（gin-fast-tenant 库）
-- 说明：
--   1. 若 ID 与现有数据冲突，请调整下方 ID 区间后再执行
--   2. 执行后需在「角色管理」中确认 role_1 / role_2 已包含新菜单
--   3. 执行后建议重启后端或触发 Casbin 策略 reload（若项目有缓存）
-- ============================================================

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- 清理旧数据（可重复执行）
-- ----------------------------
DELETE FROM `sys_menu_api` WHERE `menu_id` BETWEEN 140350 AND 140362;
DELETE FROM `sys_role_menu` WHERE `menu_id` BETWEEN 140350 AND 140362;
DELETE FROM `sys_casbin_rule` WHERE `v1` LIKE '/api/plugins/clientapp/%';
DELETE FROM `sys_menu` WHERE `id` BETWEEN 140350 AND 140362;
DELETE FROM `sys_api` WHERE `id` BETWEEN 217 AND 240;

-- ----------------------------
-- sys_api（ID: 217-240）
-- ----------------------------
INSERT INTO `sys_api` (`id`, `title`, `path`, `method`, `api_group`, `created_at`, `updated_at`, `deleted_at`, `created_by`) VALUES
(217, '客户端列表', '/api/plugins/clientapp/admin/client/list', 'GET', '客户端应用', NOW(), NOW(), NULL, 1),
(218, '客户端下拉选项', '/api/plugins/clientapp/admin/client/options', 'GET', '客户端应用', NOW(), NOW(), NULL, 1),
(219, '客户端详情', '/api/plugins/clientapp/admin/client/:id', 'GET', '客户端应用', NOW(), NOW(), NULL, 1),
(220, '新增客户端', '/api/plugins/clientapp/admin/client/add', 'POST', '客户端应用', NOW(), NOW(), NULL, 1),
(221, '编辑客户端', '/api/plugins/clientapp/admin/client/edit', 'PUT', '客户端应用', NOW(), NOW(), NULL, 1),
(222, '更新客户端状态', '/api/plugins/clientapp/admin/client/status', 'PUT', '客户端应用', NOW(), NOW(), NULL, 1),
(223, '删除客户端', '/api/plugins/clientapp/admin/client/delete', 'DELETE', '客户端应用', NOW(), NOW(), NULL, 1),

(224, '平台渠道列表', '/api/plugins/clientapp/admin/platform/list', 'GET', '客户端应用', NOW(), NOW(), NULL, 1),
(225, '平台渠道下拉选项', '/api/plugins/clientapp/admin/platform/options', 'GET', '客户端应用', NOW(), NOW(), NULL, 1),
(226, '支持的平台列表', '/api/plugins/clientapp/admin/platform/supported', 'GET', '客户端应用', NOW(), NOW(), NULL, 1),
(227, '平台渠道详情', '/api/plugins/clientapp/admin/platform/:id', 'GET', '客户端应用', NOW(), NOW(), NULL, 1),
(228, '新增平台渠道', '/api/plugins/clientapp/admin/platform/add', 'POST', '客户端应用', NOW(), NOW(), NULL, 1),
(229, '编辑平台渠道', '/api/plugins/clientapp/admin/platform/edit', 'PUT', '客户端应用', NOW(), NOW(), NULL, 1),
(230, '更新平台渠道状态', '/api/plugins/clientapp/admin/platform/status', 'PUT', '客户端应用', NOW(), NOW(), NULL, 1),
(231, '删除平台渠道', '/api/plugins/clientapp/admin/platform/delete', 'DELETE', '客户端应用', NOW(), NOW(), NULL, 1),

(232, '客户端用户列表', '/api/plugins/clientapp/admin/user/list', 'GET', '客户端应用', NOW(), NOW(), NULL, 1),
(233, '客户端用户详情', '/api/plugins/clientapp/admin/user/:id', 'GET', '客户端应用', NOW(), NOW(), NULL, 1),
(234, '代注册用户', '/api/plugins/clientapp/admin/user/add', 'POST', '客户端应用', NOW(), NOW(), NULL, 1),
(235, '编辑客户端用户', '/api/plugins/clientapp/admin/user/edit', 'PUT', '客户端应用', NOW(), NOW(), NULL, 1),
(236, '更新客户端用户状态', '/api/plugins/clientapp/admin/user/status', 'PUT', '客户端应用', NOW(), NOW(), NULL, 1),
(237, '删除客户端用户', '/api/plugins/clientapp/admin/user/delete', 'DELETE', '客户端应用', NOW(), NOW(), NULL, 1),
(238, '绑定用户身份', '/api/plugins/clientapp/admin/user/identity/bind', 'POST', '客户端应用', NOW(), NOW(), NULL, 1),
(239, '解绑用户身份', '/api/plugins/clientapp/admin/user/identity/unbind', 'DELETE', '客户端应用', NOW(), NOW(), NULL, 1),

(240, '登录日志列表', '/api/plugins/clientapp/admin/signlog/list', 'GET', '客户端应用', NOW(), NOW(), NULL, 1);

-- ----------------------------
-- sys_menu（ID: 140350-140362）
-- type: 1=目录 2=菜单 3=按钮
-- ----------------------------
INSERT INTO `sys_menu` VALUES
('140350', '0', '/clientapp', 'Clientapp', '', '', 'clientapp', '0', '0', '0', '1', '0', '', '0', 'mobile', '', '0', '1', '0', '', NOW(), NOW(), NULL, '1'),
('140351', '140350', '/clientapp/clients', 'ClientappClients', '', 'plugins/clientapp/views/client-list', 'clientapp-client', '0', '0', '0', '1', '0', '', '0', '', 'IconMobile', '1', '2', '0', '', NOW(), NOW(), NULL, '1'),
('140352', '140350', '/clientapp/users', 'ClientappUsers', '', 'plugins/clientapp/views/user-list', 'clientapp-user', '0', '0', '0', '1', '0', '', '0', '', 'IconUser', '2', '2', '0', '', NOW(), NOW(), NULL, '1'),
('140353', '140350', '/clientapp/login-log', 'ClientappLoginLog', '', 'plugins/clientapp/views/login-log', 'clientapp-log', '0', '0', '0', '1', '0', '', '0', '', 'IconHistory', '3', '2', '0', '', NOW(), NOW(), NULL, '1'),

('140354', '140351', '', '', '', '', '新增客户端', '0', '0', '0', '1', '0', '', '0', '', '', '1', '3', '0', 'plugins:clientapp:client:add', NOW(), NOW(), NULL, '1'),
('140355', '140351', '', '', '', '', '编辑客户端', '0', '0', '0', '1', '0', '', '0', '', '', '2', '3', '0', 'plugins:clientapp:client:edit', NOW(), NOW(), NULL, '1'),
('140356', '140351', '', '', '', '', '删除客户端', '0', '0', '0', '1', '0', '', '0', '', '', '3', '3', '0', 'plugins:clientapp:client:delete', NOW(), NOW(), NULL, '1'),
('140357', '140351', '', '', '', '', '新增平台渠道', '0', '0', '0', '1', '0', '', '0', '', '', '4', '3', '0', 'plugins:clientapp:platform:add', NOW(), NOW(), NULL, '1'),
('140358', '140351', '', '', '', '', '编辑平台渠道', '0', '0', '0', '1', '0', '', '0', '', '', '5', '3', '0', 'plugins:clientapp:platform:edit', NOW(), NOW(), NULL, '1'),
('140359', '140351', '', '', '', '', '删除平台渠道', '0', '0', '0', '1', '0', '', '0', '', '', '6', '3', '0', 'plugins:clientapp:platform:delete', NOW(), NOW(), NULL, '1'),

('140360', '140352', '', '', '', '', '代注册用户', '0', '0', '0', '1', '0', '', '0', '', '', '1', '3', '0', 'plugins:clientapp:user:add', NOW(), NOW(), NULL, '1'),
('140361', '140352', '', '', '', '', '编辑用户', '0', '0', '0', '1', '0', '', '0', '', '', '2', '3', '0', 'plugins:clientapp:user:edit', NOW(), NOW(), NULL, '1'),
('140362', '140352', '', '', '', '', '删除用户', '0', '0', '0', '1', '0', '', '0', '', '', '3', '3', '0', 'plugins:clientapp:user:delete', NOW(), NOW(), NULL, '1');

-- ----------------------------
-- sys_menu_api（菜单与 API 关联）
-- ----------------------------
INSERT INTO `sys_menu_api` (`menu_id`, `api_id`) VALUES
-- 客户端管理页
('140351', 217), ('140351', 218), ('140351', 219),
('140351', 224), ('140351', 225), ('140351', 226), ('140351', 227),
-- 用户管理页（依赖客户端/平台下拉）
('140352', 218), ('140352', 225),
('140352', 232), ('140352', 233),
-- 登录日志页
('140353', 218), ('140353', 240),

-- 客户端按钮
('140354', 220),
('140355', 221), ('140355', 222),
('140356', 223),
('140357', 228),
('140358', 229), ('140358', 230),
('140359', 231),

-- 用户按钮
('140360', 234), ('140360', 238),
('140361', 235), ('140361', 236), ('140361', 238), ('140361', 239),
('140362', 237);

-- ----------------------------
-- sys_casbin_rule（role_1 超级管理员 / role_2 租户管理员）
-- ID: 7600-7647
-- ----------------------------
INSERT INTO `sys_casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES
-- role_1
(7600, 'p', 'role_1', '/api/plugins/clientapp/admin/client/list', 'GET', '*', '', ''),
(7601, 'p', 'role_1', '/api/plugins/clientapp/admin/client/options', 'GET', '*', '', ''),
(7602, 'p', 'role_1', '/api/plugins/clientapp/admin/client/:id', 'GET', '*', '', ''),
(7603, 'p', 'role_1', '/api/plugins/clientapp/admin/client/add', 'POST', '*', '', ''),
(7604, 'p', 'role_1', '/api/plugins/clientapp/admin/client/edit', 'PUT', '*', '', ''),
(7605, 'p', 'role_1', '/api/plugins/clientapp/admin/client/status', 'PUT', '*', '', ''),
(7606, 'p', 'role_1', '/api/plugins/clientapp/admin/client/delete', 'DELETE', '*', '', ''),
(7607, 'p', 'role_1', '/api/plugins/clientapp/admin/platform/list', 'GET', '*', '', ''),
(7608, 'p', 'role_1', '/api/plugins/clientapp/admin/platform/options', 'GET', '*', '', ''),
(7609, 'p', 'role_1', '/api/plugins/clientapp/admin/platform/supported', 'GET', '*', '', ''),
(7610, 'p', 'role_1', '/api/plugins/clientapp/admin/platform/:id', 'GET', '*', '', ''),
(7611, 'p', 'role_1', '/api/plugins/clientapp/admin/platform/add', 'POST', '*', '', ''),
(7612, 'p', 'role_1', '/api/plugins/clientapp/admin/platform/edit', 'PUT', '*', '', ''),
(7613, 'p', 'role_1', '/api/plugins/clientapp/admin/platform/status', 'PUT', '*', '', ''),
(7614, 'p', 'role_1', '/api/plugins/clientapp/admin/platform/delete', 'DELETE', '*', '', ''),
(7615, 'p', 'role_1', '/api/plugins/clientapp/admin/user/list', 'GET', '*', '', ''),
(7616, 'p', 'role_1', '/api/plugins/clientapp/admin/user/:id', 'GET', '*', '', ''),
(7617, 'p', 'role_1', '/api/plugins/clientapp/admin/user/add', 'POST', '*', '', ''),
(7618, 'p', 'role_1', '/api/plugins/clientapp/admin/user/edit', 'PUT', '*', '', ''),
(7619, 'p', 'role_1', '/api/plugins/clientapp/admin/user/status', 'PUT', '*', '', ''),
(7620, 'p', 'role_1', '/api/plugins/clientapp/admin/user/delete', 'DELETE', '*', '', ''),
(7621, 'p', 'role_1', '/api/plugins/clientapp/admin/user/identity/bind', 'POST', '*', '', ''),
(7622, 'p', 'role_1', '/api/plugins/clientapp/admin/user/identity/unbind', 'DELETE', '*', '', ''),
(7623, 'p', 'role_1', '/api/plugins/clientapp/admin/signlog/list', 'GET', '*', '', ''),

-- role_2
(7630, 'p', 'role_2', '/api/plugins/clientapp/admin/client/list', 'GET', '*', '', ''),
(7631, 'p', 'role_2', '/api/plugins/clientapp/admin/client/options', 'GET', '*', '', ''),
(7632, 'p', 'role_2', '/api/plugins/clientapp/admin/client/:id', 'GET', '*', '', ''),
(7633, 'p', 'role_2', '/api/plugins/clientapp/admin/client/add', 'POST', '*', '', ''),
(7634, 'p', 'role_2', '/api/plugins/clientapp/admin/client/edit', 'PUT', '*', '', ''),
(7635, 'p', 'role_2', '/api/plugins/clientapp/admin/client/status', 'PUT', '*', '', ''),
(7636, 'p', 'role_2', '/api/plugins/clientapp/admin/client/delete', 'DELETE', '*', '', ''),
(7637, 'p', 'role_2', '/api/plugins/clientapp/admin/platform/list', 'GET', '*', '', ''),
(7638, 'p', 'role_2', '/api/plugins/clientapp/admin/platform/options', 'GET', '*', '', ''),
(7639, 'p', 'role_2', '/api/plugins/clientapp/admin/platform/supported', 'GET', '*', '', ''),
(7640, 'p', 'role_2', '/api/plugins/clientapp/admin/platform/:id', 'GET', '*', '', ''),
(7641, 'p', 'role_2', '/api/plugins/clientapp/admin/platform/add', 'POST', '*', '', ''),
(7642, 'p', 'role_2', '/api/plugins/clientapp/admin/platform/edit', 'PUT', '*', '', ''),
(7643, 'p', 'role_2', '/api/plugins/clientapp/admin/platform/status', 'PUT', '*', '', ''),
(7644, 'p', 'role_2', '/api/plugins/clientapp/admin/platform/delete', 'DELETE', '*', '', ''),
(7645, 'p', 'role_2', '/api/plugins/clientapp/admin/user/list', 'GET', '*', '', ''),
(7646, 'p', 'role_2', '/api/plugins/clientapp/admin/user/:id', 'GET', '*', '', ''),
(7647, 'p', 'role_2', '/api/plugins/clientapp/admin/user/add', 'POST', '*', '', ''),
(7648, 'p', 'role_2', '/api/plugins/clientapp/admin/user/edit', 'PUT', '*', '', ''),
(7649, 'p', 'role_2', '/api/plugins/clientapp/admin/user/status', 'PUT', '*', '', ''),
(7650, 'p', 'role_2', '/api/plugins/clientapp/admin/user/delete', 'DELETE', '*', '', ''),
(7651, 'p', 'role_2', '/api/plugins/clientapp/admin/user/identity/bind', 'POST', '*', '', ''),
(7652, 'p', 'role_2', '/api/plugins/clientapp/admin/user/identity/unbind', 'DELETE', '*', '', ''),
(7653, 'p', 'role_2', '/api/plugins/clientapp/admin/signlog/list', 'GET', '*', '', '');

-- ----------------------------
-- sys_role_menu（为 role_1 / role_2 分配菜单）
-- ----------------------------
INSERT INTO `sys_role_menu` (`role_id`, `menu_id`) VALUES
('1', '140350'), ('1', '140351'), ('1', '140352'), ('1', '140353'),
('1', '140354'), ('1', '140355'), ('1', '140356'), ('1', '140357'), ('1', '140358'), ('1', '140359'),
('1', '140360'), ('1', '140361'), ('1', '140362'),

('2', '140350'), ('2', '140351'), ('2', '140352'), ('2', '140353'),
('2', '140354'), ('2', '140355'), ('2', '140356'), ('2', '140357'), ('2', '140358'), ('2', '140359'),
('2', '140360'), ('2', '140361'), ('2', '140362');

SET FOREIGN_KEY_CHECKS=1;
