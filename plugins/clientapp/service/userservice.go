package service

import (
	"gin-fast/app/utils/common"
	"gin-fast/plugins/clientapp/models"
	"gin-fast/plugins/clientapp/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserService 客户端用户服务
type UserService struct {
	identityService *IdentityService
	clientService   *ClientService
}

func NewUserService() *UserService {
	return &UserService{
		identityService: NewIdentityService(),
		clientService:   NewClientService(),
	}
}

func (s *UserService) tenantScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return utils.TenantScope(c)
}

func (s *UserService) List(c *gin.Context, req models.UserListRequest) (*models.UserList, int64, error) {
	list := models.NewUserList()
	scopes := []func(*gorm.DB) *gorm.DB{s.tenantScope(c), req.Handle()}
	total, err := list.GetTotal(c, scopes...)
	if err != nil {
		return nil, 0, err
	}
	if err := list.Find(c, append(scopes, req.Paginate())...); err != nil {
		return nil, 0, err
	}
	for i := range *list {
		count, _ := models.CountUserIdentities(c, (*list)[i].ID)
		(*list)[i].IdentityCount = count
	}
	return list, total, nil
}

func (s *UserService) GetByID(c *gin.Context, id uint) (*models.User, error) {
	user := models.NewUser()
	if err := user.GetByID(c, id, s.tenantScope(c)); err != nil {
		return nil, err
	}
	if user.IsEmpty() {
		return nil, models.ErrUserNotFound
	}
	identities, err := s.identityService.ListByUserID(c, id)
	if err != nil {
		return nil, err
	}
	user.Identities = identities
	count, _ := models.CountUserIdentities(c, id)
	user.IdentityCount = count
	return user, nil
}

func (s *UserService) Create(c *gin.Context, req models.UserCreateRequest) (*models.User, error) {
	if req.RegisterClientID != nil && *req.RegisterClientID > 0 {
		if _, err := s.clientService.EnsureClientBelongsToTenant(c, *req.RegisterClientID); err != nil {
			return nil, err
		}
	}

	user := models.NewUser()
	user.TenantID = common.GetCurrentTenantID(c)
	user.Nickname = req.Nickname
	user.Avatar = req.Avatar
	user.Gender = req.Gender
	user.Status = req.Status
	if user.Status == 0 {
		user.Status = 1
	}
	user.RegisterSource = models.RegisterSourceAdmin
	user.RegisterClientID = req.RegisterClientID
	user.RegisterPlatformID = req.RegisterPlatformID
	user.Remark = req.Remark
	user.CreatedBy = common.GetCurrentUserID(c)

	if err := user.Create(c); err != nil {
		return nil, err
	}

	for _, item := range req.Identities {
		var platform *models.Platform
		if item.PlatformID != nil && *item.PlatformID > 0 {
			var err error
			platform, err = s.identityService.GetPlatform(c, *item.PlatformID)
			if err != nil {
				return nil, err
			}
			if item.ClientID == nil {
				clientID := platform.ClientID
				item.ClientID = &clientID
			}
		}
		if _, err := s.identityService.CreateIdentity(c, user.ID, item, platform); err != nil {
			return nil, err
		}
	}

	return s.GetByID(c, user.ID)
}

func (s *UserService) Update(c *gin.Context, req models.UserUpdateRequest) error {
	user := models.NewUser()
	if err := user.GetByID(c, req.ID, s.tenantScope(c)); err != nil {
		return err
	}
	if user.IsEmpty() {
		return models.ErrUserNotFound
	}
	user.Nickname = req.Nickname
	user.Avatar = req.Avatar
	user.Gender = req.Gender
	user.Status = req.Status
	user.Remark = req.Remark
	return user.Update(c)
}

func (s *UserService) UpdateStatus(c *gin.Context, id uint, status int8) error {
	user := models.NewUser()
	if err := user.GetByID(c, id, s.tenantScope(c)); err != nil {
		return err
	}
	if user.IsEmpty() {
		return models.ErrUserNotFound
	}
	user.Status = status
	return user.Update(c)
}

func (s *UserService) Delete(c *gin.Context, id uint) error {
	user := models.NewUser()
	if err := user.GetByID(c, id, s.tenantScope(c)); err != nil {
		return err
	}
	if user.IsEmpty() {
		return models.ErrUserNotFound
	}
	return user.Delete(c)
}

func (s *UserService) BindIdentity(c *gin.Context, req models.UserIdentityBindRequest) (*models.UserIdentity, error) {
	user := models.NewUser()
	if err := user.GetByID(c, req.UserID, s.tenantScope(c)); err != nil {
		return nil, err
	}
	if user.IsEmpty() {
		return nil, models.ErrUserNotFound
	}

	var platform *models.Platform
	if req.PlatformID != nil && *req.PlatformID > 0 {
		var err error
		platform, err = s.identityService.GetPlatform(c, *req.PlatformID)
		if err != nil {
			return nil, err
		}
		if req.ClientID == nil {
			clientID := platform.ClientID
			req.ClientID = &clientID
		}
	}
	return s.identityService.CreateIdentity(c, req.UserID, req.IdentityInput, platform)
}

func (s *UserService) UnbindIdentity(c *gin.Context, id uint) error {
	return s.identityService.Unbind(c, id)
}
