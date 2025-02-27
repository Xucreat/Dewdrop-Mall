package model

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// User 用户模型
type User struct {
	gorm.Model
	UserName       string `gorm:"unique"`
	Email          string
	PasswordDigest string
	NickName       string
	Status         string
	Avatar         string `gorm:"size:1000"`
	Money          string
}

const (
	PassWordCost        = 12       //密码加密难度
	Active       string = "active" //激活用户
)

// SetPassword 设置密码
func (user *User) SetPassword(password string) error {
	// 返回给定加密难度相应的密码，以字节切片的形式
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), PassWordCost)
	if err != nil {
		return err
	}
	user.PasswordDigest = string(bytes)
	return nil
}

// CheckPassword 校验密码
func (user *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordDigest), []byte(password))
	return err == nil
}

// AvatarUrl 头像地址
func (user *User) AvatarURL() string {
	signedGetURL := user.Avatar
	return signedGetURL
}
