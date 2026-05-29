package utils

import (
	"encoding/json"
	"strings"
)

var secretKeys = []string{
	"appSecret",
	"privateKey",
	"alipayPublicKey",
	"appKey",
	"secret",
}

// MaskIdentityKey 脱敏身份标识
func MaskIdentityKey(identityType, identityKey string) string {
	key := strings.TrimSpace(identityKey)
	if key == "" {
		return ""
	}
	switch identityType {
	case "phone":
		if len(key) <= 7 {
			return key
		}
		return key[:4] + "****" + key[len(key)-3:]
	case "wallet_evm":
		if len(key) <= 10 {
			return key
		}
		return key[:6] + "****" + key[len(key)-4:]
	default:
		if len(key) <= 8 {
			return key
		}
		return key[:4] + "****" + key[len(key)-4:]
	}
}

// MaskCredentials 脱敏平台凭证
func MaskCredentials(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return raw
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return raw
	}
	for _, secretKey := range secretKeys {
		if _, ok := data[secretKey]; ok {
			data[secretKey] = "******"
		}
	}
	masked, err := json.Marshal(data)
	if err != nil {
		return raw
	}
	return string(masked)
}
