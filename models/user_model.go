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
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func GetUsers() ([]User, error) {
	db := utils.GetDB()

	users := []User{}
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

	createdAt := time.Now().Format("2006-01-02 15:04:05")
	updatedAt := time.Now().Format("2006-01-02 15:04:05")

	result, err := db.Exec("INSERT INTO users (name, email, phone, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		user.Name, user.Email, user.PhoneNumber, createdAt, updatedAt)
	if err != nil {
		return User{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return User{}, err
	}

	user.ID = uint(id)
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	return user, nil
}

func GetUserByID(id uint) (User, error) {
	db := utils.GetDB()

	user := User{}
	row := db.QueryRow("SELECT * FROM users WHERE id = ?", id)
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt)
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

	user := User{}
	row := db.QueryRow("SELECT * FROM users WHERE phone_number = ?", phoneNumber)
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, nil
		}
		return User{}, err
	}

	return user, nil
}

func UpdateUser(user User) (User, error) {
	db := utils.GetDB()

	updatedAt := time.Now().Format("2006-01-02 15:04:05")
	result, err := db.Exec("UPDATE users SET name = ?, email = ?, phone_number = ?, updated_at = ? WHERE id = ?",
		user.Name, user.Email, user.PhoneNumber, updatedAt, user.ID)
	if err != nil {
		return User{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return User{}, err
	}

	if rowsAffected == 0 {
		return User{}, fmt.Errorf("no rows affected")
	}

	user.UpdatedAt = updatedAt

	return user, nil
}

func DeleteUser(id uint) error {
	db := utils.GetDB()

	result, err := db.Exec("DELETE FROM users WHERE id = ?", id)
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
