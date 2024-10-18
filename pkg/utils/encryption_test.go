package util

import (
	"fmt"
	"testing"
)

func TestEncryption(t *testing.T) {
	util := Encryption{}
	util.SetKey("正确的密钥")
	original := "1234.56"

	// 加密
	encrypted := util.AesEncoding(original)
	fmt.Printf("Encrypted: %s\n", encrypted)

	// 解密
	decrypted := util.AesDecoding(encrypted)
	if decrypted == "Decryption failed" {
		t.Errorf("Decryption failed: %s", decrypted)
	} else {
		fmt.Printf("Decrypted: %s\n", decrypted)
	}

	// 检查解密是否成功
	if original != decrypted {
		t.Errorf("Decryption mismatch. Expected %s, got %s", original, decrypted)
	}
}
