-- 将 app_key 重命名为 client_key（适用于已执行旧版 clientapp.sql 的环境）
-- 若表不存在或列已是 client_key，请忽略对应报错

ALTER TABLE `plugin_clientapp_client`
  CHANGE COLUMN `app_key` `client_key` varchar(64) NOT NULL COMMENT '客户端Key';

ALTER TABLE `plugin_clientapp_client`
  DROP INDEX `uk_tenant_app_key`,
  ADD UNIQUE KEY `uk_tenant_client_key` (`tenant_id`,`client_key`) USING BTREE;
