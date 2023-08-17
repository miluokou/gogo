package routers

import (
	"github.com/gin-gonic/gin"
	"mvc/controllers"
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

	router.GET("/propertydata", controllers.EsEnv)

	// poi 查询
	router.GET("/poi_around", controllers.PoiAround)

	//获取kafka 的主题
	router.GET("/get_topic", kafka.GetTopics)
	router.POST("/kafka/produce", kafka.ProduceMessage)
	router.POST("/kafka/consume", kafka.ConsumeMessages)
	// 添加其他路由...
	// 注意：此处不需要返回任何内容，因为我们是直接修改传入的 `router` 对象

	//触发加入生产者

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

		// 封装正常返回的格式
		response := gin.H{
			"code":    http.StatusOK,
			"message": "OK",
			"data":    c.Keys["response"],
		}
		c.JSON(http.StatusOK, response)
	}
}
