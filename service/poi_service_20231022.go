package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"mvc/utils"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// POIService20231022 POIService 提供与POI相关的服务方法
type POIService20231022 struct {
	esClient *elasticsearch.Client
}

type POIResult20231022 struct {
	POIs           []map[string]interface{}
	HouseholdsAvg  float64
	PricePerSqMAvg float64
}

// NewPOIService20231022 NewPOIService 创建一个POIService实例
func NewPOIService20231022() (*POIService, error) {
	esClient, err := utils.GetESClient()
	if err != nil {
		return nil, err
	}

	return &POIService{
		esClient: esClient,
	}, nil
}

var req = esapi.SearchRequest{
	Index: []string{"poi_2023_01"},
	Body:  nil, // 初始置为空
}

// WaitForFileDescriptors 等待足够的文件描述符可用
func WaitForFileDescriptors(desiredLimit uint64, delay time.Duration) {
	for {
		// 获取当前打开的文件描述符数量
		_, err := getCurrentFileDescriptorLimit()
		if err != nil {
			LogInfo("无法获取系统文件描述符限制：" + err.Error())
			time.Sleep(delay)
			continue
		}
		break
	}
}

// getCurrentFileDescriptorLimit 获取当前打开的文件描述符数量
func getCurrentFileDescriptorLimit() (uint64, error) {
	cmd := exec.Command("bash", "-c", "ulimit -n")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	limitStr := strings.TrimSpace(string(output))
	currentLimit, err := strconv.ParseUint(limitStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return currentLimit, nil
}

// getCurrentFileDescriptorLimit 获取当前打开的文件描述符数量
func getCurrentFileDescriptorLimitTanXing() (uint64, error) {
	cmd := exec.Command("bash", "-c", "ls /proc/self/fd | wc -l")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	limitStr := strings.TrimSpace(string(output))
	currentLimit, err := strconv.ParseUint(limitStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return currentLimit, nil
}

func (s *POIService) GetPOIsByLocationAndRadius20231022(latitude, longitude float64, radius float64) (POIResult, error) {
	// 获取系统的文件描述符数量
	desiredLimit := uint64(1000) // 期望的文件描述符限制
	waitDelay := 2 * time.Second // 等待延时

	WaitForFileDescriptors(desiredLimit, waitDelay)

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"geo_distance": map[string]interface{}{
							"distance": "0.000000000001m",
							"location": map[string]interface{}{
								"lat": latitude,
								"lon": longitude,
							},
						},
					},
				},
			},
		},
		"size": 1,
	}

	req.Body = esutil.NewJSONReader(query)

	res, err := req.Do(context.Background(), s.esClient)
	if err != nil {
		return POIResult{}, fmt.Errorf("执行搜索请求失败：%w", err)
	}
	defer res.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return POIResult{}, fmt.Errorf("解析搜索响应失败：%w", err)
	}

	hitsData, ok := response["hits"].(map[string]interface{})
	if !ok {
		LogInfo("无效的搜索响应1，响应内容为：")
		LogInfo(response)
		return POIResult{}, fmt.Errorf("无效的搜索响应1")
	}

	hits, ok := hitsData["hits"].([]interface{})
	if !ok {
		return POIResult{}, fmt.Errorf("无效的搜索响应2")
	}

	pois := make([]map[string]interface{}, len(hits))
	var householdsSum, pricePerSqMSum float64

	for i, hit := range hits {
		hitData, ok := hit.(map[string]interface{})
		if !ok {
			return POIResult{}, fmt.Errorf("无效的搜索响应3")
		}

		source, ok := hitData["_source"].(map[string]interface{})
		if !ok {
			return POIResult{}, fmt.Errorf("invalid search response4")
		}
		//LogInfo(source["location"])

		location, ok := source["location"].(map[string]interface{})
		if !ok {
			return POIResult{}, fmt.Errorf("invalid search response: location not found or invalid")
		}

		latitude, ok := location["lat"].(float64)
		if !ok {
			return POIResult{}, fmt.Errorf("invalid search response: latitude not found or invalid. Got: %v", location["lat"])
		}

		longitude, ok := location["lon"].(float64)
		if !ok {
			return POIResult{}, fmt.Errorf("invalid search response: longitude not found or invalid. Got: %v", location["lon"])
		}

		locationStr := fmt.Sprintf("%f,%f", longitude, latitude)

		poi := map[string]interface{}{
			//"id":                source["poi_id"].(float64),
			//"formatted_address": source["formatted_address"].(string),
			//"price_per_sqm":     source["price_per_sqm"],
			"location": locationStr,
		}
		pois[i] = poi

		if households, ok := source["households"].(float64); ok {
			householdsSum += households
		}

		if pricePerSqM, ok := source["price_per_sqm"].(float64); ok {
			pricePerSqMSum += pricePerSqM
		}
	}
	householdsAvg := householdsSum / float64(len(hits))
	pricePerSqMAvg := pricePerSqMSum / float64(len(hits))

	return POIResult{
		POIs:           pois,
		HouseholdsAvg:  householdsAvg,
		PricePerSqMAvg: pricePerSqMAvg,
	}, nil
}
