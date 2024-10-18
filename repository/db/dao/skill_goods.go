package dao

import (
	"context"

	"gorm.io/gorm"

	"mall/repository/db/model"
)

type SkillGoodsDao struct {
	*gorm.DB
}

func NewSkillGoodsDao(ctx context.Context) *SkillGoodsDao {
	return &SkillGoodsDao{NewDBClient(ctx)}
}

func (dao *SkillGoodsDao) Create(in *model.SkillGoods) error {
	return dao.Model(&model.SkillGoods{}).Create(&in).Error
}

func (dao *SkillGoodsDao) CreateByList(in []*model.SkillGoods) error {
	return dao.Model(&model.SkillGoods{}).Create(&in).Error
}

func (dao *SkillGoodsDao) ListSkillGoods() (resp []*model.SkillGoods, err error) {
	err = dao.Model(&model.SkillGoods{}).Where("num > 0").Find(&resp).Error
	return
}

// GetSkillGoodById 根据 SkillGoodsId 获取秒杀商品信息
func (dao *SkillGoodsDao) GetSkillGoodById(skillGoodsId uint) (*model.SkillGoods, error) {
	var skillGood model.SkillGoods
	if err := dao.DB.Where("id = ?", skillGoodsId).First(&skillGood).Error; err != nil {
		return nil, err
	}
	return &skillGood, nil
}
