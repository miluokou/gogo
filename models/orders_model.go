package models

import (
	"database/sql"
	"fmt"
	"time"

	"mvc/utils"

	_ "github.com/go-sql-driver/mysql"
)

type Order struct {
	ID            uint    `json:"id"`
	OrderNumber   string  `json:"order_number"`
	CallbackTime  string  `json:"callback_time"`
	TransactionID string  `json:"transaction_id"`
	PaymentAmount float64 `json:"payment_amount"`
	PaymentStatus string  `json:"payment_status"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	OpenID        string  `json:"open_id"` // 新增字段：微信支付回调返回的 open_id
}


func CreateOrder(order Order) (Order, error) {
	db := utils.GetDB()
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	query := "INSERT INTO orders (order_number, callback_time, transaction_id, payment_amount, payment_status, created_at, updated_at, open_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	result, err := db.Exec(query, order.OrderNumber, order.CallbackTime, order.TransactionID, order.PaymentAmount, order.PaymentStatus, timestamp, timestamp,order.OpenID)
	if err != nil {
		return Order{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Order{}, err
	}

	order.ID = uint(id)
	order.CreatedAt = timestamp
	order.UpdatedAt = timestamp

	return order, nil
}

func GetOrderByID(id uint) (Order, error) {
	db := utils.GetDB()
	query := "SELECT * FROM orders WHERE id = ?"

	order := Order{}
	err := db.QueryRow(query, id).Scan(&order.ID, &order.OrderNumber, &order.CallbackTime, &order.TransactionID, &order.PaymentAmount, &order.PaymentStatus, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Order{}, nil
		}
		return Order{}, err
	}

	return order, nil
}

func UpdateOrder(order Order) (Order, error) {
	db := utils.GetDB()
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	query := "UPDATE orders SET order_number = ?, callback_time = ?, transaction_id = ?, payment_amount = ?, payment_status = ?, updated_at = ? WHERE id = ?"
	_, err := db.Exec(query, order.OrderNumber, order.CallbackTime, order.TransactionID, order.PaymentAmount, order.PaymentStatus, timestamp, order.ID)
	if err != nil {
		return Order{}, err
	}

	order.UpdatedAt = timestamp

	return order, nil
}

func DeleteOrder(id uint) error {
	db := utils.GetDB()
	query := "DELETE FROM orders WHERE id = ?"

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

// 在 models/orders_model.go 中添加以下方法

func GetOrderByOrderNumber(orderNumber string) (Order, error) {
	db := utils.GetDB()
	query := "SELECT * FROM orders WHERE order_number = ?"

	order := Order{}
	err := db.QueryRow(query, orderNumber).Scan(&order.ID, &order.OrderNumber, &order.CallbackTime, &order.TransactionID, &order.PaymentAmount, &order.PaymentStatus, &order.CreatedAt, &order.UpdatedAt, &order.OpenID)
	if err != nil {
		if err == sql.ErrNoRows {
			return Order{}, nil
		}
		return Order{}, err
	}

	return order, nil
}

