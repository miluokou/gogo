package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gin-gonic/gin"
	"mvc/models"
	"mvc/service"
	"net/http"
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

type PlaceSearchResult struct {
	RawData map[string]interface{} `json:"-"`
	Data    []interface{}          `json:"data"`
}

func createESClient() (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
		Username:  "elastic",
		Password:  "miluokou",
	}

	return elasticsearch.NewClient(cfg)
}

func storeData(c *gin.Context, esClient *elasticsearch.Client, index, id string, data map[string]interface{}) error {
	reqData, _ := json.Marshal(data)

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(reqData),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		logError(c, "存储数据到Elasticsearch失败：%v", err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		logError(c, "存储数据失败。响应状态：%s", res.Status())
		c.JSON(res.StatusCode, gin.H{"error": res.Status()})
		return fmt.Errorf("存储数据失败。响应状态：%s", res.Status())
	}

	return nil
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

	esCfg := elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	}
	esClient, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		fmt.Printf("Failed to create Elasticsearch client: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建Elasticsearch客户端"})
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

		fmt.Println(result)
		if m, ok := result[0].(map[string]interface{}); ok {
			results = append(results, m)
		} else {
			fmt.Println("无法转换为map[string]interface{}")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法转换地理编码结果"})
			return
		}

		err = storeData(c, esClient, "poi_data_2023", "anjuke", result[0].(map[string]interface{}))
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
}
