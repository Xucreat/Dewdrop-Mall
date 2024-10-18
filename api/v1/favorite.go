package v1

import (
	"mall/consts"
	util "mall/pkg/utils"
	service2 "mall/service"

	"github.com/gin-gonic/gin"
)

// 创建收藏
func CollectFavoriteStore(c *gin.Context) {
	service := service2.FavoritesService{}
	claim, _ := util.ParseToken(c.GetHeader("Authorization"))
	if err := c.ShouldBind(&service); err == nil {
		res := service.Collect(c.Request.Context(), claim.ID)
		c.JSON(consts.StatusOK, res)
	} else {
		c.JSON(consts.IlleageRequest, ErrorResponse(err))
		util.LogrusObj.Infoln(err)
	}
}

// 收藏夹详情接口
func ShowFavoriteStores(c *gin.Context) {
	service := service2.FavoritesService{}
	claim, _ := util.ParseToken(c.GetHeader("Authorization"))
	if err := c.ShouldBind(&service); err == nil {
		res := service.Show(c.Request.Context(), claim.ID)
		c.JSON(consts.StatusOK, res)
	} else {
		c.JSON(consts.IlleageRequest, ErrorResponse(err))
		util.LogrusObj.Infoln(err)
	}
}

// 删除单个店铺
func DeleteFavoriteStore(c *gin.Context) {
	service := service2.FavoritesService{}
	claim, _ := util.ParseToken(c.GetHeader("Authorization"))
	if err := c.ShouldBind(&service); err == nil {
		res := service.Delete(c.Request.Context(), service.BossId, claim.ID)
		c.JSON(consts.StatusOK, res)
	} else {
		c.JSON(consts.IlleageRequest, ErrorResponse(err))
		util.LogrusObj.Infoln(err)
	}
}

// 批量删除店铺
func DeleteFavoriteStores(c *gin.Context) {
	service := service2.FavoritesService{}
	claim, _ := util.ParseToken(c.GetHeader("Authorization"))
	// 绑定接收多个 boss_id 的参数
	var bossIds struct {
		BossIds []uint `form:"boss_ids" json:"boss_ids"`
	}

	if err := c.ShouldBind(&bossIds); err == nil {
		res := service.DeleteBatch(c.Request.Context(), bossIds.BossIds, claim.ID)
		c.JSON(consts.StatusOK, res)
	} else {
		c.JSON(consts.IlleageRequest, ErrorResponse(err))
		util.LogrusObj.Infoln(err)
	}
}
