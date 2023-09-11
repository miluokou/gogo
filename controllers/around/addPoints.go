package around

import (
	"strings"

	"github.com/gin-gonic/gin"
	"mvc/models/orm"
	"mvc/service"
	"net/http"
)

type PointsRequest struct {
	Location   string  `json:"location"`
	Radius     string  `json:"radius"`
	PointsName *string `json:"pointsName"`
	Address    *string `json:"address"`
}

func AddPoints(c *gin.Context) {
	var request PointsRequest

	authHeader := c.Request.Header.Get("Authorization")
	secretKey := "your_secret_key"
	salt := "your_salt"
	jwtController, err := service.NewJWTController(secretKey, salt)

	claims, err := jwtController.VerifyToken(strings.TrimPrefix(authHeader, "Bearer "))
	if err != nil {
		service.LogInfo(err)
		c.Set("error", "用户token失效")
		return
	}
	userID := claims.UserID

	service.LogInfo("uid 是")
	service.LogInfo(userID)

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	location := request.Location
	radius := request.Radius
	isFavorite := 0

	data, err := orm.CreatePoint(userID, location, radius, int8(isFavorite), request.PointsName, request.Address)
	if err != nil {
		service.LogInfo(err)
		c.Set("error", "Failed to get data")
	}
	c.Set("response", data)
}
