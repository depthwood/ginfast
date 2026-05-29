package utils

import (
	"regexp"
	"strings"
)

var evmAddressPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

// NormalizeEVMAddress 规范化 EVM 地址为小写
func NormalizeEVMAddress(address string) (string, bool) {
	address = strings.TrimSpace(address)
	if !evmAddressPattern.MatchString(address) {
		return "", false
	}
	return strings.ToLower(address), true
}
