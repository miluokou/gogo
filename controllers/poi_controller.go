package controllers

import (
	"github.com/gin-gonic/gin"
	"mvc/service"
	"net/http"
	"strconv"
)

func PoiAround(c *gin.Context) {
	// 获取经纬度参数
	location := c.Query("location")
	radiusStr := c.Query("radius")

	// 解析经纬度字符串
	longitude, latitude, err := service.ParseLocation(location)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location parameter"})
		return
	}

	// 将半径参数转换为 float64 类型
	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid radius parameter"})
		return
	}

	// 创建 POIService 实例
	poiService, err := service.NewPOIService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error1"})
		return
	}

	// 调用 POIService 的方法查询 POI 点位信息
	pois, err := poiService.GetPOIsByLocationAndRadius(latitude, longitude, radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error2"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": pois,
	})
}
