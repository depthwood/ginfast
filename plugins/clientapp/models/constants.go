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

// IsSupportedPlatform 校验平台是否受支持
func IsSupportedPlatform(platform string) bool {
	for _, item := range SupportedPlatforms {
		if item == platform {
			return true
		}
	}
	return false
}
