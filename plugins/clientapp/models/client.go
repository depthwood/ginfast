package models

import (
	"context"
	"gin-fast/app/global/app"
	"gin-fast/app/models"

	"gorm.io/gorm"
)

// Client 客户端应用
type Client struct {
	models.BaseModel
	TenantID          uint       `gorm:"column:tenant_id;type:int(11);not null;comment:租户ID" json:"tenantID"`
	ClientKey         string     `gorm:"column:client_key;type:varchar(64);not null;comment:客户端Key" json:"clientKey"`
	ClientName        string     `gorm:"column:client_name;type:varchar(100);not null;comment:客户端名称" json:"clientName"`
	ClientType        string     `gorm:"column:client_type;type:varchar(32);not null;default:mini_program;comment:客户端类型" json:"clientType"`
	Status            int8       `gorm:"column:status;type:tinyint(1);not null;default:1;comment:状态" json:"status"`
	WalletEvmEnabled  int8       `gorm:"column:wallet_evm_enabled;type:tinyint(1);not null;default:0;comment:EVM钱包登录" json:"walletEvmEnabled"`
	AllowedChainIds   JSONString `gorm:"column:allowed_chain_ids;type:varchar(500);comment:允许链ID" json:"allowedChainIds"`
	WalletSignMessage string     `gorm:"column:wallet_sign_message;type:varchar(500);comment:签名模板" json:"walletSignMessage"`
	Logo              string     `gorm:"column:logo;type:varchar(500);comment:Logo" json:"logo"`
	Remark            string     `gorm:"column:remark;type:varchar(500);comment:备注" json:"remark"`
	CreatedBy         uint       `gorm:"column:created_by;type:int(11);not null;default:0;comment:创建人" json:"createdBy"`
	PlatformCount     int64      `gorm:"-" json:"platformCount,omitempty"`
	IdentityUserCount int64      `gorm:"-" json:"identityUserCount,omitempty"`
}

func (Client) TableName() string {
	return "plugin_clientapp_client"
}

func NewClient() *Client {
	return &Client{}
}

type ClientList []Client

func NewClientList() *ClientList {
	return &ClientList{}
}

func (m *Client) IsEmpty() bool {
	return m == nil || m.ID == 0
}

func (m *Client) GetByID(c context.Context, id uint, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Scopes(scopes...).First(m, id).Error
}

func (m *Client) GetByClientKey(c context.Context, tenantID uint, clientKey string, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Scopes(scopes...).Where("tenant_id = ? AND client_key = ?", tenantID, clientKey).First(m).Error
}

func (m *Client) Create(c context.Context) error {
	return app.DB().WithContext(c).Create(m).Error
}

func (m *Client) Update(c context.Context) error {
	return app.DB().WithContext(c).Save(m).Error
}

func (m *Client) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

func (l *ClientList) Find(c context.Context, scopes ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&Client{}).Scopes(scopes...).Find(l).Error
}

func (l *ClientList) GetTotal(c context.Context, scopes ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&Client{}).Scopes(scopes...).Count(&count).Error
	return count, err
}

func CountClientPlatforms(c context.Context, clientID uint) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&Platform{}).Where("client_id = ?", clientID).Count(&count).Error
	return count, err
}
