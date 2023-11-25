package orm

import (
	"mvc/service"
	"mvc/utils"
)

// LandParcel 地块数据结构
type LandParcel struct {
	ID               uint   `gorm:"primaryKey"`
	Province         string `gorm:"column:province"`
	City             string `gorm:"column:city"`
	District         string `gorm:"column:district"`
	FenceData        string `gorm:"column:fence_data"`
	Center           string `gorm:"column:center"`
	Page             string `gorm:"column:page"`
	Deal             string `gorm:"column:deal"`
	Adcode           string `gorm:"column:adcode"`
	Towncode         string `gorm:"column:towncode"`
	Township         string `gorm:"column:township"`
	FormattedAddress string `gorm:"column:formatted_address"`
}

// city := ""
// province := ""
// adcode := ""
// district := ""
// towncode := ""
// township := ""
// formatted_address := ""
// TableName 设置 LandParcel 表名为 "land_parcel"
func (LandParcel) TableName() string {
	return "land_parcel"
}

// CreateLandParcel 创建地块数据
func CreateLandParcel(province, city, district, fenceData, center, page string) error {
	ormDB := utils.GetOrmDB2()
	//_ = ormDB.AutoMigrate(&LandParcel{})
	//os.Exit(1)
	parcel := LandParcel{
		Province:  province,
		City:      city,
		District:  district,
		FenceData: fenceData,
		Center:    center,
		Page:      page,
	}
	result := ormDB.Create(&parcel)
	if result.Error != nil {
		service.LogInfo(result.Error)
		return result.Error
	}
	return nil
}

func GetLandParcel(count int) ([]LandParcel, error) {
	ormDB := utils.GetOrmDB2()
	var properties []LandParcel
	result := ormDB.Order("id DESC").Limit(count).Find(&properties)
	if result.Error != nil {
		service.LogInfo(result.Error)
		return nil, result.Error
	}
	return properties, nil
}

func GetLandParcelNotDeal(count int) ([]LandParcel, error) {
	ormDB := utils.GetOrmDB2()
	var properties []LandParcel
	result := ormDB.Where("deal = ?", "0").Order("id ASC").Limit(count).Find(&properties)
	if result.Error != nil {
		service.LogInfo(result.Error)
		return nil, result.Error
	}
	return properties, nil
}

func UpdateLandParcel(id uint, newData map[string]interface{}) error {
	ormDB := utils.GetOrmDB2()
	result := ormDB.Model(&LandParcel{}).Where("id = ?", id).Updates(newData)
	if result.Error != nil {
		service.LogInfo(result.Error)
		return result.Error
	}
	return nil
}
