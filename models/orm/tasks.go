package orm

import (
	"fmt"
	"gorm.io/gorm"
	"mvc/service"
	"time"
)

type Task struct {
	gorm.Model
	UserID         uint       `json:"user_id"`
	LocationID     int        `json:"location_id"`
	Mark           string     `json:"mark"`
	DataCompletion string     `json:"data_completion"`
	TaskType       string     `json:"task_type"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at"`
}

// CreateTask function: stores data in the database
func CreateTask(uid uint, locationID int, mark, dataCompletion, taskType string) (*Task, error) {
	err := Init()
	if err != nil {
		return nil, err
	}

	service.LogInfo("Executing CreateTask method")

	task := Task{
		UserID:         uid,
		LocationID:     locationID,
		Mark:           mark,
		DataCompletion: dataCompletion,
		TaskType:       taskType,
	}

	result := businessDb.Create(&task)
	if result.Error != nil {
		service.LogInfo("Error storing task")
		service.LogInfo(result.Error)
		return nil, result.Error
	}

	fmt.Println("Data created successfully")

	return &task, nil
}

// Set table name to "tasks"
func (Task) TableName() string {
	return "tasks"
}
