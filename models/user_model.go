package models

import (
	"database/sql"
	"fmt"
	"time"

	"mvc/utils"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone"`
	OpenId      string `json:"openid"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func GetUsers() ([]User, error) {
	db := utils.GetDB()
	query := "SELECT * FROM users"

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt)
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

func CreateUser(user User) (User, error) {
	db := utils.GetDB()
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	query := "INSERT INTO users (name, email, phone, openid,created_at, updated_at) VALUES (?, ?, ?, ?, ?,?)"
	result, err := db.Exec(query, user.Name, user.Email, user.PhoneNumber,user.OpenId, timestamp, timestamp)
	if err != nil {
		return User{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return User{}, err
	}

	user.ID = uint(id)
	user.CreatedAt = timestamp
	user.UpdatedAt = timestamp

	return user, nil
}

func GetUserByID(id uint) (User, error) {
	db := utils.GetDB()
	query := "SELECT * FROM users WHERE id = ?"

	user := User{}
	err := db.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email, &user.PhoneNumber, &user.OpenId, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, nil
		}
		return User{}, err
	}

	return user, nil
}

func GetUserByPhoneNumber(phoneNumber string) (User, error) {
	db := utils.GetDB()

	query := "SELECT id, name, email, phone, COALESCE(openid, '') as openid, created_at, updated_at FROM users WHERE phone = ?"

	user := User{}
	err := db.QueryRow(query, phoneNumber).Scan(&user.ID, &user.Name, &user.Email, &user.PhoneNumber, &user.OpenId, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, nil
		}
		fmt.Println("获取用户时出错：", err)
		return User{}, err
	}

	return user, nil
}

func UpdateUser(user User) (User, error) {
	db := utils.GetDB()
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	fmt.Printf("OpenId: %s\n", user.OpenId)
	query := "UPDATE users SET name = ?, email = ?, phone = ?, openid = ?, updated_at = ? WHERE id = ?"
	_, err := db.Exec(query, user.Name, user.Email, user.PhoneNumber, user.OpenId, timestamp, user.ID)
	if err != nil {
		return User{}, err
	}

	user.UpdatedAt = timestamp

	return user, nil
}

func DeleteUser(id uint) error {
	db := utils.GetDB()
	query := "DELETE FROM users WHERE id = ?"

	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}
