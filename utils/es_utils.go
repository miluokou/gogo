package utils

import (
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
)

var esClientPool *sync.Pool

func init() {
	maxConnections := 1 // 设置连接池的最大容量
	esClientPool = &sync.Pool{
		New: func() interface{} {
			cfg := elasticsearch.Config{
				Addresses: []string{"http://47.100.242.199:9200"}, // 替换为 Elasticsearch 实际的地址
				Username:  "elastic",                              // 替换为您的 Elasticsearch 用户名
				Password:  "miluokou",
			}
			client, err := elasticsearch.NewClient(cfg)
			if err != nil {
				panic(err) // 初始化连接池时出错，程序无法继续，可以根据实际需求选择处理方式
			}
			return client
		},
	}
	// 预先创建指定数量的连接并放入连接池中
	for i := 0; i < maxConnections; i++ {
		esClientPool.Put(esClientPool.New())
	}
}

func GetESClient() (*elasticsearch.Client, error) {
	client := esClientPool.Get().(*elasticsearch.Client)
	return client, nil
}

func ReleaseESClient(client *elasticsearch.Client) {
	esClientPool.Put(client)
}
