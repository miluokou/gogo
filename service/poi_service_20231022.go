package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"mvc/utils"
	"strconv"
	"strings"
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
	esClient, err := utils.CreateESClient()
	if err != nil {
		return nil, err
	}

	return &POIService{
		esClient: esClient,
	}, nil
}

func (s *POIService) GetPOIsByLocationAndRadius20231022(latitude, longitude float64, radius float64) (POIResult, error) {
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

	req := esapi.SearchRequest{
		Index: []string{"poi_2023_01"},
		Body:  esutil.NewJSONReader(query),
	}

	res, err := req.Do(context.Background(), s.esClient)
	if err != nil {
		return POIResult{}, fmt.Errorf("failed to execute search request: %w", err)
	}
	defer res.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return POIResult{}, fmt.Errorf("failed to parse search response: %w", err)
	}

	hitsData, ok := response["hits"].(map[string]interface{})
	if !ok {
		LogInfo("invalid search response1 的response是：")
		LogInfo(response)
		return POIResult{}, fmt.Errorf("invalid search response1")
	}

	hits, ok := hitsData["hits"].([]interface{})
	if !ok {
		return POIResult{}, fmt.Errorf("invalid search response2")
	}

	pois := make([]map[string]interface{}, len(hits))
	var householdsSum, pricePerSqMSum float64

	for i, hit := range hits {
		hitData, ok := hit.(map[string]interface{})
		if !ok {
			return POIResult{}, fmt.Errorf("invalid search response3")
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
			"id":                source["poi_id"].(float64),
			"formatted_address": source["formatted_address"].(string),
			"price_per_sqm":     source["price_per_sqm"],
			"location":          locationStr,
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

// ParseLocation20231022 ParseLocation 解析经纬度字符串，返回经度和纬度值
func ParseLocation20231022(location string) (string, string, error) {
	// 将经纬度字符串拆分为经度和纬度部分
	parts := strings.Split(location, ",")

	if len(parts) != 2 {
		return "", "", errors.New("Invalid location format")
	}

	// 解析经度和纬度值
	latitude, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return "", "", errors.New("Invalid latitude value")
	}

	longitude, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return "", "", errors.New("Invalid longitude value")
	}

	return strconv.FormatFloat(latitude, 'f', -1, 64), strconv.FormatFloat(longitude, 'f', -1, 64), nil
}
