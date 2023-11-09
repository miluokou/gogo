package orm

import (
	"fmt"
	"gorm.io/gorm"
	"mvc/service"
	"mvc/utils"
)

// POIData202301 新表的数据结构
type POIData202301 struct {
	gorm.Model
	Name        string `gorm:"column:名称"`
	Category    string `gorm:"column:大类"`
	SubCategory string `gorm:"column:中类"`
	Longitude   string `gorm:"column:经度"`
	Latitude    string `gorm:"column:纬度"`
	Province    string `gorm:"column:省份"`
	City        string `gorm:"column:城市"`
	District    string `gorm:"column:区域"`
	Deal        bool   `gorm:"column:deal"`
}

// TableName 设置 POIData202301 表名为 "poi_data_2023_01"
func (POIData202301) TableName() string {
	return "poi_data_2023_01"
}

// GetNonNullDealData 查询数据库中不为空的 Deal 数据
func GetNonNullDealData(count int) ([]POIData202301, error) {
	ormDB2 := utils.GetOrmDB2()
	var properties []POIData202301
	result := ormDB2.Where("deal IS NULL").Limit(count).Find(&properties)
	if result.Error != nil {
		service.LogInfo(result.Error)
		return nil, result.Error
	}

	fmt.Println("检索到的记录数量:", len(properties))

	return properties, nil
}

func UpdateDealField(id uint, value int) error {
	ormDB2 := utils.GetOrmDB2()
	err := ormDB2.Model(&POIData202301{}).Where("id = ?", id).Update("deal", value).Error
	return err
}
