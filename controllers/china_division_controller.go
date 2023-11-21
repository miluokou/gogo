package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"mvc/controllers/poi"
	"mvc/models/orm"
	"mvc/service"
	"os"
	"strconv"
	"strings"
	"time"
)

var centerCache = make(map[string]bool)

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
func GridPartition(pageNumber int) []LandParcel {
	cachePath := "public/data/chinaFence.json" // 存储响应数据的JSON文件路径

	var data map[string]interface{}

	// 尝试从缓存文件中读取响应数据
	if fileExists(cachePath) {
		cacheFile, err := os.Open(cachePath)
		if err == nil {
			defer cacheFile.Close()

			_ = json.NewDecoder(cacheFile).Decode(&data)
		}
	}

	var parcels []LandParcel

	gridSize := 0.03 // 修改为0.03以获得3公里的网格

	pageSize := 1000 // 每页的网格数量

	startIndex := (pageNumber - 1) * pageSize
	endIndex := pageNumber * pageSize
	// 调用生成网格数据的方法
	gridData := poi.GenerateGridData(data, gridSize, startIndex, endIndex)

	//gaoDeService := service.NewAMapService()
	for _, value := range gridData {
		centerPoint := GetCenterPoint(value)

		parcel := LandParcel{
			Province:  "",
			City:      "", // 根据实际需求设置城市名称
			District:  "", // 根据实际需求设置区县名称
			FenceData: strings.Join(gridData, "|"),
			Center:    centerPoint,
		}
		parcels = append(parcels, parcel)

		//coords := strings.Split(centerPoint, ",")
		//lon, _ := strconv.ParseFloat(coords[0], 64)
		//lat, _ := strconv.ParseFloat(coords[1], 64)
		//regeocodes, _ := gaoDeService.ReverseGeocode(lat, lon)
		//service.LogInfo(regeocodes)

		//if regeocodes != nil {
		//	addressComponent, ok := regeocodes["addressComponent"].(map[string]interface{})
		//	if ok && addressComponent["adcode"] != nil {
		//		adcodeValue := addressComponent["adcode"]
		//
		//		switch adcode := adcodeValue.(type) {
		//		case string:
		//			_, err := strconv.Atoi(adcode)
		//			if err == nil {
		//				// adcode值为数值类型的字符串
		//				service.LogInfo("adcode: " + adcode)
		//				service.LogInfo(regeocodes["addressComponent"])
		//				os.Exit(1)
		//			} else {
		//				// adcode值不是数值类型的字符串
		//				service.LogInfo(centerPoint + " adcode is not a numeric string")
		//			}
		//		default:
		//			// adcode值为其他类型
		//			service.LogInfo(centerPoint + " adcode is of invalid type")
		//		}
		//	} else {
		//		// adcode值不存在或为空
		//		service.LogInfo(centerPoint + " adcode is missing or empty")
		//	}
		//} else {
		//	service.LogInfo(centerPoint + " regeocodes is nil")
		//}

	}
	return parcels
}

// 检查文件是否存在
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// GetLandParcels 获取地块信息的控制器方法
func GetLandParcels(c *gin.Context) {

	// 创建 Cache 实例
	cache, err := service.NewCache()
	if err != nil {
		fmt.Println("Failed to create cache:", err)
		return
	}

	// 设置计数器初始值
	//err = cache.Set("counter", 0, time.Hour)
	//if err != nil {
	//	fmt.Println("Failed to set counter:", err)
	//	return
	//}

	// 增加计数器值
	err = IncreaseCounter(cache, "counter")
	if err != nil {
		fmt.Println("Failed to increase counter:", err)
		return
	}

	// 获取计数器值
	value, err := GetCounter(cache, "counter")
	if err != nil {
		fmt.Println("Failed to get counter:", err)
		return
	}

	parcels := GridPartition(value)
	str := strconv.Itoa(value)
	for _, parcel := range parcels {
		err := orm.CreateLandParcel(parcel.Province, parcel.City, parcel.District, parcel.FenceData, parcel.Center, str)
		if err != nil {
			service.LogInfo(err)
			c.JSON(500, gin.H{"error": "Failed to create land parcel"})
			return
		}
	}

	c.JSON(200, parcels)
}

func GetCenterPoint(fenceData string) string {
	// 解析围栏数据
	points := parseFenceData(fenceData)

	// 计算围栏中心点坐标
	center := calculateCenter(points)

	// 拼接成符合高德经纬度规范的字符串
	result := fmt.Sprintf("%.6f,%.6f", center.x, center.y)
	return result
}

type Point struct {
	x float64
	y float64
}

// 解析围栏数据
func parseFenceData(data string) []Point {
	var points []Point

	// 分割字符串获取每个点的坐标
	coordinates := strings.Split(data, ";")
	for _, coordinate := range coordinates {
		xy := strings.Split(coordinate, ",")
		if len(xy) == 2 {
			point := Point{}
			point.x = parseFloat(xy[0])
			point.y = parseFloat(xy[1])
			points = append(points, point)
		}
	}

	return points
}

// 将字符串转换为浮点数
func parseFloat(s string) float64 {
	val := 0.0
	fmt.Sscanf(s, "%f", &val)
	return val
}

// 计算围栏中心点坐标
func calculateCenter(points []Point) Point {
	sumX := 0.0
	sumY := 0.0

	// 求和
	for _, point := range points {
		sumX += point.x
		sumY += point.y
	}

	// 计算平均值
	centerX := sumX / float64(len(points))
	centerY := sumY / float64(len(points))

	return Point{x: centerX, y: centerY}
}

func IncreaseCounter(cache *service.Cache, key string) error {
	value, err := cache.Get(key)
	if err != nil {
		return fmt.Errorf("获取计数器值失败: %v", err)
	}

	var counter int64
	switch v := value.(type) {
	case int64:
		counter = v
	case float64:
		counter = int64(v)
	default:
		return fmt.Errorf("无效的计数器值")
	}

	counter++

	err = cache.Set(key, counter, time.Hour)
	if err != nil {
		return fmt.Errorf("设置计数器值失败: %v", err)
	}

	return nil
}

func GetCounter(cache *service.Cache, key string) (int, error) {
	value, err := cache.Get(key)
	if err != nil {
		return 0, fmt.Errorf("获取计数器值失败: %v", err)
	}

	var counter int64
	switch v := value.(type) {
	case int64:
		counter = v
	case float64:
		counter = int64(v)
	default:
		counter = 0
	}

	return int(counter), nil
}
