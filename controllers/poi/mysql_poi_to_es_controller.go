package poi

import (
	"github.com/gin-gonic/gin"
	"mvc/models/orm"
	"mvc/service"
	"net/http"
)

func MysqlPoiToES(c *gin.Context) {
	res, err := orm.GetNonNullDealData(59)
	if err != nil {
		service.LogInfo(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取属性数据111"})
		return
	}
	service.LogInfo("数据库中的数据")
	service.LogInfo(res)

	data := [][]string{} // Initialize empty data slice

	for _, item := range res {
		record := []string{
			item.Name,
			item.Category,
			item.SubCategory,
			item.Longitude,
			item.Latitude,
			item.Province,
			item.City,
			item.District,
		}
		data = append(data, record)
	}

	err = service.StoreData20231022("poi_2023_01", data)
	if err != nil {
		service.LogInfo(err)
		for _, item := range res {
			err = orm.UpdateDealField(item.ID, 0) // 假设有一个通过ID更新deal字段的函数
			if err != nil {
				service.LogInfo(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "无法更新数据库中的数据"})
				return
			}
		}
	}

	for _, item := range res {
		err = orm.UpdateDealField(item.ID, 1) // 假设有一个通过ID更新deal字段的函数
		if err != nil {
			service.LogInfo(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法更新数据库中的数据"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "属性数据获取成功111",
		"data":    res,
	})
}
