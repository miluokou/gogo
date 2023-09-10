package orm

import (
	"fmt"
	"gorm.io/gorm"
	"mvc/service"
	"time"
)

var businessDb *gorm.DB

type Points struct {
	gorm.Model
	Location   string     `json:"location"`
	Radius     string     `json:"radius"`
	PointsName string     `json:"points_name"`
	Address    string     `json:"address"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at"`
	IsFavorite int8       `json:"is_favorite"`
}

// 新增函数：将数据存储到数据库中
func CreatePoint(location, radius string, isFavorite int8, pointsName, address *string) (*Points, error) {
	err := Init()
	if err != nil {
		// 处理初始化错误
		return nil, err
	}

	service.LogInfo("已经执行到了CreatePoint 方法22222")
	point := Points{
		Location:   location,
		Radius:     radius,
		IsFavorite: isFavorite,
	}

	if pointsName != nil {
		point.PointsName = *pointsName
	}

	if address != nil {
		point.Address = *address
	}
	service.LogInfo("businessDb.Create 之前，看看之后有没有日志，没有的话说明有问题")
	result := businessDb.Create(&point)
	if result.Error != nil {
		service.LogInfo("打印存储点位的时候的异常日志")
		service.LogInfo(result.Error)
		return nil, result.Error
	}

	// 添加日志语句，用于调试
	fmt.Println("创建数据成功")

	return &point, nil
}

// 设置表名为 "points"
func (Points) TableName() string {
	return "points"
}
