package controllers

import (
	"github.com/gin-gonic/gin"
	"mvc/service"
	"strconv"
)

func PoiAround(c *gin.Context) {
	// 获取经纬度参数
	location := c.Query("location")
	radiusStr := c.Query("radius")

	// 解析经纬度字符串
	longitudeStr, latitudeStr, err := service.ParseLocation(location)
	if err != nil {
		c.Set("error", err.Error())
		return
	}

	// 将经度和纬度转换为 float64 类型
	longitude, err := strconv.ParseFloat(longitudeStr, 64)
	if err != nil {
		c.Set("error", err.Error())
		return
	}
	latitude, err := strconv.ParseFloat(latitudeStr, 64)
	if err != nil {
		c.Set("error", err.Error())
		return
	}

	// 将半径参数转换为 float64 类型
	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		c.Set("error", err.Error())
		return
	}

	// 创建 POIService 实例
	poiService, err := service.NewPOIService()
	if err != nil {
		c.Set("error", err.Error())
		return
	}

	// 调用 POIService 的方法查询 POI 点位信息
	result, err := poiService.GetPOIsByLocationAndRadius(latitude, longitude, radius)
	if err != nil {
		// 处理错误\
		c.Set("error", err.Error()) // 将异常信息存储到上下文的 Keys 中
		return
	} else {
		pois := result.POIs
		householdsAvg := result.HouseholdsAvg
		pricePerSqMAvg := result.PricePerSqMAvg

		// 使用返回值进行后续操作
		count := len(pois)
		c.Set("response", pois)
		c.Set("households_avg", householdsAvg)
		c.Set("price_per_sqm_avg", pricePerSqMAvg)
		c.Set("count", count)
	}
}
