package service

import (
	"context"
	"strconv"

	"mall/pkg/e"
	"mall/pkg/errorhandler"
	util "mall/pkg/utils"
	"mall/repository/db/dao"
	"mall/repository/db/model"
	"mall/serializer"

	"github.com/sirupsen/logrus"
)

type AddressService struct {
	Name    string `form:"name" json:"name"`
	Phone   string `form:"phone" json:"phone"`
	Address string `form:"address" json:"address"`
}

func (service *AddressService) Create(ctx context.Context, uId uint) serializer.Response {
	code := e.SUCCESS
	addressDao := dao.NewAddressDao(ctx)
	address := &model.Address{
		UserID:  uId,
		Name:    service.Name,
		Phone:   service.Phone,
		Address: service.Address,
	}
	err := addressDao.CreateAddress(address)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "AddressCreate", // 描述了日志记录发生的上下文或服务
		}).Error("Failed to create address")
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	addressDao = dao.NewAddressDaoByDB(addressDao.DB)
	var addresses []*model.Address
	addresses, err = addressDao.ListAddressByUid(uId)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "AddressCreate", // 描述了日志记录发生的上下文或服务
		}).Error("Failed to query the address list of the specified user in the database")
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	return serializer.Response{
		Status: code,
		Data:   serializer.BuildAddresses(addresses),
		Msg:    e.GetMsg(code),
	}
}

func (service *AddressService) Show(ctx context.Context, aId string) serializer.Response {
	code := e.SUCCESS
	addressDao := dao.NewAddressDao(ctx)
	addressId, _ := strconv.Atoi(aId)
	address, err := addressDao.GetAddressByAid(uint(addressId))
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "AddressShow", // 描述了日志记录发生的上下文或服务
		}).Error("Failed to get address by address id")
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	return serializer.Response{
		Status: code,
		Data:   serializer.BuildAddress(address),
		Msg:    e.GetMsg(code),
	}
}

func (service *AddressService) List(ctx context.Context, uId uint) serializer.Response {
	code := e.SUCCESS
	addressDao := dao.NewAddressDao(ctx)
	address, err := addressDao.ListAddressByUid(uId)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "AddressList", // 描述了日志记录发生的上下文或服务
		}).Error("Failed to query the address list of the specified user in the database")
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	return serializer.Response{
		Status: code,
		Data:   serializer.BuildAddresses(address),
		Msg:    e.GetMsg(code),
	}
}

func (service *AddressService) Delete(ctx context.Context, aId string) serializer.Response {
	addressDao := dao.NewAddressDao(ctx)
	code := e.SUCCESS
	addressId, _ := strconv.Atoi(aId)
	err := addressDao.DeleteAddressById(uint(addressId))
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "AddressDelete", // 描述了日志记录发生的上下文或服务
		}).Error("Failed to delete the address by address Id")
		code = e.ErrorDatabase
		// return serializer.Response{
		// 	Status: code,
		// 	Msg:    e.GetMsg(code),
		// 	Error:  err.Error(),
		// }
		return errorhandler.HandleError(err, code, nil)
	}
	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}

func (service *AddressService) Update(ctx context.Context, uid uint, aid string) serializer.Response {
	code := e.SUCCESS

	addressDao := dao.NewAddressDao(ctx)
	addressId, _ := strconv.Atoi(aid)
	// 必须三个字段全更新
	// address := &model.Address{
	// 	UserID:  uid,
	// 	Name:    service.Name,
	// 	Phone:   service.Phone,
	// 	Address: service.Address,
	// }

	// 构建一个 map 以只更新指定的字段
	updateFields := make(map[string]interface{})
	if service.Name != "" {
		updateFields["name"] = service.Name
	}
	if service.Phone != "" {
		updateFields["phone"] = service.Phone
	}
	if service.Address != "" {
		updateFields["address"] = service.Address
	}

	// 只更新非空字段
	err := addressDao.UpdateAddressFieldsById(uint(addressId), updateFields)
	if err != nil {
		return errorhandler.HandleError(err, e.ErrorDatabase, nil)
	}
	// 获取用户的所有地址列表并返回
	var addresses []*model.Address
	addressDao = dao.NewAddressDaoByDB(addressDao.DB)
	addresses, err = addressDao.ListAddressByUid(uid)
	if err != nil {
		return errorhandler.HandleError(err, e.ErrorDatabase, nil)
	}
	return serializer.Response{
		Status: code,
		Data:   serializer.BuildAddresses(addresses),
		Msg:    e.GetMsg(code),
	}
}
