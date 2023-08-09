package orm

import (
	"fmt"
	"gorm.io/gorm"
	"mvc/utils"
)

var db *gorm.DB

func Init() error {
	var err error
	db, err = utils.ConnectDatabase()
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&PropertyData{})
	if err != nil {
		return err
	}

	// 添加日志语句，用于调试
	fmt.Println("Database migration successful")

	return nil
}

type PropertyData struct {
	gorm.Model
	YearInfo      string `json:"year_info"`
	CommunityName string `json:"community_name"`
	AddressInfo   string `json:"address_info"`
	PricePerSqm   string `json:"price_per_sqm"`
	PageNumber    string `json:"page_number"`
	Deal          string `json:"deal"`
	City          string `json:"city"`
	QuText        string `json:"qu_text"`
}

// 查询所有数据
// 根据数量查询数据
func GetPropertyDataByCount(count int) ([]PropertyData, error) {
	var properties []PropertyData
	result := db.Limit(count).Find(&properties)
	if result.Error != nil {
		return nil, result.Error
	}

	// 添加日志语句，用于调试
	fmt.Println("检索到的记录数量:", len(properties))

	return properties, nil
}

// 设置表名为 "djangoproject_propertydata"
func (PropertyData) TableName() string {
	return "djangoproject_propertydata"
}
