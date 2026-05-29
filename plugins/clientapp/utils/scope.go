package utils

import (
	"gin-fast/app/utils/common"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TenantScope 租户数据隔离
func TenantScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		tenantID := common.GetCurrentTenantID(c)
		if tenantID == 0 {
			return db.Where("1 = 0")
		}
		return db.Where("tenant_id = ?", tenantID)
	}
}
