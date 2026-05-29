package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONString JSON 字符串字段
type JSONString string

func (j JSONString) Value() (driver.Value, error) {
	if j == "" {
		return nil, nil
	}
	return string(j), nil
}

func (j *JSONString) Scan(value interface{}) error {
	if value == nil {
		*j = ""
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = JSONString(v)
	case string:
		*j = JSONString(v)
	default:
		return errors.New("invalid JSONString value")
	}
	return nil
}

// MarshalJSON 序列化为 JSON
func (j JSONString) MarshalJSON() ([]byte, error) {
	if j == "" {
		return []byte("null"), nil
	}
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(j), &raw); err != nil {
		return json.Marshal(string(j))
	}
	return []byte(j), nil
}

// PlatformFeatures 平台特性配置
type PlatformFeatures struct {
	LoginEnabled   bool `json:"loginEnabled"`
	UnionIDEnabled bool `json:"unionIdEnabled"`
}

// DefaultPlatformFeatures 默认平台特性
func DefaultPlatformFeatures() PlatformFeatures {
	return PlatformFeatures{
		LoginEnabled:   true,
		UnionIDEnabled: false,
	}
}
