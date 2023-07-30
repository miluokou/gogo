package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mmcloughlin/geohash"
	"log"
	"mvc/models"
	"mvc/service"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Property struct {
	ID            int    `json:"id"`
	YearInfo      string `json:"year_info"`
	CommunityName string `json:"community_name"`
	AddressInfo   string `json:"address_info"`
	PricePerSqm   string `json:"price_per_sqm"`
	PageNumber    string `json:"page_number"`
	Deal          string `json:"deal"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

var esClient *elasticsearch.Client

func init() {
	var err error
	esClient, err = createESClient()
	if err != nil {
		log.Fatalf("Failed to create Elasticsearch client: %s", err)
	}
}

type PlaceSearchResult struct {
	RawData map[string]interface{} `json:"-"`
	Data    []interface{}          `json:"data"`
}

func createESClient() (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://47.100.242.199:9200"},
		Username:  "elastic",
		Password:  "miluokou",
	}

	return elasticsearch.NewClient(cfg)
}

func storeData(c *gin.Context, esClient *elasticsearch.Client, index string, data []interface{}) error {
	// 将data内容转换为[]map[string]interface{}类型
	var poiData []map[string]interface{}
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			poiData = append(poiData, m)
		} else {
			return errors.New("无效的数据格式")
		}
	}
	service.LogInfo("index 的值是")
	service.LogInfo(index)
	prepareData := bytes.NewReader(prepareBulkPayload(poiData))
	bulkRequest := esapi.BulkRequest{
		Index:   index,
		Body:    prepareData,
		Refresh: "true",
	}

	res, err := bulkRequest.Do(context.Background(), esClient)
	if err != nil {
		errorMsg := fmt.Errorf("存储数据到Elasticsearch失败：%v", err)
		service.LogInfo(errorMsg.Error())
		return errorMsg
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			// 写入日志文件
			service.LogInfo(closeErr)
		}
	}()

	if res.IsError() {
		errorMsg := fmt.Errorf("存储数据失败。响应状态：%s", res.Status())
		service.LogInfo(errorMsg.Error())
		c.JSON(res.StatusCode, gin.H{"error": res.Status()})
		return errorMsg
	}

	return nil
}

func prepareBulkPayload(data []map[string]interface{}) []byte {
	var bulkPayload strings.Builder

	for _, poiData := range data {
		latLng := poiData["location"].(string)
		coordinates := strings.Split(latLng, ",")
		if len(coordinates) != 2 {
			continue // 跳过无效的经纬度数据
		}
		lon, err := strconv.ParseFloat(coordinates[0], 64)
		if err != nil {
			continue // 跳过无效的经度值
		}
		lat, err := strconv.ParseFloat(coordinates[1], 64)
		if err != nil {
			continue // 跳过无效的纬度值
		}

		location := map[string]interface{}{
			"lon": lon,
			"lat": lat,
		}
		poiData["location"] = location

		geoHash := geohash.Encode(lat, lon)
		poiData["geohash"] = geoHash

		adcode, ok := poiData["adcode"].(string)
		if !ok {
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
			continue // 跳过无效的JSON序列化
		}

		bulkPayload.WriteString(`{"index":{"_index":"poi_data_2023","_id":"`)
		bulkPayload.WriteString(documentID)
		bulkPayload.WriteString(`"}}`)
		bulkPayload.WriteByte('\n')
		bulkPayload.Write(poiJSON)
		bulkPayload.WriteByte('\n')
	}

	return []byte(bulkPayload.String())
}

func generateUniqueID(adcode string) int64 {
	adcodeInt, _ := strconv.ParseInt(adcode, 10, 64)

	// 生成UUID
	uuidValue := uuid.New().ID()

	// 将adcode与UUID进行合并
	uniqueID := adcodeInt<<32 | int64(uuidValue)
	return uniqueID
}

func generateDocumentID(adcode string) string {
	id := uuid.New()
	adcodeInt, _ := strconv.ParseInt(adcode, 10, 64)
	uuidString := id.String()

	uniqueID := int64(uuidStringToInt(uuidString))<<32 | adcodeInt
	documentID := strconv.FormatInt(uniqueID, 10)
	return adcode + documentID
}

func uuidStringToInt(uuidString string) uint64 {
	uuidBytes := []byte(uuidString)
	var result uint64

	for i := 0; i < len(uuidBytes); i++ {
		result = result<<8 + uint64(uuidBytes[i])
	}

	return result
}

func EsEnv(c *gin.Context) {
	properties, err := models.GetOriginData()
	if err != nil {
		fmt.Println("错误：", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取属性数据"})
		return
	}

	apiKey := "cb3e60dc70d48516d5d19ccaa000ae37"
	gaoDeService := service.NewAMapService(apiKey)

	results := make([]map[string]interface{}, 0)

	for _, prop := range properties {
		addressInfo := strings.ReplaceAll(prop.AddressInfo, "-", "")
		address := prop.City + addressInfo + prop.CommunityName
		result, err := gaoDeService.Geocode(address)

		//这里需要判断这条数据是否在数据库中已经存在了。
		//	如果已经存在
		//		判断房价是否为0
		//         房价为0 那么直接赋值房价数据
		//         房价不为0 那么取二者的平均值
		//	如果不存在的话，还需要增加房价的数据放到其中
		//		那么直接赋值房价数据

		if err != nil {
			fmt.Println("地理编码失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "地理编码失败"})
			return
		}

		if m, ok := result[0].(map[string]interface{}); ok {
			results = append(results, m)
		} else {
			fmt.Println("无法转换为map[string]interface{}")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法转换地理编码结果"})
			return
		}

		err = storeData(c, esClient, "poi_data_2023", result)
		if err != nil {
			fmt.Printf("Failed to store data in Elasticsearch: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法存储数据到Elasticsearch"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "属性数据获取成功",
		"data":    results,
	})
}
