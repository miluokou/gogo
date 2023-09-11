package around

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"mvc/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const apiKey = "cb3e60dc70d48516d5d19ccaa000ae37"

type Data struct {
	FootbusTimeMin    int32 `json:"footbus_time_min"`
	FootsubwayTimeMin int32 `json:"footsubway_time_min"`
	DriveTrainTimeMin int32 `json:"drive_train_time_min"`
	DrivePlaneTimeMin int32 `json:"drive_plane_time_min"`
}

type TrafficRequest struct {
	Location string `json:"location"`
}

//func main() {
//	router := gin.Default()
//	router.GET("/traffic", TrafficConditions)
//	router.Run(":8080")
//}

func TrafficConditions(c *gin.Context) {
	var request TrafficRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	location := request.Location

	data, err := getTrafficConditions(location, 35000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get data"})
		return
	}
	c.Set("response", data)
	//c.JSON(http.StatusOK, data)
}

func getTrafficConditions(location string, radius int) (*Data, error) {
	data := &Data{}

	urlbus := fmt.Sprintf("https://restapi.amap.com/v3/place/around?types=150700&location=%s&key=%s&radius=%d", location, apiKey, radius)
	footbusTimeMin, err := checkTravelTime(location, 5000, urlbus, "walking")
	service.LogInfo("距离最近公交车时间")
	service.LogInfo(footbusTimeMin)
	if err != nil {
		return nil, err
	}
	data.FootbusTimeMin = footbusTimeMin

	urlsubway := fmt.Sprintf("https://restapi.amap.com/v3/place/around?types=150500&location=%s&key=%s&radius=%d&citylimit=true", location, apiKey, radius)
	footsubwayTimeMin, err := checkTravelTime(location, 5000, urlsubway, "walking")
	if err != nil {
		return nil, err
	}

	data.FootsubwayTimeMin = footsubwayTimeMin

	urlTrain := fmt.Sprintf("https://restapi.amap.com/v3/place/around?types=150200&location=%s&key=%s&radius=%d", location, apiKey, radius)
	driveTrainTimeMin, err := checkTravelTime(location, radius, urlTrain, "driving")
	if err != nil {
		return nil, err
	}
	data.DriveTrainTimeMin = driveTrainTimeMin

	urlPlane := fmt.Sprintf("https://restapi.amap.com/v3/place/around?types=150100&location=%s&key=%s&radius=%d", location, apiKey, radius)
	drivePlaneTimeMin, err := checkTravelTime(location, radius, urlPlane, "driving")
	if err != nil {
		return nil, err
	}
	data.DrivePlaneTimeMin = drivePlaneTimeMin
	return data, nil
}

func calculateDuration(route map[string]interface{}) int {
	durationSum := 0

	paths := route["paths"].([]interface{})
	for _, path := range paths {
		pathMap := path.(map[string]interface{})
		durationStr := pathMap["duration"].(string)
		duration, err := strconv.Atoi(durationStr)
		if err != nil {
			duration = 0 // 转换失败时设置默认值为0
		}
		service.LogInfo("calculateDuration 中 计算出来的时间是")
		service.LogInfo(duration)
		subPaths, ok := pathMap["steps"].(map[string]interface{})
		if ok {
			durationSum += calculateDuration(subPaths)
		}

		durationSum += int(duration)
	}

	return durationSum
}

func checkTravelTime(location string, radius int, url string, way string) (int32, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return 0, err
	}
	service.LogInfo("看是否取到了poi 数据")
	service.LogInfo(result["pois"])
	footbusTimeMin := 0
	if pois, ok := result["pois"].([]interface{}); ok && len(pois) > 0 {
		if busLocation, ok := pois[0].(map[string]interface{})["location"].(string); ok {
			urlfootBus := fmt.Sprintf("https://restapi.amap.com/v3/direction/%s?origin=%s&destination=%s&key=%s&radius=%d", way, location, busLocation, apiKey, radius)
			resp, err := http.Get(urlfootBus)
			//service.LogInfo("访问到达目的地的url：")
			//service.LogInfo(urlfootBus)
			//service.LogInfo("访问到达目的地的url结束：")
			if err != nil {
				return 0, err
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return 0, err
			}

			var result map[string]interface{}
			err = json.Unmarshal(body, &result)
			if err != nil {
				return 0, err
			}
			//service.LogInfo("访问到达目的地的url的json转化的result结果打印开始")
			//service.LogInfo(result)
			//service.LogInfo("访问到达目的地的url的json转化的result结果打印结束")

			route, ok := result["route"].(map[string]interface{})
			if !ok {
				return 0, fmt.Errorf("Failed to parse route")
			}
			//service.LogInfo("如果已经请求到这里说明已经获取到了route")
			//service.LogInfo(route)
			footbusTimeSec := calculateDuration(route)
			//footbusTimeMin = int(math.Round(float64(footbusTimeSec / 60)))
			footbusTimeMin = int(math.Ceil(float64(footbusTimeSec) / 60))
		}
	}

	return int32(footbusTimeMin), nil
}
