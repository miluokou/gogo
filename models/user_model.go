// mvc/models/user_model.go
package models

import (
	"time"
	"database/sql"
	"mvc/utils"
)

type User struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetUsers 返回所有用户
func GetUsers() ([]User, error) {
	db := utils.GetDB()

	// 执行获取用户列表的逻辑
	users := []User{}
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// CreateUser 创建新用户
func CreateUser(user User) (User, error) {
	db := utils.GetDB()

	// 执行创建用户的逻辑
	_, err := db.Exec("INSERT INTO users (name, email, created_at, updated_at) VALUES (?, ?, ?, ?)",
		user.Name, user.Email, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

// GetUserByID 根据ID获取单个用户
func GetUserByID(id uint) (User, error) {
	db := utils.GetDB()

	user := User{}
	row := db.QueryRow("SELECT * FROM users WHERE id = ?", id)
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, nil // 没有找到对应ID的用户，返回nil
		}
		return User{}, err
	}

	return user, nil
}

// UpdateUser 更新用户
func UpdateUser(user User) (User, error) {
	db := utils.GetDB()

	// 执行更新用户的逻辑
	_, err := db.Exec("UPDATE users SET name = ?, email = ?, updated_at = ? WHERE id = ?",
		user.Name, user.Email, user.UpdatedAt, user.ID)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

// DeleteUser 删除用户
func DeleteUser(id uint) error {
	db := utils.GetDB()

	// 执行删除用户的逻辑
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}

	return nil
}
