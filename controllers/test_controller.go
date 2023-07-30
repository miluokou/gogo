package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	// 	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime"
	// 	"mvc/models"
	"github.com/segmentio/kafka-go"
	"log"
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

func StoreData(c *gin.Context, esClient *elasticsearch.Client, indexName string, documentID string, data map[string]interface{}) error {
	// 构建存储请求
	requestData, _ := json.Marshal(data)

	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: documentID,
		Body:       bytes.NewReader(requestData),
		Refresh:    "true",
	}

	// 执行存储请求
	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		logWithLineNum("Failed to store data in Elasticsearch: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store data in Elasticsearch"})
		return err
	}
	defer res.Body.Close()

	// 检查响应状态码
	if res.IsError() {
		logWithLineNum("Failed to store data. Response status: %s", res.Status())
		c.JSON(res.StatusCode, gin.H{"error": res.Status()})
		return fmt.Errorf("failed to store data. Response status: %s", res.Status())
	}

	return nil
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
	if getRes.IsError() {
		logWithLineNum("Failed to retrieve data. Response status: %s", getRes.Status())
		c.JSON(getRes.StatusCode, gin.H{"error": getRes.Status()})
		return nil, fmt.Errorf("failed to retrieve data. Response status: %s", getRes.Status())
	}

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
	message := c.PostForm("message")

	// 创建Kafka生产者
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-data1",
	})

	// 发送消息到Kafka
	err := writer.WriteMessages(c.Request.Context(), kafka.Message{
		Value: []byte(message),
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "消息发送失败："+err.Error())
		return
	}

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

	// 	// 打印接收到的消息值和偏移量
	// 	fmt.Println("接收到的消息值:", string(msg.Value))
	// 	fmt.Println("消息偏移量:", msg.Offset)

	c.String(http.StatusOK, string(msg.Value))
}
