package models

import (
	"context"
	"gin-fast/app/global/app"
	"gin-fast/app/models"

	"gorm.io/gorm"
)

// Platform 平台渠道
type Platform struct {
	models.BaseModel
	TenantID        uint       `gorm:"column:tenant_id;type:int(11);not null;comment:租户ID" json:"tenantID"`
	ClientID        uint       `gorm:"column:client_id;type:int(11);not null;comment:客户端ID" json:"clientID"`
	Platform        string     `gorm:"column:platform;type:varchar(32);not null;comment:平台" json:"platform"`
	PlatformAppID   string     `gorm:"column:platform_app_id;type:varchar(128);not null;comment:平台AppID" json:"platformAppId"`
	PlatformAppName string     `gorm:"column:platform_app_name;type:varchar(100);comment:渠道名称" json:"platformAppName"`
	Credentials     JSONString `gorm:"column:credentials;type:text;comment:平台凭证" json:"credentials"`
	Features        JSONString `gorm:"column:features;type:varchar(1000);comment:平台特性" json:"features"`
	Status          int8       `gorm:"column:status;type:tinyint(1);not null;default:1;comment:状态" json:"status"`
	Remark          string     `gorm:"column:remark;type:varchar(500);comment:备注" json:"remark"`
	ClientName      string     `gorm:"-" json:"clientName,omitempty"`
	ClientKey       string     `gorm:"-" json:"clientKey,omitempty"`
}

func (Platform) TableName() string {
	return "plugin_clientapp_platform"
}

func NewPlatform() *Platform {
	return &Platform{}
}

type PlatformList []Platform

func NewPlatformList() *PlatformList {
	return &PlatformList{}
}

func (m *Platform) IsEmpty() bool {
	return m == nil || m.ID == 0
}

func (m *Platform) GetByID(c context.Context, id uint, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Scopes(scopes...).First(m, id).Error
}

func (m *Platform) Create(c context.Context) error {
	return app.DB().WithContext(c).Create(m).Error
}

func (m *Platform) Update(c context.Context) error {
	return app.DB().WithContext(c).Save(m).Error
}

func (m *Platform) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

func (l *PlatformList) Find(c context.Context, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&Platform{}).Scopes(scopes...).Find(l).Error
}

func (l *PlatformList) GetTotal(c context.Context, scopes ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&Platform{}).Scopes(scopes...).Count(&count).Error
	return count, err
}

func CountPlatformIdentities(c context.Context, platformID uint) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&UserIdentity{}).Where("platform_id = ? AND status = 1", platformID).Count(&count).Error
	return count, err
}
