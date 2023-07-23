package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"mvc/models"
    "bytes"
	"github.com/gin-gonic/gin"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
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

	results := []Property{}
	for _, prop := range properties {
		results = append(results, Property{
			ID:            int(prop.ID),
			YearInfo:      prop.YearInfo,
			CommunityName: prop.CommunityName,
			AddressInfo:   prop.AddressInfo,
			PricePerSqm:   prop.PricePerSqm,
			PageNumber:    prop.PageNumber,
			Deal:          prop.Deal,
			CreatedAt:     prop.CreatedAt,
			UpdatedAt:     prop.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "属性数据获取成功", "data": results})

// 	esClient, err := createESClient()
// 	if err != nil {
// 		logError(c, "创建Elasticsearch客户端失败：%v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建Elasticsearch客户端失败"})
// 		return
// 	}

// 	index, id := "poi_data_2023", "anjuke"

// 	data := map[string]interface{}{
// 		"field3": "value3",
// 		"field4": "中文数据测试",
// 	}

// 	if err := storeData(c, esClient, index, id, data); err != nil {
// 		return
// 	}

// 	result, err := retrieveData(c, esClient, index, id)
// 	if err != nil {
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "数据检索成功", "data": result})
}

func logError(c *gin.Context, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fmt.Println("[ERROR]", msg)
}