package config

import "testing"

func TestProtectRoundtrip(t *testing.T) {
	plain := "sk-test-secret-key"
	enc, err := ProtectSecret(plain)
	if err != nil {
		t.Fatal(err)
	}
	if enc == "" {
		t.Fatal("empty ciphertext")
	}
	out, err := UnprotectSecret(enc)
	if err != nil {
		t.Fatal(err)
	}
	if out != plain {
		t.Fatalf("got %q want %q", out, plain)
	}
}

func TestMaskAPIKey(t *testing.T) {
	m := MaskAPIKey("sk-abcdefghijklmnop")
	if m == "sk-abcdefghijklmnop" {
		t.Fatal("should mask")
	}
	if !IsMaskedAPIKey(m) {
		t.Fatal("masked flag")
	}
}
