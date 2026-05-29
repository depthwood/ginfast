package models

import (
	"context"
	"gin-fast/app/global/app"
	"gin-fast/app/models"
	"time"

	"gorm.io/gorm"
)

// User 客户端用户
type User struct {
	models.BaseModel
	TenantID           uint           `gorm:"column:tenant_id;type:int(11);not null;comment:租户ID" json:"tenantID"`
	Nickname           string         `gorm:"column:nickname;type:varchar(100);comment:昵称" json:"nickname"`
	Avatar             string         `gorm:"column:avatar;type:varchar(500);comment:头像" json:"avatar"`
	Gender             int8           `gorm:"column:gender;type:tinyint(1);not null;default:0;comment:性别" json:"gender"`
	Status             int8           `gorm:"column:status;type:tinyint(1);not null;default:1;comment:状态" json:"status"`
	RegisterSource     string         `gorm:"column:register_source;type:varchar(32);not null;default:admin;comment:注册来源" json:"registerSource"`
	RegisterClientID   *uint          `gorm:"column:register_client_id;type:int(11);comment:注册客户端ID" json:"registerClientId"`
	RegisterPlatformID *uint          `gorm:"column:register_platform_id;type:int(11);comment:注册平台ID" json:"registerPlatformId"`
	LastLoginAt        *time.Time     `gorm:"column:last_login_at;comment:最后登录时间" json:"lastLoginAt"`
	LastLoginIP        string         `gorm:"column:last_login_ip;type:varchar(64);comment:最后登录IP" json:"lastLoginIp"`
	Remark             string         `gorm:"column:remark;type:varchar(500);comment:备注" json:"remark"`
	CreatedBy          uint           `gorm:"column:created_by;type:int(11);not null;default:0;comment:创建人" json:"createdBy"`
	IdentityCount      int64          `gorm:"-" json:"identityCount,omitempty"`
	Identities         []UserIdentity `gorm:"-" json:"identities,omitempty"`
}

func (User) TableName() string {
	return "plugin_clientapp_user"
}

func NewUser() *User {
	return &User{}
}

type UserList []User

func NewUserList() *UserList {
	return &UserList{}
}

func (m *User) IsEmpty() bool {
	return m == nil || m.ID == 0
}

func (m *User) GetByID(c context.Context, id uint, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Scopes(scopes...).First(m, id).Error
}

func (m *User) Create(c context.Context) error {
	return app.DB().WithContext(c).Create(m).Error
}

func (m *User) Update(c context.Context) error {
	return app.DB().WithContext(c).Save(m).Error
}

func (m *User) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

func (l *UserList) Find(c context.Context, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&User{}).Scopes(scopes...).Find(l).Error
}

func (l *UserList) GetTotal(c context.Context, scopes ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&User{}).Scopes(scopes...).Count(&count).Error
	return count, err
}

func CountUserIdentities(c context.Context, userID uint) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&UserIdentity{}).Where("user_id = ? AND status = 1", userID).Count(&count).Error
	return count, err
}
