package service

import (
	"context"
	"mime/multipart"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"

	"mall/conf"
	"mall/consts"
	"mall/pkg/e"
	util "mall/pkg/utils"
	dao2 "mall/repository/db/dao"

	model2 "mall/repository/db/model"
	"mall/serializer"
)

// 更新商品的服务
type ProductService struct {
	Query    string `json:"query"`
	Page     int    `form:"page" json:"page"`           // 页码 让 Gin 框架能够自动绑定 HTTP 请求中的参数到结构体字段中
	PageSize int    `form:"page_size" json:"page_size"` // 每页数量
	// 新增商品更新时需要的字段
	ID            uint   `form:"id" json:"id"`
	Name          string `form:"name" json:"name"`
	CategoryID    int    `form:"category_id" json:"category_id"`
	Title         string `form:"title" json:"title"`
	Info          string `form:"info" json:"info"`
	ImgPath       string `form:"img_path" json:"img_path"`
	Price         string `form:"price" json:"price"`
	DiscountPrice string `form:"discount_price" json:"discount_price"`
	OnSale        bool   `form:"on_sale" json:"on_sale"`
	Num           int    `form:"num" json:"num"`
}
type ListProductImgService struct {
}

// 商品
func (service *ProductService) Show(ctx context.Context, id string) serializer.Response {
	code := e.SUCCESS

	pId, _ := strconv.Atoi(id)

	productDao := dao2.NewProductDao(ctx)
	product, err := productDao.GetProductById(uint(pId))
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "ProductShow", // 描述了日志记录发生的上下文或服务
		}).Error("Failed to get product by product id")
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}

	return serializer.Response{
		Status: code,
		Data:   serializer.BuildProduct(product),
		Msg:    e.GetMsg(code),
	}
}

// 创建商品
func (service *ProductService) Create(ctx context.Context, uId uint, files []*multipart.FileHeader) serializer.Response {
	code := e.SUCCESS

	// 1. 获取商家信息
	userDao := dao2.NewUserDao(ctx)
	boss, err := userDao.GetUserById(uId)
	if err != nil {
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}

	// 2. 检查是否有上传的文件
	if len(files) == 0 {
		code = e.ErrorUploadFile
		return serializer.Response{
			Status: code,
			Msg:    "No files uploaded",
			Error:  "No files were provided for the product creation",
		}
	}

	// 3. 上传封面图片: 打开第一个文件 (封面图片)
	tmp, err := files[0].Open()
	if err != nil {
		code = e.ErrorUploadFile
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	defer tmp.Close()

	var path string
	// 根据 conf.UploadModel 配置，选择上传到本地或七牛云
	if conf.UploadModel == consts.UploadModelLocal {
		path, err = util.UploadProductToLocalStatic(tmp, uId, service.Name)
	} else {
		path, err = util.UploadToQiNiu(tmp, files[0].Size)
	}
	if err != nil {
		code = e.ErrorUploadFile
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}

	// 3. 创建商品
	product := &model2.Product{
		Name:          service.Name,
		CategoryID:    uint(service.CategoryID),
		Title:         service.Title,
		Info:          service.Info,
		ImgPath:       path,
		Price:         service.Price,
		DiscountPrice: service.DiscountPrice,
		Num:           service.Num,
		OnSale:        true,
		BossID:        uId,
		BossName:      boss.UserName,
		BossAvatar:    boss.Avatar,
	}

	productDao := dao2.NewProductDao(ctx)
	err = productDao.CreateProduct(product)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "ProductCreate",
		}).Error("Failed to create product")
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}

	// 4. 并发上传商品图片
	wg := new(sync.WaitGroup)               // 并发原语:用于等待一组“工作协程”结束->用于等待所有图片上传完成
	errChan := make(chan error, len(files)) // 用于收集上传错误
	// 遍历所有图片文件
	for index, file := range files {
		wg.Add(1)
		// 使用 go 语句启动协程上传图片
		go func(file *multipart.FileHeader, index int) {
			defer wg.Done()
			num := strconv.Itoa(index) // 将索引转换为字符串，用于生成图片路径
			productImgDao := dao2.NewProductImgDaoByDB(productDao.DB)

			// 打开图片文件，获取文件读取器
			tmp, err := file.Open()
			if err != nil {
				errChan <- err
				return
			}
			defer tmp.Close() // defer 关键字用于延迟关闭文件，避免资源泄露

			// 使用 util 工具包中的函数上传图片，并记录图片路径
			if conf.UploadModel == consts.UploadModelLocal {
				path, err = util.UploadProductToLocalStatic(tmp, uId, service.Name+num)
			} else {
				path, err = util.UploadToQiNiu(tmp, file.Size)
			}
			if err != nil {
				errChan <- err
				return
			}

			// 存储图片信息
			productImg := &model2.ProductImg{
				ProductID: product.ID,
				ImgPath:   path,
			}
			//  使用商品图片 DAO 对象创建商品图片，并记录返回错误。
			if err := productImgDao.CreateProductImg(productImg); err != nil {
				errChan <- err
				return
			}
		}(file, index)
	}
	wg.Wait()

	// 检查错误
	// 错误数量最大=图片数量。因为每个协程只会返回一个错误，而协程数量和图片文件数量相同
	close(errChan)
	for err := range errChan {
		if err != nil {
			return serializer.Response{
				Status: e.ErrorUploadFile,
				Msg:    e.GetMsg(e.ErrorUploadFile),
				Error:  err.Error(),
			}
		}
	}

	return serializer.Response{
		Status: code,
		Data:   serializer.BuildProduct(product),
		Msg:    e.GetMsg(code),
	}
}

// 创建商品
// func (service *ProductService) Create(ctx context.Context, uId uint, files []*multipart.FileHeader) serializer.Response {
// 	var boss *model2.User
// 	var err error
// 	code := e.SUCCESS

// 	userDao := dao2.NewUserDao(ctx)
// 	boss, _ = userDao.GetUserById(uId)
// 	// 以第一张作为封面图
// 	tmp, _ := files[0].Open()
// 	var path string
// 	// 根据 conf.UploadModel 配置，选择上传到本地或七牛云
// 	if conf.UploadModel == consts.UploadModelLocal {
// 		path, err = util.UploadProductToLocalStatic(tmp, uId, service.Name)
// 	} else {
// 		path, err = util.UploadToQiNiu(tmp, files[0].Size)
// 	}
// 	if err != nil {
// 		code = e.ErrorUploadFile
// 		return serializer.Response{
// 			Status: code,
// 			Data:   e.GetMsg(code),
// 			Error:  path,
// 		}
// 	}
// 	// 创建商品
// 	product := &model2.Product{
// 		Name:          service.Name,
// 		CategoryID:    uint(service.CategoryID),
// 		Title:         service.Title,
// 		Info:          service.Info,
// 		ImgPath:       path,
// 		Price:         service.Price,
// 		DiscountPrice: service.DiscountPrice,
// 		Num:           service.Num,
// 		OnSale:        true,
// 		BossID:        uId,
// 		BossName:      boss.UserName,
// 		BossAvatar:    boss.Avatar,
// 	}
// 	productDao := dao2.NewProductDao(ctx)
// 	err = productDao.CreateProduct(product)
// 	if err != nil {
// 		util.LogrusObj.WithFields(logrus.Fields{
// 			"error":   err,
// 			"context": "ProductCreate", // 描述了日志记录发生的上下文或服务
// 		}).Error("Failed to create product")
// 		code = e.ErrorDatabase
// 		return serializer.Response{
// 			Status: code,
// 			Msg:    e.GetMsg(code),
// 			Error:  err.Error(),
// 		}
// 	}
// 	// 并发上传商品图片
// 	wg := new(sync.WaitGroup) // 创建 sync.WaitGroup 对象，用于等待所有图片上传完成
// 	wg.Add(len(files))
// 	for index, file := range files {
// 		num := strconv.Itoa(index)
// 		productImgDao := dao2.NewProductImgDaoByDB(productDao.DB)
// 		tmp, _ = file.Open()
// 		if conf.UploadModel == consts.UploadModelLocal {
// 			path, err = util.UploadProductToLocalStatic(tmp, uId, service.Name+num)
// 		} else {
// 			path, err = util.UploadToQiNiu(tmp, file.Size)
// 		}
// 		if err != nil {
// 			code = e.ErrorUploadFile
// 			return serializer.Response{
// 				Status: code,
// 				Data:   e.GetMsg(code),
// 				Error:  path,
// 			}
// 		}
// 		productImg := &model2.ProductImg{
// 			ProductID: product.ID,
// 			ImgPath:   path,
// 		}
// 		err = productImgDao.CreateProductImg(productImg)
// 		if err != nil {
// 			code = e.ERROR
// 			return serializer.Response{
// 				Status: code,
// 				Msg:    e.GetMsg(code),
// 				Error:  err.Error(),
// 			}
// 		}
// 		wg.Done()
// 	}

// 	wg.Wait()

// 	return serializer.Response{
// 		Status: code,
// 		Data:   serializer.BuildProduct(product),
// 		Msg:    e.GetMsg(code),
// 	}
// }

// 获取商品列表
func (service *ProductService) List(ctx context.Context) serializer.Response {
	var products []*model2.Product // 用于存储查询到的商品列表
	var total int64                // 存储商品总数
	code := e.SUCCESS              // 存储响应状态码

	// 1. 设置默认分页大小
	if service.PageSize == 0 {
		service.PageSize = 15
	}

	// 2. 条件查询: 获取查询条件：catagory_id
	condition := make(map[string]interface{}) // 存储查询条件
	if service.CategoryID != 0 {
		condition["category_id"] = service.CategoryID
	}

	productDao := dao2.NewProductDao(ctx)
	// 获取满足条件的商品总数,即有商品id的商品总数
	total, err := productDao.CountProductByCondition(condition)
	if err != nil {
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}
	}

	// 3. 并发查询商品列表
	wg := new(sync.WaitGroup) // 等待并发操作完成
	wg.Add(1)                 // 增加等待组的计数，表示有一个待完成的操作

	errChan := make(chan error, 1)
	// 使用并发协程来执行 数据库查询 ，以提高性能
	go func() {
		defer wg.Done()
		productDao = dao2.NewProductDaoByDB(productDao.DB)
		// 调用 ListProductByCondition 方法获取商品列表,实现了分页展示
		products, err = productDao.ListProductByCondition(condition, service.Page, service.PageSize)
		if err != nil {
			errChan <- err
			return
		}
	}()

	wg.Wait()
	close(errChan)

	// 检查错误
	for err := range errChan {
		if err != nil {
			return serializer.Response{
				Status: e.ErrorDatabase,
				Msg:    e.GetMsg(e.ErrorDatabase),
			}
		}
	}

	// 构建商品列表响应，并返回
	// 响应包含商品列表和商品总数
	return serializer.BuildListResponse(serializer.BuildProducts(products), uint(total))
}

// func (service *ProductService) List(ctx context.Context) serializer.Response {
// 	var products []*model2.Product
// 	var total int64
// 	code := e.SUCCESS

// 	if service.PageSize == 0 {
// 		service.PageSize = 15
// 	}
// 	condition := make(map[string]interface{})
// 	if service.CategoryID != 0 {
// 		condition["category_id"] = service.CategoryID
// 	}
// 	productDao := dao2.NewProductDao(ctx)
// 	total, err := productDao.CountProductByCondition(condition)
// 	if err != nil {
// 		code = e.ErrorDatabase
// 		return serializer.Response{
// 			Status: code,
// 			Msg:    e.GetMsg(code),
// 		}
// 	}
// 	wg := new(sync.WaitGroup)
// 	wg.Add(1)
// 	go func() {
// 		productDao = dao2.NewProductDaoByDB(productDao.DB)
// 		products, _ = productDao.ListProductByCondition(condition, service.BasePage)
// 		wg.Done()
// 	}()
// 	wg.Wait()

// 	return serializer.BuildListResponse(serializer.BuildProducts(products), uint(total))
// }

// 删除商品
func (service *ProductService) Delete(ctx context.Context, pId string) serializer.Response {
	code := e.SUCCESS

	productDao := dao2.NewProductDao(ctx)
	productId, _ := strconv.Atoi(pId)
	err := productDao.DeleteProduct(uint(productId))
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "ProductDelete", // 描述了日志记录发生的上下文或服务
		}).Error("Failed to delete product by product id")
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}

// 更新商品
func (service *ProductService) Update(ctx context.Context, pId string) serializer.Response {
	code := e.SUCCESS
	productDao := dao2.NewProductDao(ctx)

	productId, _ := strconv.Atoi(pId)
	product := &model2.Product{
		Name:       service.Name,
		CategoryID: uint(service.CategoryID),
		Title:      service.Title,
		Info:       service.Info,
		// ImgPath:       service.ImgPath,
		Price:         service.Price,
		DiscountPrice: service.DiscountPrice,
		OnSale:        service.OnSale,
	}
	err := productDao.UpdateProduct(uint(productId), product)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "ProductUpdate", // 描述了日志记录发生的上下文或服务
		}).Error("Failed to update product")
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	return serializer.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}
}

// Search 实现搜索功能，包括分页和日志记录
// func (service *ProductService) Search(ctx context.Context) serializer.Response {
// 	code := e.SUCCESS
// 	if service.PageSize == 0 {
// 		service.PageSize = 15
// 	}

//		productDao := dao2.NewProductDao(ctx)
//		products, err := productDao.SearchProduct(service.Info, service.Page, service.PageSize)
//		if err != nil {
//			util.LogrusObj.WithFields(logrus.Fields{
//				"error":   err,
//				"context": "ProductSearch",
//			}).Error("Failed to search product from database")
//			code = e.ErrorDatabase
//			return serializer.Response{
//				Status: code,
//				Msg:    e.GetMsg(code),
//				Error:  err.Error(),
//			}
//		}
//		return serializer.BuildListResponse(serializer.BuildProducts(products), uint(len(products)))
//	}
//
// 搜索商品
func (service *ProductService) Search(ctx context.Context) serializer.Response {
	code := e.SUCCESS
	if service.PageSize == 0 {
		service.PageSize = 15
	}

	productDao := dao2.NewProductDao(ctx)
	products, err := productDao.SearchProduct(service.Info, service.Page, service.PageSize, service.CategoryID)
	if err != nil {
		util.LogrusObj.WithFields(logrus.Fields{
			"error":   err,
			"context": "ProductSearch",
		}).Error("Failed to search product from database")
		code = e.ErrorDatabase
		return serializer.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	return serializer.BuildListResponse(serializer.BuildProducts(products), uint(len(products)))
}

// List 获取商品列表图片
func (service *ListProductImgService) List(ctx context.Context, pId string) serializer.Response {
	productImgDao := dao2.NewProductImgDao(ctx)
	productId, _ := strconv.Atoi(pId)
	productImgs, _ := productImgDao.ListProductImgByProductId(uint(productId))
	return serializer.BuildListResponse(serializer.BuildProductImgs(productImgs), uint(len(productImgs)))
}
