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
		c.Set("error", res.Status())
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

		poiService, _ := service.NewPOIService()
		existingData, err := poiService.GetPOIsByLocationAndRadius(lat, lon, 3)
		if err != nil {
			// Handle the error from GetPOIsByLocationAndRadius
			errorMsg := fmt.Errorf("Error checking existing data: %v", err)
			service.LogInfo(errorMsg.Error())
			continue
		}
		pois := existingData.POIs
		if len(pois) > 0 {
			service.LogInfo("已经有这条数据了，跳过了存储")
			// Data already exists, skip storage
			continue
		}

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

func ImportMysqlHoseDataToES(c *gin.Context) {
	properties, err := models.GetOriginData()
	if err != nil {
		fmt.Println("错误：", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取属性数据"})
		return
	}

	gaoDeService := service.NewAMapService()

	results := make([]interface{}, 0)

	for _, prop := range properties {
		addressInfo := ""
		if prop.AddressInfo != nil {
			addressInfo = strings.ReplaceAll(*prop.AddressInfo, "-", "")
		}
		address := ""
		if prop.City != nil && prop.CommunityName != nil {
			address = *prop.City + addressInfo + *prop.CommunityName
		} else {
			continue // 继续下一次循环
		}
		geocodes, err := gaoDeService.Geocode(address)
		if err != nil {
			service.LogInfo(err)
			continue // 继续下一次循环
		}

		for _, geocode := range geocodes {
			result := make(map[string]interface{})
			if m, ok := geocode.(map[string]interface{}); ok {
				m["price_per_sqm"] = prop.PricePerSqm // 将价格赋给result
				m["households"] = prop.HouseHolds     // 将户数数据赋给result
				result = m
			} else {
				service.LogInfo("：无法转换地理编码数据")
				continue // 继续下一次循环
			}

			duplicateData, err := findDuplicateData(esClient, "poi_data_2023", result)
			if err != nil {
				fmt.Printf("Failed to query duplicate data: %v", err)
				c.Set("error", "查询重复数据失败")
				return
			}

			if len(duplicateData) > 0 {
				service.LogInfo("发现重复数据，存储操作被中止")
				continue // 继续下一次循环
			}

			err = storeData(c, esClient, "poi_data_2023", []interface{}{result})
			if err != nil {
				service.LogInfo(result)
				fmt.Printf("Failed to store data in Elasticsearch: %v", err)
				c.Set("error", "无法存储数据到Elasticsearch")
				//return
			}
			// 存储成功的话从mysql 数据库中删除这条数据
			err = prop.Delete()
			if err != nil {
				fmt.Printf("Failed to delete data from MySQL: %v", err)
				c.Set("error", "无法从MySQL删除数据")
				return
			}
			results = append(results, result)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "属性数据获取成功",
		"data":    results,
	})
}

/**
 * findDuplicateData 判断是否在es中已经有了这条数据
 *
 *
 * @param
 *
 * @return
 */
func findDuplicateData(esClient *elasticsearch.Client, index string, data map[string]interface{}) ([]map[string]interface{}, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"term": map[string]interface{}{"location.keyword": data["location"]}},
					{"term": map[string]interface{}{"formatted_address.keyword": data["formatted_address"]}},
					{"term": map[string]interface{}{"price_per_sqm": data["price_per_sqm"]}},
				},
			},
		},
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	searchRequest := esapi.SearchRequest{
		Index: []string{index},
		Body:  bytes.NewReader(queryJSON),
	}

	res, err := searchRequest.Do(context.Background(), esClient)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("查询重复数据失败。响应状态：%s", res.Status())
	}

	var resultData map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&resultData); err != nil {
		return nil, err
	}

	hits := resultData["hits"].(map[string]interface{})["hits"].([]interface{})
	duplicateData := make([]map[string]interface{}, len(hits))
	for i, hit := range hits {
		duplicateData[i] = hit.(map[string]interface{})["_source"].(map[string]interface{})
	}

	return duplicateData, nil
}
