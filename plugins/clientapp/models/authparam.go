package models

import (
	"gin-fast/app/models"

	"github.com/gin-gonic/gin"
)

type AuthBaseRequest struct {
	TenantID   uint   `json:"tenantId" form:"tenantId"`
	TenantCode string `json:"tenantCode" form:"tenantCode"`
	ClientKey  string `json:"clientKey" form:"clientKey" validate:"required" message:"客户端Key不能为空"`
}

type MiniProgramLoginRequest struct {
	models.Validator
	AuthBaseRequest
	Platform     string `json:"platform" validate:"required" message:"平台不能为空"`
	Code         string `json:"code"`
	IdentityKey  string `json:"identityKey"`
	UnionKey     string `json:"unionKey"`
	Nickname     string `json:"nickname"`
	Avatar       string `json:"avatar"`
	Gender       int8   `json:"gender"`
	DeviceInfo   string `json:"deviceInfo"`
	LoginChannel string `json:"loginChannel"`
}

func (r *MiniProgramLoginRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

type WalletNonceRequest struct {
	models.Validator
	AuthBaseRequest
	Address string `form:"address" json:"address" validate:"required" message:"钱包地址不能为空"`
	ChainID string `form:"chainId" json:"chainId" validate:"required" message:"链ID不能为空"`
}

func (r *WalletNonceRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

type WalletLoginRequest struct {
	models.Validator
	AuthBaseRequest
	Address      string `json:"address" validate:"required" message:"钱包地址不能为空"`
	ChainID      string `json:"chainId" validate:"required" message:"链ID不能为空"`
	Nonce        string `json:"nonce" validate:"required" message:"Nonce不能为空"`
	Signature    string `json:"signature" validate:"required" message:"签名不能为空"`
	Nickname     string `json:"nickname"`
	Avatar       string `json:"avatar"`
	DeviceInfo   string `json:"deviceInfo"`
	LoginChannel string `json:"loginChannel"`
}

func (r *WalletLoginRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

type AppConfigGetRequest struct {
	models.Validator
	AuthBaseRequest
	PageCode string `json:"pageCode" form:"pageCode"`
}

func (r *AppConfigGetRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}
