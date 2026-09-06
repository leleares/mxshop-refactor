package v1

import (
	proto "mxshop/api/goods/v1"
	"mxshop/app/mxshop/api/internal/domain/request"
	service "mxshop/app/mxshop/api/internal/service"
	"mxshop/pkg/common/core"
	"mxshop/pkg/log"

	"github.com/gin-gonic/gin"
)

type GoodsController struct {
	srv service.ServiceFactory
}

func NewGoodsController(srv service.ServiceFactory) *GoodsController {
	return &GoodsController{
		srv,
	}
}

func (gc *GoodsController) List(ctx *gin.Context) {
	log.Info("goods list function called ...")

	var r request.GoodsFilter

	if err := ctx.ShouldBindQuery(&r); err != nil {
		return
	}

	gfr := proto.GoodsFilterRequest{
		IsNew:       r.IsNew,
		IsHot:       r.IsHot,
		PriceMax:    r.PriceMax,
		PriceMin:    r.PriceMin,
		TopCategory: r.TopCategory,
		Brand:       r.Brand,
		KeyWords:    r.KeyWords,
		Pages:       r.Pages,
		PagePerNums: r.PagePerNums,
	}

	goodsDTOList, err := gc.srv.Goods().List(ctx, &gfr)
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	reMap := map[string]interface{}{
		"total": goodsDTOList.Total,
	}
	goodsList := make([]interface{}, 0)
	for _, value := range goodsDTOList.Data {
		goodsList = append(goodsList, map[string]interface{}{
			"id":          value.Id,
			"name":        value.Name,
			"goods_brief": value.GoodsBrief,
			"desc":        value.GoodsDesc,
			"ship_free":   value.ShipFree,
			"images":      value.Images,
			"desc_images": value.DescImages,
			"front_image": value.GoodsFrontImage,
			"shop_price":  value.ShopPrice,
			"category": map[string]interface{}{
				"id":   value.Category.Id,
				"name": value.Category.Name,
			},
			"brand": map[string]interface{}{
				"id":   value.Brand.Id,
				"name": value.Brand.Name,
				"logo": value.Brand.Logo,
			},
			"is_hot":  value.IsHot,
			"is_new":  value.IsNew,
			"on_sale": value.OnSale,
		})
	}
	reMap["data"] = goodsList

	core.WriteResponse(ctx, nil, reMap)
}
