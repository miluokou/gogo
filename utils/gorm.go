package utils

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var ormDB1 *gorm.DB
var ormDB2 *gorm.DB

// 获取数据库连接
func GetOrmDB1() *gorm.DB {
	return ormDB1
}

// 获取数据库连接
func GetOrmDB2() *gorm.DB {
	return ormDB2
}

func init() {
	var err error
	ormDB1, err = ConnectDatabase()
	if err != nil {
		fmt.Printf("ormDB1 Failed to connect to database: %v\n", err)
		panic(err) // 连接失败时直接抛出异常
	}
	ormDB2, err = ConnectBusinessDatabase()
	if err != nil {
		fmt.Printf("ormDB2 Failed to connect to database: %v\n", err)
		panic(err) // 连接失败时直接抛出异常
	}
}

func ConnectDatabase() (*gorm.DB, error) {
	dsn := "anjuke:5YHj73mpbLmMC4A2@tcp(47.116.7.26:3306)/anjuke?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func ConnectBusinessDatabase() (*gorm.DB, error) {
	fmt.Printf("ConnectBusinessDatabase==========================")
	dsn := "go_project:rkZSJjmz5ZMSZKmm@tcp(47.116.7.26:3306)/go_project?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}
