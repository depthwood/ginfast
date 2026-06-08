package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gin-fast/app/global/app"
	"gin-fast/plugins/clientapp/models"
	"gin-fast/plugins/clientapp/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const appConfigCacheTTL = 60 * time.Second
const appConfigPageCacheTTL = 120 * time.Second

type AppConfigService struct {
	clientService *ClientService
	authService   *AuthService
}

func NewAppConfigService() *AppConfigService {
	return &AppConfigService{
		clientService: NewClientService(),
		authService:   NewAuthService(),
	}
}

func (s *AppConfigService) tenantScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return utils.TenantScope(c)
}

func (s *AppConfigService) fillClientInfo(c *gin.Context, item *models.AppConfig) {
	client := models.NewClient()
	if err := client.GetByID(c, item.ClientID, s.tenantScope(c)); err == nil && !client.IsEmpty() {
		item.ClientName = client.ClientName
		item.ClientKey = client.ClientKey
	}
}

// ==================== 基础 CRUD ====================

func (s *AppConfigService) List(c *gin.Context, req models.AppConfigListRequest) (*models.AppConfigList, int64, error) {
	list := models.NewAppConfigList()
	scopes := []func(*gorm.DB) *gorm.DB{s.tenantScope(c), req.Handle()}
	total, err := list.GetTotal(c, scopes...)
	if err != nil {
		return nil, 0, err
	}
	if err := list.Find(c, append(scopes, req.Paginate())...); err != nil {
		return nil, 0, err
	}
	for i := range *list {
		s.fillClientInfo(c, &(*list)[i])
	}
	return list, total, nil
}

func (s *AppConfigService) GetByID(c *gin.Context, id uint) (*models.AppConfig, error) {
	item := models.NewAppConfig()
	if err := item.GetByID(c, id, s.tenantScope(c)); err != nil {
		return nil, err
	}
	if item.IsEmpty() {
		return nil, models.ErrAppConfigNotFound
	}
	s.fillClientInfo(c, item)
	return item, nil
}

// Save 保存配置（支持多方案，configKey 区分方案）
func (s *AppConfigService) Save(c *gin.Context, req models.AppConfigSaveRequest) (*models.AppConfig, error) {
	client, err := s.clientService.EnsureClientBelongsToTenant(c, req.ClientID)
	if err != nil {
		return nil, err
	}

	configKey := req.ConfigKey
	if configKey == "" {
		configKey = "default"
	}

	item := models.NewAppConfig()
	if req.ID > 0 {
		if err := item.GetByID(c, req.ID, s.tenantScope(c)); err != nil {
			return nil, err
		}
		if item.IsEmpty() {
			return nil, models.ErrAppConfigNotFound
		}
	} else {
		// 按 configKey 查找已有方案，找不到则新建
		if err := item.FindByKey(c, client.TenantID, req.ClientID, configKey, s.tenantScope(c)); err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
		if item.ID == 0 {
			item.TenantID = client.TenantID
			item.ClientID = req.ClientID
			item.ConfigKey = configKey
			// 第一个方案自动激活
			item.IsActive = true
		}
	}

	item.Name = req.Name
	item.Status = req.Status
	item.Theme = models.JSONString(normalizeJSON(req.Theme, "{}"))
	item.Pages = models.JSONString(normalizeJSON(req.Pages, "[]"))
	item.FeatureFlags = models.JSONString(normalizeJSON(req.FeatureFlags, "{}"))
	item.Navigation = models.JSONString(normalizeJSON(req.Navigation, "[]"))
	item.Extra = models.JSONString(normalizeJSON(req.Extra, `{"shared":{},"pages":{}}`))
	item.CacheVersion = fmt.Sprintf("%d", time.Now().Unix())
	item.Remark = req.Remark

	if item.ID == 0 {
		if err := item.Create(c); err != nil {
			return nil, err
		}
	} else if err := item.Update(c); err != nil {
		return nil, err
	}
	s.invalidateCache(client.TenantID, req.ClientID)
	s.fillClientInfo(c, item)
	return item, nil
}

func (s *AppConfigService) Publish(c *gin.Context, req models.AppConfigSaveRequest) (*models.AppConfig, error) {
	item, err := s.Save(c, req)
	if err != nil {
		return nil, err
	}
	item.EffectiveConfig = models.JSONString(s.mergeConfig(item))
	item.EffectiveETag = hashText(string(item.EffectiveConfig))
	item.EffectiveCacheTTL = int(appConfigCacheTTL.Seconds())
	return item, nil
}

func (s *AppConfigService) UpdateStatus(c *gin.Context, id uint, status int8) error {
	item, err := s.GetByID(c, id)
	if err != nil {
		return err
	}
	item.Status = status
	item.CacheVersion = fmt.Sprintf("%d", time.Now().Unix())
	if err := item.Update(c); err != nil {
		return err
	}
	s.invalidateCache(item.TenantID, item.ClientID)
	return nil
}

func (s *AppConfigService) Delete(c *gin.Context, id uint) error {
	item, err := s.GetByID(c, id)
	if err != nil {
		return err
	}
	if err := item.Delete(c); err != nil {
		return err
	}
	s.invalidateCache(item.TenantID, item.ClientID)
	return nil
}

// ==================== 多页面装修 ====================

// GetPageConfig 获取指定页面的装修数据
func (s *AppConfigService) GetPageConfig(c *gin.Context, id uint, pageCode string) (map[string]interface{}, error) {
	item, err := s.GetByID(c, id)
	if err != nil {
		return nil, err
	}

	extra := parseJSON(string(item.Extra), map[string]interface{}{}).(map[string]interface{})
	shared := getMapOrEmpty(extra, "shared")
	pages := getMapOrEmpty(extra, "pages")

	pageCode = normalizePageCode(pageCode)
	pageData := getMapOrEmpty(pages, pageCode)

	// 合并 shared.modules + page.modules（page 覆盖 shared）
	result := map[string]interface{}{
		"pageCode": pageCode,
		"modules":  mergeModules(shared, pageData),
		"home":     getMapOrEmpty(pageData, "home"),
	}

	return result, nil
}

// SavePageConfig 保存指定页面的装修数据
func (s *AppConfigService) SavePageConfig(c *gin.Context, req models.AppConfigPageSaveRequest) (*models.AppConfig, error) {
	item, err := s.GetByID(c, req.ID)
	if err != nil {
		return nil, err
	}

	pageCode := normalizePageCode(req.PageCode)
	extra := parseJSON(string(item.Extra), map[string]interface{}{}).(map[string]interface{})

	// 确保结构存在
	if extra["shared"] == nil {
		extra["shared"] = map[string]interface{}{}
	}
	if extra["pages"] == nil {
		extra["pages"] = map[string]interface{}{}
	}
	pages := extra["pages"].(map[string]interface{})

	// 解析页面数据
	pageData := parseJSON(req.PageData, map[string]interface{}{}).(map[string]interface{})

	// 根据 pageCode 存储到对应位置
	if pageCode == "shared" {
		extra["shared"] = pageData
	} else {
		pages[pageCode] = pageData
		extra["pages"] = pages
	}

	item.Extra = models.JSONString(mustMarshal(extra))
	item.CacheVersion = fmt.Sprintf("%d", time.Now().Unix())

	if err := item.Update(c); err != nil {
		return nil, err
	}
	s.invalidateCache(item.TenantID, item.ClientID)
	s.invalidatePageCache(item.TenantID, item.ClientID, pageCode)
	s.fillClientInfo(c, item)
	return item, nil
}

// ==================== 方案切换 ====================

// ActivateScheme 激活指定配置方案
func (s *AppConfigService) ActivateScheme(c *gin.Context, id uint) (*models.AppConfig, error) {
	item, err := s.GetByID(c, id)
	if err != nil {
		return nil, err
	}
	if item.IsActive {
		return item, nil
	}

	// 先取消所有激活状态
	if err := models.DeactivateAllConfigs(c, item.TenantID, item.ClientID); err != nil {
		return nil, err
	}

	// 激活目标方案
	item.IsActive = true
	item.CacheVersion = fmt.Sprintf("%d", time.Now().Unix())
	if err := item.Update(c); err != nil {
		return nil, err
	}

	s.invalidateCache(item.TenantID, item.ClientID)
	return item, nil
}

// ==================== 快照/版本管理 ====================

// CreateSnapshot 创建配置快照
func (s *AppConfigService) CreateSnapshot(c *gin.Context, req models.AppConfigSnapshotCreateRequest) (*models.AppConfigSnapshot, error) {
	item, err := s.GetByID(c, req.ConfigID)
	if err != nil {
		return nil, err
	}

	// 生成版本号
	version := fmt.Sprintf("v%d", time.Now().Unix())

	snapshotData := map[string]interface{}{
		"theme":        parseJSON(string(item.Theme), map[string]interface{}{}),
		"pages":        parseJSON(string(item.Pages), []interface{}{}),
		"featureFlags": parseJSON(string(item.FeatureFlags), map[string]interface{}{}),
		"navigation":   parseJSON(string(item.Navigation), []interface{}{}),
		"extra":        parseJSON(string(item.Extra), map[string]interface{}{}),
		"cacheVersion": item.CacheVersion,
	}

	snapshot := models.NewAppConfigSnapshot()
	snapshot.TenantID = item.TenantID
	snapshot.ClientID = item.ClientID
	snapshot.ConfigID = item.ID
	snapshot.ConfigKey = item.ConfigKey
	snapshot.Name = req.Name
	snapshot.Version = version
	snapshot.SnapshotData = models.JSONString(mustMarshal(snapshotData))
	snapshot.Remark = req.Remark

	if err := snapshot.Create(c); err != nil {
		return nil, err
	}
	snapshot.ConfigName = item.Name
	return snapshot, nil
}

// ListSnapshots 获取快照列表
func (s *AppConfigService) ListSnapshots(c *gin.Context, req models.AppConfigSnapshotListRequest) (*models.AppConfigSnapshotList, int64, error) {
	list := models.NewAppConfigSnapshotList()

	if req.ConfigID > 0 {
		if err := list.FindByConfigID(c, req.ConfigID, s.tenantScope(c)); err != nil {
			return nil, 0, err
		}
	} else if req.ClientID > 0 {
		tenantID := utils.GetTenantID(c)
		if err := list.FindByClient(c, tenantID, req.ClientID, s.tenantScope(c)); err != nil {
			return nil, 0, err
		}
	} else {
		return list, 0, nil
	}

	total := int64(len(*list))
	return list, total, nil
}

// RestoreFromSnapshot 从快照恢复配置
func (s *AppConfigService) RestoreFromSnapshot(c *gin.Context, req models.AppConfigSnapshotRestoreRequest) (*models.AppConfig, error) {
	snapshot := models.NewAppConfigSnapshot()
	if err := snapshot.GetByID(c, req.SnapshotID, s.tenantScope(c)); err != nil {
		return nil, err
	}
	if snapshot.IsEmpty() {
		return nil, models.ErrSnapshotNotFound
	}

	// 获取关联的配置
	item, err := s.GetByID(c, snapshot.ConfigID)
	if err != nil {
		return nil, err
	}

	// 解析快照数据
	data := parseJSON(string(snapshot.SnapshotData), map[string]interface{}{}).(map[string]interface{})

	// 恢复配置
	if theme, ok := data["theme"]; ok {
		item.Theme = models.JSONString(mustMarshal(theme))
	}
	if pages, ok := data["pages"]; ok {
		item.Pages = models.JSONString(mustMarshal(pages))
	}
	if flags, ok := data["featureFlags"]; ok {
		item.FeatureFlags = models.JSONString(mustMarshal(flags))
	}
	if nav, ok := data["navigation"]; ok {
		item.Navigation = models.JSONString(mustMarshal(nav))
	}
	if extra, ok := data["extra"]; ok {
		item.Extra = models.JSONString(mustMarshal(extra))
	}
	item.CacheVersion = fmt.Sprintf("%d", time.Now().Unix())

	if err := item.Update(c); err != nil {
		return nil, err
	}

	s.invalidateCache(item.TenantID, item.ClientID)
	s.fillClientInfo(c, item)
	return item, nil
}

// DeleteSnapshot 删除快照
func (s *AppConfigService) DeleteSnapshot(c *gin.Context, id uint) error {
	snapshot := models.NewAppConfigSnapshot()
	if err := snapshot.GetByID(c, id, s.tenantScope(c)); err != nil {
		return err
	}
	if snapshot.IsEmpty() {
		return models.ErrSnapshotNotFound
	}
	return snapshot.Delete(c)
}

// ==================== AI 智能装修（多页面支持） ====================

func (s *AppConfigService) GenerateDecorationPreview(req models.AppDecorationPreviewRequest) (map[string]interface{}, error) {
	prompt := strings.ToLower(strings.TrimSpace(req.Prompt))
	pageCode := normalizePageCode(req.PageCode)

	theme := parseJSON(normalizeJSON(req.Theme, "{}"), map[string]interface{}{}).(map[string]interface{})
	pages := parseJSON(normalizeJSON(req.Pages, "[]"), []interface{}{}).([]interface{})
	featureFlags := parseJSON(normalizeJSON(req.FeatureFlags, "{}"), map[string]interface{}{}).(map[string]interface{})
	navigation := parseJSON(normalizeJSON(req.Navigation, "[]"), []interface{}{}).([]interface{})
	extra := parseJSON(normalizeJSON(req.Extra, `{"shared":{},"pages":{}}`), map[string]interface{}{}).(map[string]interface{})

	// 确保 extra 结构
	if extra["shared"] == nil {
		extra["shared"] = map[string]interface{}{}
	}
	if extra["pages"] == nil {
		extra["pages"] = map[string]interface{}{}
	}

	palette := s.pickPalette(prompt)
	for key, value := range palette {
		theme[key] = value
	}
	if _, ok := theme["brandName"]; !ok {
		theme["brandName"] = "GinFast"
	}

	pages = s.defaultPages(pages)
	navigation = s.defaultNavigation(navigation)

	// 根据 pageCode 生成对应页面的装修数据
	pageDecoration := s.buildPageDecoration(prompt, pageCode, palette)
	pagesMap := extra["pages"].(map[string]interface{})
	pagesMap[pageCode] = pageDecoration
	extra["pages"] = pagesMap

	// 根据关键词设置功能开关
	if strings.Contains(prompt, "钱包") || strings.Contains(prompt, "web3") {
		featureFlags["walletLogin"] = true
	}
	if strings.Contains(prompt, "支付") {
		featureFlags["payment"] = true
	}

	pageTitle := getPageTitle(pageCode)
	summary := s.previewSummary(prompt, pageTitle)

	return map[string]interface{}{
		"theme":        theme,
		"pages":        pages,
		"featureFlags": featureFlags,
		"navigation":   navigation,
		"extra":        extra,
		"pageCode":     pageCode,
		"summary":      summary,
	}, nil
}

// buildPageDecoration 根据页面标识生成装修数据
func (s *AppConfigService) buildPageDecoration(prompt string, pageCode string, palette map[string]interface{}) map[string]interface{} {
	switch pageCode {
	case models.PageCodeHome:
		return s.buildHomeDecoration(prompt, palette)
	case models.PageCodeWork:
		return s.buildWorkDecoration(prompt, palette)
	case models.PageCodeDiscover:
		return s.buildDiscoverDecoration(prompt, palette)
	case models.PageCodeMessage:
		return s.buildMessageDecoration(prompt, palette)
	case models.PageCodeMine:
		return s.buildMineDecoration(prompt, palette)
	default:
		return s.buildHomeDecoration(prompt, palette)
	}
}

func (s *AppConfigService) buildHomeDecoration(prompt string, palette map[string]interface{}) map[string]interface{} {
	title := "新人权益礼包"
	desc := "登录后领取专属优惠和服务权益"
	notice := "系统公告：新人权益礼包已上线，登录后即可领取。"
	if strings.Contains(prompt, "电商") || strings.Contains(prompt, "商城") {
		title = "会员专享商城"
		desc = "精选好物、优惠券和订单进度一站式查看"
		notice = "商城活动：本周会员券包限时开放领取。"
	}
	if strings.Contains(prompt, "服务") || strings.Contains(prompt, "预约") {
		title = "快捷服务预约"
		desc = "热门服务在线预约，办理进度实时提醒"
		notice = "服务提醒：预约、办理和客服进度已接入消息中心。"
	}
	primary := fmt.Sprintf("%v", palette["primaryColor"])
	return map[string]interface{}{
		"notice": notice,
		"banners": []interface{}{
			map[string]interface{}{"id": "b1", "title": title, "desc": desc, "buttonText": "立即查看", "action": "login", "mark": "NEW", "bgColor": primary},
			map[string]interface{}{"id": "b2", "title": "精选活动", "desc": "围绕你的需求推荐最新活动", "buttonText": "去看看", "action": "discover", "mark": "HOT", "bgColor": "#1677FF"},
		},
		"quickServices": []interface{}{
			map[string]interface{}{"id": "coupon", "title": "领权益", "icon": "gift", "className": "icon-red"},
			map[string]interface{}{"id": "order", "title": "查进度", "icon": "order", "className": "icon-blue"},
			map[string]interface{}{"id": "customer", "title": "联系客服", "icon": "kefu-ermai", "className": "icon-green"},
			map[string]interface{}{"id": "help", "title": "帮助中心", "icon": "question-circle", "className": "icon-purple"},
		},
		"news": []interface{}{
			map[string]interface{}{"id": "n1", "title": "服务上新", "desc": "新的装修方案已发布到 App 首页"},
			map[string]interface{}{"id": "n2", "title": "消息提醒", "desc": "登录后可接收订单进度和权益通知"},
		},
	}
}

func (s *AppConfigService) buildWorkDecoration(prompt string, palette map[string]interface{}) map[string]interface{} {
	primary := fmt.Sprintf("%v", palette["primaryColor"])
	return map[string]interface{}{
		"notice": "服务提醒：热门服务已更新，欢迎体验。",
		"categories": []interface{}{
			map[string]interface{}{"id": "cat1", "title": "热门服务", "icon": "fire", "color": primary},
			map[string]interface{}{"id": "cat2", "title": "最新上线", "icon": "new", "color": "#1677FF"},
			map[string]interface{}{"id": "cat3", "title": "限时优惠", "icon": "tag", "color": "#E53935"},
		},
		"services": []interface{}{
			map[string]interface{}{"id": "s1", "title": "在线预约", "desc": "快速预约，即时确认", "icon": "calendar"},
			map[string]interface{}{"id": "s2", "title": "进度查询", "desc": "实时追踪办理进度", "icon": "search"},
			map[string]interface{}{"id": "s3", "title": "在线客服", "desc": "7x24小时在线服务", "icon": "kefu-ermai"},
			map[string]interface{}{"id": "s4", "title": "帮助中心", "desc": "常见问题解答", "icon": "question-circle"},
		},
	}
}

func (s *AppConfigService) buildDiscoverDecoration(prompt string, palette map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"notice": "活动更新：本周精选活动已上线。",
		"featured": []interface{}{
			map[string]interface{}{"id": "f1", "title": "本周精选活动", "desc": "为你推荐的热门活动", "tag": "精选"},
			map[string]interface{}{"id": "f2", "title": "新人专享福利", "desc": "注册即享专属优惠", "tag": "NEW"},
		},
		"activities": []interface{}{
			map[string]interface{}{"id": "a1", "title": "签到有礼", "desc": "每日签到赢取积分奖励", "status": "active"},
			map[string]interface{}{"id": "a2", "title": "邀请好友", "desc": "邀请好友注册获得奖励", "status": "active"},
		},
	}
}

func (s *AppConfigService) buildMessageDecoration(prompt string, palette map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"notice": "消息中心：查看最新通知和消息。",
		"categories": []interface{}{
			map[string]interface{}{"id": "msg1", "title": "系统通知", "icon": "bell", "badge": 0},
			map[string]interface{}{"id": "msg2", "title": "服务进度", "icon": "order", "badge": 0},
			map[string]interface{}{"id": "msg3", "title": "客服消息", "icon": "chat", "badge": 0},
		},
	}
}

func (s *AppConfigService) buildMineDecoration(prompt string, palette map[string]interface{}) map[string]interface{} {
	primary := fmt.Sprintf("%v", palette["primaryColor"])
	return map[string]interface{}{
		"userCard": map[string]interface{}{
			"bgColor":    primary,
			"showAvatar": true,
			"showLevel":  true,
		},
		"menuGroups": []interface{}{
			map[string]interface{}{
				"id":    "group1",
				"title": "我的服务",
				"items": []interface{}{
					map[string]interface{}{"id": "m1", "title": "我的订单", "icon": "order"},
					map[string]interface{}{"id": "m2", "title": "我的收藏", "icon": "heart"},
					map[string]interface{}{"id": "m3", "title": "优惠券", "icon": "coupon"},
				},
			},
			map[string]interface{}{
				"id":    "group2",
				"title": "设置",
				"items": []interface{}{
					map[string]interface{}{"id": "m4", "title": "账号设置", "icon": "setting"},
					map[string]interface{}{"id": "m5", "title": "帮助中心", "icon": "question-circle"},
					map[string]interface{}{"id": "m6", "title": "关于我们", "icon": "info-circle"},
				},
			},
		},
	}
}

func (s *AppConfigService) pickPalette(prompt string) map[string]interface{} {
	if strings.Contains(prompt, "高端") || strings.Contains(prompt, "黑金") || strings.Contains(prompt, "奢华") {
		return map[string]interface{}{
			"primaryColor":      "#111827",
			"primaryLightColor": "#F8FAFC",
			"primarySoftColor":  "#E5E7EB",
			"pageBgColor":       "#F3F4F6",
			"cardBgColor":       "#FFFFFF",
			"textColor":         "#111827",
			"subTextColor":      "#6B7280",
			"borderColor":       "#E5E7EB",
		}
	}
	if strings.Contains(prompt, "清新") || strings.Contains(prompt, "绿色") || strings.Contains(prompt, "健康") {
		return map[string]interface{}{
			"primaryColor":      "#00A870",
			"primaryLightColor": "#F0FFF8",
			"primarySoftColor":  "#E1F8EF",
			"pageBgColor":       "#F6FBF8",
			"cardBgColor":       "#FFFFFF",
			"textColor":         "#173B2F",
			"subTextColor":      "#667085",
			"borderColor":       "#E3EEE9",
		}
	}
	if strings.Contains(prompt, "科技") || strings.Contains(prompt, "蓝") || strings.Contains(prompt, "saas") {
		return map[string]interface{}{
			"primaryColor":      "#1677FF",
			"primaryLightColor": "#F2F7FF",
			"primarySoftColor":  "#E8F1FF",
			"pageBgColor":       "#F6F8FC",
			"cardBgColor":       "#FFFFFF",
			"textColor":         "#1F2937",
			"subTextColor":      "#667085",
			"borderColor":       "#E5EAF3",
		}
	}
	return map[string]interface{}{
		"primaryColor":      "#E53935",
		"primaryLightColor": "#FFF7F6",
		"primarySoftColor":  "#FDECEC",
		"pageBgColor":       "#F7F8FA",
		"cardBgColor":       "#FFFFFF",
		"textColor":         "#1F2937",
		"subTextColor":      "#8A94A6",
		"borderColor":       "#EEF0F4",
	}
}

func (s *AppConfigService) defaultPages(existing []interface{}) []interface{} {
	if len(existing) > 0 {
		return existing
	}
	result := make([]interface{}, len(models.DefaultPageDefinitions))
	for i, def := range models.DefaultPageDefinitions {
		result[i] = def
	}
	return result
}

func (s *AppConfigService) defaultNavigation(existing []interface{}) []interface{} {
	if len(existing) > 0 {
		return existing
	}
	return []interface{}{
		map[string]interface{}{"id": "home", "title": "首页", "icon": "home", "activeIcon": "home-fill", "routeId": "home", "enabled": true},
		map[string]interface{}{"id": "work", "title": "服务", "icon": "grid", "activeIcon": "grid-fill", "routeId": "work", "enabled": true},
		map[string]interface{}{"id": "discover", "title": "活动", "icon": "gift", "activeIcon": "gift-fill", "routeId": "discover", "enabled": true},
		map[string]interface{}{"id": "message", "title": "消息", "icon": "chat", "activeIcon": "chat-fill", "routeId": "message", "enabled": true},
		map[string]interface{}{"id": "mine", "title": "我的", "icon": "account", "activeIcon": "account-fill", "routeId": "mine", "enabled": true},
	}
}

func (s *AppConfigService) previewSummary(prompt string, pageTitle string) string {
	if prompt == "" {
		return fmt.Sprintf("已根据默认品牌风格生成%s装修预览。", pageTitle)
	}
	return fmt.Sprintf("已根据需求生成%s装修预览：%s", pageTitle, prompt)
}

// ==================== 公开接口（C端消费） ====================

func (s *AppConfigService) PublicConfig(c *gin.Context, req models.AppConfigGetRequest) (*models.AppConfig, error) {
	tenantID, err := s.authService.ResolveTenantID(c, req.TenantID, req.TenantCode)
	if err != nil {
		return nil, err
	}
	client, err := s.authService.GetClient(c, tenantID, req.ClientKey)
	if err != nil {
		return nil, err
	}

	// 如果有 pageCode，尝试使用页面级缓存
	pageCode := normalizePageCode(req.PageCode)
	if pageCode != "" {
		return s.publicConfigWithPage(c, tenantID, client, pageCode)
	}

	// 全量配置：使用整体缓存
	cacheKey := s.cacheKey(tenantID, client.ID)
	if cached, err := app.Cache.Get(c, cacheKey); err == nil && cached != "" {
		var item models.AppConfig
		if err := json.Unmarshal([]byte(cached), &item); err == nil {
			return &item, nil
		}
	}

	item := s.findActiveOrDefault(c, tenantID, client)
	item.ClientName = client.ClientName
	item.ClientKey = client.ClientKey
	item.EffectiveConfig = models.JSONString(s.mergeConfig(item))
	item.EffectiveETag = hashText(string(item.EffectiveConfig))
	item.EffectiveCacheTTL = int(appConfigCacheTTL.Seconds())
	payload, _ := json.Marshal(item)
	_ = app.Cache.Set(c, cacheKey, string(payload), appConfigCacheTTL)
	return item, nil
}

// publicConfigWithPage 返回带指定页面数据的配置（性能优化：按需加载页面数据）
func (s *AppConfigService) publicConfigWithPage(c *gin.Context, tenantID uint, client *models.Client, pageCode string) (*models.AppConfig, error) {
	// 页面级缓存
	pageCacheKey := s.pageCacheKey(tenantID, client.ID, pageCode)
	if cached, err := app.Cache.Get(c, pageCacheKey); err == nil && cached != "" {
		var item models.AppConfig
		if err := json.Unmarshal([]byte(cached), &item); err == nil {
			return &item, nil
		}
	}

	item := s.findActiveOrDefault(c, tenantID, client)
	item.ClientName = client.ClientName
	item.ClientKey = client.ClientKey

	// 构建精简配置：主题 + 指定页面装修数据 + 导航
	mergedData := map[string]interface{}{
		"theme":        parseJSON(string(item.Theme), map[string]interface{}{}),
		"navigation":   parseJSON(string(item.Navigation), []interface{}{}),
		"pages":        parseJSON(string(item.Pages), []interface{}{}),
		"featureFlags": parseJSON(string(item.FeatureFlags), map[string]interface{}{}),
		"cacheVersion": item.CacheVersion,
	}

	// 只包含请求页面的装修数据
	extra := parseJSON(string(item.Extra), map[string]interface{}{}).(map[string]interface{})
	shared := getMapOrEmpty(extra, "shared")
	pagesMap := getMapOrEmpty(extra, "pages")
	pageData := getMapOrEmpty(pagesMap, pageCode)

	mergedData["extra"] = map[string]interface{}{
		"pages": map[string]interface{}{
			pageCode: pageData,
		},
		"shared": shared,
	}

	effectiveJSON := mustMarshal(mergedData)
	item.EffectiveConfig = models.JSONString(effectiveJSON)
	item.EffectiveETag = hashText(effectiveJSON)
	item.EffectiveCacheTTL = int(appConfigPageCacheTTL.Seconds())

	payload, _ := json.Marshal(item)
	_ = app.Cache.Set(c, pageCacheKey, string(payload), appConfigPageCacheTTL)
	return item, nil
}

// findActiveOrDefault 查找激活的配置，找不到则返回默认配置
func (s *AppConfigService) findActiveOrDefault(c *gin.Context, tenantID uint, client *models.Client) *models.AppConfig {
	item := models.NewAppConfig()

	// 优先查找激活方案
	err := item.FindActive(c, tenantID, client.ID, s.tenantScope(c))
	if err == nil && !item.IsEmpty() {
		return item
	}

	// 回退到 default key
	err = item.FindDefault(c, tenantID, client.ID, s.tenantScope(c))
	if err == nil && !item.IsEmpty() {
		return item
	}

	// 都没有，返回默认配置
	return s.defaultConfig(tenantID, client)
}

func (s *AppConfigService) defaultConfig(tenantID uint, client *models.Client) *models.AppConfig {
	pagesJSON, _ := json.Marshal(models.DefaultPageDefinitions)
	return &models.AppConfig{
		TenantID:     tenantID,
		ClientID:     client.ID,
		ConfigKey:    "default",
		Name:         "默认配置",
		Status:       1,
		IsActive:     true,
		Theme:        models.JSONString(`{"primaryColor":"#165dff","darkMode":false}`),
		Pages:        models.JSONString(string(pagesJSON)),
		FeatureFlags: models.JSONString(`{}`),
		Navigation:   models.JSONString(`[]`),
		Extra:        models.JSONString(`{"shared":{},"pages":{}}`),
		CacheVersion: "default",
	}
}

func (s *AppConfigService) mergeConfig(item *models.AppConfig) string {
	data := map[string]interface{}{
		"theme":        parseJSON(string(item.Theme), map[string]interface{}{}),
		"pages":        parseJSON(string(item.Pages), []interface{}{}),
		"featureFlags": parseJSON(string(item.FeatureFlags), map[string]interface{}{}),
		"navigation":   parseJSON(string(item.Navigation), []interface{}{}),
		"extra":        parseJSON(string(item.Extra), map[string]interface{}{}),
		"cacheVersion": item.CacheVersion,
	}
	result, _ := json.Marshal(data)
	return string(result)
}

// ==================== 缓存管理 ====================

func (s *AppConfigService) cacheKey(tenantID, clientID uint) string {
	return fmt.Sprintf("clientapp:app_config:%d:%d", tenantID, clientID)
}

func (s *AppConfigService) pageCacheKey(tenantID, clientID uint, pageCode string) string {
	return fmt.Sprintf("clientapp:app_config:%d:%d:page:%s", tenantID, clientID, pageCode)
}

func (s *AppConfigService) invalidateCache(tenantID, clientID uint) {
	_ = app.Cache.Del(context.Background(), s.cacheKey(tenantID, clientID))
	// 同时清除所有页面缓存
	for _, page := range []string{"home", "work", "discover", "message", "mine"} {
		_ = app.Cache.Del(context.Background(), s.pageCacheKey(tenantID, clientID, page))
	}
}

func (s *AppConfigService) invalidatePageCache(tenantID, clientID uint, pageCode string) {
	// 清除全量缓存（因为全量配置包含了页面数据）
	_ = app.Cache.Del(context.Background(), s.cacheKey(tenantID, clientID))
	// 清除对应页面缓存
	_ = app.Cache.Del(context.Background(), s.pageCacheKey(tenantID, clientID, pageCode))
}

// ==================== 工具函数 ====================

func normalizeJSON(raw, fallback string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	var v interface{}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return fallback
	}
	data, _ := json.Marshal(v)
	return string(data)
}

func parseJSON(raw string, fallback interface{}) interface{} {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	var v interface{}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return fallback
	}
	return v
}

func hashText(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])[:16]
}

func mustMarshal(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func normalizePageCode(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" {
		return models.PageCodeHome
	}
	return code
}

func getPageTitle(pageCode string) string {
	for _, def := range models.DefaultPageDefinitions {
		if def["id"] == pageCode {
			return def["title"].(string)
		}
	}
	return pageCode
}

func getMapOrEmpty(parent map[string]interface{}, key string) map[string]interface{} {
	if parent == nil {
		return map[string]interface{}{}
	}
	v, ok := parent[key]
	if !ok || v == nil {
		return map[string]interface{}{}
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

// mergeModules 合并 shared 模块和页面级模块（页面覆盖共享）
func mergeModules(shared, pageData map[string]interface{}) []interface{} {
	sharedModules := getModulesList(shared, "modules")
	pageModules := getModulesList(pageData, "modules")

	if len(pageModules) > 0 {
		return pageModules
	}
	return sharedModules
}

func getModulesList(data map[string]interface{}, key string) []interface{} {
	if data == nil {
		return nil
	}
	v, ok := data[key]
	if !ok || v == nil {
		return nil
	}
	if list, ok := v.([]interface{}); ok {
		return list
	}
	return nil
}
