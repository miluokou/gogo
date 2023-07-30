package utils

import (
	"github.com/elastic/go-elasticsearch/v8"
)

func CreateESClient() (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://47.100.242.199:9200"},
		Username:  "elastic",
		Password:  "miluokou",
	}

	return elasticsearch.NewClient(cfg)
}
