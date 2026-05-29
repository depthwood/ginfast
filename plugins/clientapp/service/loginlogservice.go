package service

import (
	"gin-fast/plugins/clientapp/models"
	"gin-fast/plugins/clientapp/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LoginLogService 登录日志服务
type LoginLogService struct {
	clientService   *ClientService
	platformService *PlatformService
}

func NewLoginLogService() *LoginLogService {
	return &LoginLogService{
		clientService:   NewClientService(),
		platformService: NewPlatformService(),
	}
}

func (s *LoginLogService) tenantScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return utils.TenantScope(c)
}

func (s *LoginLogService) List(c *gin.Context, req models.LoginLogListRequest) (*models.LoginLogList, int64, error) {
	list := models.NewLoginLogList()
	scopes := []func(*gorm.DB) *gorm.DB{s.tenantScope(c), req.Handle()}
	total, err := list.GetTotal(c, scopes...)
	if err != nil {
		return nil, 0, err
	}
	if err := list.Find(c, append(scopes, req.Paginate())...); err != nil {
		return nil, 0, err
	}
	for i := range *list {
		if (*list)[i].ClientID != nil && *(*list)[i].ClientID > 0 {
			client, err := s.clientService.GetByID(c, *(*list)[i].ClientID)
			if err == nil && client != nil {
				(*list)[i].ClientName = client.ClientName
			}
		}
		if (*list)[i].PlatformID != nil && *(*list)[i].PlatformID > 0 {
			platform, err := s.platformService.GetByID(c, *(*list)[i].PlatformID)
			if err == nil && platform != nil {
				(*list)[i].PlatformName = platform.PlatformAppName
			}
		}
	}
	return list, total, nil
}
