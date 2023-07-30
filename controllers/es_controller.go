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
	"mvc/utils"
	"net/http"
	"strconv"
	"strings"
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
var Logger *log.Logger

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

func storeData(c *gin.Context, esClient *elasticsearch.Client, index, id string, data []interface{}) error {
	// 将data内容转换为[]map[string]interface{}类型
	var poiData []map[string]interface{}
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			poiData = append(poiData, m)
		} else {
			return errors.New("无效的数据格式")
		}
	}
	logInfo(c, poiData)
	logInfo(c, "index 的值是")
	logInfo(c, index)
	prepareData := bytes.NewReader(prepareBulkPayload(poiData))
	logInfo(c, "准备后的值是：")
	logInfo(c, prepareData)
	bulkRequest := esapi.BulkRequest{
		Index:   index,
		Body:    prepareData,
		Refresh: "true",
	}

	res, err := bulkRequest.Do(context.Background(), esClient)
	if err != nil {
		errorMsg := fmt.Errorf("存储数据到Elasticsearch失败：%v", err)
		logError(c, errorMsg.Error())
		return errorMsg
	}
	logInfo(c, res)
	defer res.Body.Close()

	if res.IsError() {
		errorMsg := fmt.Errorf("存储数据失败。响应状态：%s", res.Status())
		logError(c, errorMsg.Error())
		c.JSON(res.StatusCode, gin.H{"error": res.Status()})
		return errorMsg
	}

	return nil
}

// 准备批量操作的payload
func prepareBulkPayload(data []map[string]interface{}) []byte {
	var bulkPayload strings.Builder

	uniqueValues := make(map[interface{}]struct{})

	for _, poiData := range data {
		latLng := poiData["location"].(string)
		coordinates := strings.Split(latLng, ",")
		lat, _ := strconv.ParseFloat(coordinates[0], 64)
		lon, _ := strconv.ParseFloat(coordinates[1], 64)
		geoHash := geohash.Encode(lat, lon)
		poiData["geohash"] = geoHash

		uniqueID := uuid.New().String()
		poiData["poi_id"] = uniqueID

		// 检查指定字段上是否已存在相同值
		if fieldValue, exists := poiData["fieldName"]; exists {
			if _, isDuplicate := uniqueValues[fieldValue]; isDuplicate {
				continue // 跳过存储重复数据
			}
		}

		// 将字段值添加到去重映射中
		if fieldValue, exists := poiData["fieldName"]; exists {
			uniqueValues[fieldValue] = struct{}{}
		}

		poiJSON, _ := json.Marshal(poiData)

		bulkPayload.WriteString(`{"index":{}}`)
		bulkPayload.WriteByte('\n')
		bulkPayload.Write(poiJSON)
		bulkPayload.WriteByte('\n')
	}

	return []byte(bulkPayload.String())
}

func retrieveData(c *gin.Context, esClient *elasticsearch.Client, index, id string) (map[string]interface{}, error) {
	getReq := esapi.GetRequest{
		Index:      index,
		DocumentID: id,
	}

	getRes, err := getReq.Do(context.Background(), esClient)
	if err != nil {
		logError(c, "从Elasticsearch检索数据失败：%v", err)
		return nil, err
	}
	defer getRes.Body.Close()

	if getRes.IsError() {
		logError(c, "检索数据失败。响应状态：%s", getRes.Status())
		c.JSON(getRes.StatusCode, gin.H{"error": getRes.Status()})
		return nil, fmt.Errorf("检索数据失败。响应状态：%s", getRes.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(getRes.Body).Decode(&result); err != nil {
		logError(c, "解码响应体失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解码响应体失败"})
		return nil, err
	}

	return result, nil
}

func EsEnv(c *gin.Context) {
	properties, err := models.GetOriginData()
	if err != nil {
		fmt.Println("错误：", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取属性数据"})
		return
	}

	apiKey := "cb3e60dc70d48516d5d19ccaa000ae37"
	service := service.NewAMapService(apiKey)

	results := make([]map[string]interface{}, 0)

	for _, prop := range properties {
		addressInfo := strings.ReplaceAll(prop.AddressInfo, "-", "")
		address := prop.City + addressInfo + prop.CommunityName
		result, err := service.Geocode(address)

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

		err = storeData(c, esClient, "poi_data_2023", "anjuke", result)
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

func logError(c *gin.Context, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fmt.Println("[ERROR]", msg)
	utils.InitLogger()

	// 将utils.Logger赋值给全局的Logger变量
	Logger = utils.Logger

	// 使用日志记录器进行日志输出
	Logger.Println(msg)
}

func logInfo(c *gin.Context, v ...interface{}) {
	message := fmt.Sprint(v...)
	fmt.Println("[INFO]", message)
	utils.InitLogger()

	// 将utils.Logger赋值给全局的Logger变量
	Logger = utils.Logger

	// 使用日志记录器进行日志输出
	Logger.Println(message)
}
