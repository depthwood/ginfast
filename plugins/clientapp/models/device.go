package models

import (
	"context"
	"gin-fast/app/global/app"
	"gin-fast/app/models"
	"time"

	"gorm.io/gorm"
)

// ClientDevice 客户端设备（自动发现注册）
type ClientDevice struct {
	models.BaseModel
	DeviceUUID   string     `gorm:"column:device_uuid;type:varchar(64);not null;comment:设备唯一标识" json:"deviceUUID"`
	TenantID     uint       `gorm:"column:tenant_id;type:int(11) unsigned;not null;default:0;comment:租户ID" json:"tenantID"`
	ClientID     uint       `gorm:"column:client_id;type:int(11) unsigned;not null;default:0;comment:客户端ID" json:"clientID"`
	DeviceName   string     `gorm:"column:device_name;type:varchar(100);comment:设备名称" json:"deviceName"`
	Platform     string     `gorm:"column:platform;type:varchar(32);not null;default:android;comment:平台:android/ios/web" json:"platform"`
	AppVersion   string     `gorm:"column:app_version;type:varchar(32);comment:APP版本" json:"appVersion"`
	DeviceInfo   JSONString `gorm:"column:device_info;type:text;comment:设备详情(JSON)" json:"deviceInfo"`
	Status       int8       `gorm:"column:status;type:tinyint(1);not null;default:1;comment:状态 0停用 1正常 2待审核" json:"status"`
	RegisteredAt time.Time  `gorm:"column:registered_at;comment:注册时间" json:"registeredAt"`
	LastSeenAt   time.Time  `gorm:"column:last_seen_at;comment:最后活跃时间" json:"lastSeenAt"`
	Remark       string     `gorm:"column:remark;type:varchar(500);comment:备注" json:"remark"`
	// 虚拟字段
	ClientName string `gorm:"-" json:"clientName,omitempty"`
	ClientKey  string `gorm:"-" json:"clientKey,omitempty"`
}

func (ClientDevice) TableName() string {
	return "plugin_clientapp_device"
}

func NewClientDevice() *ClientDevice {
	return &ClientDevice{}
}

type ClientDeviceList []ClientDevice

func NewClientDeviceList() *ClientDeviceList {
	return &ClientDeviceList{}
}

func (m *ClientDevice) IsEmpty() bool {
	return m == nil || m.ID == 0
}

func (m *ClientDevice) GetByID(c context.Context, id uint, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Scopes(scopes...).First(m, id).Error
}

func (m *ClientDevice) GetByUUID(c context.Context, deviceUUID string, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Scopes(scopes...).Where("device_uuid = ?", deviceUUID).First(m).Error
}

func (m *ClientDevice) Create(c context.Context) error {
	return app.DB().WithContext(c).Create(m).Error
}

func (m *ClientDevice) Update(c context.Context) error {
	return app.DB().WithContext(c).Save(m).Error
}

func (m *ClientDevice) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

func (l *ClientDeviceList) Find(c context.Context, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&ClientDevice{}).Scopes(scopes...).Find(l).Error
}

func (l *ClientDeviceList) GetTotal(c context.Context, scopes ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&ClientDevice{}).Scopes(scopes...).Count(&count).Error
	return count, err
}
