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

func (s *AppConfigService) Save(c *gin.Context, req models.AppConfigSaveRequest) (*models.AppConfig, error) {
	client, err := s.clientService.EnsureClientBelongsToTenant(c, req.ClientID)
	if err != nil {
		return nil, err
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
		if err := item.FindDefault(c, client.TenantID, req.ClientID, s.tenantScope(c)); err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
		if item.ID == 0 {
			item.TenantID = client.TenantID
			item.ClientID = req.ClientID
			item.ConfigKey = "default"
		}
	}

	item.Name = req.Name
	item.Status = req.Status
	item.Theme = models.JSONString(normalizeJSON(req.Theme, "{}"))
	item.Pages = models.JSONString(normalizeJSON(req.Pages, "[]"))
	item.FeatureFlags = models.JSONString(normalizeJSON(req.FeatureFlags, "{}"))
	item.Navigation = models.JSONString(normalizeJSON(req.Navigation, "[]"))
	item.Extra = models.JSONString(normalizeJSON(req.Extra, "{}"))
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

func (s *AppConfigService) GenerateDecorationPreview(req models.AppDecorationPreviewRequest) (map[string]interface{}, error) {
	prompt := strings.ToLower(strings.TrimSpace(req.Prompt))
	theme := parseJSON(normalizeJSON(req.Theme, "{}"), map[string]interface{}{}).(map[string]interface{})
	pages := parseJSON(normalizeJSON(req.Pages, "[]"), []interface{}{}).([]interface{})
	featureFlags := parseJSON(normalizeJSON(req.FeatureFlags, "{}"), map[string]interface{}{}).(map[string]interface{})
	navigation := parseJSON(normalizeJSON(req.Navigation, "[]"), []interface{}{}).([]interface{})
	extra := parseJSON(normalizeJSON(req.Extra, "{}"), map[string]interface{}{}).(map[string]interface{})

	palette := s.pickPalette(prompt)
	for key, value := range palette {
		theme[key] = value
	}
	if _, ok := theme["brandName"]; !ok {
		theme["brandName"] = "GinFast"
	}

	pages = s.defaultPages(pages)
	navigation = s.defaultNavigation(navigation)
	extra["home"] = s.buildHomeDecoration(prompt, palette)

	if strings.Contains(prompt, "钱包") || strings.Contains(prompt, "web3") {
		featureFlags["walletLogin"] = true
	}
	if strings.Contains(prompt, "支付") {
		featureFlags["payment"] = true
	}

	return map[string]interface{}{
		"theme":        theme,
		"pages":        pages,
		"featureFlags": featureFlags,
		"navigation":   navigation,
		"extra":        extra,
		"summary":      s.previewSummary(prompt),
	}, nil
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
	return []interface{}{
		map[string]interface{}{"id": "home", "title": "首页", "path": "/pages/index/index", "enabled": true},
		map[string]interface{}{"id": "work", "title": "服务", "path": "/pages/work/work", "enabled": true},
		map[string]interface{}{"id": "discover", "title": "活动", "path": "/pages/discover/discover", "enabled": true},
		map[string]interface{}{"id": "message", "title": "消息", "path": "/pages/message/message", "enabled": true},
		map[string]interface{}{"id": "mine", "title": "我的", "path": "/pages/mine/mine", "enabled": true},
	}
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

func (s *AppConfigService) previewSummary(prompt string) string {
	if prompt == "" {
		return "已根据默认品牌风格生成首页装修预览。"
	}
	return "已根据需求生成首页装修预览：" + prompt
}

func (s *AppConfigService) PublicConfig(c *gin.Context, req models.AppConfigGetRequest) (*models.AppConfig, error) {
	tenantID, err := s.authService.ResolveTenantID(c, req.TenantID, req.TenantCode)
	if err != nil {
		return nil, err
	}
	client, err := s.authService.GetClient(c, tenantID, req.ClientKey)
	if err != nil {
		return nil, err
	}
	cacheKey := s.cacheKey(tenantID, client.ID)
	if cached, err := app.Cache.Get(c, cacheKey); err == nil && cached != "" {
		var item models.AppConfig
		if err := json.Unmarshal([]byte(cached), &item); err == nil {
			return &item, nil
		}
	}
	item := models.NewAppConfig()
	err = app.DB().WithContext(c).Where("tenant_id = ? AND client_id = ? AND config_key = ? AND status = 1",
		tenantID, client.ID, "default").First(item).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if item.ID == 0 {
		item = s.defaultConfig(tenantID, client)
	}
	item.ClientName = client.ClientName
	item.ClientKey = client.ClientKey
	item.EffectiveConfig = models.JSONString(s.mergeConfig(item))
	item.EffectiveETag = hashText(string(item.EffectiveConfig))
	item.EffectiveCacheTTL = int(appConfigCacheTTL.Seconds())
	payload, _ := json.Marshal(item)
	_ = app.Cache.Set(c, cacheKey, string(payload), appConfigCacheTTL)
	return item, nil
}

func (s *AppConfigService) defaultConfig(tenantID uint, client *models.Client) *models.AppConfig {
	return &models.AppConfig{
		TenantID:     tenantID,
		ClientID:     client.ID,
		ConfigKey:    "default",
		Name:         "默认配置",
		Status:       1,
		Theme:        models.JSONString(`{"primaryColor":"#165dff","darkMode":false}`),
		Pages:        models.JSONString(`[]`),
		FeatureFlags: models.JSONString(`{}`),
		Navigation:   models.JSONString(`[]`),
		Extra:        models.JSONString(`{}`),
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

func (s *AppConfigService) cacheKey(tenantID, clientID uint) string {
	return fmt.Sprintf("clientapp:app_config:%d:%d", tenantID, clientID)
}

func (s *AppConfigService) invalidateCache(tenantID, clientID uint) {
	_ = app.Cache.Del(context.Background(), s.cacheKey(tenantID, clientID))
}

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
