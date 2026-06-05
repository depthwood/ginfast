package models

import (
	"gin-fast/app/models"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var clientKeyPattern = regexp.MustCompile(`^[a-z0-9-]{3,64}$`)

// ClientListRequest 客户端列表请求
type ClientListRequest struct {
	models.BasePaging
	models.Validator
	ClientKey  *string `form:"clientKey" json:"clientKey"`
	ClientName *string `form:"clientName" json:"clientName"`
	ClientType *string `form:"clientType" json:"clientType"`
	Status     *int8   `form:"status" json:"status"`
}

func (r *ClientListRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

func (r *ClientListRequest) Handle() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.ClientKey != nil && *r.ClientKey != "" {
			db = db.Where("client_key LIKE ?", "%"+*r.ClientKey+"%")
		}
		if r.ClientName != nil && *r.ClientName != "" {
			db = db.Where("client_name LIKE ?", "%"+*r.ClientName+"%")
		}
		if r.ClientType != nil && *r.ClientType != "" {
			db = db.Where("client_type = ?", *r.ClientType)
		}
		if r.Status != nil {
			db = db.Where("status = ?", *r.Status)
		}
		return db
	}
}

// ClientCreateRequest 创建客户端请求
type ClientCreateRequest struct {
	models.Validator
	ClientKey         string  `json:"clientKey" validate:"required" message:"客户端Key不能为空"`
	ClientName        string  `json:"clientName" validate:"required" message:"客户端名称不能为空"`
	ClientType        string  `json:"clientType" validate:"required" message:"客户端类型不能为空"`
	Status            int8    `json:"status"`
	WalletEvmEnabled  int8    `json:"walletEvmEnabled"`
	AllowedChainIds   *string `json:"allowedChainIds"`
	WalletSignMessage string  `json:"walletSignMessage"`
	Logo              string  `json:"logo"`
	Remark            string  `json:"remark"`
}

func (r *ClientCreateRequest) Validate(c *gin.Context) error {
	if err := r.Validator.Check(c, r); err != nil {
		return err
	}
	r.ClientKey = strings.TrimSpace(strings.ToLower(r.ClientKey))
	if !clientKeyPattern.MatchString(r.ClientKey) {
		return ErrInvalidClientKey
	}
	return nil
}

// ClientUpdateRequest 更新客户端请求
type ClientUpdateRequest struct {
	models.Validator
	ID                uint    `json:"id" validate:"required" message:"ID不能为空"`
	ClientName        string  `json:"clientName" validate:"required" message:"客户端名称不能为空"`
	ClientType        string  `json:"clientType" validate:"required" message:"客户端类型不能为空"`
	Status            int8    `json:"status"`
	WalletEvmEnabled  int8    `json:"walletEvmEnabled"`
	AllowedChainIds   *string `json:"allowedChainIds"`
	WalletSignMessage string  `json:"walletSignMessage"`
	Logo              string  `json:"logo"`
	Remark            string  `json:"remark"`
}

func (r *ClientUpdateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// ClientStatusRequest 更新客户端状态
type ClientStatusRequest struct {
	models.Validator
	ID     uint `json:"id" validate:"required" message:"ID不能为空"`
	Status int8 `json:"status" validate:"required" message:"状态不能为空"`
}

func (r *ClientStatusRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// ClientDeleteRequest 删除客户端
type ClientDeleteRequest struct {
	models.Validator
	ID uint `json:"id" validate:"required" message:"ID不能为空"`
}

func (r *ClientDeleteRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// ClientGetByIDRequest 客户端详情
type ClientGetByIDRequest struct {
	models.Validator
	ID uint `uri:"id" validate:"required" message:"ID不能为空"`
}

func (r *ClientGetByIDRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// PlatformListRequest 平台渠道列表
type PlatformListRequest struct {
	models.BasePaging
	models.Validator
	ClientID        *uint   `form:"clientId" json:"clientId"`
	Platform        *string `form:"platform" json:"platform"`
	PlatformAppID   *string `form:"platformAppId" json:"platformAppId"`
	PlatformAppName *string `form:"platformAppName" json:"platformAppName"`
	Status          *int8   `form:"status" json:"status"`
}

func (r *PlatformListRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

func (r *PlatformListRequest) Handle() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.ClientID != nil && *r.ClientID > 0 {
			db = db.Where("client_id = ?", *r.ClientID)
		}
		if r.Platform != nil && *r.Platform != "" {
			db = db.Where("platform = ?", *r.Platform)
		}
		if r.PlatformAppID != nil && *r.PlatformAppID != "" {
			db = db.Where("platform_app_id LIKE ?", "%"+*r.PlatformAppID+"%")
		}
		if r.PlatformAppName != nil && *r.PlatformAppName != "" {
			db = db.Where("platform_app_name LIKE ?", "%"+*r.PlatformAppName+"%")
		}
		if r.Status != nil {
			db = db.Where("status = ?", *r.Status)
		}
		return db
	}
}

// PlatformCreateRequest 创建平台渠道
type PlatformCreateRequest struct {
	models.Validator
	ClientID        uint   `json:"clientId" validate:"required" message:"客户端ID不能为空"`
	Platform        string `json:"platform" validate:"required" message:"平台不能为空"`
	PlatformAppID   string `json:"platformAppId" validate:"required" message:"平台AppID不能为空"`
	PlatformAppName string `json:"platformAppName"`
	Credentials     string `json:"credentials"`
	Features        string `json:"features"`
	Status          int8   `json:"status"`
	Remark          string `json:"remark"`
}

func (r *PlatformCreateRequest) Validate(c *gin.Context) error {
	if err := r.Validator.Check(c, r); err != nil {
		return err
	}
	if !IsSupportedPlatform(r.Platform) {
		return ErrUnsupportedPlatform
	}
	return nil
}

// PlatformUpdateRequest 更新平台渠道
type PlatformUpdateRequest struct {
	models.Validator
	ID              uint   `json:"id" validate:"required" message:"ID不能为空"`
	PlatformAppName string `json:"platformAppName"`
	Credentials     string `json:"credentials"`
	Features        string `json:"features"`
	Status          int8   `json:"status"`
	Remark          string `json:"remark"`
}

func (r *PlatformUpdateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// PlatformStatusRequest 更新平台状态
type PlatformStatusRequest struct {
	models.Validator
	ID     uint `json:"id" validate:"required" message:"ID不能为空"`
	Status int8 `json:"status" validate:"required" message:"状态不能为空"`
}

func (r *PlatformStatusRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// PlatformDeleteRequest 删除平台渠道
type PlatformDeleteRequest struct {
	models.Validator
	ID uint `json:"id" validate:"required" message:"ID不能为空"`
}

func (r *PlatformDeleteRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// PlatformGetByIDRequest 平台详情
type PlatformGetByIDRequest struct {
	models.Validator
	ID uint `uri:"id" validate:"required" message:"ID不能为空"`
}

func (r *PlatformGetByIDRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// UserListRequest 用户列表
type UserListRequest struct {
	models.BasePaging
	models.Validator
	Nickname       *string `form:"nickname" json:"nickname"`
	Status         *int8   `form:"status" json:"status"`
	RegisterSource *string `form:"registerSource" json:"registerSource"`
	ClientID       *uint   `form:"clientId" json:"clientId"`
}

func (r *UserListRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

func (r *UserListRequest) Handle() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.Nickname != nil && *r.Nickname != "" {
			db = db.Where("nickname LIKE ?", "%"+*r.Nickname+"%")
		}
		if r.Status != nil {
			db = db.Where("status = ?", *r.Status)
		}
		if r.RegisterSource != nil && *r.RegisterSource != "" {
			db = db.Where("register_source = ?", *r.RegisterSource)
		}
		if r.ClientID != nil && *r.ClientID > 0 {
			db = db.Where("register_client_id = ?", *r.ClientID)
		}
		return db
	}
}

// IdentityInput 身份绑定输入
type IdentityInput struct {
	IdentityType string `json:"identityType" validate:"required" message:"身份类型不能为空"`
	IdentityKey  string `json:"identityKey" validate:"required" message:"身份标识不能为空"`
	PlatformID   *uint  `json:"platformId"`
	ClientID     *uint  `json:"clientId"`
	ProviderID   string `json:"providerId"`
	UnionKey     string `json:"unionKey"`
	Platform     string `json:"-"`
}

// UserCreateRequest 代注册用户
type UserCreateRequest struct {
	models.Validator
	Nickname           string          `json:"nickname" validate:"required" message:"昵称不能为空"`
	Avatar             string          `json:"avatar"`
	Gender             int8            `json:"gender"`
	Status             int8            `json:"status"`
	Remark             string          `json:"remark"`
	RegisterClientID   *uint           `json:"registerClientId"`
	RegisterPlatformID *uint           `json:"registerPlatformId"`
	Identities         []IdentityInput `json:"identities"`
}

func (r *UserCreateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// UserUpdateRequest 更新用户
type UserUpdateRequest struct {
	models.Validator
	ID       uint   `json:"id" validate:"required" message:"ID不能为空"`
	Nickname string `json:"nickname" validate:"required" message:"昵称不能为空"`
	Avatar   string `json:"avatar"`
	Gender   int8   `json:"gender"`
	Status   int8   `json:"status"`
	Remark   string `json:"remark"`
}

func (r *UserUpdateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// UserStatusRequest 更新用户状态
type UserStatusRequest struct {
	models.Validator
	ID     uint `json:"id" validate:"required" message:"ID不能为空"`
	Status int8 `json:"status" validate:"required" message:"状态不能为空"`
}

func (r *UserStatusRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// UserDeleteRequest 删除用户
type UserDeleteRequest struct {
	models.Validator
	ID uint `json:"id" validate:"required" message:"ID不能为空"`
}

func (r *UserDeleteRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// UserGetByIDRequest 用户详情
type UserGetByIDRequest struct {
	models.Validator
	ID uint `uri:"id" validate:"required" message:"ID不能为空"`
}

func (r *UserGetByIDRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// UserIdentityBindRequest 绑定身份
type UserIdentityBindRequest struct {
	models.Validator
	UserID uint `json:"userId" validate:"required" message:"用户ID不能为空"`
	IdentityInput
}

func (r *UserIdentityBindRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// UserIdentityUnbindRequest 解绑身份
type UserIdentityUnbindRequest struct {
	models.Validator
	ID uint `json:"id" validate:"required" message:"身份ID不能为空"`
}

func (r *UserIdentityUnbindRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// LoginLogListRequest 登录日志列表
type LoginLogListRequest struct {
	models.BasePaging
	models.Validator
	ClientID     *uint   `form:"clientId" json:"clientId"`
	PlatformID   *uint   `form:"platformId" json:"platformId"`
	Platform     *string `form:"platform" json:"platform"`
	UserID       *uint   `form:"userId" json:"userId"`
	LoginChannel *string `form:"loginChannel" json:"loginChannel"`
	Status       *int8   `form:"status" json:"status"`
}

func (r *LoginLogListRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

func (r *LoginLogListRequest) Handle() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.ClientID != nil && *r.ClientID > 0 {
			db = db.Where("client_id = ?", *r.ClientID)
		}
		if r.PlatformID != nil && *r.PlatformID > 0 {
			db = db.Where("platform_id = ?", *r.PlatformID)
		}
		if r.Platform != nil && *r.Platform != "" {
			db = db.Where("platform = ?", *r.Platform)
		}
		if r.UserID != nil && *r.UserID > 0 {
			db = db.Where("user_id = ?", *r.UserID)
		}
		if r.LoginChannel != nil && *r.LoginChannel != "" {
			db = db.Where("login_channel = ?", *r.LoginChannel)
		}
		if r.Status != nil {
			db = db.Where("status = ?", *r.Status)
		}
		return db
	}
}

// AppConfigListRequest App界面配置列表
type AppConfigListRequest struct {
	models.BasePaging
	models.Validator
	ClientID *uint   `form:"clientId" json:"clientId"`
	Name     *string `form:"name" json:"name"`
	Status   *int8   `form:"status" json:"status"`
}

func (r *AppConfigListRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

func (r *AppConfigListRequest) Handle() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.ClientID != nil && *r.ClientID > 0 {
			db = db.Where("client_id = ?", *r.ClientID)
		}
		if r.Name != nil && *r.Name != "" {
			db = db.Where("name LIKE ?", "%"+*r.Name+"%")
		}
		if r.Status != nil {
			db = db.Where("status = ?", *r.Status)
		}
		return db
	}
}

type AppConfigSaveRequest struct {
	models.Validator
	ID           uint   `json:"id"`
	ClientID     uint   `json:"clientId" validate:"required" message:"客户端ID不能为空"`
	Name         string `json:"name" validate:"required" message:"配置名称不能为空"`
	Status       int8   `json:"status"`
	Theme        string `json:"theme"`
	Pages        string `json:"pages"`
	FeatureFlags string `json:"featureFlags"`
	Navigation   string `json:"navigation"`
	Extra        string `json:"extra"`
	Remark       string `json:"remark"`
}

func (r *AppConfigSaveRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

type AppDecorationPreviewRequest struct {
	models.Validator
	Prompt       string `json:"prompt" validate:"required" message:"装修需求不能为空"`
	Theme        string `json:"theme"`
	Pages        string `json:"pages"`
	FeatureFlags string `json:"featureFlags"`
	Navigation   string `json:"navigation"`
	Extra        string `json:"extra"`
}

func (r *AppDecorationPreviewRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

type AppConfigStatusRequest struct {
	models.Validator
	ID     uint `json:"id" validate:"required" message:"ID不能为空"`
	Status int8 `json:"status" validate:"required" message:"状态不能为空"`
}

func (r *AppConfigStatusRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

type AppConfigDeleteRequest struct {
	models.Validator
	ID uint `json:"id" validate:"required" message:"ID不能为空"`
}

func (r *AppConfigDeleteRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

type AppConfigGetByIDRequest struct {
	models.Validator
	ID uint `uri:"id" validate:"required" message:"ID不能为空"`
}

func (r *AppConfigGetByIDRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}
