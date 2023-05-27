// mvc/routers/router.go
package routers

import (
	"mvc/controllers"

	"github.com/gin-gonic/gin"
	"fmt"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	fmt.Println("Setting up routes...")

// 注册路由
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, users!",
		})
	})
	// 用户相关路由
	userGroup := router.Group("/users")
	{
		userGroup.GET("/all", controllers.GetAllUsers)
		userGroup.POST("", controllers.CreateUser)
		userGroup.GET("/:id", controllers.GetUserByID)
		// 添加其他用户相关的路由
	}

	// 添加日志输出语句
	router.Use(func(c *gin.Context) {
		fmt.Println("Request passed through the router")
		c.Next()
	})

	return router
}

