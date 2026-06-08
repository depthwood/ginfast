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
	cc.Success(c, gin.H{"id": item.ID, "configKey": item.ConfigKey, "isActive": item.IsActive})
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

// ==================== 多页面装修接口 ====================

// GetPageConfig 获取指定页面的装修数据
func (cc *AppConfigController) GetPageConfig(c *gin.Context) {
	var req models.AppConfigPageGetRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	pageData, err := cc.appConfigService.GetPageConfig(c, req.ID, req.PageCode)
	if err != nil {
		cc.FailAndAbort(c, "获取页面装修数据失败", err)
	}
	cc.Success(c, pageData)
}

// SavePageConfig 保存指定页面的装修数据
func (cc *AppConfigController) SavePageConfig(c *gin.Context) {
	var req models.AppConfigPageSaveRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	item, err := cc.appConfigService.SavePageConfig(c, req)
	if err != nil {
		cc.FailAndAbort(c, "保存页面装修数据失败", err)
	}
	cc.Success(c, gin.H{"id": item.ID, "cacheVersion": item.CacheVersion})
}

// ==================== 方案切换接口 ====================

// ActivateScheme 激活指定配置方案
func (cc *AppConfigController) ActivateScheme(c *gin.Context) {
	var req models.AppConfigActivateRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	item, err := cc.appConfigService.ActivateScheme(c, req.ID)
	if err != nil {
		cc.FailAndAbort(c, "激活配置方案失败", err)
	}
	cc.Success(c, gin.H{"id": item.ID, "configKey": item.ConfigKey, "isActive": item.IsActive})
}

// ==================== 快照/版本管理接口 ====================

// CreateSnapshot 创建配置快照
func (cc *AppConfigController) CreateSnapshot(c *gin.Context) {
	var req models.AppConfigSnapshotCreateRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	snapshot, err := cc.appConfigService.CreateSnapshot(c, req)
	if err != nil {
		cc.FailAndAbort(c, "创建快照失败", err)
	}
	cc.Success(c, snapshot)
}

// ListSnapshots 获取快照列表
func (cc *AppConfigController) ListSnapshots(c *gin.Context) {
	var req models.AppConfigSnapshotListRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	list, total, err := cc.appConfigService.ListSnapshots(c, req)
	if err != nil {
		cc.FailAndAbort(c, "获取快照列表失败", err)
	}
	cc.Success(c, gin.H{"list": list, "total": total})
}

// RestoreFromSnapshot 从快照恢复配置
func (cc *AppConfigController) RestoreFromSnapshot(c *gin.Context) {
	var req models.AppConfigSnapshotRestoreRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	item, err := cc.appConfigService.RestoreFromSnapshot(c, req)
	if err != nil {
		cc.FailAndAbort(c, "从快照恢复失败", err)
	}
	cc.Success(c, gin.H{"id": item.ID, "cacheVersion": item.CacheVersion})
}

// DeleteSnapshot 删除快照
func (cc *AppConfigController) DeleteSnapshot(c *gin.Context) {
	var req models.AppConfigSnapshotDeleteRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	if err := cc.appConfigService.DeleteSnapshot(c, req.ID); err != nil {
		cc.FailAndAbort(c, "删除快照失败", err)
	}
	cc.SuccessWithMessage(c, "快照删除成功")
}
