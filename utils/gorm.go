package utils

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectDatabase() (*gorm.DB, error) {
	dsn := "anjuke:5YHj73mpbLmMC4A2@tcp(47.100.242.199:3306)/anjuke?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func ConnectBusinessDatabase() (*gorm.DB, error) {
	fmt.Printf("ConnectBusinessDatabase==========================")
	dsn := "go_project:rkZSJjmz5ZMSZKmm@tcp(47.100.242.199:3306)/go_project?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}
