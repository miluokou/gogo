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

type GeocodeResult struct {
	RawData map[string]interface{} `json:"-"` // 存储原始数据的 map
}

func NewAMapService() *AMapService {
	return &AMapService{
		APIKey: "cb3e60dc70d48516d5d19ccaa000ae37",
	}
}

func (s *AMapService) Geocode(address string) ([]interface{}, error) {
	baseURL := "https://restapi.amap.com/v3/geocode/geo"
	apiURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	queryString := apiURL.Query()
	queryString.Set("key", s.APIKey)
	queryString.Set("address", address)

	apiURL.RawQuery = queryString.Encode()

	resp, err := http.Get(apiURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("获取地理编码数据失败")
	}

	var result map[string]interface{}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	geocodes, ok := result["geocodes"].([]interface{})
	if !ok {
		return nil, errors.New("无法提取地理编码数据")
	}

	return geocodes, nil
}
