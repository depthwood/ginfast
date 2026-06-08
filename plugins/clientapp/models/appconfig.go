package models

import (
	"context"
	"gin-fast/app/global/app"
	"gin-fast/app/models"

	"gorm.io/gorm"
)

type AppConfig struct {
	models.BaseModel
	TenantID          uint       `gorm:"column:tenant_id;type:int(11);not null;comment:租户ID" json:"tenantID"`
	ClientID          uint       `gorm:"column:client_id;type:int(11);not null;comment:客户端ID" json:"clientId"`
	ConfigKey         string     `gorm:"column:config_key;type:varchar(64);not null;default:default;comment:配置Key(方案标识)" json:"configKey"`
	Name              string     `gorm:"column:name;type:varchar(100);not null;comment:配置名称" json:"name"`
	Status            int8       `gorm:"column:status;type:tinyint(1);not null;default:1;comment:状态 0停用 1启用" json:"status"`
	IsActive          bool       `gorm:"column:is_active;type:tinyint(1);not null;default:0;comment:是否为当前激活方案" json:"isActive"`
	Theme             JSONString `gorm:"column:theme;type:text;comment:主题配置" json:"theme"`
	Pages             JSONString `gorm:"column:pages;type:text;comment:页面开放配置" json:"pages"`
	FeatureFlags      JSONString `gorm:"column:feature_flags;type:text;comment:功能开关" json:"featureFlags"`
	Navigation        JSONString `gorm:"column:navigation;type:text;comment:导航配置" json:"navigation"`
	Extra             JSONString `gorm:"column:extra;type:longtext;comment:扩展配置(多页面装修数据)" json:"extra"`
	CacheVersion      string     `gorm:"column:cache_version;type:varchar(64);comment:缓存版本" json:"cacheVersion"`
	Remark            string     `gorm:"column:remark;type:varchar(500);comment:备注" json:"remark"`
	ClientName        string     `gorm:"-" json:"clientName,omitempty"`
	ClientKey         string     `gorm:"-" json:"clientKey,omitempty"`
	EffectiveConfig   JSONString `gorm:"-" json:"effectiveConfig,omitempty"`
	EffectiveETag     string     `gorm:"-" json:"effectiveETag,omitempty"`
	EffectiveCacheTTL int        `gorm:"-" json:"cacheTtl,omitempty"`
}

func (AppConfig) TableName() string {
	return "plugin_clientapp_app_config"
}

func NewAppConfig() *AppConfig {
	return &AppConfig{}
}

type AppConfigList []AppConfig

func NewAppConfigList() *AppConfigList {
	return &AppConfigList{}
}

func (m *AppConfig) IsEmpty() bool {
	return m == nil || m.ID == 0
}

func (m *AppConfig) GetByID(c context.Context, id uint, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Scopes(scopes...).First(m, id).Error
}

func (m *AppConfig) FindDefault(c context.Context, tenantID uint, clientID uint, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Scopes(scopes...).
		Where("tenant_id = ? AND client_id = ? AND config_key = ?", tenantID, clientID, "default").
		First(m).Error
}

func (m *AppConfig) FindActive(c context.Context, tenantID uint, clientID uint, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Scopes(scopes...).
		Where("tenant_id = ? AND client_id = ? AND is_active = 1 AND status = 1", tenantID, clientID).
		First(m).Error
}

func (m *AppConfig) FindByKey(c context.Context, tenantID uint, clientID uint, configKey string, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Scopes(scopes...).
		Where("tenant_id = ? AND client_id = ? AND config_key = ?", tenantID, clientID, configKey).
		First(m).Error
}

func DeactivateAllConfigs(c context.Context, tenantID uint, clientID uint) error {
	return app.DB().WithContext(c).
		Model(&AppConfig{}).
		Where("tenant_id = ? AND client_id = ? AND is_active = 1", tenantID, clientID).
		Update("is_active", false).Error
}

func (m *AppConfig) Create(c context.Context) error {
	return app.DB().WithContext(c).Create(m).Error
}

func (m *AppConfig) Update(c context.Context) error {
	return app.DB().WithContext(c).Save(m).Error
}

func (m *AppConfig) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

func (l *AppConfigList) Find(c context.Context, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&AppConfig{}).Scopes(scopes...).Order("id desc").Find(l).Error
}

func (l *AppConfigList) GetTotal(c context.Context, scopes ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&AppConfig{}).Scopes(scopes...).Count(&count).Error
	return count, err
}
