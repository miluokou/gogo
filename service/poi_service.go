package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"mvc/utils"
)

// POIService 提供与POI相关的服务方法
type POIService struct {
	esClient *elasticsearch.Client
}

// NewPOIService 创建一个POIService实例
func NewPOIService() (*POIService, error) {
	esClient, err := utils.CreateESClient()
	if err != nil {
		return nil, err
	}

	return &POIService{
		esClient: esClient,
	}, nil
}

// GetPOIsByLocationAndRadius 根据经纬度和半径查询POI点位信息
func (s *POIService) GetPOIsByLocationAndRadius(latitude, longitude float64, radius float64) ([]map[string]interface{}, error) {
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
							"distance": fmt.Sprintf("%fm", radius),
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

	// 创建查询请求
	req := esapi.SearchRequest{
		Index: []string{"poi_data_2023"},
		Body:  esutil.NewJSONReader(query),
	}

	// 执行查询请求
	res, err := req.Do(context.Background(), s.esClient)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search request: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			// 写入日志文件
			LogInfo(closeErr)
		}
	}()

	// 解析查询响应
	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// 解析查询结果
	hitsData, ok := response["hits"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid search response")
	}

	hits, ok := hitsData["hits"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid search response")
	}

	pois := make([]map[string]interface{}, len(hits))
	for i, hit := range hits {
		hitData, ok := hit.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid search response")
		}

		source, ok := hitData["_source"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid search response")
		}

		poi := map[string]interface{}{
			"ID":               source["poi_id"].(float64),
			"FormattedAddress": source["formatted_address"].(string),
			"Location":         source["location"].(map[string]interface{})["lat"].(float64),
		}
		pois[i] = poi
	}

	return pois, nil
}

// ParseLocation 解析经纬度字符串，返回经度和纬度值
func ParseLocation(location string) (float64, float64, error) {
	// 将经纬度字符串拆分为经度和纬度部分
	parts := strings.Split(location, ",")

	if len(parts) != 2 {
		return 0, 0, errors.New("Invalid location format")
	}

	// 解析经度和纬度值
	latitude, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, errors.New("Invalid latitude value")
	}

	longitude, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, errors.New("Invalid longitude value")
	}

	return latitude, longitude, nil
}
