package orm

import (
	"mvc/service"
	"mvc/utils"
)

// LandParcel 地块数据结构
type LandParcel struct {
	ID        uint   `gorm:"primaryKey"`
	Province  string `gorm:"column:province"`
	City      string `gorm:"column:city"`
	District  string `gorm:"column:district"`
	FenceData string `gorm:"column:fence_data"`
	Center    string `gorm:"column:center"`
}

// TableName 设置 LandParcel 表名为 "land_parcel"
func (LandParcel) TableName() string {
	return "land_parcel"
}

// CreateLandParcel 创建地块数据
func CreateLandParcel(province, city, district, fenceData, center string) error {
	ormDB := utils.GetOrmDB2()
	//_ = ormDB.AutoMigrate(&LandParcel{})
	//os.Exit(1)
	parcel := LandParcel{
		Province:  province,
		City:      city,
		District:  district,
		FenceData: fenceData,
		Center:    center,
	}
	result := ormDB.Create(&parcel)
	if result.Error != nil {
		service.LogInfo(result.Error)
		return result.Error
	}
	return nil
}
