//go:build !windows

package config

func dpapiProtect(plain []byte) ([]byte, error) { return plain, nil }
func dpapiUnprotect(blob []byte) ([]byte, error) { return blob, nil }
