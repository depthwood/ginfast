package models

import (
	"context"
	"gin-fast/app/global/app"
	"gin-fast/app/models"
	"time"

	"gorm.io/gorm"
)

// UserIdentity 用户身份绑定
type UserIdentity struct {
	models.BaseModel
	TenantID     uint       `gorm:"column:tenant_id;type:int(11);not null;comment:租户ID" json:"tenantID"`
	UserID       uint       `gorm:"column:user_id;type:int(11);not null;comment:用户ID" json:"userID"`
	ClientID     *uint      `gorm:"column:client_id;type:int(11);comment:客户端ID" json:"clientId"`
	PlatformID   *uint      `gorm:"column:platform_id;type:int(11);comment:平台渠道ID" json:"platformId"`
	Platform     string     `gorm:"column:platform;type:varchar(32);comment:平台码" json:"platform"`
	IdentityType string     `gorm:"column:identity_type;type:varchar(32);not null;comment:身份类型" json:"identityType"`
	IdentityKey  string     `gorm:"column:identity_key;type:varchar(128);not null;comment:身份标识" json:"identityKey"`
	ProviderID   string     `gorm:"column:provider_id;type:varchar(128);comment:辅助ID" json:"providerId"`
	UnionKey     string     `gorm:"column:union_key;type:varchar(128);comment:UnionID" json:"unionKey"`
	ExtraData    JSONString `gorm:"column:extra_data;type:text;comment:扩展数据" json:"extraData"`
	VerifiedAt   *time.Time `gorm:"column:verified_at;comment:验证时间" json:"verifiedAt"`
	Status       int8       `gorm:"column:status;type:tinyint(1);not null;default:1;comment:状态" json:"status"`
	PlatformName string     `gorm:"-" json:"platformAppName,omitempty"`
	ClientName   string     `gorm:"-" json:"clientName,omitempty"`
}

func (UserIdentity) TableName() string {
	return "plugin_clientapp_user_identity"
}

func NewUserIdentity() *UserIdentity {
	return &UserIdentity{}
}

type UserIdentityList []UserIdentity

func NewUserIdentityList() *UserIdentityList {
	return &UserIdentityList{}
}

func (m *UserIdentity) IsEmpty() bool {
	return m == nil || m.ID == 0
}

func (m *UserIdentity) GetByID(c context.Context, id uint, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Scopes(scopes...).First(m, id).Error
}

func (m *UserIdentity) Create(c context.Context) error {
	return app.DB().WithContext(c).Create(m).Error
}

func (m *UserIdentity) Update(c context.Context) error {
	return app.DB().WithContext(c).Save(m).Error
}

func (m *UserIdentity) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

func (l *UserIdentityList) Find(c context.Context, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&UserIdentity{}).Scopes(scopes...).Find(l).Error
}

func (l *UserIdentityList) GetTotal(c context.Context, scopes ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&UserIdentity{}).Scopes(scopes...).Count(&count).Error
	return count, err
}

func ExistsIdentity(c context.Context, tenantID uint, identityType, platform, identityKey, providerID string, platformID uint) (bool, error) {
	var count int64
	query := app.DB().WithContext(c).Model(&UserIdentity{}).
		Where("tenant_id = ? AND identity_type = ? AND platform = ? AND identity_key = ? AND status = 1",
			tenantID, identityType, platform, identityKey, providerID)
	if platformID > 0 {
		query = query.Where("platform_id = ?", platformID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}
