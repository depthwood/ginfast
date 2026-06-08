package models

import (
	"gin-fast/app/models"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DeviceDiscoverRequest APP端自动发现请求
type DeviceDiscoverRequest struct {
	models.Validator
	DeviceUUID string `json:"deviceUUID" validate:"required" message:"设备标识不能为空"`
	DeviceName string `json:"deviceName"`
	Platform   string `json:"platform"`
	AppVersion string `json:"appVersion"`
	DeviceInfo string `json:"deviceInfo"`
}

func (r *DeviceDiscoverRequest) Validate(c *gin.Context) error {
	if err := r.Validator.Check(c, r); err != nil {
		return err
	}
	r.DeviceUUID = strings.TrimSpace(r.DeviceUUID)
	if r.Platform == "" {
		r.Platform = "android"
	}
	return nil
}

// DeviceListRequest 设备列表请求（管理后台）
type DeviceListRequest struct {
	models.BasePaging
	models.Validator
	DeviceUUID *string `form:"deviceUUID" json:"deviceUUID"`
	ClientID   *uint   `form:"clientId" json:"clientId"`
	Platform   *string `form:"platform" json:"platform"`
	Status     *int8   `form:"status" json:"status"`
}

func (r *DeviceListRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

func (r *DeviceListRequest) Handle() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.DeviceUUID != nil && *r.DeviceUUID != "" {
			db = db.Where("device_uuid LIKE ?", "%"+*r.DeviceUUID+"%")
		}
		if r.ClientID != nil && *r.ClientID > 0 {
			db = db.Where("client_id = ?", *r.ClientID)
		}
		if r.Platform != nil && *r.Platform != "" {
			db = db.Where("platform = ?", *r.Platform)
		}
		if r.Status != nil {
			db = db.Where("status = ?", *r.Status)
		}
		return db
	}
}

// DeviceStatusRequest 设备状态更新
type DeviceStatusRequest struct {
	models.Validator
	ID     uint `json:"id" validate:"required" message:"ID不能为空"`
	Status int8 `json:"status" validate:"required" message:"状态不能为空"`
}

func (r *DeviceStatusRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// DeviceUpdateRequest 设备信息更新（管理后台绑定客户端）
type DeviceUpdateRequest struct {
	models.Validator
	ID       uint   `json:"id" validate:"required" message:"ID不能为空"`
	ClientID uint   `json:"clientId"`
	Remark   string `json:"remark"`
}

func (r *DeviceUpdateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// DeviceDeleteRequest 删除设备
type DeviceDeleteRequest struct {
	models.Validator
	ID uint `json:"id" validate:"required" message:"ID不能为空"`
}

func (r *DeviceDeleteRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// DeviceHeartbeatRequest APP端心跳请求（轻量级更新 last_seen_at）
type DeviceHeartbeatRequest struct {
	models.Validator
	DeviceUUID string `json:"deviceUUID" validate:"required" message:"设备标识不能为空"`
}

func (r *DeviceHeartbeatRequest) Validate(c *gin.Context) error {
	if err := r.Validator.Check(c, r); err != nil {
		return err
	}
	r.DeviceUUID = strings.TrimSpace(r.DeviceUUID)
	return nil
}
