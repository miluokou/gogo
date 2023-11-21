package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	client *redis.Client
}

func NewCache() (*Cache, error) {
	// 创建 Redis 客户端连接
	client := redis.NewClient(&redis.Options{
		Addr:     "47.116.7.26:6379", // 根据实际情况修改 Redis 服务器地址和端口
		Password: "47116727",         // 如果有密码，请提供正确的密码
		DB:       0,                  // 可选：根据实际情况选择 Redis 数据库索引
	})

	// 创建上下文
	ctx := context.Background()

	// 检查 Redis 连接是否成功
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	cache := &Cache{client: client}
	return cache, nil
}

func (c *Cache) Set(key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %v", err)
	}

	// 创建上下文
	ctx := context.Background()

	err = c.client.Set(ctx, key, jsonData, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache data in Redis: %v", err)
	}

	return nil
}

func (c *Cache) Get(key string) (interface{}, error) {
	// 创建上下文
	ctx := context.Background()

	jsonData, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("key not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get cache data from Redis: %v", err)
	}

	var value interface{}
	err = json.Unmarshal(jsonData, &value)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %v", err)
	}

	return value, nil
}
