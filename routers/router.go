package routers

import (
	"github.com/gin-gonic/gin"
	"mvc/controllers"
	"mvc/controllers/around"
	"mvc/controllers/job"
	"mvc/controllers/kafka"
	"mvc/controllers/ocr"
	"mvc/controllers/poi"
	"mvc/controllers/poi/csv"
	"mvc/generate"
	"net/http"
)

func SetupRouter(router *gin.Engine) {
	// 添加处理静态文件的中间件
	// 添加处理静态文件的中间件
	router.Static("/static", "./static")

	router.LoadHTMLGlob("view/*")
	//web文件
	router.GET("/web/china_fence", func(c *gin.Context) {
		// 使用HTML模板渲染网页
		c.HTML(http.StatusOK, "chinafence.html", gin.H{
			"title": "Hello, World!",
		})
	})
	//获取高德网格数据
	router.GET("/china_girds", poi.CalculateGrid)

	//基础报告页面1
	router.GET("/base_report", func(c *gin.Context) {
		// 使用HTML模板渲染网页
		c.HTML(http.StatusOK, "basereport.html", gin.H{
			"title": "Hello, World!",
		})
	})
	//chromedp 试一下chromedp是否能把网页转化为pdf
	router.POST("/chromedp", generate.CreateAst)
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

	// poi 查询
	router.GET("/poi_around", controllers.PoiAround)

	//获取kafka 的主题
	router.GET("/get_topic", kafka.GetTopics)
	router.POST("/kafka/produce", kafka.ProduceMessage)
	router.POST("/kafka/consume", kafka.ConsumeMessages)

	//附近交通情况
	router.POST("/around/traffic/conditions", around.TrafficConditions)

	//10 添加点位
	router.POST("/addPoints", around.AddPoints)

	//11 暂无数据的任务添加
	router.POST("/addNoDataJobs", job.AddNoDataJobs)

	//ocr识别，图片转化成excel 谷歌方案
	router.POST("/ocr", ocr.ConvertToCSV)
	//百度cor识别方案
	//router.POST("/ocr", ocr.ConvertToCSV)

	// 注意：此处不需要返回任何内容，因为我们是直接修改传入的 `router` 对象
	//导入csv文件到es中
	router.GET("/poi/import_poi_data_into_es", poi.CsvToPoi)

	//把大的poi文件拆分成小文件 SplitCSV
	router.POST("/poi/csv/split", csv.SplitFiles)

	//把python存储的mysql的房源数据导入到es中
	router.GET("/propertydata", controllers.ImportMysqlHoseDataToES)

	//mysql 中的poi 数据 导入到es中
	router.GET("/mysql_poi_to_es", poi.MysqlPoiToES)

	//数据采集，中国分块
	router.GET("/china_division", controllers.GetLandParcels)

	//逆地理编码分块中心点
	router.GET("/reverse_geocoding_block_center_point", controllers.ReverseGeocodingBlockCenterPoint)

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
