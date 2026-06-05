package models

import (
	"context"
	"gin-fast/app/global/app"
	"time"

	"gorm.io/gorm"
)

// LoginLog 登录日志
type LoginLog struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	TenantID     uint       `gorm:"column:tenant_id;type:int(11);not null;comment:租户ID" json:"tenantID"`
	ClientID     *uint      `gorm:"column:client_id;type:int(11);comment:客户端ID" json:"clientId"`
	PlatformID   *uint      `gorm:"column:platform_id;type:int(11);comment:平台渠道ID" json:"platformId"`
	Platform     string     `gorm:"column:platform;type:varchar(32);comment:平台码" json:"platform"`
	UserID       *uint      `gorm:"column:user_id;type:int(11);comment:用户ID" json:"userId"`
	IdentityType string     `gorm:"column:identity_type;type:varchar(32);comment:身份类型" json:"identityType"`
	IdentityKey  string     `gorm:"column:identity_key;type:varchar(128);comment:身份标识" json:"identityKey"`
	LoginChannel string     `gorm:"column:login_channel;type:varchar(32);comment:登录渠道" json:"loginChannel"`
	IP           string     `gorm:"column:ip;type:varchar(64);comment:IP" json:"ip"`
	UserAgent    string     `gorm:"column:user_agent;type:varchar(500);comment:UserAgent" json:"userAgent"`
	DeviceInfo   JSONString `gorm:"column:device_info;type:text;comment:设备信息" json:"deviceInfo"`
	Status       int8       `gorm:"column:status;type:tinyint(1);not null;default:1;comment:状态" json:"status"`
	FailReason   string     `gorm:"column:fail_reason;type:varchar(255);comment:失败原因" json:"failReason"`
	CreatedAt    time.Time  `gorm:"column:created_at;comment:创建时间" json:"createdAt"`
	ClientName   string     `gorm:"-" json:"clientName,omitempty"`
	PlatformName string     `gorm:"-" json:"platformAppName,omitempty"`
}

func (LoginLog) TableName() string {
	return "plugin_clientapp_login_log"
}

func NewLoginLog() *LoginLog {
	return &LoginLog{}
}

type LoginLogList []LoginLog

func NewLoginLogList() *LoginLogList {
	return &LoginLogList{}
}

func (m *LoginLog) Create(c context.Context) error {
	return app.DB().WithContext(c).Create(m).Error
}

func (l *LoginLogList) Find(c context.Context, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&LoginLog{}).Scopes(scopes...).Order("id desc").Find(l).Error
}

func (l *LoginLogList) GetTotal(c context.Context, scopes ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&LoginLog{}).Scopes(scopes...).Count(&count).Error
	return count, err
}
