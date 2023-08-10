package jobs

import (
	"encoding/json"
	"fmt"
	"mvc/models/orm"
	"mvc/service"
)

func MysqlToKafka() {
	// 初始化 ORM
	err := orm.Init()
	if err != nil {
		// 处理初始化错误
		return
	}

	count := 1 // 设置要检索的记录数
	properties, err := orm.GetPropertyDataByCount(count)
	if err != nil {
		fmt.Println("Error retrieving property data:", err)
		return
	}

	for _, property := range properties {
		propertyJSON, err := convertPropertyToJSON(property) // 将 property 转换为 JSON 字符串
		if err != nil {
			fmt.Println("Error converting property to JSON:", err)
			continue
		}
		service.ProduceMessage(string(propertyJSON)) // 将转换后的 JSON 字符串传递给 ProduceMessage 方法
	}

	service.LogInfo(properties)
}

func convertPropertyToJSON(property orm.PropertyData) ([]byte, error) {
	propertyJSON, err := json.Marshal(property) // 将 property 转换为 JSON 格式的字节数组
	if err != nil {
		return nil, err
	}
	return propertyJSON, nil
}
