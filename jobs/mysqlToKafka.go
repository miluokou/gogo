package jobs

import (
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

	// 从 djangoproject_propertydata 表中读取

	count := 10 // Set the number of records you want to retrieve
	properties, err := orm.GetPropertyDataByCount(count)
	if err != nil {
		fmt.Println("Error retrieving property data:", err)
		return
	}

	for _, property := range properties {
		fmt.Println(property)
	}

	service.LogInfo(properties)
}
