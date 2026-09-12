package config

import (
	"encoding/base64"
	"runtime"
	"strings"
)

const secretPlainPrefix = "plain:"

// ProtectSecret 保护 API Key。Windows 使用 DPAPI；其它平台使用 plain: 前缀（仍避免与明文 json 字段混用）。
func ProtectSecret(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	if runtime.GOOS == "windows" {
		blob, err := dpapiProtect([]byte(plain))
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(blob), nil
	}
	return secretPlainPrefix + plain, nil
}

// UnprotectSecret 还原 API Key。兼容尚未加密的历史明文。
func UnprotectSecret(stored string) (string, error) {
	if stored == "" {
		return "", nil
	}
	if strings.HasPrefix(stored, secretPlainPrefix) {
		return strings.TrimPrefix(stored, secretPlainPrefix), nil
	}
	if runtime.GOOS == "windows" {
		raw, err := base64.StdEncoding.DecodeString(stored)
		if err != nil {
			// 旧版明文
			return stored, nil
		}
		plain, err := dpapiUnprotect(raw)
		if err != nil {
			return stored, nil
		}
		return string(plain), nil
	}
	return stored, nil
}

func MaskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	r := []rune(key)
	if len(r) <= 8 {
		return "****"
	}
	return string(r[:4]) + "****" + string(r[len(r)-4:])
}

func IsMaskedAPIKey(key string) bool {
	return key == "" || strings.Contains(key, "****")
}
