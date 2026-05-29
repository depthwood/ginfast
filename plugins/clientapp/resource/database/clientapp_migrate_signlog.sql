-- signlog 权限迁移（loginlog -> signlog）
-- 适用：已将后端/前端接口改为 signlog，但数据库仍为 loginlog 的环境
-- 执行后请重启后端，或等待 Casbin 自动重载（默认约 120 秒）

-- 1. 更新 sys_api 路径
UPDATE `sys_api`
SET `path` = '/api/plugins/clientapp/admin/signlog/list', `updated_at` = NOW()
WHERE `path` = '/api/plugins/clientapp/admin/loginlog/list';

-- 2. 更新已有 Casbin 策略路径
UPDATE `sys_casbin_rule`
SET `v1` = '/api/plugins/clientapp/admin/signlog/list'
WHERE `v1` = '/api/plugins/clientapp/admin/loginlog/list';

-- 3. 若从未写入过 signlog 策略，则为 role_1 / role_2 补插入
INSERT INTO `sys_casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`)
SELECT 7623, 'p', 'role_1', '/api/plugins/clientapp/admin/signlog/list', 'GET', '*', '', ''
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM `sys_casbin_rule`
    WHERE `v0` = 'role_1' AND `v1` = '/api/plugins/clientapp/admin/signlog/list' AND `v2` = 'GET'
);

INSERT INTO `sys_casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`)
SELECT 7653, 'p', 'role_2', '/api/plugins/clientapp/admin/signlog/list', 'GET', '*', '', ''
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM `sys_casbin_rule`
    WHERE `v0` = 'role_2' AND `v1` = '/api/plugins/clientapp/admin/signlog/list' AND `v2` = 'GET'
);

-- 4. 若 sys_api 中 signlog 记录缺失，则补插入（ID 240）
INSERT INTO `sys_api` (`id`, `title`, `path`, `method`, `api_group`, `created_at`, `updated_at`, `deleted_at`, `created_by`)
SELECT 240, '登录日志列表', '/api/plugins/clientapp/admin/signlog/list', 'GET', '客户端应用', NOW(), NOW(), NULL, 1
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM `sys_api` WHERE `id` = 240 OR `path` = '/api/plugins/clientapp/admin/signlog/list'
);

-- 5. 确保登录日志菜单关联 signlog API
INSERT IGNORE INTO `sys_menu_api` (`menu_id`, `api_id`) VALUES ('140353', 240);
