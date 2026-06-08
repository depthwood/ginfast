package models

import (
	"context"
	"gin-fast/app/global/app"
	"gin-fast/app/models"

	"gorm.io/gorm"
)

// AppConfigSnapshot 配置快照，用于版本管理和回滚
type AppConfigSnapshot struct {
	models.BaseModel
	TenantID     uint       `gorm:"column:tenant_id;type:int(11);not null;comment:租户ID" json:"tenantID"`
	ClientID     uint       `gorm:"column:client_id;type:int(11);not null;comment:客户端ID" json:"clientId"`
	ConfigID     uint       `gorm:"column:config_id;type:int(11);not null;comment:关联的AppConfig ID" json:"configId"`
	ConfigKey    string     `gorm:"column:config_key;type:varchar(64);not null;comment:方案标识" json:"configKey"`
	Name         string     `gorm:"column:name;type:varchar(100);not null;comment:快照名称" json:"name"`
	Version      string     `gorm:"column:version;type:varchar(32);not null;comment:版本号" json:"version"`
	SnapshotData JSONString `gorm:"column:snapshot_data;type:longtext;not null;comment:完整配置快照(JSON)" json:"snapshotData"`
	Remark       string     `gorm:"column:remark;type:varchar(500);comment:备注" json:"remark"`
	ConfigName   string     `gorm:"-" json:"configName,omitempty"`
}

func (AppConfigSnapshot) TableName() string {
	return "plugin_clientapp_app_config_snapshot"
}

func NewAppConfigSnapshot() *AppConfigSnapshot {
	return &AppConfigSnapshot{}
}

type AppConfigSnapshotList []AppConfigSnapshot

func NewAppConfigSnapshotList() *AppConfigSnapshotList {
	return &AppConfigSnapshotList{}
}

func (m *AppConfigSnapshot) IsEmpty() bool {
	return m == nil || m.ID == 0
}

func (m *AppConfigSnapshot) GetByID(c context.Context, id uint, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Scopes(scopes...).First(m, id).Error
}

func (m *AppConfigSnapshot) Create(c context.Context) error {
	return app.DB().WithContext(c).Create(m).Error
}

func (m *AppConfigSnapshot) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

func (l *AppConfigSnapshotList) FindByConfigID(c context.Context, configID uint, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&AppConfigSnapshot{}).Scopes(scopes...).
		Where("config_id = ?", configID).
		Order("id desc").Find(l).Error
}

func (l *AppConfigSnapshotList) FindByClient(c context.Context, tenantID uint, clientID uint, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&AppConfigSnapshot{}).Scopes(scopes...).
		Where("tenant_id = ? AND client_id = ?", tenantID, clientID).
		Order("id desc").Find(l).Error
}

func (l *AppConfigSnapshotList) GetTotal(c context.Context, scopes ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&AppConfigSnapshot{}).Scopes(scopes...).Count(&count).Error
	return count, err
}
