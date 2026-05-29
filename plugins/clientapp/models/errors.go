package models

import "errors"

var (
	ErrInvalidClientKey    = errors.New("客户端Key格式不正确，仅支持小写字母、数字和连字符，长度3-64")
	ErrUnsupportedPlatform = errors.New("不支持的平台类型")
	ErrClientNotFound      = errors.New("客户端不存在")
	ErrPlatformNotFound    = errors.New("平台渠道不存在")
	ErrUserNotFound        = errors.New("用户不存在")
	ErrIdentityNotFound    = errors.New("身份绑定不存在")
	ErrClientKeyExists     = errors.New("客户端Key已存在")
	ErrPlatformExists      = errors.New("该平台AppID已存在")
	ErrIdentityExists      = errors.New("身份标识已存在")
	ErrClientHasPlatform   = errors.New("客户端下存在平台渠道，无法删除")
	ErrPlatformHasIdentity = errors.New("平台渠道下存在身份绑定，无法删除")
	ErrInvalidIdentityType = errors.New("不支持的身份类型")
	ErrInvalidPhone        = errors.New("手机号格式不正确，请使用E.164格式")
	ErrInvalidEVMAddress   = errors.New("EVM钱包地址格式不正确")
	ErrPlatformRequired    = errors.New("小程序身份必须指定平台渠道")
	ErrInvalidPlatform     = errors.New("平台渠道无效")
)
