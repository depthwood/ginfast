package models

// 客户端类型
const (
	ClientTypeMiniProgram = "mini_program"
	ClientTypeH5          = "h5"
	ClientTypeNative      = "native"
)

// 小程序平台
const (
	PlatformWechat   = "wechat"
	PlatformAlipay   = "alipay"
	PlatformDouyin   = "douyin"
	PlatformBaidu    = "baidu"
	PlatformQQ       = "qq"
	PlatformKuaishou = "kuaishou"
	PlatformFeishu   = "feishu"
	PlatformJD       = "jd"
)

// SupportedPlatforms 支持的小程序平台列表
var SupportedPlatforms = []string{
	PlatformWechat,
	PlatformAlipay,
	PlatformDouyin,
	PlatformBaidu,
	PlatformQQ,
	PlatformKuaishou,
	PlatformFeishu,
	PlatformJD,
}

// 身份类型
const (
	IdentityTypeMpOpenID  = "mp_openid"
	IdentityTypeMpUnionID = "mp_unionid"
	IdentityTypeWalletEVM = "wallet_evm"
	IdentityTypePhone     = "phone"
)

// 注册来源
const (
	RegisterSourceAdmin     = "admin"
	RegisterSourceMP        = "mp"
	RegisterSourceWalletEVM = "wallet_evm"
	RegisterSourcePhone     = "phone"
)

// 设备平台
const (
	DevicePlatformAndroid = "android"
	DevicePlatformIOS     = "ios"
	DevicePlatformWeb     = "web"
)

// 设备状态
const (
	DeviceStatusDisabled = 0
	DeviceStatusActive   = 1
	DeviceStatusPending  = 2
)

// 装修页面标识
const (
	PageCodeHome     = "home"
	PageCodeWork     = "work"
	PageCodeDiscover = "discover"
	PageCodeMessage  = "message"
	PageCodeMine     = "mine"
)

// DefaultPageDefinitions 默认页面定义列表
var DefaultPageDefinitions = []map[string]interface{}{
	{"id": PageCodeHome, "title": "首页", "path": "/pages/index/index", "enabled": true},
	{"id": PageCodeWork, "title": "服务", "path": "/pages/work/work", "enabled": true},
	{"id": PageCodeDiscover, "title": "活动", "path": "/pages/discover/discover", "enabled": true},
	{"id": PageCodeMessage, "title": "消息", "path": "/pages/message/message", "enabled": true},
	{"id": PageCodeMine, "title": "我的", "path": "/pages/mine/mine", "enabled": true},
}

// AppConfigStatus
const (
	ConfigStatusDisabled = 0
	ConfigStatusEnabled  = 1
)

// IsSupportedPlatform 校验平台是否受支持
func IsSupportedPlatform(platform string) bool {
	for _, item := range SupportedPlatforms {
		if item == platform {
			return true
		}
	}
	return false
}
