package service

import (
	"encoding/json"
	"gin-fast/app/utils/common"
	"gin-fast/plugins/clientapp/models"
	"gin-fast/plugins/clientapp/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// IdentityService 身份绑定服务
type IdentityService struct{}

func NewIdentityService() *IdentityService {
	return &IdentityService{}
}

func (s *IdentityService) tenantScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return utils.TenantScope(c)
}

func (s *IdentityService) normalizePhone(phone string) (string, error) {
	phone = strings.TrimSpace(phone)
	if !strings.HasPrefix(phone, "+") {
		if strings.HasPrefix(phone, "86") {
			phone = "+" + phone
		} else if len(phone) == 11 && strings.HasPrefix(phone, "1") {
			phone = "+86" + phone
		}
	}
	if !strings.HasPrefix(phone, "+") || len(phone) < 8 {
		return "", models.ErrInvalidPhone
	}
	return phone, nil
}

func (s *IdentityService) ValidateAndNormalize(input models.IdentityInput, platform *models.Platform) (models.IdentityInput, error) {
	input.IdentityType = strings.TrimSpace(input.IdentityType)
	input.IdentityKey = strings.TrimSpace(input.IdentityKey)
	input.ProviderID = strings.TrimSpace(input.ProviderID)
	input.UnionKey = strings.TrimSpace(input.UnionKey)

	switch input.IdentityType {
	case models.IdentityTypeMpOpenID, models.IdentityTypeMpUnionID:
		if input.PlatformID == nil || *input.PlatformID == 0 {
			return input, models.ErrPlatformRequired
		}
		if platform == nil || platform.IsEmpty() {
			return input, models.ErrInvalidPlatform
		}
		input.Platform = platform.Platform
		if input.ProviderID == "" {
			input.ProviderID = platform.PlatformAppID
		}
	case models.IdentityTypeWalletEVM:
		address, ok := utils.NormalizeEVMAddress(input.IdentityKey)
		if !ok {
			return input, models.ErrInvalidEVMAddress
		}
		input.IdentityKey = address
		if input.ProviderID == "" {
			input.ProviderID = "1"
		}
	case models.IdentityTypePhone:
		phone, err := s.normalizePhone(input.IdentityKey)
		if err != nil {
			return input, err
		}
		input.IdentityKey = phone
	default:
		return input, models.ErrInvalidIdentityType
	}
	return input, nil
}

func (s *IdentityService) CreateIdentity(c *gin.Context, userID uint, input models.IdentityInput, platform *models.Platform) (*models.UserIdentity, error) {
	normalized, err := s.ValidateAndNormalize(input, platform)
	if err != nil {
		return nil, err
	}

	tenantID := common.GetCurrentTenantID(c)
	platformID := uint(0)
	if normalized.PlatformID != nil {
		platformID = *normalized.PlatformID
	}
	exists, err := models.ExistsIdentity(c, tenantID, normalized.IdentityType, normalized.Platform, normalized.IdentityKey, normalized.ProviderID, platformID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, models.ErrIdentityExists
	}

	now := time.Now()
	identity := models.NewUserIdentity()
	identity.TenantID = tenantID
	identity.UserID = userID
	identity.ClientID = normalized.ClientID
	identity.PlatformID = normalized.PlatformID
	identity.Platform = normalized.Platform
	identity.IdentityType = normalized.IdentityType
	identity.IdentityKey = normalized.IdentityKey
	identity.ProviderID = normalized.ProviderID
	identity.UnionKey = normalized.UnionKey
	identity.VerifiedAt = &now
	identity.Status = 1

	if err := identity.Create(c); err != nil {
		return nil, err
	}
	return identity, nil
}

func (s *IdentityService) ListByUserID(c *gin.Context, userID uint) ([]models.UserIdentity, error) {
	list := models.NewUserIdentityList()
	err := list.Find(c, s.tenantScope(c), func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND status = 1", userID).Order("id desc")
	})
	if err != nil {
		return nil, err
	}
	result := make([]models.UserIdentity, len(*list))
	for i, item := range *list {
		result[i] = item
		result[i].IdentityKey = utils.MaskIdentityKey(item.IdentityType, item.IdentityKey)
		if item.PlatformID != nil && *item.PlatformID > 0 {
			platform := models.NewPlatform()
			if err := platform.GetByID(c, *item.PlatformID, s.tenantScope(c)); err == nil && !platform.IsEmpty() {
				result[i].PlatformName = platform.PlatformAppName
			}
		}
	}
	return result, nil
}

func (s *IdentityService) Unbind(c *gin.Context, id uint) error {
	identity := models.NewUserIdentity()
	if err := identity.GetByID(c, id, s.tenantScope(c)); err != nil {
		return err
	}
	if identity.IsEmpty() {
		return models.ErrIdentityNotFound
	}
	identity.Status = 0
	return identity.Update(c)
}

func (s *IdentityService) GetPlatform(c *gin.Context, platformID uint) (*models.Platform, error) {
	platform := models.NewPlatform()
	if err := platform.GetByID(c, platformID, s.tenantScope(c)); err != nil {
		return nil, err
	}
	if platform.IsEmpty() {
		return nil, models.ErrPlatformNotFound
	}
	return platform, nil
}

func ParseFeatures(raw string) models.PlatformFeatures {
	features := models.DefaultPlatformFeatures()
	if strings.TrimSpace(raw) == "" {
		return features
	}
	_ = json.Unmarshal([]byte(raw), &features)
	return features
}

func BuildFeaturesJSON(features models.PlatformFeatures) string {
	data, _ := json.Marshal(features)
	return string(data)
}
