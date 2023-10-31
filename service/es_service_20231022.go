package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/mmcloughlin/geohash"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PlaceSearchResult20231022 struct {
	RawData map[string]interface{} `json:"-"`
	Data    []interface{}          `json:"data"`
}

var esClient20231022 *elasticsearch.Client

func createESClient20231022() (*elasticsearch.Client, error) {
	if esClient20231022 != nil {
		return esClient20231022, nil
	}

	cfg := elasticsearch.Config{
		Addresses: []string{"http://47.100.242.199:9200"}, // 替换为 Elasticsearch 实际的地址
		Username:  "elastic",                              // 替换为您的 Elasticsearch 用户名
		Password:  "miluokou",
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

var StoreData20231022Group sync.WaitGroup
var StoreData20231022Semaphore = make(chan struct{}, 9)

func StoreData20231022(index string, data [][]string) error {
	// 将每条记录转换为map[string]interface{}
	var err error
	esClient20231022, err = createESClient20231022()
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
		//LogInfo(fmt.Sprintf("当前记录：%v", record))
		//LogInfo(fmt.Sprintf("当前item记录：%v", item))

		poiData = append(poiData, item)
	}
	StoreData20231022Group.Add(1)
	// 获取信号量，限制并发数量
	StoreData20231022Semaphore <- struct{}{}

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

/**
* 这个方法应该只是拼接一下数据
 */
var waitGroup sync.WaitGroup
var semaphore = make(chan struct{}, 9) // 设置并发请求数量为10

func prepareBulkPayload20231022(data []map[string]interface{}) []byte {
	var bulkPayload strings.Builder

	for _, poiData := range data {

		lon, err := strconv.ParseFloat(poiData["longitude"].(string), 64)
		if err != nil {
			continue // 跳过无效的经度值
		}
		lat, err := strconv.ParseFloat(poiData["latitude"].(string), 64)
		if err != nil {
			continue // 跳过无效的纬度值
		}

		location := map[string]interface{}{
			"lon": lon,
			"lat": lat,
		}
		poiData["location"] = location

		poiService, _ := NewPOIService20231022()
		waitGroup.Add(1)
		// 获取信号量，限制并发数量
		semaphore <- struct{}{}
		existingData, err := poiService.GetPOIsByLocationAndRadius20231022(lat, lon, 5000)

		if err != nil {
			// Handle the error from GetPOIsByLocationAndRadius
			errorMsg := fmt.Errorf("Error checking existing data: %v", err)
			LogInfo(errorMsg.Error())
			LogInfo(existingData)

			<-semaphore // 释放信号量，允许下一个请求
			waitGroup.Done()

			continue
		}

		pois := existingData.POIs
		if len(pois) > 0 {
			LogInfo("已经有这条数据了，跳过了存储")
			// Data already exists, skip storage

			<-semaphore // 释放信号量，允许下一个请求
			waitGroup.Done()

			continue
		}

		gaoDeService := NewAMapService()
		regeocodes, err := gaoDeService.ReverseGeocode(lat, lon)

		if err != nil {
			LogInfo("逆地理编码失败")
			LogInfo(err)

			<-semaphore // 释放信号量，允许下一个请求
			waitGroup.Done()

			continue
		}
		poiData["adcode"] = regeocodes["addressComponent"].(map[string]interface{})["adcode"]

		geoHash := geohash.Encode(lat, lon)
		poiData["geohash"] = geoHash

		adcode, ok := poiData["adcode"].(string)
		if !ok {
			<-semaphore // 释放信号量，允许下一个请求
			waitGroup.Done()

			continue // 跳过无效的adcode值
		}
		uniqueID := generateUniqueID(adcode)
		poiData["poi_id"] = uniqueID

		documentID := generateDocumentID(adcode)

		currentTime := time.Now().Format("2006-01-02 15:04:05")

		poiData["created_at"] = currentTime
		poiData["updated_at"] = currentTime

		poiJSON, err := json.Marshal(poiData)
		if err != nil {
			<-semaphore // 释放信号量，允许下一个请求
			waitGroup.Done()

			continue // 跳过无效的JSON序列化
		}

		bulkPayload.WriteString(`{"index":{"_index":"poi_2023_01","_id":"`)
		bulkPayload.WriteString(documentID)
		bulkPayload.WriteString(`"}}`)
		bulkPayload.WriteByte('\n')
		bulkPayload.Write(poiJSON)
		bulkPayload.WriteByte('\n')

		<-semaphore // 释放信号量，允许下一个请求
		waitGroup.Done()
	}

	waitGroup.Wait() // 等待所有请求完成

	return []byte(bulkPayload.String())
}
