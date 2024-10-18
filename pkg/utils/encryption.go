package util

/*将ECB模式改为CTR模式, 更安全的加密模式*/
import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
)

var Encrypt *Encryption

// AES 加密算法
type Encryption struct {
	key string
}

func init() {
	Encrypt = NewEncryption()
}

func NewEncryption() *Encryption {
	return &Encryption{}
}

// 填充密码长度
func PadPwd(srcByte []byte, blockSize int) []byte {
	padNum := blockSize - len(srcByte)%blockSize
	ret := bytes.Repeat([]byte{byte(padNum)}, padNum)
	srcByte = append(srcByte, ret...)
	return srcByte
}

// 加密
func (k *Encryption) AesEncoding(src string) string {
	srcByte := []byte(src)
	// 确保 key 长度为 16 字节
	NewKey := PadKeyTo16Bytes(k.key)

	block, err := aes.NewCipher([]byte(NewKey))
	if err != nil {
		return src
	}
	// // 密码填充(使用ECB模式)
	// NewSrcByte := PadPwd(srcByte, block.BlockSize()) // 由于字节长度不够，所以要进行字节的填充
	// dst := make([]byte, len(NewSrcByte))
	// block.Encrypt(dst, NewSrcByte)
	// // base64 编码
	// pwd := base64.StdEncoding.EncodeToString(dst)
	// return pwd

	// 使用 CTR 模式加密，需要初始化向量 (IV)
	iv := make([]byte, block.BlockSize())
	if _, err := rand.Read(iv); err != nil {
		return src // 如果生成随机 IV 出错，返回原始字符串
	}

	// 使用 CTR 模式加密
	stream := cipher.NewCTR(block, iv)
	encrypted := make([]byte, len(srcByte))
	stream.XORKeyStream(encrypted, srcByte)

	// 拼接 IV 和加密后的数据，以便解密时可以使用相同的 IV
	finalData := append(iv, encrypted...)

	// base64 编码
	pwd := base64.StdEncoding.EncodeToString(finalData)
	return pwd
}

// 将密钥填充为 16 字节
func PadKeyTo16Bytes(key string) string {
	// if len(key) >= 16 {
	//     return key[:16]
	// }
	// 使用 "0" 字符填充，直到达到 16 字节
	return key + strings.Repeat("0", 16-len(key))
}

// 去掉填充的部分
func UnPadPwd(dst []byte) ([]byte, error) {
	if len(dst) <= 0 {
		return dst, errors.New("长度有误")
	}
	// 去掉的长度
	unpadNum := int(dst[len(dst)-1])
	strErr := "error"
	op := []byte(strErr)
	if len(dst) < unpadNum {
		return op, nil
	}
	str := dst[:(len(dst) - unpadNum)]
	return str, nil
}

// 解密
func (k *Encryption) AesDecoding(pwd string) string {
	// Base64 解码
	pwdByte, err := base64.StdEncoding.DecodeString(pwd)
	if err != nil {
		return "Decryption failed  解码失败，返回错误提示" // 解码失败，返回错误提示
	}

	// 确保 key 长度为 16 字节
	NewKey := PadKeyTo16Bytes(k.key)
	// 初始化 AES Cipher
	// 创建一个新的 AES 加密/解密器，使用 k.key(注册时的key,相当于支付密码) 作为密钥
	// 如果 k.key（用于 AES 解密的密钥）不正确，NewCipher 将返回错误，导致返回原始的 pwd 值。
	block, errBlock := aes.NewCipher([]byte(NewKey))
	if errBlock != nil {
		return "Decryption failed 密钥错误，返回错误提示" // 密钥错误，返回错误提示
	}
	// // 解密
	// // 创建一个和密文长度相同的字节切片 dst，用于存放解密后的结果
	// dst := make([]byte, len(pwdByte))
	// block.Decrypt(dst, pwdByte) // 使用 block.Decrypt 进行解密操作，将 pwdByte（加密的字节数组）解密到 dst 中;这个方法要求输入数据的长度必须是 AES 块大小（16 字节）的整数倍
	// // 填充的要去掉
	// dst, err = UnPadPwd(dst)
	// if err != nil {
	// 	return "0"
	// }
	// return string(dst)
	// 提取 IV 和加密数据

	// 提取 IV 和加密数据
	if len(pwdByte) < block.BlockSize() {
		return "Decryption failed 密文长度不正确" // 密文长度不正确
	}

	iv := pwdByte[:block.BlockSize()]
	encryptedData := pwdByte[block.BlockSize():]

	// 使用 CTR 模式解密
	stream := cipher.NewCTR(block, iv)
	decrypted := make([]byte, len(encryptedData))
	stream.XORKeyStream(decrypted, encryptedData)

	// 检查解密结果是否是有效的浮点数
	_, err = strconv.ParseFloat(string(decrypted), 64)
	if err != nil {
		return "Decryption failed 无法正确解析为金额" // 如果解密后的内容无法正确解析为金额，返回错误提示
	}

	return string(decrypted)

}

// set方法
func (k *Encryption) SetKey(key string) {
	k.key = key
}
