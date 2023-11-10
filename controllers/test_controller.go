package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gin-gonic/gin"
	"mvc/jobs"
	"net/http"
	"runtime"
)

//func createESClient1() (*elasticsearch.Client, error) {
//	cfg := elasticsearch.Config{
//		Addresses: []string{"http://47.116.7.26:9200"}, // 替换为 Elasticsearch 实际的地址
//		Username:  "elastic",                              // 替换为您的 Elasticsearch 用户名
//		Password:  "miluokou",
//	}
//
//	esClient, err := elasticsearch.NewClient(cfg)
//	if err != nil {
//		return nil, err
//	}
//
//	return esClient, nil
//}

func logWithLineNum(format string, a ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("[%s:%d] %s\n", file, line, fmt.Sprintf(format, a...))
}

func RetrieveData(c *gin.Context, esClient *elasticsearch.Client, indexName string, documentID string) (map[string]interface{}, error) {
	// 读取数据从 Elasticsearch
	getReq := esapi.GetRequest{
		Index:      indexName,
		DocumentID: documentID,
	}

	// 执行读取请求
	getRes, err := getReq.Do(context.Background(), esClient)
	if err != nil {
		logWithLineNum("Failed to retrieve data from Elasticsearch: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve data from Elasticsearch"})
		return nil, err
	}
	defer getRes.Body.Close()
	// 检查响应状态码

	// 解析读取的响应数据
	var result map[string]interface{}
	if err := json.NewDecoder(getRes.Body).Decode(&result); err != nil {
		logWithLineNum("Failed to decode response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode response body"})
		return nil, err
	}

	return result, nil
}

func TestEnv(c *gin.Context) {
	//获取安居客原始表

	//你地理编码

	// 创建 Elasticsearch 客户端
	//esClient, err := createESClient1()
	//if err != nil {
	//	logWithLineNum("Failed to create Elasticsearch client: %v", err)
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Elasticsearch client"})
	//	return
	//}

	// 存储数据到 Elasticsearch
	indexName := "your-index-name"
	documentID := "your-document-id" // 可选，如果未提供，Elasticsearch 将自动生成一个文档 ID
	// 读取数据从 Elasticsearch
	result, err := RetrieveData(c, esClient, indexName, documentID)
	if err != nil {
		return
	}

	// 返回存储和读取的结果给客户端
	c.JSON(http.StatusOK, gin.H{"message": "Data stored and retrieved successfully", "data": result})
}

func TestEnvProduce(c *gin.Context) {
	// 从请求中获取消息内容
	jobs.MysqlToKafka()
	c.String(http.StatusOK, "消息发送成功")
}

func TestEnvConsume(c *gin.Context) {

	jobs.KafkaToEs()
	c.String(http.StatusOK, "消费的请求成功，具体消费情况要看日志")
}
