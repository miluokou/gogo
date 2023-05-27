// mvc/utils/database.go

package utils

import (
    "database/sql"
    "log"
    _ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

// InitDatabase 初始化数据库连接
func InitDatabase() {
    var err error

   db, err = sql.Open("mysql", "go_project:MkyhbWcyYTiYZYMm@tcp(47.100.242.199:3306)/go_project")
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // 设置数据库连接池最大连接数
    db.SetMaxOpenConns(10)

    // 检查数据库连接是否正常
    if err = db.Ping(); err != nil {
        log.Fatal("Failed to ping database:", err)
    }

    log.Println("Connected to database!")
}

// GetDB 返回数据库连接对象
func GetDB() *sql.DB {
    return db
}
