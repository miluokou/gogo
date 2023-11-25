package poi

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"mvc/models/orm"
	"net/http"
	"strconv"
	"strings"
)

func CalculateGrid(c *gin.Context) {
	resp, err := http.Get("https://restapi.amap.com/v3/config/district?key=cb3e60dc70d48516d5d19ccaa000ae37&keywords=%E4%B8%AD%E5%9B%BD&subdistrict=0&extensions=all")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch fence data"})
		return
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse JSON"})
		return
	}

	districts := data["districts"].([]interface{})
	if len(districts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No districts found"})
		return
	}

	//gridSize := 0.03 // 修改为0.03以获得3公里的网格

	pageSize := 10000 // 每页的网格数量
	//pageNumber, _ := strconv.Atoi(c.Query("page")) // 获取当前页码，默认为第1页

	//startIndex := (pageNumber - 1) * pageSize
	//endIndex := pageNumber * pageSize
	// 调用生成网格数据的方法
	//gridData := GenerateGridData(data, gridSize, startIndex, endIndex)

	// 从mysql 中获取最新的网格的计算情况
	gridData := GenerateGridDataFromMysql(pageSize)

	data["districts"].([]interface{})[0].(map[string]interface{})["polyline"] = strings.Join(gridData, "|")

	c.JSON(http.StatusOK, data)
}

// 生成网格数据的方法
func GenerateGridData(data map[string]interface{}, gridSize float64, startIndex int, endIndex int) []string {
	count := 0
	district := data["districts"].([]interface{})[0].(map[string]interface{})
	polyline := district["polyline"].(string)
	fenceData := strings.Split(polyline, "|")

	var maxFencePoints []string
	for _, fence := range fenceData {
		points := strings.Split(fence, ";")
		if len(points) > len(maxFencePoints) {
			maxFencePoints = points
		}
	}

	var minLat, minLng, maxLat, maxLng float64
	for _, point := range maxFencePoints {
		latLng := strings.Split(point, ",")
		lat, lng := latLng[1], latLng[0]
		latFloat, _ := strconv.ParseFloat(lat, 64)
		lngFloat, _ := strconv.ParseFloat(lng, 64)
		if minLat == 0 || latFloat < minLat {
			minLat = latFloat
		}
		if minLng == 0 || lngFloat < minLng {
			minLng = lngFloat
		}
		if maxLat == 0 || latFloat > maxLat {
			maxLat = latFloat
		}
		if maxLng == 0 || lngFloat > maxLng {
			maxLng = lngFloat
		}
	}

	gridCountLat := int((maxLat - minLat) / gridSize)
	gridCountLng := int((maxLng - minLng) / gridSize)

	var gridData []string
	for lat := 0; lat < gridCountLat; lat++ {
		for lng := 0; lng < gridCountLng; lng++ {
			gridMinLat := minLat + float64(lat)*gridSize
			gridMaxLat := gridMinLat + gridSize
			gridMinLng := minLng + float64(lng)*gridSize
			gridMaxLng := gridMinLng + gridSize

			// 确保网格完全在围栏内部
			if gridMinLat < minLat {
				gridMinLat = minLat
			}
			if gridMaxLat > maxLat {
				gridMaxLat = maxLat
			}
			if gridMinLng < minLng {
				gridMinLng = minLng
			}
			if gridMaxLng > maxLng {
				gridMaxLng = maxLng
			}

			if gridMaxLat > maxLat || gridMaxLng > maxLng {
				continue
			}

			gridFence := fmt.Sprintf("%f,%f;%f,%f;%f,%f;%f,%f;%f,%f;",
				gridMinLng, gridMinLat,
				gridMaxLng, gridMinLat,
				gridMaxLng, gridMaxLat,
				gridMinLng, gridMaxLat,
				gridMinLng, gridMinLat)

			count++
			if count >= startIndex && count <= endIndex {
				gridData = append(gridData, gridFence)
			}
		}
	}

	return gridData
}

func GenerateGridDataFromMysql(pageSize int) []string {
	var gridData []string
	properties, _ := orm.GetLandParcel(pageSize)
	for _, value := range properties {
		gridData = append(gridData, value.FenceData)
	}
	return gridData
}
