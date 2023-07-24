package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type AMapService struct {
	APIKey string // 高德地图API密钥
}

type PlaceSearchResult struct {
	RawData map[string]interface{} `json:"-"` // 存储原始数据的 map
}

func NewAMapService(apiKey string) *AMapService {
	return &AMapService{
		APIKey: apiKey,
	}
}

func (s *AMapService) PlaceSearch(keywords string, city string) (*PlaceSearchResult, error) {
	baseURL := "https://restapi.amap.com/v5/place/text"
	apiURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	queryString := apiURL.Query()
	queryString.Set("key", s.APIKey)
	queryString.Set("keywords", keywords)
	queryString.Set("city", city)

	apiURL.RawQuery = queryString.Encode()

	resp, err := http.Get(apiURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("获取关键字搜索数据失败")
	}

	var result PlaceSearchResult
	err = json.NewDecoder(resp.Body).Decode(&result.RawData)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
