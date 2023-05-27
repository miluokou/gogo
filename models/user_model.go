package models

import (
	"database/sql"
	"errors"
	"time"
    "database/sql/driver"
	"mvc/utils"
)

type CustomTime struct {
	time.Time
}

func (ct *CustomTime) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		ct.Time = v
	case []byte:
		t, err := time.Parse("2006-01-02 15:04:05", string(v))
		if err != nil {
			return err
		}
		ct.Time = t
	case nil:
		ct.Time = time.Time{}
	default:
		return errors.New("failed to scan CustomTime")
	}

	return nil
}

func (ct CustomTime) Value() (driver.Value, error) {
	if ct.IsZero() {
		return nil, nil
	}
	return ct.Time, nil
}

type User struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	CreatedAt CustomTime `json:"created_at"`
	UpdatedAt CustomTime `json:"updated_at"`
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

func CreateUser(user User) (User, error) {
	db := utils.GetDB()

	_, err := db.Exec("INSERT INTO users (name, email, created_at, updated_at) VALUES (?, ?, ?, ?)",
		user.Name, user.Email, user.CreatedAt.Time, user.UpdatedAt.Time)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func GetUserByID(id uint) (User, error) {
	db := utils.GetDB()

	user := User{}
	row := db.QueryRow("SELECT * FROM users WHERE id = ?", id)
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
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

	_, err := db.Exec("UPDATE users SET name = ?, email = ?, updated_at = ? WHERE id = ?",
		user.Name, user.Email, user.UpdatedAt.Time, user.ID)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func DeleteUser(id uint) error {
	db := utils.GetDB()

	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}

	return nil
}
