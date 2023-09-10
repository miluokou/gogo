package around

import (
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

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	location := request.Location
	radius := request.Radius
	isFavorite := 0

	data, err := orm.CreatePoint(location, radius, int8(isFavorite), request.PointsName, request.Address)
	if err != nil {
		service.LogInfo(err)
		c.Set("error", "Failed to get data")
		//c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get data"})
		//return
	}
	c.Set("response", data)
	//c.JSON(http.StatusOK, data)
}
