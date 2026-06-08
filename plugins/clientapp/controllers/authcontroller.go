package controllers

import (
	"gin-fast/app/controllers"
	"gin-fast/plugins/clientapp/models"
	"gin-fast/plugins/clientapp/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	controllers.Common
	authService      *service.AuthService
	appConfigService *service.AppConfigService
}

func NewAuthController() *AuthController {
	return &AuthController{
		authService:      service.NewAuthService(),
		appConfigService: service.NewAppConfigService(),
	}
}

func (ac *AuthController) MiniProgramLogin(c *gin.Context) {
	var req models.MiniProgramLoginRequest
	if err := req.Validate(c); err != nil {
		ac.FailAndAbort(c, err.Error(), err)
	}
	result, err := ac.authService.MiniProgramLogin(c, req)
	if err != nil {
		ac.FailAndAbort(c, err.Error(), err)
	}
	ac.Success(c, result)
}

func (ac *AuthController) WalletNonce(c *gin.Context) {
	var req models.WalletNonceRequest
	if err := req.Validate(c); err != nil {
		ac.FailAndAbort(c, err.Error(), err)
	}
	result, err := ac.authService.GenerateWalletNonce(c, req)
	if err != nil {
		ac.FailAndAbort(c, err.Error(), err)
	}
	ac.Success(c, result)
}

func (ac *AuthController) WalletLogin(c *gin.Context) {
	var req models.WalletLoginRequest
	if err := req.Validate(c); err != nil {
		ac.FailAndAbort(c, err.Error(), err)
	}
	result, err := ac.authService.WalletLogin(c, req)
	if err != nil {
		ac.FailAndAbort(c, err.Error(), err)
	}
	ac.Success(c, result)
}

func (ac *AuthController) AppConfig(c *gin.Context) {
	var req models.AppConfigGetRequest
	if err := req.Validate(c); err != nil {
		ac.FailAndAbort(c, err.Error(), err)
	}
	result, err := ac.appConfigService.PublicConfig(c, req)
	if err != nil {
		ac.FailAndAbort(c, err.Error(), err)
	}
	c.Header("ETag", result.EffectiveETag)
	c.Header("Cache-Control", "public, max-age=60")
	ac.Success(c, result)
}
