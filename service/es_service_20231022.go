package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/mmcloughlin/geohash"
	"log"
	"mvc/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PlaceSearchResult20231022 struct {
	RawData map[string]interface{} `json:"-"`
	Data    []interface{}          `json:"data"`
}

var StoreData20231022Group sync.WaitGroup
var StoreData20231022Semaphore = make(chan struct{}, 1)

func StoreData20231022(index string, data [][]string) error {
	// 将每条记录转换为map[string]interface{}
	var err error
	esClient20231022, err := utils.GetESClient()
	if err != nil {
		log.Fatalf("无法创建Elasticsearch客户端：%s", err)
	}

	var poiData []map[string]interface{}
	for _, record := range data {
		if len(record) != 8 {
			return errors.New("无效的数据格式")
		}
		item := map[string]interface{}{
			"name":      record[0],
			"category1": record[1],
			"category2": record[2],
			"longitude": record[3],
			"latitude":  record[4],
			"province":  record[5],
			"city":      record[6],
			"region":    record[7],
		}

		poiData = append(poiData, item)
	}

	StoreData20231022Group.Add(1)
	// 获取信号量，限制并发数量
	StoreData20231022Semaphore <- struct{}{}
	LogInfo("组合好了一波 poiData 开始向prepareBulkPayload20231022 方法中传入开始预处理")
	prepareDataBefore := prepareBulkPayload20231022(poiData)

	<-StoreData20231022Semaphore // 释放信号量，允许下一个请求
	StoreData20231022Group.Done()
	//LogInfo("准备好的数据格式是")
	//LogInfo(prepareDataBefore)
	prepareData := bytes.NewReader(prepareDataBefore)
	bulkRequest := esapi.BulkRequest{
		Index:   index,
		Body:    prepareData,
		Refresh: "true",
	}

	res, err := bulkRequest.Do(context.Background(), esClient20231022)
	if err != nil {
		errorMsg := fmt.Errorf("StoreData20231022存储数据到Elasticsearch失败：%v", err)
		LogInfo(errorMsg.Error())
		return errorMsg
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			// 写入日志文件
			LogInfo(closeErr)
		}
	}()

	if res.IsError() {
		errorMsg := fmt.Errorf("StoreData20231022存储数据失败。响应状态：%s", res.Status())
		LogInfo(errorMsg.Error())
		return errorMsg
	}

	return nil
}

var dataCache sync.Map // 并发安全的缓存结构
var requestCount = struct {
	sync.Map
	sync.Mutex
}{}
var requestCountMutex sync.Mutex // 互斥锁

func prepareBulkPayload20231022(data []map[string]interface{}) []byte {
	var bulkPayload strings.Builder
	var counter int
	var cacheKey string

	for _, poiData := range data {
		cacheKey := fmt.Sprintf("%v", poiData) // 以poiData作为缓存的key

		// 检查是否已经缓存过该数据
		if cachedResult, found := dataCache.Load(cacheKey); found {
			// 如果找到缓存结果，直接返回
			return cachedResult.([]byte)
		}

		lon, _ := strconv.ParseFloat(poiData["longitude"].(string), 64)
		lat, _ := strconv.ParseFloat(poiData["latitude"].(string), 64)

		key := fmt.Sprintf("%.6f_%.6f", lat, lon)

		requestCount.Lock()
		count, _ := requestCount.Map.LoadOrStore(key, 0) // 使用 LoadOrStore 方法获取计数值，如果不存在则初始化为 0
		requestCount.Map.Store(key, count.(int)+1)
		if count.(int) > 0 {
			LogInfo(fmt.Sprintf("经纬度 %.6f, %.6f 的请求次数大于1，跳过处理", lat, lon))
			requestCount.Unlock()
			continue
		}
		requestCount.Unlock()

		location := map[string]interface{}{"lon": lon, "lat": lat}
		poiData["location"] = location

		poiService, _ := NewPOIService20231022()
		existingData, err := poiService.GetPOIsByLocationAndRadius20231022(lat, lon, 5000)

		if err != nil {
			errorMsg := fmt.Errorf("Error checking existing data: %v", err)
			LogInfo(errorMsg.Error())
			LogInfo(existingData)
			continue
		}

		pois := existingData.POIs
		if len(pois) > 0 {
			LogInfo(fmt.Sprintf("已经有这条数据了，跳过存储，经纬度 %.6f, %.6f 的，跳过处理", lat, lon))
			LogInfo(pois)
			continue
		}

		gaoDeService := NewAMapService()
		regeocodes, err := gaoDeService.ReverseGeocode(lat, lon)

		if err != nil {
			LogInfo("逆地理编码失败")
			LogInfo(err)
			continue
		}

		poiData["adcode"] = regeocodes["addressComponent"].(map[string]interface{})["adcode"]

		geoHash := geohash.Encode(lat, lon)
		poiData["geohash"] = geoHash

		adcode, ok := poiData["adcode"].(string)
		if !ok {
			continue
		}
		uniqueID := generateUniqueID(adcode)
		poiData["poi_id"] = uniqueID

		documentID := generateDocumentID(adcode)
		currentTime := time.Now().Format("2006-01-02 15:04:05")

		poiData["created_at"] = currentTime
		poiData["updated_at"] = currentTime

		poiJSON, _ := json.Marshal(poiData)
		bulkPayload.WriteString(fmt.Sprintf(`{"index":{"_index":"poi_2023_01","_id":"%s"}}`, documentID))
		bulkPayload.WriteString("\n")
		bulkPayload.Write(poiJSON)
		bulkPayload.WriteString("\n")

		counter++
		LogInfo(fmt.Sprintf("处理完成第 %d 条数据", counter))
	}

	dataCache.Store(cacheKey, []byte(bulkPayload.String())) // 将缓存设置放在循环外面

	return []byte(bulkPayload.String())
}
