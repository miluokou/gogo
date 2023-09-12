package job

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mvc/models/orm"
	"mvc/service"
	"net/http"
	"strconv"
)

var businessDb *gorm.DB

type Task struct {
	UserID         uint
	LocationID     string `json:"location_id"`
	Mark           string `json:"mark"`
	DataCompletion string `json:"data_completion"`
	TaskType       string `json:"task_type"`
}

func AddNoDataJobs(c *gin.Context) {
	var req Task
	authHeader := c.Request.Header.Get("Authorization")
	uid, _ := service.GetUidByToken(authHeader)

	err := c.BindJSON(&req)
	if err != nil {
		// Handle JSON binding error
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	locationID, err := strconv.Atoi(req.LocationID)
	if err != nil {
		// Handle location_id conversion error
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location_id format"})
		return
	}

	task := orm.Task{
		UserID:         uid,
		LocationID:     locationID,
		Mark:           req.Mark,
		DataCompletion: req.DataCompletion,
		TaskType:       req.TaskType,
	}

	_, err = orm.CreateTask(task.UserID, task.LocationID, task.Mark, task.DataCompletion, task.TaskType)
	if err != nil {
		service.LogInfo("Error storing task")
		service.LogInfo(err.Error())
		c.Set("error", "Failed to store task")
		return
	}

	c.Set("message", "Task added successfully")
}
