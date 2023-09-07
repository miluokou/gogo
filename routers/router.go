package routers

import (
	"github.com/gin-gonic/gin"
	"mvc/controllers"
	"mvc/controllers/around"
	"mvc/controllers/kafka"
	"net/http"
)

func SetupRouter(router *gin.Engine) {
	// 添加中间件
	router.Use(ResponseMiddleware())

	router.POST("/wechat/login", controllers.WeChatLogin)
	router.POST("/wechatpay/notify/payNotify/wxpay-mp", controllers.WechatCallback)

	// 用户相关路由
	userGroup := router.Group("/users")
	{
		userGroup.GET("/all", controllers.GetAllUsers)
		userGroup.POST("", controllers.CreateUser)
		userGroup.GET("/:id", controllers.GetUserByID)
		// 添加其他用户相关的路由
	}

	// 注册根路由
	router.GET("/test", controllers.TestEnv)

	router.GET("/produce", controllers.TestEnvProduce)
	router.GET("/consume", controllers.TestEnvConsume)

	//把mysql的房源数据导入到es中
	router.GET("/propertydata", controllers.ImportMysqlHoseDataToES)

	// poi 查询
	router.GET("/poi_around", controllers.PoiAround)

	//获取kafka 的主题
	router.GET("/get_topic", kafka.GetTopics)
	router.POST("/kafka/produce", kafka.ProduceMessage)
	router.POST("/kafka/consume", kafka.ConsumeMessages)

	//附近交通情况
	router.POST("/around/traffic/conditions", around.TrafficConditions)
	// 添加其他路由...
	// 注意：此处不需要返回任何内容，因为我们是直接修改传入的 `router` 对象

}

func ResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 处理异常报错的格式
		if err, exists := c.Get("error"); exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}

		count, ok := c.Keys["count"].(int)
		if !ok {
			count = 0
		}

		householdsAvg, ok := c.Keys["households_avg"]
		if !ok {
			householdsAvg = 0
		}

		pricePerSqMAvg, ok := c.Keys["price_per_sqm_avg"]
		if !ok {
			pricePerSqMAvg = 0
		}

		data, ok := c.Keys["response"]
		if !ok {
			data = []interface{}{}
		}
		// 封装正常返回的格式
		response := gin.H{
			"code":              http.StatusOK,
			"message":           "OK",
			"count":             count,
			"households_avg":    householdsAvg,
			"price_per_sqm_avg": pricePerSqMAvg,
			"data":              data,
		}
		c.JSON(http.StatusOK, response)
	}
}
