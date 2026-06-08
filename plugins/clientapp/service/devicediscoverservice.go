package service

import (
	"context"
	"fmt"
	"gin-fast/app/global/app"
	appModels "gin-fast/app/models"
	clientModels "gin-fast/plugins/clientapp/models"
	cutils "gin-fast/plugins/clientapp/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DeviceDiscoverService 设备自动发现服务
// 职责：设备注册/更新 + 关联客户端解析 + 租户解析
// 不负责：AppConfig 加载（APP 应在发现完成后通过 /app/config 端点获取）
type DeviceDiscoverService struct {
	clientService *ClientService
}

func NewDeviceDiscoverService() *DeviceDiscoverService {
	return &DeviceDiscoverService{
		clientService: NewClientService(),
	}
}

// DiscoverResponse 自动发现响应（精简版：设备 + 客户端 + 租户）
type DiscoverResponse struct {
	DeviceUUID   string `json:"deviceUUID"`
	DeviceName   string `json:"deviceName"`
	Status       int8   `json:"status"`
	ClientID     uint   `json:"clientID"`
	ClientKey    string `json:"clientKey"`
	ClientName   string `json:"clientName"`
	TenantID     uint   `json:"tenantID"`
	TenantCode   string `json:"tenantCode"`
	TenantName   string `json:"tenantName"`
	RegisteredAt string `json:"registeredAt"`
	LastSeenAt   string `json:"lastSeenAt"`
	IsNew        bool   `json:"isNew"`
}

// Discover 设备自动发现（核心流程）
// 流程：解析租户 → 确保客户端 → 注册/更新设备 → 返回关联信息
func (s *DeviceDiscoverService) Discover(c *gin.Context, req clientModels.DeviceDiscoverRequest) (*DiscoverResponse, error) {
	log := app.ZapLog.Sugar()
	log.Infof("[设备发现] 开始: uuid=%s, platform=%s, deviceName=%s", req.DeviceUUID, req.Platform, req.DeviceName)

	// 1. 解析租户
	tenantID := s.resolveTenantFromHeader(c)
	clientKeyHeader := c.GetHeader("Clientid")
	log.Infof("[设备发现] 租户解析: X-Tenant-Id=%s -> tenantID=%d, Clientid=%s",
		c.GetHeader("X-Tenant-Id"), tenantID, clientKeyHeader)

	// 2. 确保客户端存在
	client, clientErr := s.ensureClientForTenant(c, tenantID, clientKeyHeader)
	if clientErr != nil {
		log.Errorf("[设备发现] 客户端创建失败: tenantID=%d, err=%v", tenantID, clientErr)
		return nil, fmt.Errorf("确保客户端存在失败: %w", clientErr)
	}
	log.Infof("[设备发现] 客户端就绪: id=%d, key=%s, tenantID=%d", client.ID, client.ClientKey, client.TenantID)

	// 3. 查找或注册设备
	device, isNew, err := s.findOrCreateDevice(c, req, client)
	if err != nil {
		return nil, err
	}

	// 4. 已注册设备检查状态
	if !isNew {
		if device.Status == clientModels.DeviceStatusDisabled {
			return nil, clientModels.ErrDeviceDisabled
		}
		// 更新活跃信息和设备信息
		s.updateDeviceActivity(c, device, req)
	}

	log.Infof("[设备发现] 完成: uuid=%s, client=%s, tenantID=%d, isNew=%v",
		device.DeviceUUID, client.ClientKey, client.TenantID, isNew)

	return s.buildResponse(c, device, client, isNew)
}

// findOrCreateDevice 查找已有设备或创建新设备
func (s *DeviceDiscoverService) findOrCreateDevice(c *gin.Context, req clientModels.DeviceDiscoverRequest, client *clientModels.Client) (*clientModels.ClientDevice, bool, error) {
	log := app.ZapLog.Sugar()

	device := clientModels.NewClientDevice()
	err := device.GetByUUID(c, req.DeviceUUID)

	if err == nil && !device.IsEmpty() {
		log.Infof("[设备发现] 已有设备: id=%d, uuid=%s, status=%d", device.ID, device.DeviceUUID, device.Status)
		return device, false, nil
	}

	// 创建设备记录
	now := time.Now()
	device = clientModels.NewClientDevice()
	device.DeviceUUID = req.DeviceUUID
	device.TenantID = client.TenantID
	device.ClientID = client.ID
	device.DeviceName = req.DeviceName
	device.Platform = req.Platform
	device.AppVersion = req.AppVersion
	if req.DeviceInfo != "" {
		device.DeviceInfo = clientModels.JSONString(req.DeviceInfo)
	}
	device.Status = clientModels.DeviceStatusActive
	device.RegisteredAt = now
	device.LastSeenAt = now

	log.Infof("[设备发现] 准备写入设备: uuid=%s, tenantID=%d, clientID=%d",
		device.DeviceUUID, device.TenantID, device.ClientID)

	if createErr := device.Create(c); createErr != nil {
		// 处理并发创建（同UUID同时注册）
		existing := clientModels.NewClientDevice()
		if getErr := existing.GetByUUID(c, req.DeviceUUID); getErr == nil && !existing.IsEmpty() {
			log.Infof("[设备发现] 并发创建已存在: id=%d, uuid=%s", existing.ID, existing.DeviceUUID)
			return existing, false, nil
		}
		log.Errorf("[设备发现] 设备写入失败: uuid=%s, err=%v", req.DeviceUUID, createErr)
		return nil, false, fmt.Errorf("设备写入数据库失败: %w", createErr)
	}

	log.Infof("[设备发现] 设备注册成功: id=%d, uuid=%s", device.ID, device.DeviceUUID)
	return device, true, nil
}

// updateDeviceActivity 更新已注册设备的活跃信息
func (s *DeviceDiscoverService) updateDeviceActivity(c context.Context, device *clientModels.ClientDevice, req clientModels.DeviceDiscoverRequest) {
	log := app.ZapLog.Sugar()
	device.LastSeenAt = time.Now()
	changed := false
	if req.AppVersion != "" && req.AppVersion != device.AppVersion {
		device.AppVersion = req.AppVersion
		changed = true
	}
	if req.DeviceInfo != "" {
		device.DeviceInfo = clientModels.JSONString(req.DeviceInfo)
		changed = true
	}
	if req.DeviceName != "" && req.DeviceName != device.DeviceName {
		device.DeviceName = req.DeviceName
		changed = true
	}
	// LastSeenAt 已变更，始终需要更新
	if updateErr := device.Update(c); updateErr != nil {
		log.Warnf("[设备发现] 设备活跃信息更新失败: id=%d, err=%v", device.ID, updateErr)
	} else if changed {
		log.Infof("[设备发现] 设备活跃信息已更新: id=%d", device.ID)
	}
}

// resolveTenantFromHeader 从请求头解析租户ID
func (s *DeviceDiscoverService) resolveTenantFromHeader(c *gin.Context) uint {
	log := app.ZapLog.Sugar()
	headerValue := c.GetHeader("X-Tenant-Id")
	if headerValue == "" {
		log.Info("[resolveTenant] X-Tenant-Id 为空，返回 tenantID=0")
		return 0
	}
	if id, parseErr := strconv.ParseUint(headerValue, 10, 64); parseErr == nil {
		log.Infof("[resolveTenant] 数字解析成功: %s -> %d", headerValue, id)
		return uint(id)
	}
	tenant := appModels.NewTenant()
	if findErr := tenant.FindByCode(c, headerValue); findErr == nil && !tenant.IsEmpty() {
		log.Infof("[resolveTenant] 编码查找成功: %s -> id=%d", headerValue, tenant.ID)
		return tenant.ID
	}
	log.Warnf("[resolveTenant] 无法解析租户: header=%s, 返回 tenantID=0", headerValue)
	return 0
}

// ensureClientForTenant 确保指定租户下存在默认客户端
func (s *DeviceDiscoverService) ensureClientForTenant(c context.Context, tenantID uint, clientKey string) (*clientModels.Client, error) {
	log := app.ZapLog.Sugar()

	// 如果前端传了 clientKey，先尝试查找已有客户端
	if clientKey != "" {
		existing := clientModels.NewClient()
		findErr := existing.GetByClientKey(c, tenantID, clientKey)
		if findErr == nil && !existing.IsEmpty() && existing.Status == 1 {
			log.Infof("[ensureClient] 找到已有客户端: id=%d, key=%s, tenantID=%d", existing.ID, existing.ClientKey, tenantID)
			return existing, nil
		}
		log.Infof("[ensureClient] 未找到客户端: key=%s, tenantID=%d, err=%v", clientKey, tenantID, findErr)
	}

	// 查找或创建默认客户端
	defaultClientKey := "auto-discover"
	found := clientModels.NewClient()
	findErr := found.GetByClientKey(c, tenantID, defaultClientKey)
	if findErr == nil && !found.IsEmpty() {
		log.Infof("[ensureClient] 找到默认客户端: id=%d, key=%s, tenantID=%d", found.ID, found.ClientKey, tenantID)
		return found, nil
	}

	// 不存在则自动创建
	newClient := clientModels.NewClient()
	newClient.TenantID = tenantID
	newClient.ClientKey = defaultClientKey
	newClient.ClientName = "自动发现客户端"
	newClient.ClientType = clientModels.ClientTypeNative
	newClient.Status = 1
	newClient.Remark = "系统自动创建，用于设备自动发现注册"

	log.Infof("[ensureClient] 准备创建默认客户端: key=%s, tenantID=%d", defaultClientKey, tenantID)
	if createErr := newClient.Create(c); createErr != nil {
		log.Warnf("[ensureClient] 创建失败(可能并发): key=%s, tenantID=%d, err=%v", defaultClientKey, tenantID, createErr)
		exists := clientModels.NewClient()
		if getErr := exists.GetByClientKey(c, tenantID, defaultClientKey); getErr == nil && !exists.IsEmpty() {
			log.Infof("[ensureClient] 并发创建已存在: id=%d, key=%s", exists.ID, exists.ClientKey)
			return exists, nil
		}
		return nil, fmt.Errorf("创建默认客户端失败: %w", createErr)
	}

	log.Infof("[ensureClient] 默认客户端创建成功: id=%d, key=%s, tenantID=%d", newClient.ID, newClient.ClientKey, tenantID)
	return newClient, nil
}

// buildResponse 构造发现响应（纯设备+客户端+租户信息，不含AppConfig）
func (s *DeviceDiscoverService) buildResponse(c context.Context, device *clientModels.ClientDevice, client *clientModels.Client, isNew bool) (*DiscoverResponse, error) {
	resp := &DiscoverResponse{
		DeviceUUID:   device.DeviceUUID,
		DeviceName:   device.DeviceName,
		Status:       device.Status,
		ClientID:     client.ID,
		ClientKey:    client.ClientKey,
		ClientName:   client.ClientName,
		TenantID:     client.TenantID,
		RegisteredAt: device.RegisteredAt.Format(time.RFC3339),
		LastSeenAt:   device.LastSeenAt.Format(time.RFC3339),
		IsNew:        isNew,
	}

	// 查找租户信息
	if client.TenantID > 0 {
		tenant := appModels.NewTenant()
		if err := tenant.FindByID(c, client.TenantID); err == nil && !tenant.IsEmpty() {
			resp.TenantCode = tenant.Code
			resp.TenantName = tenant.Name
		}
	}

	return resp, nil
}

// ---- 管理后台设备管理 ----

// DeviceService 设备管理服务（管理后台）
type DeviceService struct{}

func NewDeviceService() *DeviceService {
	return &DeviceService{}
}

func (s *DeviceService) tenantScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return cutils.TenantScope(c)
}

// List 设备列表（分页）
func (s *DeviceService) List(c *gin.Context, req clientModels.DeviceListRequest) (*clientModels.ClientDeviceList, int64, error) {
	list := clientModels.NewClientDeviceList()
	scopes := []func(*gorm.DB) *gorm.DB{s.tenantScope(c), req.Handle()}
	total, err := list.GetTotal(c, scopes...)
	if err != nil {
		return nil, 0, err
	}
	if err := list.Find(c, append(scopes, req.Paginate())...); err != nil {
		return nil, 0, err
	}
	for i := range *list {
		client := clientModels.NewClient()
		if err := client.GetByID(c, (*list)[i].ClientID); err == nil {
			(*list)[i].ClientName = client.ClientName
			(*list)[i].ClientKey = client.ClientKey
		}
	}
	return list, total, nil
}

// UpdateStatus 更新设备状态
func (s *DeviceService) UpdateStatus(c *gin.Context, id uint, status int8) error {
	device := clientModels.NewClientDevice()
	if err := device.GetByID(c, id, s.tenantScope(c)); err != nil {
		return err
	}
	if device.IsEmpty() {
		return clientModels.ErrDeviceNotFound
	}
	device.Status = status
	return device.Update(c)
}

// BindClient 绑定设备到指定客户端
func (s *DeviceService) BindClient(c *gin.Context, id uint, clientID uint, remark string) error {
	device := clientModels.NewClientDevice()
	if err := device.GetByID(c, id, s.tenantScope(c)); err != nil {
		return err
	}
	if device.IsEmpty() {
		return clientModels.ErrDeviceNotFound
	}
	if clientID > 0 {
		device.ClientID = clientID
	}
	if remark != "" {
		device.Remark = remark
	}
	return device.Update(c)
}

// Delete 删除设备
func (s *DeviceService) Delete(c *gin.Context, id uint) error {
	device := clientModels.NewClientDevice()
	if err := device.GetByID(c, id, s.tenantScope(c)); err != nil {
		return err
	}
	if device.IsEmpty() {
		return clientModels.ErrDeviceNotFound
	}
	return device.Delete(c)
}

// Heartbeat 设备心跳（轻量更新 last_seen_at）
func (s *DeviceService) Heartbeat(c *gin.Context, deviceUUID string) error {
	device := clientModels.NewClientDevice()
	if err := device.GetByUUID(c, deviceUUID); err != nil || device.IsEmpty() {
		return clientModels.ErrDeviceNotFound
	}
	if device.Status == clientModels.DeviceStatusDisabled {
		return clientModels.ErrDeviceDisabled
	}
	device.LastSeenAt = time.Now()
	return device.Update(c)
}
