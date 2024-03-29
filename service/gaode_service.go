package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type AMapService struct {
	APIKey        string                   // 高德地图API密钥
	Concurrent    int                      // 最大并发数
	RateLimiter   *RateLimiter             // 速率限制器
	WaitCond      *sync.Cond               // 条件变量用于等待数据
	WaitLock      sync.Mutex               // 互斥锁用于保护条件变量
	PendingCount  int                      // 等待处理的请求数
	Cache         map[string]GeocodeResult // 缓存
	CacheMutex    sync.RWMutex             // 读写锁用于保护缓存
	CacheDuration time.Duration            // 缓存过期时间
}

type RateLimiter struct {
	TokenBucket chan struct{} // 令牌桶通道
}

type GeocodeResult struct {
	RawData    map[string]interface{} `json:"-"` // 存储原始数据的 map
	Expiration time.Time              // 过期时间
}

func NewAMapService() *AMapService {
	return &AMapService{
		APIKey:        "cb3e60dc70d48516d5d19ccaa000ae37",
		Concurrent:    10, // 最大并发数
		RateLimiter:   NewRateLimiter(10),
		WaitCond:      sync.NewCond(&sync.Mutex{}),
		Cache:         make(map[string]GeocodeResult),
		CacheDuration: 24 * time.Hour, // 一天的过期时间
	}
}

func NewRateLimiter(concurrent int) *RateLimiter {
	return &RateLimiter{
		TokenBucket: make(chan struct{}, concurrent),
	}
}

func (rl *RateLimiter) Allow() {
	rl.TokenBucket <- struct{}{}
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
	coordinateKey := fmt.Sprintf("%.6f,%.6f", latitude, longitude)

	// 先尝试从缓存中获取数据
	s.CacheMutex.RLock()
	cachedResult, ok := s.Cache[coordinateKey]
	s.CacheMutex.RUnlock()

	if ok && !cachedResult.Expiration.Before(time.Now()) {
		LogInfo("走的缓存")
		return cachedResult.RawData, nil
	}
	s.WaitLock.Lock()
	if s.PendingCount >= s.Concurrent {
		s.WaitCond.Wait() // 等待空闲槽位
	}
	s.PendingCount++
	s.WaitLock.Unlock()

	s.RateLimiter.Allow() // 限流

	baseURL := "https://restapi.amap.com/v3/geocode/regeo"
	apiURL, err := url.Parse(baseURL)
	//time.Sleep(time.Duration(rand.Intn(300)+500) * time.Millisecond) // 随机延迟0.5到2秒

	if err != nil {
		return nil, err
	}

	defer func() {
		s.WaitLock.Lock()
		s.PendingCount--
		s.WaitCond.Signal() // 释放槽位
		s.WaitLock.Unlock()
	}()

	queryString := apiURL.Query()
	queryString.Set("key", s.APIKey)
	queryString.Set("location", fmt.Sprintf("%.6f,%.6f", longitude, latitude))

	query := queryString.Encode()
	apiURL.RawQuery = query
	resp, err := http.Get(apiURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("获取逆地理编码数据失败")
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
		return nil, errors.New("无法提取逆地理编码数据")
	}

	// 将数据存储到缓存中
	s.CacheMutex.Lock()
	s.Cache[coordinateKey] = GeocodeResult{
		RawData:    regeocodes,
		Expiration: time.Now().Add(s.CacheDuration),
	}
	s.CacheMutex.Unlock()

	return regeocodes, nil
}

/**
* 这个方法主要是处理不限流，因为限流好像会导致mysql每次导入只能导入进去10条
 */
func (s *AMapService) ReverseGeocodeNoLimit(latitude, longitude float64) (map[string]interface{}, error) {
	coordinateKey := fmt.Sprintf("%.6f,%.6f", latitude, longitude)

	// 先尝试从缓存中获取数据
	s.CacheMutex.RLock()
	cachedResult, ok := s.Cache[coordinateKey]
	s.CacheMutex.RUnlock()

	if ok && !cachedResult.Expiration.Before(time.Now()) {
		LogInfo("走的缓存")
		return cachedResult.RawData, nil
	}

	baseURL := "https://restapi.amap.com/v3/geocode/regeo"
	apiURL, err := url.Parse(baseURL)
	//time.Sleep(time.Duration(rand.Intn(300)+500) * time.Millisecond) // 随机延迟0.5到2秒

	if err != nil {
		return nil, err
	}

	queryString := apiURL.Query()
	queryString.Set("key", s.APIKey)
	queryString.Set("location", fmt.Sprintf("%.6f,%.6f", longitude, latitude))

	query := queryString.Encode()
	apiURL.RawQuery = query
	resp, err := http.Get(apiURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("获取逆地理编码数据失败")
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
		return nil, errors.New("无法提取逆地理编码数据")
	}

	s.Cache[coordinateKey] = GeocodeResult{
		RawData:    regeocodes,
		Expiration: time.Now().Add(s.CacheDuration),
	}

	return regeocodes, nil
}
