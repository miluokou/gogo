package utils

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var db2 *sql.DB

// InitDatabase 初始化数据库连接
func InitDatabase() {
	var err error
	var err2 error

	db, err = sql.Open("mysql", "go_project:rkZSJjmz5ZMSZKmm@tcp(47.100.242.199:3306)/go_project")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	db2, err2 = sql.Open("mysql", "anjuke:5YHj73mpbLmMC4A2@tcp(47.100.242.199:3306)/anjuke")
	if err2 != nil {
		log.Fatal("Failed to connect to database:", err2)
	}

	// 设置数据库连接池最大连接数
	db.SetMaxOpenConns(10)
	db2.SetMaxOpenConns(10)

	// 检查数据库连接是否正常
	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	if err2 = db2.Ping(); err2 != nil {
		log.Fatal("Failed to ping database:", err2)
	}

	log.Println("数据库连接成功!")
}

// GetDB 返回数据库连接对象
func GetDB() *sql.DB {
	return db
}

// GetDB2 返回第二个数据库连接对象
func GetDB2() *sql.DB {
	return db2
}
