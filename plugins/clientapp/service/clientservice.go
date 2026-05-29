package service

import (
	"errors"
	"gin-fast/app/utils/common"
	"gin-fast/plugins/clientapp/models"
	"gin-fast/plugins/clientapp/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ClientService 客户端服务
type ClientService struct{}

func NewClientService() *ClientService {
	return &ClientService{}
}

func (s *ClientService) tenantScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return utils.TenantScope(c)
}

func (s *ClientService) List(c *gin.Context, req models.ClientListRequest) (*models.ClientList, int64, error) {
	list := models.NewClientList()
	scopes := []func(*gorm.DB) *gorm.DB{s.tenantScope(c), req.Handle()}
	total, err := list.GetTotal(c, scopes...)
	if err != nil {
		return nil, 0, err
	}
	if err := list.Find(c, append(scopes, req.Paginate())...); err != nil {
		return nil, 0, err
	}
	for i := range *list {
		platformCount, _ := models.CountClientPlatforms(c, (*list)[i].ID)
		(*list)[i].PlatformCount = platformCount
	}
	return list, total, nil
}

func (s *ClientService) GetByID(c *gin.Context, id uint) (*models.Client, error) {
	client := models.NewClient()
	if err := client.GetByID(c, id, s.tenantScope(c)); err != nil {
		return nil, err
	}
	if client.IsEmpty() {
		return nil, models.ErrClientNotFound
	}
	platformCount, _ := models.CountClientPlatforms(c, client.ID)
	client.PlatformCount = platformCount
	return client, nil
}

func (s *ClientService) Create(c *gin.Context, req models.ClientCreateRequest) (*models.Client, error) {
	tenantID := common.GetCurrentTenantID(c)
	exists := models.NewClient()
	if err := exists.GetByClientKey(c, tenantID, req.ClientKey, s.tenantScope(c)); err != nil {
		return nil, err
	}
	if !exists.IsEmpty() {
		return nil, models.ErrClientKeyExists
	}

	client := models.NewClient()
	client.TenantID = tenantID
	client.ClientKey = req.ClientKey
	client.ClientName = req.ClientName
	client.ClientType = req.ClientType
	client.Status = req.Status
	if client.Status == 0 && req.Status != 0 {
		client.Status = req.Status
	}
	if client.Status == 0 {
		client.Status = 1
	}
	client.WalletEvmEnabled = req.WalletEvmEnabled
	if req.AllowedChainIds != nil {
		client.AllowedChainIds = models.JSONString(*req.AllowedChainIds)
	}
	client.WalletSignMessage = req.WalletSignMessage
	client.Logo = req.Logo
	client.Remark = req.Remark
	client.CreatedBy = common.GetCurrentUserID(c)

	if err := client.Create(c); err != nil {
		return nil, err
	}
	return client, nil
}

func (s *ClientService) Update(c *gin.Context, req models.ClientUpdateRequest) error {
	client := models.NewClient()
	if err := client.GetByID(c, req.ID, s.tenantScope(c)); err != nil {
		return err
	}
	if client.IsEmpty() {
		return models.ErrClientNotFound
	}

	client.ClientName = req.ClientName
	client.ClientType = req.ClientType
	client.Status = req.Status
	client.WalletEvmEnabled = req.WalletEvmEnabled
	if req.AllowedChainIds != nil {
		client.AllowedChainIds = models.JSONString(*req.AllowedChainIds)
	}
	client.WalletSignMessage = req.WalletSignMessage
	client.Logo = req.Logo
	client.Remark = req.Remark
	return client.Update(c)
}

func (s *ClientService) UpdateStatus(c *gin.Context, id uint, status int8) error {
	client := models.NewClient()
	if err := client.GetByID(c, id, s.tenantScope(c)); err != nil {
		return err
	}
	if client.IsEmpty() {
		return models.ErrClientNotFound
	}
	client.Status = status
	return client.Update(c)
}

func (s *ClientService) Delete(c *gin.Context, id uint) error {
	client := models.NewClient()
	if err := client.GetByID(c, id, s.tenantScope(c)); err != nil {
		return err
	}
	if client.IsEmpty() {
		return models.ErrClientNotFound
	}
	count, err := models.CountClientPlatforms(c, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return models.ErrClientHasPlatform
	}
	return client.Delete(c)
}

func (s *ClientService) ListAll(c *gin.Context) ([]models.Client, error) {
	list := models.NewClientList()
	if err := list.Find(c, s.tenantScope(c), func(db *gorm.DB) *gorm.DB {
		return db.Where("status = 1").Order("id desc")
	}); err != nil {
		return nil, err
	}
	return []models.Client(*list), nil
}

func (s *ClientService) EnsureClientBelongsToTenant(c *gin.Context, clientID uint) (*models.Client, error) {
	if clientID == 0 {
		return nil, errors.New("客户端ID不能为空")
	}
	return s.GetByID(c, clientID)
}
