package controllers

import (
	"gin-fast/app/controllers"
	"gin-fast/plugins/clientapp/models"
	"gin-fast/plugins/clientapp/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PlatformController 平台渠道控制器
type PlatformController struct {
	controllers.Common
	platformService *service.PlatformService
}

func NewPlatformController() *PlatformController {
	return &PlatformController{
		platformService: service.NewPlatformService(),
	}
}

func (pc *PlatformController) List(c *gin.Context) {
	var req models.PlatformListRequest
	if err := req.Validate(c); err != nil {
		pc.FailAndAbort(c, err.Error(), err)
	}
	list, total, err := pc.platformService.List(c, req)
	if err != nil {
		pc.FailAndAbort(c, "获取平台渠道列表失败", err)
	}
	pc.Success(c, gin.H{"list": list, "total": total})
}

func (pc *PlatformController) GetByID(c *gin.Context) {
	var req models.PlatformGetByIDRequest
	if err := req.Validate(c); err != nil {
		pc.FailAndAbort(c, err.Error(), err)
	}
	platform, err := pc.platformService.GetByID(c, req.ID)
	if err != nil {
		pc.FailAndAbort(c, err.Error(), err)
	}
	pc.Success(c, platform)
}

func (pc *PlatformController) Create(c *gin.Context) {
	var req models.PlatformCreateRequest
	if err := req.Validate(c); err != nil {
		pc.FailAndAbort(c, err.Error(), err)
	}
	platform, err := pc.platformService.Create(c, req)
	if err != nil {
		pc.FailAndAbort(c, err.Error(), err)
	}
	pc.Success(c, gin.H{"id": platform.ID})
}

func (pc *PlatformController) Update(c *gin.Context) {
	var req models.PlatformUpdateRequest
	if err := req.Validate(c); err != nil {
		pc.FailAndAbort(c, err.Error(), err)
	}
	if err := pc.platformService.Update(c, req); err != nil {
		pc.FailAndAbort(c, err.Error(), err)
	}
	pc.SuccessWithMessage(c, "更新成功")
}

func (pc *PlatformController) UpdateStatus(c *gin.Context) {
	var req models.PlatformStatusRequest
	if err := req.Validate(c); err != nil {
		pc.FailAndAbort(c, err.Error(), err)
	}
	if err := pc.platformService.UpdateStatus(c, req.ID, req.Status); err != nil {
		pc.FailAndAbort(c, err.Error(), err)
	}
	pc.SuccessWithMessage(c, "状态更新成功")
}

func (pc *PlatformController) Delete(c *gin.Context) {
	var req models.PlatformDeleteRequest
	if err := req.Validate(c); err != nil {
		pc.FailAndAbort(c, err.Error(), err)
	}
	if err := pc.platformService.Delete(c, req.ID); err != nil {
		pc.FailAndAbort(c, err.Error(), err)
	}
	pc.SuccessWithMessage(c, "删除成功")
}

func (pc *PlatformController) Options(c *gin.Context) {
	clientID, _ := strconv.ParseUint(c.Query("clientId"), 10, 64)
	list, err := pc.platformService.ListByClientID(c, uint(clientID))
	if err != nil {
		pc.FailAndAbort(c, "获取平台渠道选项失败", err)
	}
	pc.Success(c, gin.H{"list": list})
}

func (pc *PlatformController) SupportedPlatforms(c *gin.Context) {
	pc.Success(c, gin.H{"list": models.SupportedPlatforms})
}
