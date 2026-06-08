package controllers

import (
	"fmt"
	"gin-fast/app/controllers"
	"gin-fast/plugins/clientapp/models"
	"gin-fast/plugins/clientapp/service"
	"time"

	"github.com/gin-gonic/gin"
)

// DeviceController 设备自动发现控制器
type DeviceController struct {
	controllers.Common
	discoverService *service.DeviceDiscoverService
	deviceService   *service.DeviceService
}

func NewDeviceController() *DeviceController {
	return &DeviceController{
		discoverService: service.NewDeviceDiscoverService(),
		deviceService:   service.NewDeviceService(),
	}
}

// Discover APP端自动发现（公开接口）
func (dc *DeviceController) Discover(c *gin.Context) {
	var req models.DeviceDiscoverRequest
	if err := req.Validate(c); err != nil {
		dc.FailAndAbort(c, err.Error(), err)
	}
	resp, err := dc.discoverService.Discover(c, req)
	if err != nil {
		dc.FailAndAbort(c, err.Error(), err)
	}
	dc.Success(c, resp)
}

// List 获取设备列表（管理后台）
func (dc *DeviceController) List(c *gin.Context) {
	var req models.DeviceListRequest
	if err := req.Validate(c); err != nil {
		dc.FailAndAbort(c, err.Error(), err)
	}
	list, total, err := dc.deviceService.List(c, req)
	if err != nil {
		dc.FailAndAbort(c, "获取设备列表失败", err)
	}
	dc.Success(c, gin.H{"list": list, "total": total})
}

// UpdateStatus 更新设备状态
func (dc *DeviceController) UpdateStatus(c *gin.Context) {
	var req models.DeviceStatusRequest
	if err := req.Validate(c); err != nil {
		dc.FailAndAbort(c, err.Error(), err)
	}
	if err := dc.deviceService.UpdateStatus(c, req.ID, req.Status); err != nil {
		dc.FailAndAbort(c, err.Error(), err)
	}
	dc.SuccessWithMessage(c, "状态更新成功")
}

// BindClient 绑定设备到指定客户端
func (dc *DeviceController) BindClient(c *gin.Context) {
	var req models.DeviceUpdateRequest
	if err := req.Validate(c); err != nil {
		dc.FailAndAbort(c, err.Error(), err)
	}
	if err := dc.deviceService.BindClient(c, req.ID, req.ClientID, req.Remark); err != nil {
		dc.FailAndAbort(c, err.Error(), err)
	}
	dc.SuccessWithMessage(c, "绑定成功")
}

// Delete 删除设备
func (dc *DeviceController) Delete(c *gin.Context) {
	var req models.DeviceDeleteRequest
	if err := req.Validate(c); err != nil {
		dc.FailAndAbort(c, err.Error(), err)
	}
	if err := dc.deviceService.Delete(c, req.ID); err != nil {
		dc.FailAndAbort(c, err.Error(), err)
	}
	dc.SuccessWithMessage(c, "删除成功")
}

// Heartbeat APP端心跳（公开接口，轻量更新 last_seen_at）
func (dc *DeviceController) Heartbeat(c *gin.Context) {
	var req models.DeviceHeartbeatRequest
	if err := req.Validate(c); err != nil {
		dc.FailAndAbort(c, err.Error(), err)
	}
	if err := dc.deviceService.Heartbeat(c, req.DeviceUUID); err != nil {
		dc.FailAndAbort(c, err.Error(), err)
	}
	dc.SuccessWithMessage(c, "ok")
}

// SimulateDiscover 模拟设备发现（管理后台测试用）
func (dc *DeviceController) SimulateDiscover(c *gin.Context) {
	var req struct {
		DeviceName string `json:"deviceName"`
		Platform   string `json:"platform"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.DeviceName == "" {
		req.DeviceName = "模拟测试设备"
	}
	if req.Platform == "" {
		req.Platform = "android"
	}

	uuid := dc.generateTestUUID()
	discoverReq := models.DeviceDiscoverRequest{
		DeviceUUID: uuid,
		DeviceName: req.DeviceName,
		Platform:   req.Platform,
		AppVersion: "1.0.0-test",
		DeviceInfo: `{"deviceBrand":"Test","deviceModel":"Simulator","osName":"Android","osVersion":"14"}`,
	}

	resp, err := dc.discoverService.Discover(c, discoverReq)
	if err != nil {
		dc.FailAndAbort(c, err.Error(), err)
	}
	dc.Success(c, resp)
}

func (dc *DeviceController) generateTestUUID() string {
	now := time.Now()
	return fmt.Sprintf("test-%08x-%04x-%04x-%04x-%012x",
		uint32(now.Unix()),
		uint16(now.UnixMilli()&0xFFFF),
		uint16(now.UnixNano()&0x0FFF)|0x4000,
		uint16(now.UnixMicro()&0x3FFF)|0x8000,
		uint64(now.UnixNano())&0xFFFFFFFFFFFF,
	)
}
