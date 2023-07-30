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
	longitude, latitude, err := service.ParseLocation(location)

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
	pois, err := poiService.GetPOIsByLocationAndRadius(latitude, longitude, radius)
	if err != nil {
		c.Set("error", err.Error()) // 将异常信息存储到上下文的 Keys 中
		return
	}

	c.Set("response", pois)
}
