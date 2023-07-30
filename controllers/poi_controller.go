package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/gin-gonic/gin"
	"mvc/service"
	"mvc/utils"
	"strconv"
	"strings"
)

func PoiAround(c *gin.Context) {
	// 获取经纬度参数
	location := c.Query("location")

	// 解析经纬度字符串
	latLng := ParseLatLng(location)
	if latLng == nil {
		c.JSON(400, gin.H{"error": "Invalid location parameter"})
		return
	}
	longitude, latitude := latLng.Lat, latLng.Lon

	// 创建 Elasticsearch 客户端连接
	esClient, err := utils.CreateESClient()
	if err != nil {
		fmt.Println("Failed to connect to Elasticsearch:", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	// 构建查询条件
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"match_all": map[string]interface{}{}},
				},
				"filter": []map[string]interface{}{
					{
						"geo_distance": map[string]interface{}{
							"distance": "5km",
							"location": map[string]float64{
								"lat": latitude,
								"lon": longitude,
							},
						},
					},
				},
			},
		},
	}

	// 打印查询内容
	queryBytes, _ := json.MarshalIndent(query, "", "  ")
	fmt.Println("Query:", string(queryBytes))

	// 创建查询请求
	req := esapi.SearchRequest{
		Index: []string{"poi_data_2023"},
		Body:  esutil.NewJSONReader(query),
	}

	// 执行查询请求
	res, err := req.Do(c.Request.Context(), esClient)
	service.LogInfo(res)
	if err != nil {
		fmt.Println("Error executing search request:", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	defer res.Body.Close()
	service.LogInfo(res)

	// 解析查询响应
	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		fmt.Println("Error parsing search response:", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	// 解析查询结果
	hitsData, ok := response["hits"].(map[string]interface{})
	if !ok {
		c.JSON(500, gin.H{"error": "Invalid search response"})
		return
	}

	hits, ok := hitsData["hits"].([]interface{})
	if !ok {
		c.JSON(500, gin.H{"error": "Invalid search response"})
		return
	}

	pois := make([]map[string]interface{}, len(hits))
	for i, hit := range hits {
		hitData, ok := hit.(map[string]interface{})
		if !ok {
			c.JSON(500, gin.H{"error": "Invalid search response"})
			return
		}

		source, ok := hitData["_source"].(map[string]interface{})
		if !ok {
			c.JSON(500, gin.H{"error": "Invalid search response"})
			return
		}

		poi := map[string]interface{}{
			"ID":               source["poi_id"].(float64),
			"FormattedAddress": source["formatted_address"].(string),
			"Location":         source["location"].(map[string]interface{})["lat"].(float64),
		}
		pois[i] = poi
	}

	// 返回查询结果
	c.JSON(200, pois)
}

// LatLng 表示经纬度坐标
type LatLng struct {
	Lat float64 `json:"lat"` // 纬度
	Lon float64 `json:"lon"` // 经度
}

// ParseLatLng 解析经纬度字符串为经纬度坐标对象
func ParseLatLng(location string) *LatLng {
	coords := strings.Split(location, ",")
	if len(coords) != 2 {
		return nil
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(coords[0]), 64)
	if err != nil {
		return nil
	}

	lon, err := strconv.ParseFloat(strings.TrimSpace(coords[1]), 64)
	if err != nil {
		return nil
	}

	return &LatLng{Lat: lat, Lon: lon}
}
