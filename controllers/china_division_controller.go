package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"mvc/models/orm"
	"mvc/service"
	"net/http"
)

var centerCache = make(map[string]bool)

// getFenceData 调用高德API获取围栏数据
func getFenceData(province string) (string, error) {
	apiKey := "a7e7f4627788a84cc785e95ee11a4bb4" // 替换为您的高德API密钥
	url := fmt.Sprintf("https://restapi.amap.com/v3/config/district?key=%s&keywords=%s&subdistrict=0", apiKey, province)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// LandParcel 地块数据结构
type LandParcel struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Province  string `gorm:"column:province" json:"province"`
	City      string `gorm:"column:city" json:"city"`
	District  string `gorm:"column:district" json:"district"`
	FenceData string `gorm:"column:fence_data" json:"fenceData"`
	Center    string `gorm:"column:center" json:"center"`
}

// GridPartition 划分网格
func GridPartition() []LandParcel {
	minLongitude := 73.446960  // 中国范围内最小经度（高德坐标系）
	maxLongitude := 135.085591 // 中国范围内最大经度（高德坐标系）
	minLatitude := 3.862440    // 中国范围内最小纬度（高德坐标系）
	maxLatitude := 53.559093   // 中国范围内最大纬度（高德坐标系）

	gridSize := 0.0411714 // 网格大小，约等于2.9公里（在高德坐标系中的经纬度差值）

	var parcels []LandParcel

	// 遍历经度和纬度范围，并创建网格单元
	for longitude := minLongitude; longitude < maxLongitude; longitude += gridSize {
		for latitude := minLatitude; latitude < maxLatitude; latitude += gridSize {
			center := fmt.Sprintf("%.6f,%.6f", longitude+gridSize/2, latitude+gridSize/2)
			if centerCache[center] {
				continue // 如果中心点已在缓存中存在，则跳过重复的地块
			}

			fenceData := fmt.Sprintf("%.6f,%.6f;%.6f,%.6f;%.6f,%.6f;%.6f,%.6f", longitude, latitude, longitude+gridSize, latitude, longitude+gridSize, latitude+gridSize, longitude, latitude+gridSize)

			parcel := LandParcel{
				Province:  "中国",
				City:      "", // 根据实际需求设置城市名称
				District:  "", // 根据实际需求设置区县名称
				FenceData: fenceData,
				Center:    center,
			}
			parcels = append(parcels, parcel)

			// 将中心点添加到缓存
			centerCache[center] = true
		}
	}

	return parcels
}

// GetLandParcels 获取地块信息的控制器方法
func GetLandParcels(c *gin.Context) {
	parcels := GridPartition()

	for _, parcel := range parcels {
		err := orm.CreateLandParcel(parcel.Province, parcel.City, parcel.District, parcel.FenceData, parcel.Center)
		if err != nil {
			service.LogInfo(err)
			c.JSON(500, gin.H{"error": "Failed to create land parcel"})
			return
		}
	}

	c.JSON(200, parcels)
}
