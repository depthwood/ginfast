package utils

import (
	"gin-fast/app/global/app"
	"gin-fast/app/utils/common"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TenantScope 租户数据隔离
func TenantScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		userID := common.GetCurrentUserID(c)
		if userID > 0 && common.IsSkipAuthUser(userID) {
			return db
		}

		tenantID := common.GetCurrentTenantID(c)
		if tenantID == 0 {
			app.ZapLog.Debug("clientapp tenant scope skipped for global tenant user")
			return db
		}
		return db.Where("tenant_id = ?", tenantID)
	}
}
