package service

import (
	"gin-fast/app/global/app"
	"gin-fast/app/utils/common"
	"gin-fast/plugins/clientapp/models"
	"gin-fast/plugins/clientapp/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PlatformService 平台渠道服务
type PlatformService struct {
	clientService *ClientService
}

func NewPlatformService() *PlatformService {
	return &PlatformService{
		clientService: NewClientService(),
	}
}

func (s *PlatformService) tenantScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return utils.TenantScope(c)
}

func (s *PlatformService) maskPlatform(platform *models.Platform) {
	platform.Credentials = models.JSONString(utils.MaskCredentials(string(platform.Credentials)))
}

func (s *PlatformService) fillClientInfo(c *gin.Context, list *models.PlatformList) {
	for i := range *list {
		client := models.NewClient()
		if err := client.GetByID(c, (*list)[i].ClientID, s.tenantScope(c)); err == nil && !client.IsEmpty() {
			(*list)[i].ClientName = client.ClientName
			(*list)[i].ClientKey = client.ClientKey
		}
		s.maskPlatform(&(*list)[i])
	}
}

func (s *PlatformService) List(c *gin.Context, req models.PlatformListRequest) (*models.PlatformList, int64, error) {
	list := models.NewPlatformList()
	scopes := []func(*gorm.DB) *gorm.DB{s.tenantScope(c), req.Handle()}
	total, err := list.GetTotal(c, scopes...)
	if err != nil {
		return nil, 0, err
	}
	if err := list.Find(c, append(scopes, req.Paginate())...); err != nil {
		return nil, 0, err
	}
	s.fillClientInfo(c, list)
	return list, total, nil
}

func (s *PlatformService) GetByID(c *gin.Context, id uint) (*models.Platform, error) {
	platform := models.NewPlatform()
	if err := platform.GetByID(c, id, s.tenantScope(c)); err != nil {
		return nil, err
	}
	if platform.IsEmpty() {
		return nil, models.ErrPlatformNotFound
	}
	client := models.NewClient()
	if err := client.GetByID(c, platform.ClientID, s.tenantScope(c)); err == nil && !client.IsEmpty() {
		platform.ClientName = client.ClientName
		platform.ClientKey = client.ClientKey
	}
	s.maskPlatform(platform)
	return platform, nil
}

func (s *PlatformService) Create(c *gin.Context, req models.PlatformCreateRequest) (*models.Platform, error) {
	client, err := s.clientService.EnsureClientBelongsToTenant(c, req.ClientID)
	if err != nil {
		return nil, err
	}
	if client.IsEmpty() {
		return nil, models.ErrClientNotFound
	}

	tenantID := common.GetCurrentTenantID(c)
	var count int64
	err = app.DB().WithContext(c).Model(&models.Platform{}).
		Scopes(s.tenantScope(c)).
		Where("client_id = ? AND platform = ? AND platform_app_id = ?", req.ClientID, req.Platform, req.PlatformAppID).
		Count(&count).Error
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, models.ErrPlatformExists
	}

	features := req.Features
	if strings.TrimSpace(features) == "" {
		features = BuildFeaturesJSON(models.DefaultPlatformFeatures())
	}

	platform := models.NewPlatform()
	platform.TenantID = tenantID
	platform.ClientID = req.ClientID
	platform.Platform = req.Platform
	platform.PlatformAppID = strings.TrimSpace(req.PlatformAppID)
	platform.PlatformAppName = req.PlatformAppName
	platform.Credentials = models.JSONString(req.Credentials)
	platform.Features = models.JSONString(features)
	platform.Status = req.Status
	if platform.Status == 0 {
		platform.Status = 1
	}
	platform.Remark = req.Remark

	if err := platform.Create(c); err != nil {
		return nil, err
	}
	s.maskPlatform(platform)
	platform.ClientName = client.ClientName
	platform.ClientKey = client.ClientKey
	return platform, nil
}

func (s *PlatformService) Update(c *gin.Context, req models.PlatformUpdateRequest) error {
	platform := models.NewPlatform()
	if err := platform.GetByID(c, req.ID, s.tenantScope(c)); err != nil {
		return err
	}
	if platform.IsEmpty() {
		return models.ErrPlatformNotFound
	}

	platform.PlatformAppName = req.PlatformAppName
	if strings.TrimSpace(req.Credentials) != "" && !strings.Contains(req.Credentials, "******") {
		platform.Credentials = models.JSONString(req.Credentials)
	}
	if strings.TrimSpace(req.Features) != "" {
		platform.Features = models.JSONString(req.Features)
	}
	platform.Status = req.Status
	platform.Remark = req.Remark
	return platform.Update(c)
}

func (s *PlatformService) UpdateStatus(c *gin.Context, id uint, status int8) error {
	platform := models.NewPlatform()
	if err := platform.GetByID(c, id, s.tenantScope(c)); err != nil {
		return err
	}
	if platform.IsEmpty() {
		return models.ErrPlatformNotFound
	}
	platform.Status = status
	return platform.Update(c)
}

func (s *PlatformService) Delete(c *gin.Context, id uint) error {
	platform := models.NewPlatform()
	if err := platform.GetByID(c, id, s.tenantScope(c)); err != nil {
		return err
	}
	if platform.IsEmpty() {
		return models.ErrPlatformNotFound
	}
	count, err := models.CountPlatformIdentities(c, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return models.ErrPlatformHasIdentity
	}
	return platform.Delete(c)
}

func (s *PlatformService) ListByClientID(c *gin.Context, clientID uint) ([]models.Platform, error) {
	list := models.NewPlatformList()
	if err := list.Find(c, s.tenantScope(c), func(db *gorm.DB) *gorm.DB {
		query := db.Where("status = 1").Order("id desc")
		if clientID > 0 {
			query = query.Where("client_id = ?", clientID)
		}
		return query
	}); err != nil {
		return nil, err
	}
	s.fillClientInfo(c, list)
	return []models.Platform(*list), nil
}
