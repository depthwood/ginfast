package controllers

import (
	"gin-fast/app/controllers"
	"gin-fast/plugins/clientapp/models"
	"gin-fast/plugins/clientapp/service"

	"github.com/gin-gonic/gin"
)

type AppConfigController struct {
	controllers.Common
	appConfigService *service.AppConfigService
}

func NewAppConfigController() *AppConfigController {
	return &AppConfigController{
		appConfigService: service.NewAppConfigService(),
	}
}

func (cc *AppConfigController) List(c *gin.Context) {
	var req models.AppConfigListRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	list, total, err := cc.appConfigService.List(c, req)
	if err != nil {
		cc.FailAndAbort(c, "获取App配置列表失败", err)
	}
	cc.Success(c, gin.H{"list": list, "total": total})
}

func (cc *AppConfigController) GetByID(c *gin.Context) {
	var req models.AppConfigGetByIDRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	item, err := cc.appConfigService.GetByID(c, req.ID)
	if err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	cc.Success(c, item)
}

func (cc *AppConfigController) Save(c *gin.Context) {
	var req models.AppConfigSaveRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	item, err := cc.appConfigService.Save(c, req)
	if err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	cc.Success(c, gin.H{"id": item.ID})
}

func (cc *AppConfigController) Publish(c *gin.Context) {
	var req models.AppConfigSaveRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	item, err := cc.appConfigService.Publish(c, req)
	if err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	cc.Success(c, item)
}

func (cc *AppConfigController) UpdateStatus(c *gin.Context) {
	var req models.AppConfigStatusRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	if err := cc.appConfigService.UpdateStatus(c, req.ID, req.Status); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	cc.SuccessWithMessage(c, "状态更新成功")
}

func (cc *AppConfigController) Delete(c *gin.Context) {
	var req models.AppConfigDeleteRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	if err := cc.appConfigService.Delete(c, req.ID); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	cc.SuccessWithMessage(c, "删除成功")
}

func (cc *AppConfigController) DecorationPreview(c *gin.Context) {
	var req models.AppDecorationPreviewRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	preview, err := cc.appConfigService.GenerateDecorationPreview(req)
	if err != nil {
		cc.FailAndAbort(c, "生成装修预览失败", err)
	}
	cc.Success(c, preview)
}
