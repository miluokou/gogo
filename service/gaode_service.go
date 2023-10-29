package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

type AMapService struct {
	APIKey       string       // 高德地图API密钥
	Concurrent   int          // 最大并发数
	RateLimiter  *RateLimiter // 速率限制器
	WaitCond     *sync.Cond   // 条件变量用于等待数据
	WaitLock     sync.Mutex   // 互斥锁用于保护条件变量
	PendingCount int          // 等待处理的请求数
}

type RateLimiter struct {
	TokenBucket chan struct{} // 令牌桶通道
}

type GeocodeResult struct {
	RawData map[string]interface{} `json:"-"` // 存储原始数据的 map
}

func NewRateLimiter(concurrent int) *RateLimiter {
	return &RateLimiter{
		TokenBucket: make(chan struct{}, concurrent),
	}
}

func (rl *RateLimiter) Allow() {
	rl.TokenBucket <- struct{}{}
}

func NewAMapService() *AMapService {
	return &AMapService{
		APIKey:      "cb3e60dc70d48516d5d19ccaa000ae37",
		Concurrent:  99,
		RateLimiter: NewRateLimiter(99),
		WaitCond:    sync.NewCond(&sync.Mutex{}),
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

func (s *AMapService) ReverseGeocode(latitude, longitude float64) (map[string]interface{}, error) {
	s.WaitLock.Lock()
	if s.PendingCount >= s.Concurrent {
		s.WaitCond.Wait() // 等待空闲槽位
	}
	s.PendingCount++
	s.WaitLock.Unlock()

	defer func() {
		s.WaitLock.Lock()
		s.PendingCount--
		s.WaitCond.Signal() // 释放槽位
		s.WaitLock.Unlock()
	}()

	s.RateLimiter.Allow() // 限流

	baseURL := "https://restapi.amap.com/v3/geocode/regeo"
	apiURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	queryString := apiURL.Query()
	queryString.Set("key", s.APIKey)
	queryString.Set("location", fmt.Sprintf("%.6f,%.6f", longitude, latitude))

	apiURL.RawQuery = queryString.Encode()
	resp, err := http.Get(apiURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("获取逆地理编码数据失败1")
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	regeocodes, ok := result["regeocode"].(map[string]interface{})
	if !ok {
		LogInfo("高德返回的结果是:")
		LogInfo(result)
		return nil, errors.New("无法提取逆地理编码数据2")
	}

	return regeocodes, nil
}
