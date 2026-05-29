package controllers

import (
	"gin-fast/app/controllers"
	"gin-fast/plugins/clientapp/models"
	"gin-fast/plugins/clientapp/service"

	"github.com/gin-gonic/gin"
)

// LoginLogController 登录日志控制器
type LoginLogController struct {
	controllers.Common
	loginLogService *service.LoginLogService
}

func NewLoginLogController() *LoginLogController {
	return &LoginLogController{
		loginLogService: service.NewLoginLogService(),
	}
}

func (lc *LoginLogController) List(c *gin.Context) {
	var req models.LoginLogListRequest
	if err := req.Validate(c); err != nil {
		lc.FailAndAbort(c, err.Error(), err)
	}
	list, total, err := lc.loginLogService.List(c, req)
	if err != nil {
		lc.FailAndAbort(c, "获取登录日志失败", err)
	}
	lc.Success(c, gin.H{"list": list, "total": total})
}
