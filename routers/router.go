package routers

import (
	"github.com/gin-gonic/gin"
	"mvc/controllers"
)

func SetupRouter(router *gin.Engine) {
	

	router.POST("/wechat/login", controllers.WeChatLogin)
	router.POST("/wechatpay/notify/payNotify/wxpay-mp",controllers.WechatCallback)
	
	// 用户相关路由
	userGroup := router.Group("/users")
	{
		userGroup.GET("/all", controllers.GetAllUsers)
		userGroup.POST("", controllers.CreateUser)
		userGroup.GET("/:id", controllers.GetUserByID)
		// 添加其他用户相关的路由
	}

	// 注册根路由
	router.GET("/test",  controllers.TestEnv)
	router.GET("/produce", controllers.TestEnvProduce)
	router.GET("/consume", controllers.TestEnvConsume)
	
	router.GET("/propertydata", controllers.EsEnv)

	// 添加其他路由...

	// 添加中间件或其他配置

	// 注意：此处不需要返回任何内容，因为我们是直接修改传入的 `router` 对象
}
