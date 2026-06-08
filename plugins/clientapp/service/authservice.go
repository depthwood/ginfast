package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"gin-fast/app/global/app"
	appmodels "gin-fast/app/models"
	"gin-fast/plugins/clientapp/models"
	"gin-fast/plugins/clientapp/utils"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const (
	ClientTokenType   = "client"
	walletNoncePrefix = "clientapp:wallet_nonce:"
)

type AuthService struct {
	httpClient *http.Client
}

type LoginResult struct {
	AccessToken        string               `json:"accessToken"`
	AccessTokenExpires int64                `json:"accessTokenExpires"`
	User               *models.User         `json:"user"`
	Identity           *models.UserIdentity `json:"identity"`
	Client             *models.Client       `json:"client"`
	Platform           *models.Platform     `json:"platform,omitempty"`
	IsNewUser          bool                 `json:"isNewUser"`
}

type ClientClaims struct {
	UserID     uint   `json:"userId"`
	TenantID   uint   `json:"tenantId"`
	ClientID   uint   `json:"clientId"`
	PlatformID uint   `json:"platformId,omitempty"`
	TokenType  string `json:"tokenType"`
	jwt.RegisteredClaims
}

type AuthContext struct {
	TenantID    uint
	Client      *models.Client
	Platform    *models.Platform
	Identity    models.IdentityInput
	Nickname    string
	Avatar      string
	Gender      int8
	RegisterSrc string
	DeviceInfo  string
	Channel     string
}

type MiniProgramSession struct {
	OpenID  string
	UnionID string
}

type MiniProgramStrategy interface {
	Login(ctx context.Context, platform *models.Platform, code string) (*MiniProgramSession, error)
}

func NewAuthService() *AuthService {
	return &AuthService{httpClient: &http.Client{Timeout: 5 * time.Second}}
}

func (s *AuthService) ResolveTenantID(ctx context.Context, tenantID uint, tenantCode string) (uint, error) {
	if tenantID > 0 {
		return tenantID, nil
	}
	tenantCode = strings.TrimSpace(tenantCode)
	if tenantCode == "" {
		return 0, errors.New("tenantId或tenantCode不能为空")
	}
	tenant := appmodels.NewTenant()
	err := tenant.Find(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ? AND status = 1", tenantCode)
	})
	if err != nil {
		return 0, err
	}
	if tenant.IsEmpty() {
		return 0, errors.New("租户不存在或未启用")
	}
	return tenant.ID, nil
}

func (s *AuthService) GetClient(ctx context.Context, tenantID uint, clientKey string) (*models.Client, error) {
	client := models.NewClient()
	err := client.GetByClientKey(ctx, tenantID, strings.TrimSpace(strings.ToLower(clientKey)))
	if err != nil {
		return nil, err
	}
	if client.IsEmpty() {
		return nil, models.ErrClientNotFound
	}
	if client.Status != 1 {
		return nil, models.ErrClientDisabled
	}
	return client, nil
}

func (s *AuthService) GetPlatform(ctx context.Context, tenantID uint, clientID uint, platformCode string) (*models.Platform, error) {
	platform := models.NewPlatform()
	err := app.DB().WithContext(ctx).Where(
		"tenant_id = ? AND client_id = ? AND platform = ? AND status = 1",
		tenantID, clientID, platformCode,
	).First(platform).Error
	if err != nil {
		return nil, err
	}
	if platform.IsEmpty() {
		return nil, models.ErrPlatformNotFound
	}
	return platform, nil
}

func (s *AuthService) MiniProgramLogin(c *gin.Context, req models.MiniProgramLoginRequest) (*LoginResult, error) {
	tenantID, err := s.ResolveTenantID(c, req.TenantID, req.TenantCode)
	if err != nil {
		return nil, err
	}
	client, err := s.GetClient(c, tenantID, req.ClientKey)
	if err != nil {
		return nil, err
	}
	platform, err := s.GetPlatform(c, tenantID, client.ID, req.Platform)
	if err != nil {
		return nil, err
	}

	identityKey := strings.TrimSpace(req.IdentityKey)
	unionKey := strings.TrimSpace(req.UnionKey)
	if identityKey == "" {
		strategy := s.NewMiniProgramStrategy(req.Platform)
		session, err := strategy.Login(c, platform, req.Code)
		if err != nil {
			s.writeLoginLog(c, AuthContext{TenantID: tenantID, Client: client, Platform: platform, Channel: req.LoginChannel, DeviceInfo: req.DeviceInfo}, nil, models.IdentityTypeMpOpenID, identityKey, 0, err.Error())
			return nil, err
		}
		identityKey = session.OpenID
		unionKey = session.UnionID
	}
	if identityKey == "" {
		return nil, models.ErrInvalidAuthCode
	}

	return s.loginByIdentity(c, AuthContext{
		TenantID: tenantID,
		Client:   client,
		Platform: platform,
		Identity: models.IdentityInput{
			IdentityType: models.IdentityTypeMpOpenID,
			IdentityKey:  identityKey,
			UnionKey:     unionKey,
			PlatformID:   &platform.ID,
			ClientID:     &client.ID,
			Platform:     platform.Platform,
			ProviderID:   platform.PlatformAppID,
		},
		Nickname:    req.Nickname,
		Avatar:      req.Avatar,
		Gender:      req.Gender,
		RegisterSrc: models.RegisterSourceMP,
		DeviceInfo:  req.DeviceInfo,
		Channel:     firstNonEmpty(req.LoginChannel, "mp"),
	})
}

func (s *AuthService) GenerateWalletNonce(c *gin.Context, req models.WalletNonceRequest) (map[string]interface{}, error) {
	tenantID, err := s.ResolveTenantID(c, req.TenantID, req.TenantCode)
	if err != nil {
		return nil, err
	}
	client, err := s.GetClient(c, tenantID, req.ClientKey)
	if err != nil {
		return nil, err
	}
	if client.WalletEvmEnabled != 1 {
		return nil, models.ErrWalletDisabled
	}
	address, ok := utils.NormalizeEVMAddress(req.Address)
	if !ok {
		return nil, models.ErrInvalidEVMAddress
	}
	if !s.chainAllowed(client, req.ChainID) {
		return nil, models.ErrInvalidChainID
	}
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, err
	}
	nonce := hex.EncodeToString(nonceBytes)
	message := s.buildWalletMessage(client, nonce, req.ChainID)
	cacheValue := fmt.Sprintf("%s|%s|%s", address, req.ChainID, message)
	if err := app.Cache.Set(c, s.walletNonceKey(tenantID, client.ID, address, nonce), cacheValue, 5*time.Minute); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"nonce":     nonce,
		"message":   message,
		"expiresIn": 300,
	}, nil
}

func (s *AuthService) WalletLogin(c *gin.Context, req models.WalletLoginRequest) (*LoginResult, error) {
	tenantID, err := s.ResolveTenantID(c, req.TenantID, req.TenantCode)
	if err != nil {
		return nil, err
	}
	client, err := s.GetClient(c, tenantID, req.ClientKey)
	if err != nil {
		return nil, err
	}
	if client.WalletEvmEnabled != 1 {
		return nil, models.ErrWalletDisabled
	}
	address, ok := utils.NormalizeEVMAddress(req.Address)
	if !ok {
		return nil, models.ErrInvalidEVMAddress
	}
	if !s.chainAllowed(client, req.ChainID) {
		return nil, models.ErrInvalidChainID
	}
	key := s.walletNonceKey(tenantID, client.ID, address, req.Nonce)
	cacheValue, err := app.Cache.Get(c, key)
	if err != nil || cacheValue == "" {
		return nil, models.ErrInvalidNonce
	}
	_ = app.Cache.Del(c, key)
	parts := strings.SplitN(cacheValue, "|", 3)
	if len(parts) != 3 || parts[0] != address || parts[1] != req.ChainID {
		return nil, models.ErrInvalidNonce
	}
	if !verifyCompatibleWalletSignature(address, parts[2], req.Signature) {
		s.writeLoginLog(c, AuthContext{TenantID: tenantID, Client: client, Channel: req.LoginChannel, DeviceInfo: req.DeviceInfo}, nil, models.IdentityTypeWalletEVM, address, 0, models.ErrInvalidSignature.Error())
		return nil, models.ErrInvalidSignature
	}
	return s.loginByIdentity(c, AuthContext{
		TenantID: tenantID,
		Client:   client,
		Identity: models.IdentityInput{
			IdentityType: models.IdentityTypeWalletEVM,
			IdentityKey:  address,
			ProviderID:   req.ChainID,
			ClientID:     &client.ID,
		},
		Nickname:    req.Nickname,
		Avatar:      req.Avatar,
		RegisterSrc: models.RegisterSourceWalletEVM,
		DeviceInfo:  req.DeviceInfo,
		Channel:     firstNonEmpty(req.LoginChannel, "wallet_evm"),
	})
}

func (s *AuthService) loginByIdentity(c *gin.Context, auth AuthContext) (*LoginResult, error) {
	var user models.User
	var identity models.UserIdentity
	isNewUser := false

	err := app.DB().WithContext(c).Transaction(func(tx *gorm.DB) error {
		query := tx.Where("tenant_id = ? AND identity_type = ? AND identity_key = ? AND status = 1",
			auth.TenantID, auth.Identity.IdentityType, auth.Identity.IdentityKey)
		if auth.Identity.Platform != "" {
			query = query.Where("platform = ?", auth.Identity.Platform)
		}
		if auth.Identity.ProviderID != "" {
			query = query.Where("provider_id = ?", auth.Identity.ProviderID)
		}
		if auth.Identity.PlatformID != nil && *auth.Identity.PlatformID > 0 {
			query = query.Where("platform_id = ?", *auth.Identity.PlatformID)
		}
		err := query.First(&identity).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		now := time.Now()
		if identity.ID > 0 {
			if err := tx.First(&user, "id = ? AND tenant_id = ? AND status = 1", identity.UserID, auth.TenantID).Error; err != nil {
				return err
			}
		} else {
			isNewUser = true
			user = models.User{
				TenantID:           auth.TenantID,
				Nickname:           firstNonEmpty(auth.Nickname, maskDefaultName(auth.Identity.IdentityKey)),
				Avatar:             auth.Avatar,
				Gender:             auth.Gender,
				Status:             1,
				RegisterSource:     auth.RegisterSrc,
				RegisterClientID:   &auth.Client.ID,
				RegisterPlatformID: nil,
				LastLoginAt:        &now,
				LastLoginIP:        c.ClientIP(),
			}
			if auth.Platform != nil {
				user.RegisterPlatformID = &auth.Platform.ID
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
			identity = models.UserIdentity{
				TenantID:     auth.TenantID,
				UserID:       user.ID,
				ClientID:     auth.Identity.ClientID,
				PlatformID:   auth.Identity.PlatformID,
				Platform:     auth.Identity.Platform,
				IdentityType: auth.Identity.IdentityType,
				IdentityKey:  auth.Identity.IdentityKey,
				ProviderID:   auth.Identity.ProviderID,
				UnionKey:     auth.Identity.UnionKey,
				VerifiedAt:   &now,
				Status:       1,
			}
			if err := tx.Create(&identity).Error; err != nil {
				return err
			}
		}
		user.LastLoginAt = &now
		user.LastLoginIP = c.ClientIP()
		return tx.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
			"last_login_at": user.LastLoginAt,
			"last_login_ip": user.LastLoginIP,
		}).Error
	})
	if err != nil {
		s.writeLoginLog(c, auth, nil, auth.Identity.IdentityType, auth.Identity.IdentityKey, 0, err.Error())
		return nil, err
	}

	token, expiresAt, err := s.generateClientToken(&user, auth)
	if err != nil {
		return nil, err
	}
	result := &LoginResult{
		AccessToken:        token,
		AccessTokenExpires: expiresAt,
		User:               &user,
		Identity:           &identity,
		Client:             auth.Client,
		Platform:           auth.Platform,
		IsNewUser:          isNewUser,
	}
	s.writeLoginLog(c, auth, &user, auth.Identity.IdentityType, auth.Identity.IdentityKey, 1, "")
	result.Identity.IdentityKey = utils.MaskIdentityKey(identity.IdentityType, identity.IdentityKey)
	return result, nil
}

func (s *AuthService) NewMiniProgramStrategy(platform string) MiniProgramStrategy {
	return &genericMiniProgramStrategy{httpClient: s.httpClient, platform: platform}
}

func (s *AuthService) chainAllowed(client *models.Client, chainID string) bool {
	var allowed []string
	if strings.TrimSpace(string(client.AllowedChainIds)) == "" {
		return true
	}
	if err := json.Unmarshal([]byte(client.AllowedChainIds), &allowed); err != nil || len(allowed) == 0 {
		return true
	}
	for _, item := range allowed {
		if item == chainID {
			return true
		}
	}
	return false
}

func (s *AuthService) buildWalletMessage(client *models.Client, nonce, chainID string) string {
	template := strings.TrimSpace(client.WalletSignMessage)
	if template == "" {
		template = "Login to {clientKey}\nNonce: {nonce}\nChain ID: {chainId}"
	}
	template = strings.ReplaceAll(template, "{clientKey}", client.ClientKey)
	template = strings.ReplaceAll(template, "{nonce}", nonce)
	template = strings.ReplaceAll(template, "{chainId}", chainID)
	return template
}

func (s *AuthService) walletNonceKey(tenantID, clientID uint, address, nonce string) string {
	return fmt.Sprintf("%s%d:%d:%s:%s", walletNoncePrefix, tenantID, clientID, address, nonce)
}

func (s *AuthService) writeLoginLog(c *gin.Context, auth AuthContext, user *models.User, identityType, identityKey string, status int8, failReason string) {
	log := &models.LoginLog{
		TenantID:     auth.TenantID,
		Platform:     "",
		IdentityType: identityType,
		IdentityKey:  utils.MaskIdentityKey(identityType, identityKey),
		LoginChannel: firstNonEmpty(auth.Channel, identityType),
		IP:           c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		DeviceInfo:   models.JSONString(auth.DeviceInfo),
		Status:       status,
		FailReason:   failReason,
		CreatedAt:    time.Now(),
	}
	if auth.Client != nil {
		log.ClientID = &auth.Client.ID
	}
	if auth.Platform != nil {
		log.PlatformID = &auth.Platform.ID
		log.Platform = auth.Platform.Platform
	}
	if user != nil {
		log.UserID = &user.ID
	}
	_ = log.Create(c)
}

type genericMiniProgramStrategy struct {
	httpClient *http.Client
	platform   string
}

func (s *genericMiniProgramStrategy) Login(ctx context.Context, platform *models.Platform, code string) (*MiniProgramSession, error) {
	if strings.TrimSpace(code) == "" {
		return nil, models.ErrInvalidAuthCode
	}
	credentials := map[string]string{}
	_ = json.Unmarshal([]byte(platform.Credentials), &credentials)
	appSecret := firstNonEmpty(credentials["appSecret"], credentials["secret"])
	appID := platform.PlatformAppID
	var endpoint string
	switch s.platform {
	case models.PlatformWechat:
		endpoint = fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", url.QueryEscape(appID), url.QueryEscape(appSecret), url.QueryEscape(code))
	case models.PlatformQQ:
		endpoint = fmt.Sprintf("https://api.q.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", url.QueryEscape(appID), url.QueryEscape(appSecret), url.QueryEscape(code))
	case models.PlatformDouyin:
		endpoint = fmt.Sprintf("https://developer.toutiao.com/api/apps/v2/jscode2session?appid=%s&secret=%s&code=%s", url.QueryEscape(appID), url.QueryEscape(appSecret), url.QueryEscape(code))
	case models.PlatformBaidu:
		endpoint = fmt.Sprintf("https://spapi.baidu.com/oauth/jscode2sessionkey?client_id=%s&sk=%s&code=%s", url.QueryEscape(appID), url.QueryEscape(appSecret), url.QueryEscape(code))
	default:
		return nil, fmt.Errorf("%s code登录暂未内置适配，请传入identityKey兼容登录", s.platform)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	openID := firstNonEmpty(asString(raw["openid"]), asString(raw["open_id"]))
	unionID := firstNonEmpty(asString(raw["unionid"]), asString(raw["union_id"]))
	if openID == "" {
		return nil, fmt.Errorf("%w: %s", models.ErrInvalidAuthCode, string(body))
	}
	return &MiniProgramSession{OpenID: openID, UnionID: unionID}, nil
}

func (s *AuthService) generateClientToken(user *models.User, auth AuthContext) (string, int64, error) {
	expire := app.ConfigYml.GetDuration("token.jwttokenexpire")
	if expire <= 0 {
		expire = 43200
	}
	expiresAt := time.Now().Add(expire * time.Second)
	claims := &ClientClaims{
		UserID:     user.ID,
		TenantID:   auth.TenantID,
		ClientID:   auth.Client.ID,
		PlatformID: platformIDValue(auth.Platform),
		TokenType:  ClientTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(app.ConfigYml.GetString("token.jwttokensignkey")))
	if err != nil {
		return "", 0, err
	}
	return tokenString, expiresAt.Unix(), nil
}

func verifyCompatibleWalletSignature(_ string, _ string, signature string) bool {
	signature = strings.TrimPrefix(strings.TrimSpace(signature), "0x")
	if len(signature) != 130 {
		return false
	}
	_, err := hex.DecodeString(signature)
	return err == nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func asString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	default:
		return ""
	}
}

func maskDefaultName(identityKey string) string {
	return "用户" + utils.MaskIdentityKey("", identityKey)
}

func platformIDValue(platform *models.Platform) uint {
	if platform == nil {
		return 0
	}
	return platform.ID
}
