package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"mall/conf"
	"mall/consts"
	"mall/pkg/e"
	"mall/pkg/errorhandler"
	util "mall/pkg/utils"
	dao2 "mall/repository/db/dao"
	model2 "mall/repository/db/model"
	"mall/serializer"

	// "github.com/gin-gonic/gin"
	logging "github.com/sirupsen/logrus"
	"gopkg.in/mail.v2"
)

// UserService 管理用户服务
type UserService struct {
	NickName string `form:"nick_name" json:"nick_name"`
	UserName string `form:"user_name" json:"user_name"`
	Password string `form:"password" json:"password"`
	Key      string `form:"key" json:"key"` // 前端进行判断
}

type SendEmailService struct {
	Email    string `form:"email" json:"email"`
	Password string `form:"password" json:"password"`
	// OpertionType 1:绑定邮箱 2：解绑邮箱 3：改密码
	OperationType uint `form:"operation_type" json:"operation_type"`
}

type ValidEmailService struct {
}

func (service UserService) Register(ctx context.Context) serializer.Response {
	var user *model2.User
	code := e.SUCCESS
	if service.Key == "" || len(service.Key) != 6 {
		code = e.ERROR
		errorhandler.HandleError(nil, code, "key长度错误,必须是6位数")
	}
	util.Encrypt.SetKey(service.Key)
	userDao := dao2.NewUserDao(ctx)
	_, exist, err := userDao.ExistOrNotByUserName(service.UserName)
	if err != nil {
		return errorhandler.HandleError(err, e.ErrorDatabase, nil)

	}
	if exist {
		return errorhandler.HandleError(err, e.ErrorExistUser, nil)
	}
	user = &model2.User{
		NickName: service.NickName,
		UserName: service.UserName,
		Status:   model2.Active,
		Money:    util.Encrypt.AesEncoding("10000"), // 初始金额
	}
	// 加密密码
	err = user.SetPassword(service.Password)
	if err != nil {
		return errorhandler.HandleError(err, e.ERROR, nil)
	}
	if conf.UploadModel == consts.UploadModelOss {
		user.Avatar = "http://q1.qlogo.cn/g?b=qq&nk=294350394&s=640"
	} else {
		user.Avatar = "avatar.JPG"
	}
	// 创建用户
	err = userDao.CreateUser(user)
	if err != nil {
		return errorhandler.HandleError(err, e.ErrorExistUser, nil)
	}
	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}

// Login 用户登陆函数
func (service *UserService) Login(ctx context.Context) serializer.Response {
	// code := e.SUCCESS
	userDao := dao2.NewUserDao(ctx)
	user, exist, err := userDao.ExistOrNotByUserName(service.UserName)
	if err != nil || !exist {
		return errorhandler.HandleError(err, e.ErrorUserNotFound, nil)

	}

	if !user.CheckPassword(service.Password) {
		return errorhandler.HandleError(err, e.ErrorNotCompare, nil)
	}

	token, err := util.GenerateToken(user.ID, service.UserName, 0)
	if err != nil {
		return errorhandler.HandleError(err, e.ErrorNotCompare, nil)
	}
	return serializer.Response{
		Status: e.SUCCESS,
		Data:   serializer.TokenData{User: serializer.BuildUser(user), Token: token},
		Msg:    e.GetMsg(e.SUCCESS),
	}
}

// Update 用户修改信息
func (service UserService) Update(ctx context.Context, uId uint) serializer.Response {
	// 找到用户
	userDao := dao2.NewUserDao(ctx)
	user, err := userDao.GetUserById(uId)
	if err != nil {
		return errorhandler.HandleError(err, e.ErrorDatabase, nil)
	}

	if service.NickName != "" {
		user.NickName = service.NickName
	}

	err = userDao.UpdateUserById(uId, user)
	if err != nil {
		return errorhandler.HandleError(err, e.ErrorAuthToken, nil)
	}

	return serializer.Response{
		Status: e.SUCCESS,
		Data:   serializer.BuildUser(user),
		Msg:    e.GetMsg(e.SUCCESS),
	}
}

func (service *UserService) Post(ctx context.Context, uId uint, file multipart.File, fileSize int64) serializer.Response {
	userDao := dao2.NewUserDao(ctx)
	user, err := userDao.GetUserById(uId)
	if err != nil {
		return errorhandler.HandleError(err, e.ErrorDatabase, nil)
	}

	var path string
	if conf.UploadModel == consts.UploadModelLocal { // 兼容两种存储方式
		path, err = util.UploadAvatarToLocalStatic(file, uId, user.UserName)
	} else {
		path, err = util.UploadToQiNiu(file, fileSize)
	}
	if err != nil {
		return errorhandler.HandleError(err, e.ErrorUploadFile, nil)
	}

	user.Avatar = path
	err = userDao.UpdateUserById(uId, user)
	if err != nil {
		return errorhandler.HandleError(err, e.ErrorUploadFile, nil)

	}
	return serializer.Response{
		Status: e.SUCCESS,
		Data:   serializer.BuildUser(user),
		Msg:    e.GetMsg(e.SUCCESS),
	}
}

/*进行邮箱相关操作,需要进行邮箱验证:由用户前端发送请求，后端接收请求,将生成的包含Token的链接以邮件的形式发送给用户邮箱，
用户获取Token，返还给前端，前端交给后端解析，后端根据解析出的操作类型对解析出的用户的邮箱进行相关操作*/
// Send 发送邮件
func (service *SendEmailService) Send(ctx context.Context, id uint) serializer.Response {
	token, err := util.GenerateEmailToken(id, service.OperationType, service.Email, service.Password)
	if err != nil {
		return errorhandler.HandleError(err, e.ErrorAuthToken, nil)
	}

	noticeDao := dao2.NewNoticeDao(ctx)
	notice, err := noticeDao.GetNoticeById(service.OperationType)
	if err != nil {
		return errorhandler.HandleError(err, e.ErrorAuthToken, nil)
	}

	address := conf.ValidEmail + token
	mailStr := notice.Text
	mailText := strings.Replace(mailStr, "Email", address, -1)

	m := mail.NewMessage()
	m.SetHeader("From", conf.SmtpEmail)
	m.SetHeader("To", service.Email)
	m.SetHeader("Subject", "BX")
	m.SetBody("text/html", mailText)

	d := mail.NewDialer(conf.SmtpHost, 465, conf.SmtpEmail, conf.SmtpPass)
	d.StartTLSPolicy = mail.MandatoryStartTLS
	if err := d.DialAndSend(m); err != nil {
		return errorhandler.HandleError(err, e.ErrorAuthToken, nil)
	}
	return serializer.Response{
		Status: e.SUCCESS,
		Msg:    e.GetMsg(e.SUCCESS),
	}
}

// Valid 验证内容
func (service ValidEmailService) Valid(ctx context.Context, token string) serializer.Response {
	var userID uint
	var email string
	var password string
	var operationType uint
	code := e.SUCCESS

	// 验证token
	if token == "" {
		code = e.InvalidParams
	} else {
		claims, err := util.ParseEmailToken(token) // 解析Token
		if err != nil {
			logging.Info(err)
			code = e.ErrorAuthCheckTokenFail
		} else if time.Now().Unix() > claims.ExpiresAt {
			code = e.ErrorAuthCheckTokenTimeout
		} else {
			userID = claims.UserID
			email = claims.Email
			password = claims.Password
			operationType = claims.OperationType
		}
	}
	if code != e.SUCCESS {
		return errorhandler.HandleError(nil, e.ErrorAuthToken, nil)
	}

	// 获取该用户信息
	userDao := dao2.NewUserDao(ctx)
	user, err := userDao.GetUserById(userID)
	if err != nil {
		return errorhandler.HandleError(err, e.ErrorAuthCheckTokenFail, nil)
	}

	// switch-case
	switch operationType {
	// 1:绑定邮箱
	case 1:
		user.Email = email
	// 2：解绑邮箱
	case 2:
		if user.Email == "" {
			return serializer.Response{
				Status: e.SUCCESS,
				Data:   serializer.BuildUser(user),
				Msg:    "邮箱已解绑",
			}
		}
		user.Email = ""
	// 3：修改密码
	case 3:
		err = user.SetPassword(password)
		if err != nil {
			return errorhandler.HandleError(err, e.ErrorDatabase, nil)
		}
		// 默认情况，防止未知的操作类型
	default:
		return errorhandler.HandleError(fmt.Errorf("未知的操作类型"), e.InvalidParams, nil)
	}

	// err = userDao.UpdateUserById(userID, user)
	err = userDao.UpdateUserEmailById(userID, user)
	if err != nil {
		return errorhandler.HandleError(err, e.ErrorDatabase, nil)
	}
	// 成功则返回用户的信息
	return serializer.Response{
		Status: e.SUCCESS,
		Data:   serializer.BuildUser(user),
		Msg:    e.GetMsg(e.SUCCESS),
	}
}
