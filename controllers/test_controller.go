package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"log"
	"mvc/jobs"
	"net/http"
	"runtime"
)

func createESClient1() (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://47.100.242.199:9200"}, // 替换为 Elasticsearch 实际的地址
		Username:  "elastic",                              // 替换为您的 Elasticsearch 用户名
		Password:  "miluokou",
	}

	esClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return esClient, nil
}

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
	esClient, err := createESClient1()
	if err != nil {
		logWithLineNum("Failed to create Elasticsearch client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Elasticsearch client"})
		return
	}

	// 存储数据到 Elasticsearch
	indexName := "your-index-name"
	documentID := "your-document-id" // 可选，如果未提供，Elasticsearch 将自动生成一个文档 ID

	// 	data := map[string]interface{}{
	// 		"field3": "value3",
	// 		"field4": "中文数据测试",
	// 	}

	// 	err = StoreData(c, esClient, indexName, documentID, data)
	// 	if err != nil {
	// 		return
	// 	}

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
	// 创建Kafka消费者
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-data1",
		GroupID: "my-group",
	})

	// 从Kafka中读取消息
	msg, err := reader.ReadMessage(c.Request.Context())
	if err != nil {
		log.Println("消息消费失败:", err)
		c.String(http.StatusInternalServerError, "消息消费失败")
		return
	}

	fmt.Printf("主题：%s，分区：%d\n", msg.Topic, msg.Partition)
	fmt.Printf("偏移量：%d\n", msg.Offset)
	fmt.Printf("键：%s\n", string(msg.Key))
	fmt.Printf("值：%s\n", string(msg.Value))
	fmt.Println("头部信息:")
	for _, header := range msg.Headers {
		fmt.Printf("%s: %s\n", header.Key, string(header.Value))
	}

	c.String(http.StatusOK, string(msg.Value))
}
